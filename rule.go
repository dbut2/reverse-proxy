package rp

import (
	"net/http"
	"net/url"
	"slices"
	"strings"
)

// Rule represents the structure for request matching and modification.
// It consists of a Matcher, which determines if a rule is applicable for a request,
// and a Modifier, which optionally alters the outgoing request when the rule criteria is met.
type Rule struct {
	// Matcher determines if the rule is applicable to a given request. It must be defined for every rule.
	Matcher Matcher

	// Modifier is an optional function to alter the outgoing request when the Matcher's criteria is satisfied.
	Modifier Modifier
}

// Always creates a rule that always matches any request.
func Always() Rule {
	return Rule{
		Matcher: func(r *http.Request) bool {
			return true
		},
	}
}

// HostMatches creates a rule that matches a request based on its host.
func HostMatches(host string) Rule {
	u, _ := url.Parse(host)
	return Rule{
		Matcher: func(r *http.Request) bool {
			return r.URL.Host == u.Host
		},
	}
}

// PathIsAt creates a rule that matches if the request path starts with the specified path.
// When the rule is satisfied, it also trims the matched path from the outgoing request.
func PathIsAt(path string) Rule {
	return Rule{
		Matcher: func(r *http.Request) bool {
			return strings.HasPrefix(r.URL.Path, path)
		},
		Modifier: func(r *http.Request) {
			r.URL.Path = strings.TrimPrefix(r.URL.Path, path)
		},
	}
}

// IPMatches creates a rule that matches a request based on the client IP derived from the X-Forwarded-For header.
func IPMatches(clientIP string) Rule {
	return Rule{
		Matcher: func(r *http.Request) bool {
			xff := r.Header.Get("X-Forwarded-For")
			ips := strings.Split(xff, ",")
			return slices.Contains(ips, clientIP)
		},
	}
}

// MethodMatches creates a rule that matches a request based on its HTTP method.
func MethodMatches(method string) Rule {
	return Rule{
		Matcher: func(r *http.Request) bool {
			return r.Method == method
		},
	}
}

// HasHeader creates a rule that matches if the specified header key is present in the request.
func HasHeader(header string) Rule {
	return Rule{
		Matcher: func(r *http.Request) bool {
			_, ok := r.Header[header]
			return ok
		},
	}
}

// HeaderContains creates a rule that matches if the specified header key-value pair exists in the request.
func HeaderContains(header string, value string) Rule {
	return Rule{
		Matcher: func(r *http.Request) bool {
			headerValues, ok := r.Header[header]
			if !ok {
				return false
			}
			return slices.Contains(headerValues, value)
		},
	}
}

// HasQueryParam creates a rule that matches if the specified query parameter key exists in the request.
func HasQueryParam(param string) Rule {
	return Rule{
		Matcher: func(r *http.Request) bool {
			values := r.URL.Query()[param]
			return len(values) > 0
		},
	}
}

// QueryParamContains creates a rule that matches if the specified query parameter key-value pair exists in the request.
func QueryParamContains(param string, value string) Rule {
	return Rule{
		Matcher: func(r *http.Request) bool {
			values := r.URL.Query()[param]
			return slices.Contains(values, value)
		},
	}
}

// HostPathIsAt creates a rule that matches a request based on both its host and path.
func HostPathIsAt(hostpath string) Rule {
	u, _ := url.Parse(hostpath)
	return AllOf(HostMatches(u.Host), PathIsAt(u.Path))
}

// AllOf creates a composite rule that matches only if all of the provided rules are satisfied.
// Modifiers of individual rules are applied in the order they are provided.
func AllOf(rules ...Rule) Rule {
	return Rule{
		Matcher: func(r *http.Request) bool {
			for _, rule := range rules {
				if !rule.Matcher(r) {
					return false
				}
			}
			return true
		},
		Modifier: func(r *http.Request) {
			for _, rule := range rules {
				if rule.Modifier != nil {
					rule.Modifier(r)
				}
			}
		},
	}
}

// AnyOf creates a composite rule that matches if any of the provided rules are satisfied.
// The modifier of the first matching rule is used.
func AnyOf(rules ...Rule) Rule {
	var modifier Modifier
	return Rule{
		Matcher: func(r *http.Request) bool {
			for _, rule := range rules {
				if rule.Matcher(r) {
					modifier = rule.Modifier
					return true
				}
			}
			return false
		},
		Modifier: modifier,
	}
}
