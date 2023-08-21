package rp

import (
	"net/http"
	"net/url"
	"slices"
	"strings"
)

// Rule is a boolean function if a match condition is found
// in request is for matching
// out request is for modifying outgoing request, as is in PathIsAt
type Rule func(in *http.Request, out *http.Request) bool

// Always always matches
func Always() Rule {
	return func(in *http.Request, out *http.Request) bool {
		return true
	}
}

// HostMatches matches on the request host
func HostMatches(host string) Rule {
	u, _ := url.Parse(host)
	return func(in *http.Request, out *http.Request) bool {
		return in.URL.Host == u.Host
	}
}

// PathIsAt matches if request path is prepended by given path
// also trims that path from outgoing request
func PathIsAt(path string) Rule {
	return func(in *http.Request, out *http.Request) bool {
		hasPrefix := strings.HasPrefix(in.URL.Path, path)
		if !hasPrefix {
			return false
		}
		out.URL.Path = strings.TrimPrefix(out.URL.Path, path)
		return true
	}
}

// IPMatches matches on the client IP from the X-Forwarded-For header
func IPMatches(clientIP string) Rule {
	return func(in *http.Request, out *http.Request) bool {
		xff := in.Header.Get("X-Forwarded-For")
		ips := strings.Split(xff, ",")
		return slices.Contains(ips, clientIP)
	}
}

// MethodMatches matches on the request method
func MethodMatches(method string) Rule {
	return func(in *http.Request, out *http.Request) bool {
		return in.Method == method
	}
}

// HasHeader matches if header key exists in request
func HasHeader(header string) Rule {
	return func(in *http.Request, out *http.Request) bool {
		_, ok := in.Header[header]
		return ok
	}
}

// HeaderContains matches if header kv pair exists in request
func HeaderContains(header string, value string) Rule {
	return func(in *http.Request, out *http.Request) bool {
		headerValues, ok := in.Header[header]
		if !ok {
			return false
		}
		return slices.Contains(headerValues, value)
	}
}

// HasQueryParam matches if query param key exists in request
func HasQueryParam(param string) Rule {
	return func(in *http.Request, out *http.Request) bool {
		values := in.URL.Query()[param]
		return len(values) > 0
	}
}

// QueryParamContains matches if query param kv pair exists in request
func QueryParamContains(param string, value string) Rule {
	return func(in *http.Request, out *http.Request) bool {
		values := in.URL.Query()[param]
		return slices.Contains(values, value)
	}
}

// HostPathIsAt matches on HostMatches and PathIsAt
func HostPathIsAt(hostpath string) Rule {
	u, _ := url.Parse(hostpath)
	return AllOf(HostMatches(u.Host), PathIsAt(u.Path))
}

// AllOf matches if all rules match
func AllOf(rules ...Rule) Rule {
	return func(in *http.Request, out *http.Request) bool {
		for _, rule := range rules {
			if !rule(in, out) {
				return false
			}
		}
		return true
	}
}

// AnyOf matches if any rules match
func AnyOf(rules ...Rule) Rule {
	return func(in *http.Request, out *http.Request) bool {
		for _, rule := range rules {
			if !rule(in, out) {
				continue
			}
			return true
		}
		return false
	}
}
