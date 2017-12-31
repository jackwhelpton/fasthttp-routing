// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package slash

import (
	"testing"

	"github.com/jackwhelpton/fasthttp-routing"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

func TestRemover(t *testing.T) {
	h := Remover(fasthttp.StatusMovedPermanently)

	var ctx fasthttp.RequestCtx
	ctx.Request.Header.SetMethod("GET")
	ctx.Request.SetRequestURI("/users/")
	c := routing.NewContext(&ctx)
	err := h(c)
	assert.Nil(t, err, "return value is nil")
	assert.Equal(t, fasthttp.StatusMovedPermanently, c.Response.StatusCode())
	assert.Equal(t, "/users", string(c.Response.Header.Peek("Location")))

	ctx.Request.SetRequestURI("/")
	ctx.Response.Reset()
	h(c)
	assert.Equal(t, fasthttp.StatusOK, c.Response.StatusCode())
	assert.Equal(t, "", string(c.Response.Header.Peek("Location")))

	ctx.Request.SetRequestURI("/users")
	ctx.Response.Reset()
	h(c)
	assert.Equal(t, fasthttp.StatusOK, c.Response.StatusCode())
	assert.Equal(t, "", string(c.Response.Header.Peek("Location")))

	ctx.Request.Header.Reset()
	ctx.Request.Header.SetMethod("POST")
	ctx.Request.SetRequestURI("/users/")
	ctx.Response.Reset()
	h(c)
	assert.Equal(t, fasthttp.StatusTemporaryRedirect, c.Response.StatusCode())
	assert.Equal(t, "/users", string(c.Response.Header.Peek("Location")))
}
