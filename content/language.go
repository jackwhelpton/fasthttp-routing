// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package content

import (
	"golang.org/x/text/language"

	"github.com/jackwhelpton/fasthttp-routing"
	"github.com/valyala/fasthttp"
)

// Language is the key used to store and retrieve the chosen language in routing.Context
const Language = "Language"

// LanguageNegotiator returns a content language negotiation handler.
//
// The method takes a list of languages (locale IDs) that are supported by the application.
// The negotiator will determine the best language to use by checking the Accept-Language request header.
// If no match is found, the first language will be used.
//
// In a handler, you can access the chosen language through routing.Context like the following:
//
//     func(c *routing.Context) error {
//         language := c.Get(content.Language).(string)
//     }
//
// If you do not specify languages, the negotiator will set the language to be "en-US".
func LanguageNegotiator(languages ...string) routing.Handler {
	if len(languages) == 0 {
		languages = []string{"en-US"}
	}
	defaultLanguage := languages[0]

	return func(c *routing.Context) error {
		language := negotiateLanguage(c.RequestCtx, languages, defaultLanguage)
		c.Set(Language, language)
		return nil
	}
}

// negotiateLanguage negotiates the acceptable language according to the Accept-Language HTTP header.
func negotiateLanguage(ctx *fasthttp.RequestCtx, offers []string, defaultOffer string) string {
	tags := make([]language.Tag, 0, len(offers))
	for _, o := range offers {
		if t, err := language.Parse(o); err == nil {
			tags = append(tags, t)
		}
	}

	accept, _, _ := language.ParseAcceptLanguage(string(ctx.Request.Header.Peek("Accept-Language")))
	if _, i, c := language.NewMatcher(tags).Match(accept...); c != language.No {
		return offers[i]
	}
	return defaultOffer
}
