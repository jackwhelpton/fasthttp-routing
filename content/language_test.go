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

func TestLanguageNegotiator(t *testing.T) {
	var ctx fasthttp.RequestCtx
	ctx.Request.Header.SetMethod("GET")
	ctx.Request.SetRequestURI("/users/")
	ctx.Request.Header.Set("Accept-Language", "ru-RU;q=0.6,ru;q=0.5,zh-CN;q=1.0,zh;q=0.9")

	// test no arguments
	h := LanguageNegotiator()
	c := routing.NewContext(&ctx)
	assert.Nil(t, h(c))
	assert.Equal(t, "en-US", c.Get(Language))

	h = LanguageNegotiator("ru-RU", "ru", "zh", "zh-CN")
	c = routing.NewContext(&ctx)
	assert.Nil(t, h(c))
	assert.Equal(t, "zh-CN", c.Get(Language))

	h = LanguageNegotiator("en", "en-US")
	c = routing.NewContext(&ctx)
	assert.Nil(t, h(c))
	assert.Equal(t, "en", c.Get(Language))

	ctx.Request.Header.Set("Accept-Language", "ru-RU;q=0")
	c = routing.NewContext(&ctx)
	h = LanguageNegotiator("en", "ru-RU")
	assert.Nil(t, h(c))
	assert.Equal(t, "en", c.Get(Language))
}
