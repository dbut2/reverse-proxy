package rp

import (
	"context"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"google.golang.org/api/idtoken"
)

type Selector struct {
	rule    Rule
	service string
}

func Select(service string, rule Rule) Selector {
	return Selector{rule: rule, service: service}
}

func New(selectors ...Selector) *httputil.ReverseProxy {
	return &httputil.ReverseProxy{
		Director: func(r *http.Request) {
			var targetURL string
			var modifier func()

			for _, selector := range selectors {
				match, mod := selector.rule(r)
				if !match {
					continue
				}

				targetURL = selector.service
				modifier = mod

				break
			}

			if targetURL == "" {
				return
			}

			target, _ := url.Parse(targetURL)
			path := target.Path + r.URL.Path
			r.URL = target
			r.URL.Path = path
			r.Header.Set("X-Forwarded-Host", r.Host)
			r.Host = target.Host

			if modifier != nil {
				modifier()
			}

			ctx := context.Background()
			tokenSource, err := idtoken.NewTokenSource(ctx, targetURL)
			if err != nil {
				log.Printf("Failed to create token source: %v\n", err)
				return
			}

			token, err := tokenSource.Token()
			if err != nil {
				log.Printf("Failed to obtain an OIDC token: %v\n", err)
				return
			}

			r.Header.Add("Authorization", "Bearer "+token.AccessToken)
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			http.Error(w, "Not found", http.StatusNotFound)
		},
	}
}
