package routing

import (
	"errors"
	"fmt"
	"testing"

	"github.com/erikdubbelboer/fasthttp"
	"github.com/stretchr/testify/assert"
)

func TestContextParam(t *testing.T) {
	c := NewContext(nil, nil)
	values := []string{"a", "b", "c", "d"}

	c.pvalues = values
	c.pnames = nil
	assert.Equal(t, "", c.Param(""))
	assert.Equal(t, "", c.Param("Name"))

	c.pnames = []string{"Name", "Age"}
	assert.Equal(t, "", c.Param(""))
	assert.Equal(t, "a", c.Param("Name"))
	assert.Equal(t, "b", c.Param("Age"))
	assert.Equal(t, "", c.Param("Xyz"))
}

func TestContextSetParam(t *testing.T) {
	c := NewContext(nil, nil)
	c.pnames = []string{"Name", "Age"}
	c.pvalues = []string{"abc", "123"}
	assert.Equal(t, "abc", c.Param("Name"))
	c.SetParam("Name", "xyz")
	assert.Equal(t, "xyz", c.Param("Name"))
	assert.Equal(t, "", c.Param("unknown"))
	c.SetParam("unknown", "xyz")
	assert.Equal(t, "xyz", c.Param("unknown"))
}

func TestContextInit(t *testing.T) {
	c := NewContext(nil)
	assert.Nil(t, c.RequestCtx)
	assert.Equal(t, 0, len(c.handlers))
	var ctx fasthttp.RequestCtx
	ctx.Request.Header.SetMethod("GET")
	ctx.Request.SetRequestURI("/users/")
	c.init(&ctx)
	assert.NotNil(t, &c.Response)
	assert.NotNil(t, &c.Request)
	assert.Equal(t, -1, c.index)
	assert.Nil(t, c.data)
}

func TestContextURL(t *testing.T) {
	router := New()
	router.Get("/users/<id:\\d+>/<action>/*").Name("users")
	c := &Context{router: router}
	assert.Equal(t, "/users/123/address/", c.URL("users", "id", 123, "action", "address"))
	assert.Equal(t, "", c.URL("abc", "id", 123, "action", "address"))
}

func TestContextGetSet(t *testing.T) {
	c := NewContext(nil, nil)
	c.init(nil)
	assert.Nil(t, c.Get("abc"))
	c.Set("abc", "123")
	c.Set("xyz", 123)
	assert.Equal(t, "123", c.Get("abc").(string))
	assert.Equal(t, 123, c.Get("xyz").(int))
}

func TestContextQueryForm(t *testing.T) {

	var ctx fasthttp.RequestCtx
	ctx.Request.Header.SetMethod("POST")
	ctx.Request.SetRequestURI("http://www.google.com/search?q=foo&q=bar&both=x&prio=1&empty=not")
	ctx.Request.SetBodyString("z=post&both=y&prio=2&empty=")
	ctx.Request.Header.SetContentType("application/x-www-form-urlencoded; param=value")

	c := NewContext(&ctx)
	assert.Equal(t, "foo", c.Query("q"))
	assert.Equal(t, "", c.Query("z"))
	assert.Equal(t, "123", c.Query("z", "123"))
	assert.Equal(t, "not", c.Query("empty", "123"))
	assert.Equal(t, "post", c.PostForm("z"))
	assert.Equal(t, "", c.PostForm("x"))
	assert.Equal(t, "123", c.PostForm("q", "123"))
	assert.Equal(t, "", c.PostForm("empty", "123"))
	assert.Equal(t, "y", c.Form("both"))
	assert.Equal(t, "", c.Form("x"))
	assert.Equal(t, "123", c.Form("x", "123"))
}

func TestContextNextAbort(t *testing.T) {
	c := testNewContext(
		testNormalHandler("a"),
		testNormalHandler("b"),
		testNormalHandler("c"),
	)
	assert.Nil(t, c.Next())
	assert.Equal(t, "<a/><b/><c/>", string(c.Response.Body()))

	c = testNewContext(
		testNextHandler("a"),
		testNextHandler("b"),
		testNextHandler("c"),
	)
	assert.Nil(t, c.Next())
	assert.Equal(t, "<a><b><c></c></b></a>", string(c.Response.Body()))

	c = testNewContext(
		testNextHandler("a"),
		testAbortHandler("b"),
		testNormalHandler("c"),
	)
	assert.Nil(t, c.Next())
	assert.Equal(t, "<a><b/></a>", string(c.Response.Body()))

	c = testNewContext(
		testNextHandler("a"),
		testErrorHandler("b"),
		testNormalHandler("c"),
	)
	err := c.Next()
	if assert.NotNil(t, err) {
		assert.Equal(t, "error:b", err.Error())
	}
	assert.Equal(t, "<a><b/></a>", string(c.Response.Body()))
}

func testNewContext(handlers ...Handler) *Context {
	var ctx fasthttp.RequestCtx
	ctx.Request.Header.SetMethod("GET")
	ctx.Request.Header.SetRequestURI("http://127.0.0.1/users")
	c := NewContext(&ctx)
	c.handlers = handlers
	return c
}

func testNextHandler(tag string) Handler {
	return func(c *Context) error {
		fmt.Fprintf(c.RequestCtx, "<%v>", tag)
		err := c.Next()
		fmt.Fprintf(c.RequestCtx, "</%v>", tag)
		return err
	}
}

func testAbortHandler(tag string) Handler {
	return func(c *Context) error {
		fmt.Fprintf(c.RequestCtx, "<%v/>", tag)
		c.Abort()
		return nil
	}
}

func testErrorHandler(tag string) Handler {
	return func(c *Context) error {
		fmt.Fprintf(c.RequestCtx, "<%v/>", tag)
		return errors.New("error:" + tag)
	}
}

func testNormalHandler(tag string) Handler {
	return func(c *Context) error {
		fmt.Fprintf(c.RequestCtx, "<%v/>", tag)
		return nil
	}
}
