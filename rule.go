package rp

import (
	"net/http"
	"strings"
)

type Rule func(r *http.Request) (bool, func())

func noEdit() {}

func noMatch() (bool, func()) { return false, nil }

func BaseRule() Rule {
	return func(r *http.Request) (bool, func()) {
		return true, noEdit
	}
}

func PathRule(path string) Rule {
	return func(r *http.Request) (bool, func()) {
		hasPrefix := strings.HasPrefix(r.URL.Path, path)

		if !hasPrefix {
			return noMatch()
		}

		return true, func() {
			r.URL.Path = strings.TrimPrefix(r.URL.Path, path)
		}
	}
}

func IPRule(clientIP string) Rule {
	return func(r *http.Request) (bool, func()) {
		xff := r.Header.Get("X-Forwarded-For")
		ips := strings.Split(xff, ",")
		reqClientIP := ""
		if len(ips) > 0 {
			reqClientIP = strings.TrimSpace(ips[0])
		}
		return reqClientIP == clientIP, noEdit
	}
}

func HeaderRule(header string) Rule {
	return func(r *http.Request) (bool, func()) {
		_, ok := r.Header[header]
		return ok, noEdit
	}
}

func HeaderMatchesRule(header string, value string) Rule {
	return func(r *http.Request) (bool, func()) {
		headerValues, ok := r.Header[header]
		if !ok {
			return noMatch()
		}
		for _, headerValue := range headerValues {
			if headerValue == value {
				return true, noEdit
			}
		}
		return noMatch()
	}
}

func QueryParamRule(param string) Rule {
	return func(r *http.Request) (bool, func()) {
		values := r.URL.Query()[param]
		if len(values) > 0 {
			return true, noEdit
		}
		return noMatch()
	}
}

func HostnameRule(hostname string) Rule {
	return func(r *http.Request) (bool, func()) {
		return r.URL.Hostname() == hostname, noEdit
	}
}

func MethodRule(method string) Rule {
	return func(r *http.Request) (bool, func()) {
		return r.Method == method, noEdit
	}
}

func AllRule(rules []Rule) Rule {
	return func(r *http.Request) (bool, func()) {
		modifiers := make([]func(), len(rules))

		for i, rule := range rules {
			match, modifier := rule(r)
			if !match {
				return noMatch()
			}

			modifiers[i] = modifier
		}

		return true, func() {
			for _, modifier := range modifiers {
				modifier()
			}
		}
	}
}

func AnyRule(rules []Rule) Rule {
	return func(r *http.Request) (bool, func()) {
		for _, rule := range rules {
			match, modifier := rule(r)
			if !match {
				continue
			}

			return true, modifier
		}

		return noMatch()
	}
}
