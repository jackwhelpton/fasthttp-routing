package routing

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/valyala/fasthttp"

	"github.com/stretchr/testify/assert"
)

func TestRouterNotFound(t *testing.T) {
	r := New()
	h := func(c *Context) error {
		fmt.Fprint(c.RequestCtx, "ok")
		return nil
	}
	r.Get("/users", h)
	r.Post("/users", h)
	r.NotFound(MethodNotAllowedHandler, NotFoundHandler)

	var ctx fasthttp.RequestCtx
	ctx.Request.Header.SetMethod("GET")
	ctx.Request.SetRequestURI("/users")
	r.HandleRequest(&ctx)
	assert.Equal(t, "ok", string(ctx.Response.Body()), "response body")
	assert.Equal(t, http.StatusOK, ctx.Response.Header.StatusCode(), "HTTP status code")

	ctx = fasthttp.RequestCtx{}
	ctx.Request.Header.SetMethod("PUT")
	ctx.Request.SetRequestURI("/users")
	r.HandleRequest(&ctx)
	assert.Equal(t, "GET, OPTIONS, POST", string(ctx.Response.Header.Peek("Allow")), "Allow header")
	assert.Equal(t, http.StatusMethodNotAllowed, ctx.Response.Header.StatusCode(), "HTTP status code")

	ctx = fasthttp.RequestCtx{}
	ctx.Request.Header.SetMethod("OPTIONS")
	ctx.Request.SetRequestURI("/users")
	r.HandleRequest(&ctx)
	assert.Equal(t, "GET, OPTIONS, POST", string(ctx.Response.Header.Peek("Allow")), "Allow header")
	assert.Equal(t, http.StatusOK, ctx.Response.Header.StatusCode(), "HTTP status code")

	ctx = fasthttp.RequestCtx{}
	ctx.Request.Header.SetMethod("GET")
	ctx.Request.SetRequestURI("/users/")
	r.HandleRequest(&ctx)
	assert.Equal(t, "", string(ctx.Response.Header.Peek("Allow")), "Allow header")
	assert.Equal(t, http.StatusNotFound, ctx.Response.Header.StatusCode(), "HTTP status code")

	r.IgnoreTrailingSlash = true
	ctx = fasthttp.RequestCtx{}
	ctx.Request.Header.SetMethod("GET")
	ctx.Request.SetRequestURI("/users/")
	r.HandleRequest(&ctx)
	assert.Equal(t, "ok", string(ctx.Response.Body()), "response body")
	assert.Equal(t, http.StatusOK, ctx.Response.Header.StatusCode(), "HTTP status code")
}

func TestRouterUse(t *testing.T) {
	r := New()
	assert.Equal(t, 2, len(r.notFoundHandlers))
	r.Use(NotFoundHandler)
	assert.Equal(t, 3, len(r.notFoundHandlers))
}

func TestRouterRoute(t *testing.T) {
	r := New()
	r.Get("/users").Name("users")
	assert.NotNil(t, r.Route("users"))
	assert.Nil(t, r.Route("users2"))
}

func TestRouterAdd(t *testing.T) {
	r := New()
	assert.Equal(t, 0, r.maxParams)
	r.add("GET", "/users/<id>", nil)
	assert.Equal(t, 1, r.maxParams)
}

func TestRouterFind(t *testing.T) {
	r := New()
	r.add("GET", "/users/<id>", []Handler{NotFoundHandler})
	handlers, params := r.Find("GET", "/users/1")
	assert.Equal(t, 1, len(handlers))
	if assert.Equal(t, 1, len(params)) {
		assert.Equal(t, "1", params["id"])
	}
}

func TestRouterNormalizeRequestPath(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"/", "/"},
		{"/users", "/users"},
		{"/users/", "/users"},
		{"/users//", "/users"},
		{"///", "/"},
	}
	r := New()
	r.IgnoreTrailingSlash = true
	for _, test := range tests {
		result := r.normalizeRequestPath(test.path)
		assert.Equal(t, test.expected, result)
	}
}

func TestRouterHandleError(t *testing.T) {
	r := New()
	c := &Context{}
	r.handleError(c, errors.New("abc"))
	assert.Equal(t, http.StatusInternalServerError, c.Response.Header.StatusCode())

	c = &Context{}
	r.handleError(c, NewHTTPError(http.StatusNotFound))
	assert.Equal(t, http.StatusNotFound, c.Response.Header.StatusCode())
}

func TestRequestHandler(t *testing.T) {
	var ctx fasthttp.RequestCtx
	ctx.Request.Header.SetMethod("GET")
	ctx.Request.SetRequestURI("/users/")
	c := NewContext(&ctx)

	h := RequestHandlerFunc(func(c *fasthttp.RequestCtx) { c.NotFound() })
	assert.Nil(t, h(c))
	assert.Equal(t, http.StatusNotFound, ctx.Response.Header.StatusCode())
}
