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

func findSelector(selectors []*Selector, in, out *http.Request) (*Selector, bool) {
	for _, selector := range selectors {
		if selector.Rule(in, out) {
			return selector, true
		}
	}
	return nil, false
}

// Selector contains information for selecting and modifying requests
type Selector struct {
	Rule      Rule
	Url       *url.URL
	Modifiers []func(r *http.Request)
}

// Select returns a selector to the address for matching on when rule
func Select(address string, when Rule, opts ...SelectOption) *Selector {
	serviceURL, err := url.Parse(address)
	if err != nil {
		panic(err.Error())
	}
	s := &Selector{Rule: when, Url: serviceURL}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

// SelectOption modifies the selector
type SelectOption func(*Selector)

// WithOIDC sets the authorization header using an OIDC token generated for the service
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
