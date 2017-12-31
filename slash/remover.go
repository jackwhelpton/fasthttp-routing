// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package slash provides a trailing slash remover handler for the fasthttp-routing package.
package slash

import (
	"strings"

	"github.com/jackwhelpton/fasthttp-routing"
	"github.com/valyala/fasthttp"
)

// Remover returns a handler that removes the trailing slash (if any) from the requested URL.
// The handler will redirect the browser to the new URL without the trailing slash.
// The status parameter should be either fasthttp.StatusMovedPermanently (301) or fasthttp.StatusFound (302),
// which is to be used for redirecting GET requests. For other requests, the status code will be
// fasthttp.StatusTemporaryRedirect (307).
// If the original URL has no trailing slash, the handler will do nothing. For example,
//
//     import (
//         "github.com/jackwhelpton/fasthttp-routing"
//         "github.com/jackwhelpton/fasthttp-routing/slash"
//         "github.com/valyala/fasthttp"
//     )
//
//     r := routing.New()
//     r.Use(slash.Remover(fasthttp.StatusMovedPermanently))
//
// Note that Remover relies on HTTP redirection to remove the trailing slashes.
// If you do not want redirection, please set `Router.IgnoreTrailingSlash` to be true without using Remover.
func Remover(status int) routing.Handler {
	return func(c *routing.Context) error {
		if path := string(c.Request.URI().Path()); path != "/" && strings.HasSuffix(path, "/") {
			if !c.IsGet() {
				status = fasthttp.StatusTemporaryRedirect
			}
			// c.Redirect performs additional path normalization that is not desired,
			// so the Location header and status code are set explicitly instead.
			c.Response.Header.Set("Location", strings.TrimRight(path, "/"))
			c.Response.SetStatusCode(status)
			c.Abort()
		}
		return nil
	}
}
