// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package cors

import (
	"testing"
	"time"

	"github.com/jackwhelpton/fasthttp-routing"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

func TestBuildAllowMap(t *testing.T) {
	m := buildAllowMap("", false)
	assert.Equal(t, 0, len(m))

	m = buildAllowMap("", true)
	assert.Equal(t, 0, len(m))

	m = buildAllowMap("GET , put", false)
	assert.Equal(t, 2, len(m))
	assert.True(t, m["GET"])
	assert.True(t, m["PUT"])
	assert.False(t, m["put"])

	m = buildAllowMap("GET , put", true)
	assert.Equal(t, 2, len(m))
	assert.True(t, m["GET"])
	assert.False(t, m["PUT"])
	assert.True(t, m["put"])
}

func TestOptionsInit(t *testing.T) {
	opts := &Options{
		AllowHeaders: "Accept, Accept-Language",
		AllowMethods: "PATCH, PUT",
		AllowOrigins: "https://example.com",
	}
	opts.init()
	assert.Equal(t, 2, len(opts.allowHeaderMap))
	assert.Equal(t, 2, len(opts.allowMethodMap))
	assert.Equal(t, 1, len(opts.allowOriginMap))
}

func TestOptionsIsOriginAllowed(t *testing.T) {
	tests := []struct {
		id      string
		allowed string
		origin  string
		result  bool
	}{
		{"t1", "*", "http://example.com", true},
		{"t2", "null", "http://example.com", false},
		{"t3", "http://foo.com", "http://example.com", false},
		{"t4", "http://example.com", "http://example.com", true},
	}

	for _, test := range tests {
		opts := &Options{AllowOrigins: test.allowed}
		opts.init()
		assert.Equal(t, test.result, opts.isOriginAllowed(test.origin), test.id)
	}
}

func TestOptionsSetOriginHeaders(t *testing.T) {
	headers := &fasthttp.ResponseHeader{}
	opts := &Options{
		AllowOrigins:     "https://example.com, https://foo.com",
		AllowCredentials: false,
	}
	opts.setOriginHeader("https://example.com", headers)
	assert.Equal(t, "https://example.com", string(headers.Peek(headerAllowOrigin)))
	assert.Equal(t, "", string(headers.Peek(headerAllowCredentials)))

	headers = &fasthttp.ResponseHeader{}
	opts = &Options{
		AllowOrigins:     "*",
		AllowCredentials: false,
	}
	opts.setOriginHeader("https://example.com", headers)
	assert.Equal(t, "*", string(headers.Peek(headerAllowOrigin)))
	assert.Equal(t, "", string(headers.Peek(headerAllowCredentials)))

	headers = &fasthttp.ResponseHeader{}
	opts = &Options{
		AllowOrigins:     "https://example.com, https://foo.com",
		AllowCredentials: true,
	}
	opts.setOriginHeader("https://example.com", headers)
	assert.Equal(t, "https://example.com", string(headers.Peek(headerAllowOrigin)))
	assert.Equal(t, "true", string(headers.Peek(headerAllowCredentials)))

	headers = &fasthttp.ResponseHeader{}
	opts = &Options{
		AllowOrigins:     "*",
		AllowCredentials: true,
	}
	opts.setOriginHeader("https://example.com", headers)
	assert.Equal(t, "https://example.com", string(headers.Peek(headerAllowOrigin)))
	assert.Equal(t, "true", string(headers.Peek(headerAllowCredentials)))
}

func TestOptionsSetActualHeaders(t *testing.T) {
	headers := &fasthttp.ResponseHeader{}
	opts := &Options{
		AllowOrigins:     "https://example.com, https://foo.com",
		AllowCredentials: false,
		ExposeHeaders:    "X-Ping, X-Pong",
	}
	opts.init()
	opts.setActualHeaders("https://example.com", headers)
	assert.Equal(t, "https://example.com", string(headers.Peek(headerAllowOrigin)))
	assert.Equal(t, "X-Ping, X-Pong", string(headers.Peek(headerExposeHeaders)))

	opts.ExposeHeaders = ""
	headers = &fasthttp.ResponseHeader{}
	opts.setActualHeaders("https://example.com", headers)
	assert.Equal(t, "https://example.com", string(headers.Peek(headerAllowOrigin)))
	assert.Equal(t, "", string(headers.Peek(headerExposeHeaders)))

	headers = &fasthttp.ResponseHeader{}
	opts.setActualHeaders("https://bar.com", headers)
	assert.Equal(t, "", string(headers.Peek(headerAllowOrigin)))
}

func TestOptionsIsPreflightAllowed(t *testing.T) {
	opts := &Options{
		AllowOrigins:     "https://example.com, https://foo.com",
		AllowMethods:     "PUT, PATCH",
		AllowCredentials: false,
		ExposeHeaders:    "X-Ping, X-Pong",
	}
	opts.init()
	allowed, headers := opts.isPreflightAllowed("https://foo.com", "PUT", "")
	assert.True(t, allowed)
	assert.Equal(t, "", headers)

	opts = &Options{
		AllowOrigins: "https://example.com, https://foo.com",
		AllowMethods: "PUT, PATCH",
	}
	opts.init()
	allowed, headers = opts.isPreflightAllowed("https://foo.com", "DELETE", "")
	assert.False(t, allowed)
	assert.Equal(t, "", headers)

	opts = &Options{
		AllowOrigins: "https://example.com, https://foo.com",
		AllowMethods: "PUT, PATCH",
		AllowHeaders: "X-Ping, X-Pong",
	}
	opts.init()
	allowed, headers = opts.isPreflightAllowed("https://foo.com", "PUT", "X-Unknown")
	assert.False(t, allowed)
	assert.Equal(t, "", headers)
}

func TestOptionsSetPreflightHeaders(t *testing.T) {
	headers := &fasthttp.ResponseHeader{}
	opts := &Options{
		AllowOrigins:     "https://example.com, https://foo.com",
		AllowMethods:     "PUT, PATCH",
		AllowHeaders:     "X-Ping, X-Pong",
		AllowCredentials: false,
		ExposeHeaders:    "X-Ping, X-Pong",
		MaxAge:           time.Duration(100) * time.Second,
	}
	opts.init()
	opts.setPreflightHeaders("https://bar.com", "PUT", "", headers)
	assert.Equal(t, "", string(headers.Peek(headerAllowOrigin)))
	assert.Equal(t, "", string(headers.Peek(headerAllowMethods)))
	assert.Equal(t, "", string(headers.Peek(headerMaxAge)))
	assert.Equal(t, "", string(headers.Peek(headerAllowHeaders)))

	headers = &fasthttp.ResponseHeader{}
	opts.setPreflightHeaders("https://foo.com", "PUT", "X-Pong", headers)
	assert.Equal(t, "https://foo.com", string(headers.Peek(headerAllowOrigin)))
	assert.Equal(t, "PUT, PATCH", string(headers.Peek(headerAllowMethods)))
	assert.Equal(t, "100", string(headers.Peek(headerMaxAge)))
	assert.Equal(t, "X-Pong", string(headers.Peek(headerAllowHeaders)))

	headers = &fasthttp.ResponseHeader{}
	opts = &Options{
		AllowOrigins: "*",
		AllowMethods: "*",
		AllowHeaders: "*",
	}
	opts.init()
	opts.setPreflightHeaders("https://bar.com", "PUT", "X-Pong", headers)
	assert.Equal(t, "*", string(headers.Peek(headerAllowOrigin)))
	assert.Equal(t, "PUT", string(headers.Peek(headerAllowMethods)))
	assert.Equal(t, "X-Pong", string(headers.Peek(headerAllowHeaders)))
}

func TestHandlers(t *testing.T) {
	h := Handler(Options{
		AllowOrigins: "https://example.com, https://foo.com",
		AllowMethods: "PUT, PATCH",
	})
	var ctx fasthttp.RequestCtx
	ctx.Request.Header.SetMethod("OPTIONS")
	ctx.Request.SetRequestURI("/users/")
	ctx.Request.Header.Set("Origin", "https://example.com")
	ctx.Request.Header.Set("Access-Control-Request-Method", "PATCH")
	c := routing.NewContext(&ctx)
	assert.Nil(t, h(c))
	assert.Equal(t, "https://example.com", string(c.Response.Header.Peek(headerAllowOrigin)))

	ctx.Request.Header.Reset()
	ctx.Request.Header.SetMethod("PATCH")
	ctx.Request.Header.Set("Origin", "https://example.com")
	ctx.Response.Reset()
	assert.Nil(t, h(c))
	assert.Equal(t, "https://example.com", string(c.Response.Header.Peek(headerAllowOrigin)))

	ctx.Request.Header.Reset()
	ctx.Request.Header.SetMethod("PATCH")
	ctx.Response.Reset()
	assert.Nil(t, h(c))
	assert.Equal(t, "", string(c.Response.Header.Peek(headerAllowOrigin)))

	ctx.Request.Header.Reset()
	ctx.Request.Header.SetMethod("OPTIONS")
	ctx.Request.Header.Set("Origin", "https://example.com")
	ctx.Response.Reset()
	assert.Nil(t, h(c))
	assert.Equal(t, "", string(c.Response.Header.Peek(headerAllowOrigin)))
}
