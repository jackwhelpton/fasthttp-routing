package fault

import (
	"bytes"
	"errors"
	"testing"

	"github.com/jackwhelpton/fasthttp-routing"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

func TestErrorHandler(t *testing.T) {
	var buf bytes.Buffer
	h := ErrorHandler(getLogger(&buf))

	var ctx fasthttp.RequestCtx
	ctx.Request.Header.SetMethod("GET")
	ctx.Request.SetRequestURI("/users/")
	c := routing.NewContext(&ctx, h, handler1, handler2)
	assert.Nil(t, c.Next())
	assert.Equal(t, fasthttp.StatusInternalServerError, c.Response.StatusCode())
	assert.Equal(t, "abc", string(c.Response.Body()))
	assert.Equal(t, "abc", buf.String())

	buf.Reset()
	ctx.Response.Reset()
	c = routing.NewContext(&ctx, h, handler2)
	assert.Nil(t, c.Next())
	assert.Equal(t, fasthttp.StatusOK, c.Response.StatusCode())
	assert.Equal(t, "test", string(c.Response.Body()))
	assert.Equal(t, "", buf.String())

	buf.Reset()
	ctx.Response.Reset()
	h = ErrorHandler(getLogger(&buf), convertError)
	c = routing.NewContext(&ctx, h, handler1, handler2)
	assert.Nil(t, c.Next())
	assert.Equal(t, fasthttp.StatusInternalServerError, c.Response.StatusCode())
	assert.Equal(t, "123", string(c.Response.Body()))
	assert.Equal(t, "abc", buf.String())

	buf.Reset()
	ctx.Response.Reset()
	h = ErrorHandler(nil)
	c = routing.NewContext(&ctx, h, handler1, handler2)
	assert.Nil(t, c.Next())
	assert.Equal(t, fasthttp.StatusInternalServerError, c.Response.StatusCode())
	assert.Equal(t, "abc", string(c.Response.Body()))
	assert.Equal(t, "", buf.String())
}

func Test_writeError(t *testing.T) {
	var ctx fasthttp.RequestCtx
	ctx.Request.Header.SetMethod("GET")
	ctx.Request.SetRequestURI("/users/")
	c := routing.NewContext(&ctx)
	writeError(c, errors.New("abc"))
	assert.Equal(t, fasthttp.StatusInternalServerError, ctx.Response.StatusCode())
	assert.Equal(t, "abc", string(ctx.Response.Body()))

	ctx.Response.Reset()
	writeError(c, routing.NewHTTPError(fasthttp.StatusNotFound, "xyz"))
	assert.Equal(t, fasthttp.StatusNotFound, ctx.Response.StatusCode())
	assert.Equal(t, "xyz", string(ctx.Response.Body()))
}

func convertError(c *routing.Context, err error) error {
	return errors.New("123")
}
