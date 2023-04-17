// Package reverseproxy provides a simple reverse proxy implementation
// with customizable routing rules.
package reverseproxy

import (
	"context"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"google.golang.org/api/idtoken"
)

// New creates a new reverse proxy instance with the specified routing rules.
// It forwards incoming requests to the target service based on the rules.
// The first rule that matches the request will be used to determine the target service.
// The proxy also handles OIDC token generation and adds the "Authorization" header to the request.
func New(rules []Rule) *httputil.ReverseProxy {
	return &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			var targetURL string
			for _, rule := range rules {
				if service, matches := rule(req); matches {
					targetURL = service
					break
				}
			}

			target, _ := url.Parse(targetURL)
			req.URL.Scheme = target.Scheme
			req.URL.Host = target.Host
			req.Header.Set("X-Forwarded-Host", req.Host)
			req.Host = target.Host

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

			req.Header.Add("Authorization", "Bearer "+token.AccessToken)
		},
	}
}
