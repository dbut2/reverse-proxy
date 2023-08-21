package rp

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"google.golang.org/api/idtoken"
)

type Selector struct {
	rule    Rule
	service *url.URL
	opts    []Option
}

func Select(service string, rule Rule, opts ...Option) Selector {
	serviceURL, err := url.Parse(service)
	if err != nil {
		panic(err.Error())
	}
	return Selector{rule: rule, service: serviceURL, opts: opts}
}

type Option func(*http.Request)

func WithOIDC() Option {
	return func(r *http.Request) {
		var audience url.URL

		audience.Scheme = r.URL.Scheme
		audience.User = r.URL.User
		audience.Host = r.URL.Host

		tokenSource, err := idtoken.NewTokenSource(r.Context(), audience.String())
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
	}
}

func New(selectors ...Selector) *httputil.ReverseProxy {
	return &httputil.ReverseProxy{
		Rewrite: func(r *httputil.ProxyRequest) {
			var targetURL *url.URL
			var opts []Option

			matched := false
			for _, selector := range selectors {
				match := selector.rule(r.In, r.Out)
				if !match {
					continue
				}
				matched = true

				targetURL = selector.service
				opts = selector.opts

				break
			}

			if !matched {
				return
			}

			path, _ := url.JoinPath(targetURL.Path, r.Out.URL.Path)

			r.SetXForwarded()
			r.SetURL(targetURL)

			r.Out.URL.Path = path

			for _, opt := range opts {
				opt(r.Out)
			}
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			http.Error(w, "Not found", http.StatusNotFound)
		},
	}
}
