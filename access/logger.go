// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package access provides an access logging handler for the ozzo routing package.
package access

import (
	"fmt"
	"strings"
	"time"

	"github.com/jackwhelpton/fasthttp-routing/v2"
	"github.com/valyala/fasthttp"
)

// LogFunc logs a message using the given format and optional arguments.
// The usage of format and arguments is similar to that for fmt.Printf().
// LogFunc should be thread safe.
type LogFunc func(format string, a ...interface{})

// LogWriterFunc takes in the request and responseWriter objects as well
// as a float64 containing the elapsed time since the request first passed
// through this middleware and does whatever log writing it wants with that
// information.
// LogWriterFunc should be thread safe.
type LogWriterFunc func(ctx *fasthttp.RequestCtx, elapsed float64)

// CustomLogger returns a handler that calls the LogWriterFunc passed to it for every request.
// The LogWriterFunc is provided with the http.Request and LogResponseWriter objects for the
// request, as well as the elapsed time since the request first came through the middleware.
// LogWriterFunc can then do whatever logging it needs to do.
//
//     import (
//         "log"
//         "net/http"
//
//         "github.com/jackwhelpton/fasthttp-routing/v2"
//         "github.com/jackwhelpton/fasthttp-routing/v2/access"
//     )
//
//     func myCustomLogger(req http.Context, res access.LogResponseWriter, elapsed int64) {
//         // Do something with the request, response, and elapsed time data here
//     }
//     r := routing.New()
//     r.Use(access.CustomLogger(myCustomLogger))
func CustomLogger(loggerFunc LogWriterFunc) routing.Handler {
	return func(c *routing.Context) error {
		startTime := time.Now()

		err := c.Next()

		elapsed := float64(time.Since(startTime).Nanoseconds()) / 1e6
		loggerFunc(c.RequestCtx, elapsed)

		return err
	}
}

// Logger returns a handler that logs a message for every request.
// The access log messages contain information including client IPs, time used to serve each request, request line,
// response status and size.
//
//     import (
//         "log"
//         "github.com/jackwhelpton/fasthttp-routing/v2"
//         "github.com/jackwhelpton/fasthttp-routing/v2/access"
//     )
//
//     r := routing.New()
//     r.Use(access.Logger(log.Printf))
func Logger(logf LogFunc) routing.Handler {
	var logger = func(ctx *fasthttp.RequestCtx, elapsed float64) {
		ip := GetClientIP(ctx)
		req := fmt.Sprintf("%s %s %s", string(ctx.Request.Header.Method()), string(ctx.RequestURI()), string(ctx.Request.URI().Scheme()))
		logf(`[%s] [%.3fms] %s %d %d`, ip, elapsed, req, ctx.Response.StatusCode(), len(ctx.Response.Body()))

	}
	return CustomLogger(logger)
}

// GetClientIP returns the originating IP for a request.
func GetClientIP(ctx *fasthttp.RequestCtx) string {
	ip := string(ctx.Request.Header.Peek("X-Real-IP"))
	if ip == "" {
		ip = string(ctx.Request.Header.Peek("X-Forwarded-For"))
		if ip == "" {
			ip = ctx.RemoteAddr().String()
		}
	}
	if colon := strings.LastIndex(ip, ":"); colon != -1 {
		ip = ip[:colon]
	}
	return ip
}
