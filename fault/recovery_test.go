// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package fault

import (
	"bytes"
	"errors"
	"fmt"
	"testing"

	"github.com/jackwhelpton/fasthttp-routing"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

func TestRecovery(t *testing.T) {
	var buf bytes.Buffer
	h := Recovery(getLogger(&buf))

	var ctx fasthttp.RequestCtx
	ctx.Request.Header.SetMethod("GET")
	ctx.Request.SetRequestURI("/users/")
	c := routing.NewContext(&ctx, h, handler1, handler2)
	assert.Nil(t, c.Next())
	assert.Equal(t, fasthttp.StatusInternalServerError, ctx.Response.StatusCode())
	assert.Equal(t, "abc", string(ctx.Response.Body()))
	assert.Equal(t, "abc", buf.String())

	buf.Reset()
	ctx.Response.Reset()
	c = routing.NewContext(&ctx, h, handler2)
	assert.Nil(t, c.Next())
	assert.Equal(t, fasthttp.StatusOK, ctx.Response.StatusCode())
	assert.Equal(t, "test", string(ctx.Response.Body()))
	assert.Equal(t, "", buf.String())

	buf.Reset()
	ctx.Response.Reset()
	c = routing.NewContext(&ctx, h, handler3, handler2)
	assert.Nil(t, c.Next())
	assert.Equal(t, fasthttp.StatusInternalServerError, ctx.Response.StatusCode())
	assert.Equal(t, "xyz", string(ctx.Response.Body()))
	assert.Contains(t, buf.String(), "recovery_test.go")
	assert.Contains(t, buf.String(), "xyz")

	buf.Reset()
	ctx.Response.Reset()
	c = routing.NewContext(&ctx, h, handler4, handler2)
	assert.Nil(t, c.Next())
	assert.Equal(t, fasthttp.StatusBadRequest, ctx.Response.StatusCode())
	assert.Equal(t, "123", string(ctx.Response.Body()))
	assert.Contains(t, buf.String(), "recovery_test.go")
	assert.Contains(t, buf.String(), "123")

	buf.Reset()
	ctx.Response.Reset()
	h = Recovery(getLogger(&buf), convertError)
	c = routing.NewContext(&ctx, h, handler3, handler2)
	assert.Nil(t, c.Next())
	assert.Equal(t, fasthttp.StatusInternalServerError, ctx.Response.StatusCode())
	assert.Equal(t, "123", string(ctx.Response.Body()))
	assert.Contains(t, buf.String(), "recovery_test.go")
	assert.Contains(t, buf.String(), "xyz")

	buf.Reset()
	ctx.Response.Reset()
	h = Recovery(getLogger(&buf), convertError)
	c = routing.NewContext(&ctx, h, handler1, handler2)
	assert.Nil(t, c.Next())
	assert.Equal(t, fasthttp.StatusInternalServerError, ctx.Response.StatusCode())
	assert.Equal(t, "123", string(ctx.Response.Body()))
	assert.Equal(t, "abc", buf.String())
}

func getLogger(buf *bytes.Buffer) LogFunc {
	return func(format string, a ...interface{}) {
		fmt.Fprintf(buf, format, a...)
	}
}

func handler1(c *routing.Context) error {
	return errors.New("abc")
}

func handler2(c *routing.Context) error {
	c.Write("test")
	return nil
}

func handler3(c *routing.Context) error {
	panic("xyz")
}

func handler4(c *routing.Context) error {
	panic(routing.NewHTTPError(fasthttp.StatusBadRequest, "123"))
}
