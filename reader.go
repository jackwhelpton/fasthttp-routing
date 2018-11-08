package routing

import (
	"bytes"
	"encoding"
	"encoding/json"
	"encoding/xml"
	"errors"
	"reflect"
	"strconv"

	"github.com/valyala/fasthttp"
)

// MIME types used when doing request data reading and response data writing.
const (
	MIME_JSON           = "application/json"
	MIME_XML            = "application/xml"
	MIME_XML2           = "text/xml"
	MIME_HTML           = "text/html"
	MIME_FORM           = "application/x-www-form-urlencoded"
	MIME_MULTIPART_FORM = "multipart/form-data"
)

var (
	textUnmarshalerType = reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()
)

// DataReader is used by Context.Read() to read data from an HTTP request.
type DataReader interface {
	// Read reads from the given HTTP request and populate the specified data.
	Read(*fasthttp.RequestCtx, interface{}) error
}

var (
	// DataReaders lists all supported content types and the corresponding data readers.
	// Context.Read() will choose a matching reader from this list according to the "Content-Type"
	// header from the current request.
	// You may modify this variable to add new supported content types.
	DataReaders = map[string]DataReader{
		MIME_FORM:           &FormDataReader{},
		MIME_MULTIPART_FORM: &FormDataReader{},
		MIME_JSON:           &JSONDataReader{},
		MIME_XML:            &XMLDataReader{},
		MIME_XML2:           &XMLDataReader{},
	}
	// DefaultFormDataReader is the reader used when there is no matching reader in DataReaders
	// or if the current request is a GET request.
	DefaultFormDataReader DataReader = &FormDataReader{}
)

// JSONDataReader reads the request body as JSON-formatted data.
type JSONDataReader struct{}

func (r *JSONDataReader) Read(ctx *fasthttp.RequestCtx, data interface{}) error {
	return json.Unmarshal(ctx.PostBody(), data)
}

// XMLDataReader reads the request body as XML-formatted data.
type XMLDataReader struct{}

func (r *XMLDataReader) Read(ctx *fasthttp.RequestCtx, data interface{}) error {
	return xml.Unmarshal(ctx.PostBody(), data)
}

// FormDataReader reads the query parameters and request body as form data.
type FormDataReader struct{}

func (r *FormDataReader) Read(ctx *fasthttp.RequestCtx, data interface{}) error {
	f := make(map[string][]string)
	if ctx.IsPost() || ctx.IsPut() || bytes.Equal(ctx.Method(), strPatch) {
		ctx.PostArgs().VisitAll(func(key, value []byte) {
			k := string(key)
			f[k] = append(f[k], string(value))
		})
	}
	ctx.QueryArgs().VisitAll(func(key, value []byte) {
		k := string(key)
		f[k] = append(f[k], string(value))
	})
	return ReadFormData(f, data)
}

const formTag = "form"

// ReadFormData populates the data variable with the data from the given form values.
func ReadFormData(form map[string][]string, data interface{}) error {
	rv := reflect.ValueOf(data)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return errors.New("data must be a pointer")
	}
	rv = indirect(rv)
	if rv.Kind() != reflect.Struct {
		return errors.New("data must be a pointer to a struct")
	}

	return readForm(form, "", rv)
}

func readForm(form map[string][]string, prefix string, rv reflect.Value) error {
	rv = indirect(rv)
	rt := rv.Type()
	n := rt.NumField()
	for i := 0; i < n; i++ {
		field := rt.Field(i)
		tag := field.Tag.Get(formTag)

		// only handle anonymous or exported fields
		if !field.Anonymous && field.PkgPath != "" || tag == "-" {
			continue
		}

		ft := field.Type
		if ft.Kind() == reflect.Ptr {
			ft = ft.Elem()
		}

		name := tag
		if name == "" && !field.Anonymous {
			name = field.Name
		}
		if name != "" && prefix != "" {
			name = prefix + "." + name
		}

		// check if type implements a known type, like encoding.TextUnmarshaler
		if ok, err := readFormFieldKnownType(form, name, rv.Field(i)); err != nil {
			return err
		} else if ok {
			continue
		}

		if ft.Kind() != reflect.Struct {
			if err := readFormField(form, name, rv.Field(i)); err != nil {
				return err
			}
			continue
		}

		if name == "" {
			name = prefix
		}
		if err := readForm(form, name, rv.Field(i)); err != nil {
			return err
		}
	}
	return nil
}

func readFormFieldKnownType(form map[string][]string, name string, rv reflect.Value) (bool, error) {
	value, ok := form[name]
	if !ok {
		return false, nil
	}
	rv = indirect(rv)
	rt := rv.Type()

	// check if type implements encoding.TextUnmarshaler
	if rt.Implements(textUnmarshalerType) {
		return true, rv.Interface().(encoding.TextUnmarshaler).UnmarshalText([]byte(value[0]))
	} else if reflect.PtrTo(rt).Implements(textUnmarshalerType) {
		return true, rv.Addr().Interface().(encoding.TextUnmarshaler).UnmarshalText([]byte(value[0]))
	}
	return false, nil
}

func readFormField(form map[string][]string, name string, rv reflect.Value) error {
	value, ok := form[name]
	if !ok {
		return nil
	}
	rv = indirect(rv)
	if rv.Kind() != reflect.Slice {
		return setFormFieldValue(rv, value[0])
	}

	n := len(value)
	slice := reflect.MakeSlice(rv.Type(), n, n)
	for i := 0; i < n; i++ {
		if err := setFormFieldValue(slice.Index(i), value[i]); err != nil {
			return err
		}
	}
	rv.Set(slice)
	return nil
}

func setFormFieldValue(rv reflect.Value, value string) error {
	switch rv.Kind() {
	case reflect.Bool:
		if value == "" {
			value = "false"
		}
		v, err := strconv.ParseBool(value)
		if err == nil {
			rv.SetBool(v)
		}
		return err
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if value == "" {
			value = "0"
		}
		v, err := strconv.ParseInt(value, 10, 64)
		if err == nil {
			rv.SetInt(v)
		}
		return err
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if value == "" {
			value = "0"
		}
		v, err := strconv.ParseUint(value, 10, 64)
		if err == nil {
			rv.SetUint(v)
		}
		return err
	case reflect.Float32, reflect.Float64:
		if value == "" {
			value = "0"
		}
		v, err := strconv.ParseFloat(value, 64)
		if err == nil {
			rv.SetFloat(v)
		}
		return err
	case reflect.String:
		rv.SetString(value)
		return nil
	default:
		return errors.New("Unknown type: " + rv.Kind().String())
	}
}

// indirect dereferences pointers and returns the actual value it points to.
// If a pointer is nil, it will be initialized with a new value.
func indirect(v reflect.Value) reflect.Value {
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		v = v.Elem()
	}
	return v
}
