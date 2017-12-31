// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package file

import (
	"strings"
	"testing"

	"github.com/jackwhelpton/fasthttp-routing"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

type (
	testLogger struct{ *testing.T }
)

func (l *testLogger) Printf(format string, v ...interface{}) {
	l.Logf(format, v...)
}

func TestParsePathMap(t *testing.T) {
	tests := []struct {
		id       string
		pathMap  PathMap
		from, to string
	}{
		{"t1", PathMap{}, "", ""},
		{"t2", PathMap{"/": ""}, "/", ""},
		{"t3", PathMap{"/": "ui/dist"}, "/", "ui/dist"},
		{"t4", PathMap{"/abc/123": "ui123/abc", "/abc": "/ui/abc", "/abc/xyz": "/xyzui/abc"}, "/abc,/abc/123,/abc/xyz", "/ui/abc,ui123/abc,/xyzui/abc"},
	}
	for _, test := range tests {
		af, at := parsePathMap(test.pathMap)
		assert.Equal(t, test.from, strings.Join(af, ","), test.id)
		assert.Equal(t, test.to, strings.Join(at, ","), test.id)
	}
}

func TestMatchPath(t *testing.T) {
	tests := []struct {
		id        string
		from, to  []string
		url, path string
		found     bool
	}{
		{"t1", []string{}, []string{}, "", "", false},

		{"t2.1", []string{"/"}, []string{"/www"}, "", "", false},
		{"t2.2", []string{"/"}, []string{"/www"}, "/", "/www", true},
		{"t2.3", []string{"/"}, []string{"/www"}, "/index", "/wwwindex", true},
		{"t2.4", []string{"/"}, []string{"/www/"}, "/index", "/www/index", true},
		{"t2.5", []string{"/"}, []string{"/www/"}, "/index/", "/www/index/", true},
		{"t2.6", []string{"/"}, []string{"/www/"}, "index", "", false},
		{"t2.7", []string{""}, []string{""}, "/", "/", true},
		{"t2.7", []string{""}, []string{""}, "/index.html", "/index.html", true},

		{"t3.1", []string{"/", "/css", "/js"}, []string{"/www/others", "/www/ui/css", "/www/ui/js"}, "", "", false},
		{"t3.2", []string{"/", "/css", "/js"}, []string{"/www/others", "/www/ui/css", "/www/ui/js"}, "/", "/www/others", true},
		{"t3.3", []string{"/", "/css", "/js"}, []string{"/www/others", "/www/ui/css", "/www/ui/js"}, "/css", "/www/ui/css", true},
		{"t3.4", []string{"/", "/css", "/js"}, []string{"/www/others", "/www/ui/css", "/www/ui/js"}, "/abc", "/www/othersabc", true},
		{"t3.5", []string{"/", "/css", "/js"}, []string{"/www/others", "/www/ui/css", "/www/ui/js"}, "/css2", "/www/ui/css2", true},

		{"t4.1", []string{"/css/abc", "/css"}, []string{"/www/abc", "/www/css"}, "/css/abc", "/www/css/abc", true},
	}
	for _, test := range tests {
		path, found := matchPath(test.url, test.from, test.to)
		assert.Equal(t, test.found, found, test.id)
		if found {
			assert.Equal(t, test.path, path, test.id)
		}
	}
}

func TestContent(t *testing.T) {
	h := Content("testdata/index.html")
	var ctx fasthttp.RequestCtx
	ctx.Request.Header.SetMethod("GET")
	ctx.Request.SetRequestURI("/index.html")
	c := routing.NewContext(&ctx)
	err := h(c)
	assert.Nil(t, err)
	assert.Equal(t, "hello\n", string(ctx.Response.Body()))

	h = Content("testdata/index.go")
	ctx.Response.Reset()
	err = h(c)
	if assert.NotNil(t, err) {
		assert.Equal(t, fasthttp.StatusNotFound, err.(routing.HTTPError).StatusCode())
	}

	h = Content("testdata/css")
	ctx.Response.Reset()
	err = h(c)
	if assert.NotNil(t, err) {
		assert.Equal(t, fasthttp.StatusNotFound, err.(routing.HTTPError).StatusCode())
	}

	h = Content("testdata/index.html")
	ctx.Request.Header.Reset()
	ctx.Request.Header.SetMethod("POST")
	ctx.Response.Reset()
	err = h(c)
	if assert.NotNil(t, err) {
		assert.Equal(t, fasthttp.StatusMethodNotAllowed, err.(routing.HTTPError).StatusCode())
	}
}

func TestServer(t *testing.T) {
	h := Server(PathMap{"/css": "/testdata/css"})
	tests := []struct {
		id          string
		method, url string
		status      int
		body        string
	}{
		{"t1", "GET", "/css/main.css", 0, "body {}\n"},
		{"t2", "HEAD", "/css/main.css", 0, ""},
		{"t3", "GET", "/css/main2.css", fasthttp.StatusNotFound, ""},
		{"t4", "POST", "/css/main.css", fasthttp.StatusMethodNotAllowed, ""},
		{"t5", "GET", "/css", fasthttp.StatusNotFound, ""},
	}

	ctx := fasthttp.RequestCtx{}
	ctx.Init(&fasthttp.Request{}, nil, newTestLogger(t))
	c := routing.NewContext(&ctx)

	for _, test := range tests {
		ctx.Request.Header.Reset()
		ctx.Request.Header.SetMethod(test.method)
		ctx.Request.SetRequestURI(test.url)
		ctx.Response.Reset()

		err := h(c)
		if test.status == 0 {
			assert.Nil(t, err, test.id)
			assert.Equal(t, test.body, string(c.Response.Body()), test.id)
		} else if assert.NotNil(t, err, test.id) {
			assert.Equal(t, test.status, err.(routing.HTTPError).StatusCode(), test.id)
		}
	}

	h = Server(PathMap{"/css": "/testdata/css"}, ServerOptions{
		IndexFile: "index.html",
		Allow: func(c *routing.Context, path string) bool {
			return path != "/testdata/css/main.css"
		},
	})

	ctx.Request.SetRequestURI("/css/main.css")
	ctx.Response.Reset()
	err := h(c)
	assert.NotNil(t, err)

	ctx.Request.SetRequestURI("/css")
	ctx.Response.Reset()
	err = h(c)
	assert.Nil(t, err)
	assert.Equal(t, "css.html\n", string(ctx.Response.Body()))

	{
		// with CatchAll option
		h = Server(PathMap{"/css": "/testdata/css"}, ServerOptions{
			IndexFile:    "index.html",
			CatchAllFile: "testdata/index.html",
			Allow: func(c *routing.Context, path string) bool {
				return path != "/testdata/css/main.css"
			},
		})

		ctx := fasthttp.RequestCtx{}
		ctx.Init(&fasthttp.Request{}, nil, newTestLogger(t))
		ctx.Request.Header.SetMethod("GET")
		ctx.Request.SetRequestURI("/css/main.css")
		c := routing.NewContext(&ctx)

		err := h(c)
		assert.NotNil(t, err)

		ctx.Request.SetRequestURI("/css")
		ctx.Response.Reset()
		err = h(c)
		assert.Nil(t, err)
		assert.Equal(t, "css.html\n", string(ctx.Response.Body()))

		ctx.Request.SetRequestURI("/css2")
		ctx.Response.Reset()
		err = h(c)
		assert.Nil(t, err)
		assert.Equal(t, "hello\n", string(ctx.Response.Body()))
	}
}

func newTestLogger(t *testing.T) *testLogger {
	return &testLogger{t}
}
