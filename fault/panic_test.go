package fault

import (
	"bytes"
	"testing"

	"github.com/jackwhelpton/fasthttp-routing"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

func TestPanicHandler(t *testing.T) {
	var buf bytes.Buffer
	h := PanicHandler(getLogger(&buf))

	var ctx fasthttp.RequestCtx
	ctx.Request.Header.SetMethod("GET")
	ctx.Request.SetRequestURI("/users/")
	c := routing.NewContext(&ctx, h, handler3, handler2)
	err := c.Next()
	if assert.NotNil(t, err) {
		assert.Equal(t, "xyz", err.Error())
	}
	assert.NotEqual(t, "", buf.String())

	buf.Reset()
	ctx.Response.Reset()
	c = routing.NewContext(&ctx, h, handler2)
	assert.Nil(t, c.Next())
	assert.Equal(t, "", buf.String())

	buf.Reset()
	ctx.Response.Reset()
	h2 := ErrorHandler(getLogger(&buf))
	c = routing.NewContext(&ctx, h2, h, handler3, handler2)
	assert.Nil(t, c.Next())
	assert.Equal(t, fasthttp.StatusInternalServerError, ctx.Response.StatusCode())
	assert.Equal(t, "xyz", string(ctx.Response.Body()))
	assert.Contains(t, buf.String(), "recovery_test.go")
	assert.Contains(t, buf.String(), "xyz")
}
