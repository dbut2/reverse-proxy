// This example demonstrates how to create and use a reverse proxy
// with customizable routing rules. In this example, we create a
// reverse proxy that routes requests from a specific client IP
// to a private URL, and all other requests to a public URL.
//
// You can run this example by setting the environment variables
// PUBLIC_URL, PRIVATE_URL, PRIVATE_CLIENT_ID, and PORT.
package main

import (
	"log"
	"net/http"
	"os"

	"github.com/dbut2/cloud-run-reverse-proxy"
)

func main() {
	// Load environment variables
	publicURL := os.Getenv("PUBLIC_URL")
	privateURL := os.Getenv("PRIVATE_URL")
	privateClientIP := os.Getenv("PRIVATE_CLIENT_ID")
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Define routing rules
	rules := []reverseproxy.Rule{
		reverseproxy.IPRule(privateClientIP, privateURL),
		reverseproxy.BaseRule(publicURL),
	}

	// Create the reverse proxy with the defined rules
	proxy := reverseproxy.New(rules)

	// Register the reverse proxy as an HTTP handler
	http.HandleFunc("/", proxy.ServeHTTP)

	// Start the HTTP server
	log.Printf("Starting reverse proxy on :%s\n", port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		panic(err)
	}
}
