// Package rp provides a set of utilities to set up and modify behavior of reverse proxies.
package rp

import (
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"

	"google.golang.org/api/idtoken"
)

// New creates a new reverse proxy instance configured with the provided selectors.
// The selectors determine which rules apply to which incoming requests.
func New(selectors ...*Selector) *httputil.ReverseProxy {
	return &httputil.ReverseProxy{
		Rewrite: func(r *httputil.ProxyRequest) {
			selector, matched := findSelector(selectors, r.In)
			if !matched {
				return
			}

			r.SetXForwarded()
			path, _ := url.JoinPath("/", selector.Url.Path, r.Out.URL.Path)
			r.SetURL(selector.Url)
			r.Out.URL.Path = path

			for _, modifier := range selector.Modifiers {
				modifier(r.Out)
			}
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			http.Error(w, "Not found", http.StatusNotFound)
		},
	}
}

// findSelector searches through the provided selectors and finds the first one that matches
// the given request. It returns the matched selector and a boolean indicating if a match was found.
func findSelector(selectors []*Selector, r *http.Request) (*Selector, bool) {
	for _, selector := range selectors {
		if selector.Matcher(r) {
			return selector, true
		}
	}
	return nil, false
}

// Matcher is a function type that defines the criteria for whether a selector should be applied to a request.
type Matcher func(r *http.Request) bool

// Modifier is a function type that describes how to alter an outgoing request before it's sent.
type Modifier func(r *http.Request)

// Selector defines the criteria and actions for selecting and modifying requests in the reverse proxy.
// It contains a Matcher to decide if the Selector applies to an incoming request,
// a destination URL to which the request should be sent, and a list of Modifiers to apply
// to the outgoing request.
type Selector struct {
	Matcher   Matcher
	Url       *url.URL
	Modifiers []Modifier
}

// Select creates a new selector with the given address, rule, and optional selector modifications.
// The "when" rule determines when this selector should be applied to incoming requests.
func Select(address string, when Rule, opts ...SelectOption) *Selector {
	serviceURL, err := url.Parse(address)
	if err != nil {
		panic(err.Error())
	}
	s := &Selector{
		Matcher: when.Matcher,
		Url:     serviceURL,
	}
	if when.Modifier != nil {
		s.Modifiers = append(s.Modifiers, when.Modifier)
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

// SelectOption is a function type that describes how to modify the behavior of a selector.
type SelectOption func(*Selector)

// WithOIDC returns a SelectOption that modifies the selector to set the authorization header
// of the outgoing request using an OIDC token generated for the target service.
func WithOIDC() SelectOption {
	return func(s *Selector) {
		modifier := func(r *http.Request) {
			tokenSource, err := idtoken.NewTokenSource(r.Context(), s.Url.String())
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

		s.Modifiers = append(s.Modifiers, modifier)
	}
}
