// Package rp_test contains the integration tests for the reverse-proxy package.
// These tests setup a simulated proxy environment with test backend services
// and validate the behavior of the reverse proxy under different routing scenarios.
package rp_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/dbut2/reverse-proxy"
)

// setupService creates a mock HTTP server that simulates a backend service.
// This is used instead of actual services for unit testing. It accepts a serviceName,
// which is returned in the response to identify which server handled the request.
func setupService(serviceName string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintf(w, "Hello from %s!", serviceName)
	}))
}

// TestReverseProxy performs various tests to ensure the reverse proxy routes incoming requests
// to the correct backend services as defined by the selectors based on request paths.
func TestReverseProxy(t *testing.T) {
	// Setup mock backend services.
	service1 := setupService("Service1")
	defer service1.Close()

	service2 := setupService("Service2")
	defer service2.Close()

	// Configure reverse proxy with rules for selecting the backends based on the request path.
	selections := []*rp.Selector{
		rp.Select(service1.URL, rp.PathIsAt("/service1")), // Route to Service1 if path starts with '/service1'
		rp.Select(service2.URL, rp.PathIsAt("/service2")), // Route to Service2 if path starts with '/service2'
	}

	// Initialize the reverse proxy with the provided selectors.
	proxy := httptest.NewServer(rp.New(selections...))
	defer proxy.Close()

	// Create an HTTP client to simulate external requests to the reverse proxy.
	client := &http.Client{}

	t.Run("Request to Service1", func(t *testing.T) {
		// Expect requests to '/service1' to be routed to Service1.
		resp, err := client.Get(proxy.URL + "/service1")
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Read the response and verify it comes from Service1.
		body, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()

		assert.Equal(t, "Hello from Service1!", string(body))
	})

	t.Run("Request to Service2", func(t *testing.T) {
		// Expect requests to '/service2' to be routed to Service2.
		resp, err := client.Get(proxy.URL + "/service2")
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Read the response and verify it comes from Service2.
		body, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()

		assert.Equal(t, "Hello from Service2!", string(body))
	})

	t.Run("Request to unknown path", func(t *testing.T) {
		// Expect requests to an unknown path to receive a 'Not Found' status code.
		resp, err := client.Get(proxy.URL + "/unknown")
		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}
