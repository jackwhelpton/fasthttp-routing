// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package content

import (
	"testing"

	"github.com/jackwhelpton/fasthttp-routing"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

func TestJSONFormatter(t *testing.T) {
	var ctx fasthttp.RequestCtx
	w := &JSONDataWriter{}
	w.SetHeader(&ctx.Response.Header)
	err := w.Write(&ctx, "xyz")
	assert.Nil(t, err)
	assert.Equal(t, "application/json", string(ctx.Response.Header.ContentType()))
	assert.Equal(t, "\"xyz\"\n", string(ctx.Response.Body()))
}

func TestXMLFormatter(t *testing.T) {
	var ctx fasthttp.RequestCtx
	w := &XMLDataWriter{}
	w.SetHeader(&ctx.Response.Header)
	err := w.Write(&ctx, "xyz")
	assert.Nil(t, err)
	assert.Equal(t, "application/xml; charset=UTF-8", string(ctx.Response.Header.ContentType()))
	assert.Equal(t, "<string>xyz</string>", string(ctx.Response.Body()))
}

func TestHTMLFormatter(t *testing.T) {
	var ctx fasthttp.RequestCtx
	w := &HTMLDataWriter{}
	w.SetHeader(&ctx.Response.Header)
	err := w.Write(&ctx, "xyz")
	assert.Nil(t, err)
	assert.Equal(t, "text/html; charset=UTF-8", string(ctx.Response.Header.ContentType()))
	assert.Equal(t, "xyz", string(ctx.Response.Body()))
}

func TestTypeNegotiator(t *testing.T) {
	var ctx fasthttp.RequestCtx
	ctx.Request.Header.SetMethod("GET")
	ctx.Request.SetRequestURI("/users/")
	ctx.Request.Header.Set("Accept", "application/xml")
	c := routing.NewContext(&ctx)

	// test no arguments
	h := TypeNegotiator()
	assert.Nil(t, h(c))
	c.Write("xyz")
	assert.Equal(t, "text/html; charset=UTF-8", string(c.Response.Header.ContentType()))
	assert.Equal(t, "xyz", string(c.Response.Body()))
	ctx.Response.Reset()

	// test format chosen based on Accept
	h = TypeNegotiator(JSON, XML)
	assert.Nil(t, h(c))
	assert.Nil(t, c.Write("xyz"))
	assert.Equal(t, "application/xml; charset=UTF-8", string(c.Response.Header.ContentType()))
	assert.Equal(t, "<string>xyz</string>", string(c.Response.Body()))
	ctx.Response.Reset()

	// test default format used when no match
	ctx.Request.Header.Set("Accept", "application/pdf")
	assert.Nil(t, h(c))
	assert.Nil(t, c.Write("xyz"))
	assert.Equal(t, "application/json", string(c.Response.Header.ContentType()))
	assert.Equal(t, "\"xyz\"\n", string(c.Response.Body()))

	assert.Panics(t, func() {
		TypeNegotiator("unknown")
	})
}

var (
	v1JSON = "application/json;v=1"
	v2JSON = "application/json;v=2"
)

type JSONDataWriter1 struct {
	JSONDataWriter
}

func (w *JSONDataWriter1) SetHeader(h *fasthttp.ResponseHeader) {
	h.SetContentType(v1JSON)
}

type JSONDataWriter2 struct {
	JSONDataWriter
}

func (w *JSONDataWriter2) SetHeader(h *fasthttp.ResponseHeader) {
	h.SetContentType(v2JSON)
}

func TestTypeNegotiatorWithVersion(t *testing.T) {
	var ctx fasthttp.RequestCtx
	ctx.Request.Header.SetMethod("GET")
	ctx.Request.SetRequestURI("/users/")
	ctx.Request.Header.Set("Accept", "application/xml,"+v1JSON)
	c := routing.NewContext(&ctx)

	// test no arguments
	h := TypeNegotiator()
	assert.Nil(t, h(c))
	c.Write("xyz")
	assert.Equal(t, "text/html; charset=UTF-8", string(c.Response.Header.ContentType()))
	assert.Equal(t, "xyz", string(c.Response.Body()))

	DataWriters[v1JSON] = &JSONDataWriter1{}
	DataWriters[v2JSON] = &JSONDataWriter2{}

	// test format chosen based on Accept
	ctx.Response.Reset()
	h = TypeNegotiator(v2JSON, v1JSON, XML)
	assert.Nil(t, h(c))
	assert.Nil(t, c.Write("xyz"))
	assert.Equal(t, v1JSON, string(c.Response.Header.ContentType()))
	assert.Equal(t, `"xyz"`+"\n", string(c.Response.Body()))

	// test default format used when no match
	ctx.Response.Reset()
	ctx.Request.Header.Set("Accept", "application/pdf")
	assert.Nil(t, h(c))
	assert.Nil(t, c.Write("xyz"))
	assert.Equal(t, v2JSON, string(c.Response.Header.ContentType()))
	assert.Equal(t, "\"xyz\"\n", string(c.Response.Body()))

	assert.Panics(t, func() {
		TypeNegotiator("unknown")
	})
}
