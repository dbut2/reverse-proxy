package rp

import (
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"

	"google.golang.org/api/idtoken"
)

func New(selectors ...*Selector) *httputil.ReverseProxy {
	return &httputil.ReverseProxy{
		Rewrite: func(r *httputil.ProxyRequest) {
			selector, matched := findSelector(selectors, r.In, r.Out)
			if !matched {
				return
			}

			r.SetXForwarded()
			path, _ := url.JoinPath("/", selector.url.Path, r.Out.URL.Path)
			r.SetURL(selector.url)
			r.Out.URL.Path = path

			for _, modifier := range selector.modifiers {
				modifier(r.Out)
			}
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			http.Error(w, "Not found", http.StatusNotFound)
		},
	}
}

func findSelector(selectors []*Selector, in, out *http.Request) (*Selector, bool) {
	for _, selector := range selectors {
		if selector.rule(in, out) {
			return selector, true
		}
	}
	return nil, false
}

// Selector contains information for selecting and modifying requests
type Selector struct {
	rule      Rule
	url       *url.URL
	modifiers []func(r *http.Request)
}

// Select returns a selector to the address for matching on when rule
func Select(address string, when Rule, opts ...SelectOption) *Selector {
	serviceURL, err := url.Parse(address)
	if err != nil {
		panic(err.Error())
	}
	s := &Selector{rule: when, url: serviceURL}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

type Modifier func(r *http.Request)

func (s *Selector) Modify(m Modifier) {
	s.modifiers = append(s.modifiers, m)
}

// SelectOption modifies the selector
type SelectOption func(*Selector)

// WithOIDC sets the authorization header using an OIDC token generated for the service
func WithOIDC() SelectOption {
	return func(s *Selector) {
		modifier := func(r *http.Request) {
			tokenSource, err := idtoken.NewTokenSource(r.Context(), s.url.String())
			if err != nil {
				slog.Error("failed to create token source", slog.Any("error", err))
				return
			}

			token, err := tokenSource.Token()
			if err != nil {
				slog.Error("failed to obtain an OIDC token", slog.Any("error", err))
				return
			}

			r.Header.Add("Authorization", "Bearer "+token.AccessToken)
		}

		s.modifiers = append(s.modifiers, modifier)
	}
}
