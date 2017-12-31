// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package access

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"testing"

	"github.com/jackwhelpton/fasthttp-routing"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

func TestCustomLogger(t *testing.T) {
	var buf bytes.Buffer
	var customFunc = func(ctx *fasthttp.RequestCtx, elapsed float64) {
		logf := getLogger(&buf)
		ip := GetClientIP(ctx)
		req := fmt.Sprintf("%s %s %s", string(ctx.Request.Header.Method()), string(ctx.RequestURI()), string(ctx.Request.URI().Scheme()))
		logf(`[%s] [%.3fms] %s %d %d`, ip, elapsed, req, ctx.Response.StatusCode(), len(ctx.Response.Body()))
	}
	h := CustomLogger(customFunc)
	ctx := &fasthttp.RequestCtx{}
	ip, _ := net.ResolveIPAddr("ip", "192.168.100.1")
	ctx.Init(&fasthttp.Request{}, ip, nil)
	ctx.Request.Header.SetMethod("GET")
	ctx.Request.SetRequestURI("http://127.0.0.1/users")

	c := routing.NewContext(ctx, h, handler1)
	assert.NotNil(t, c.Next())
	assert.Contains(t, buf.String(), "GET http://127.0.0.1/users")
}

func TestLogger(t *testing.T) {
	var buf bytes.Buffer
	h := Logger(getLogger(&buf))
	ctx := &fasthttp.RequestCtx{}
	ip, _ := net.ResolveIPAddr("ip", "192.168.100.1")
	ctx.Init(&fasthttp.Request{}, ip, nil)
	ctx.Request.Header.SetMethod("GET")
	ctx.Request.SetRequestURI("http://127.0.0.1/users")

	c := routing.NewContext(ctx, h, handler1)
	assert.NotNil(t, c.Next())
	assert.Contains(t, buf.String(), "GET http://127.0.0.1/users")
}

func TestGetClientIP(t *testing.T) {
	ctx := &fasthttp.RequestCtx{}
	ip, _ := net.ResolveIPAddr("ip", "192.168.100.3")
	ctx.Init(&fasthttp.Request{}, ip, nil)
	ctx.Request.Header.SetMethod("GET")
	ctx.Request.SetRequestURI("/users/")
	ctx.Request.Header.Set("X-Real-IP", "192.168.100.1")
	ctx.Request.Header.Set("X-Forwarded-For", "192.168.100.2")

	assert.Equal(t, "192.168.100.1", GetClientIP(ctx))
	ctx.Request.Header.Del("X-Real-IP")
	assert.Equal(t, "192.168.100.2", GetClientIP(ctx))
	ctx.Request.Header.Del("X-Forwarded-For")
	assert.Equal(t, "192.168.100.3", GetClientIP(ctx))

	tcp, _ := net.ResolveTCPAddr("tcp", "192.168.100.3:8080")
	ctx.Request.Header.Reset()
	ctx.Init(&ctx.Request, tcp, nil)
	assert.Equal(t, "192.168.100.3", GetClientIP(ctx))
}

func getLogger(buf *bytes.Buffer) LogFunc {
	return func(format string, a ...interface{}) {
		fmt.Fprintf(buf, format, a...)
	}
}

func handler1(c *routing.Context) error {
	return errors.New("abc")
}
