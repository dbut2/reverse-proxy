package rp

import (
	"net/http"
	"net/url"
	"strings"
)

type Rule func(in *http.Request, out *http.Request) bool

func BaseRule() Rule {
	return func(in *http.Request, out *http.Request) bool {
		return true
	}
}

func PathRule(path string) Rule {
	return func(in *http.Request, out *http.Request) bool {
		hasPrefix := strings.HasPrefix(in.URL.Path, path)

		if !hasPrefix {
			return false
		}

		out.URL.Path = strings.TrimPrefix(out.URL.Path, path)
		return true
	}
}

func IPRule(clientIP string) Rule {
	return func(in *http.Request, out *http.Request) bool {
		xff := in.Header.Get("X-Forwarded-For")
		ips := strings.Split(xff, ",")
		reqClientIP := ""
		if len(ips) > 0 {
			reqClientIP = strings.TrimSpace(ips[0])
		}
		return reqClientIP == clientIP
	}
}

func HeaderRule(header string) Rule {
	return func(in *http.Request, out *http.Request) bool {
		_, ok := in.Header[header]
		return ok
	}
}

func HeaderMatchesRule(header string, value string) Rule {
	return func(in *http.Request, out *http.Request) bool {
		headerValues, ok := in.Header[header]
		if !ok {
			return false
		}
		for _, headerValue := range headerValues {
			if headerValue == value {
				return true
			}
		}
		return false
	}
}

func QueryParamRule(param string) Rule {
	return func(in *http.Request, out *http.Request) bool {
		values := in.URL.Query()[param]
		if len(values) > 0 {
			return true
		}
		return false
	}
}

func HostRule(host string) Rule {
	u, _ := url.Parse(host)

	return func(in *http.Request, out *http.Request) bool {
		return in.URL.Host == u.Host
	}
}

func HostPathRule(hostnamepath string) Rule {
	u, _ := url.Parse(hostnamepath)
	return AllRule(HostRule(u.Host), PathRule(u.Path))
}

func MethodRule(method string) Rule {
	return func(in *http.Request, out *http.Request) bool {
		return in.Method == method
	}
}

func AllRule(rules ...Rule) Rule {
	return func(in *http.Request, out *http.Request) bool {
		for _, rule := range rules {
			if !rule(in, out) {
				return false
			}
		}

		return true
	}
}

func AnyRule(rules ...Rule) Rule {
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
