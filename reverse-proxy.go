package main

import (
	"context"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

	"google.golang.org/api/idtoken"
)

func main() {
	// Read the public and private service URLs, and allowed client IP from environment variables
	publicURL := os.Getenv("PUBLIC_URL")
	privateURL := os.Getenv("PRIVATE_URL")
	privateClientIP := os.Getenv("PRIVATE_CLIENT_ID")

	// Read the PORT environment variable, defaulting to 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Create the reverse proxy
	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			// Extract the client IP address from the X-Forwarded-For header
			xff := req.Header.Get("X-Forwarded-For")
			ips := strings.Split(xff, ",")
			clientIP := ""
			if len(ips) > 0 {
				clientIP = strings.TrimSpace(ips[0])
			}

			// Determine the target URL based on the client IP address
			var targetURL string
			switch clientIP {
			case privateClientIP:
				targetURL = privateURL
			default:
				targetURL = publicURL
			}

			// Update the request to point to the target service
			target, _ := url.Parse(targetURL)
			req.URL.Scheme = target.Scheme
			req.URL.Host = target.Host
			req.Header.Set("X-Forwarded-Host", req.Host)
			req.Host = target.Host

			// Create a new token source and obtain an OIDC token
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

			// Add the Authorization header to the request
			req.Header.Add("Authorization", "Bearer "+token.AccessToken)
		},
	}

	// Register the reverse proxy handler
	http.HandleFunc("/", proxy.ServeHTTP)

	// Start the reverse proxy server
	log.Printf("Starting reverse proxy on :%s\n", port)
	err := http.ListenAndServe(":"+port, proxy)
	if err != nil {
		panic(err)
	}
}
