// Package reverseproxy provides a simple reverse proxy implementation
// with customizable routing rules.
package reverseproxy

import (
	"net/http"
	"strings"
)

// Rule represents a routing rule for the reverse proxy.
// It returns the target service URL and a boolean indicating
// whether the rule matches the request or not.
type Rule func(r *http.Request) (string, bool)

// BaseRule creates a basic routing rule that matches all requests
// and routes them to the specified service.
func BaseRule(service string) Rule {
	return func(r *http.Request) (string, bool) {
		return service, true
	}
}

// PathRule creates a routing rule that matches requests based on their path.
// If the request path starts with the specified path, the rule matches
// and routes the request to the specified service.
func PathRule(path string, service string) Rule {
	return func(r *http.Request) (string, bool) {
		return service, strings.HasPrefix(r.URL.Path, path)
	}
}

// IPRule creates a routing rule that matches requests based on their client IP.
// If the request comes from the specified client IP, the rule matches
// and routes the request to the specified service.
// The client IP is extracted from the "X-Forwarded-For" header.
func IPRule(clientIP string, service string) Rule {
	return func(r *http.Request) (string, bool) {
		xff := r.Header.Get("X-Forwarded-For")
		ips := strings.Split(xff, ",")
		reqClientIP := ""
		if len(ips) > 0 {
			reqClientIP = strings.TrimSpace(ips[0])
		}
		return service, reqClientIP == clientIP
	}
}

// HeaderRule creates a routing rule that matches requests based on the
// presence of a specific header. If the request contains the specified
// header, the rule matches and routes the request to the specified service.
func HeaderRule(header string, service string) Rule {
	return func(r *http.Request) (string, bool) {
		_, ok := r.Header[header]
		return service, ok
	}
}

// HeaderMatchesRule creates a routing rule that matches requests based on
// a specific header and its value. If the request contains the specified
// header and at least one of its values matches the provided value, the rule
// matches and routes the request to the specified service.
func HeaderMatchesRule(header string, value string, service string) Rule {
	return func(r *http.Request) (string, bool) {
		headerValues, ok := r.Header[header]
		if !ok {
			return service, false
		}
		for _, headerValue := range headerValues {
			if headerValue == value {
				return service, true
			}
		}
		return service, false
	}
}
