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
// Selectors are used to direct and possibly modify incoming requests according to predetermined rules.
// The resulting reverse proxy handles requests according to these selectors.
//
// Usage example:
//
//     selectors := []*Selector{
//         Select("http://example.com", PathIsAt("/api")),
//         // More selectors here...
//     }
//     proxy := New(selectors...)
//
// params:
// - selectors: A variadic slice of pointers to Selector objects that determine the matching criteria
//               and the modifications to be made to incoming requests.
//
// return:
// - A pointer to an http.ReverseProxy that routes and modifies requests following the defined selectors.
func New(selectors ...*Selector) *httputil.ReverseProxy {
	return &httputil.ReverseProxy{
		Rewrite: func(r *httputil.ProxyRequest) {
			selector, matched := findSelector(selectors, r.In)
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

// findSelector searches through the provided selectors and finds the first one that matches
// the given request. It returns the matched selector and a boolean indicating if a match was found.
//
// params:
// - selectors: A slice of pointers to Selector objects among which to find a match.
// - r: The incoming HTTP request to match against the selectors.
//
// return:
// - *Selector: The first selector that matches the request, otherwise nil.
// - bool: A boolean indicating whether a match was found.
func findSelector(selectors []*Selector, r *http.Request) (*Selector, bool) {
	for _, selector := range selectors {
		if selector.matcher(r) {
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
// to the outgoing request before it is sent to the target service.
type Selector struct {
	matcher   Matcher
	url       *url.URL
	modifiers []Modifier
}

// Select creates a new selector with the given address, rule, and optional selector modifications (options).
// The "when" rule contains the conditions under which this selector should be applied to incoming requests.
//
// Usage example:
//
//     selector := Select("http://backend-service/api", PathIsAt("/api"), WithOIDC())
//
// params:
// - address: The target URL as a string where requests matching the rule will be directed.
// - when: A Rule object that defines the conditions for the match.
// - opts: Variadic optional selector options to modify how a matching request is handled.
//
// return:
//
// A pointer to a Selector that is configured with the target URL, the matching rule,
// and any provided modification functions.
func Select(address string, when Rule, opts ...SelectOption) *Selector {
	serviceURL, err := url.Parse(address)
	if err != nil {
		panic(err.Error())
	}
	s := &Selector{
		matcher: when.Matcher,
		url:     serviceURL,
	}
	if when.Modifier != nil {
		s.modifiers = append(s.modifiers, when.Modifier)
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

// Modify appends a new modifier to the selector's list of modifiers.
// This allows on-the-fly modification of the selector's behavior.
//
// params:
// - m: The Modifier function that alters the outgoing request.
func (s *Selector) Modify(m Modifier) {
	s.modifiers = append(s.modifiers, m)
}

// SelectOption defines a type for functions that customize a selector.
// These functions can add modifications to selectors that are applied to outgoing requests.
type SelectOption func(*Selector)

// WithOIDC constructs a SelectOption that augments a selector to attach
// an OIDC token as the authorization header for the outgoing request,
// intended for the target service.
//
// Usage example:
//
//     selector := Select("https://secure-service", Always(), WithOIDC())
//
// return:
// A SelectOption that configures the Selector to attach an OIDC token to the outgoing request header.
func WithOIDC() SelectOption {
	return func(s *Selector) {
		s.Modify(func(r *http.Request) {
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
		})
	}
}
