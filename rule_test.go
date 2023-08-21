package rp_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/dbut2/reverse-proxy"
)

func TestBaseRule(t *testing.T) {
	rule := rp.Always()

	req, _ := http.NewRequest("GET", "http://localhost:8000/test", nil)

	t.Run("Always match", func(t *testing.T) {
		match := rule(req, nil)
		assert.True(t, match)
	})
}

func TestPathRule(t *testing.T) {
	path := "/test"
	rule := rp.PathIsAt(path)

	reqMatch, _ := http.NewRequest("GET", "http://localhost:8000/test/hello", nil)
	reqNoMatch, _ := http.NewRequest("GET", "http://localhost:8000/other/hello", nil)

	t.Run("PathIsAt match", func(t *testing.T) {
		match := rule(reqMatch, reqMatch)
		assert.True(t, match)
	})

	t.Run("PathIsAt no match", func(t *testing.T) {
		match := rule(reqNoMatch, reqNoMatch)
		assert.False(t, match)
	})
}

func TestIPRule(t *testing.T) {
	clientIP := "192.168.1.2"
	rule := rp.IPMatches(clientIP)

	reqMatch, _ := http.NewRequest("GET", "http://localhost:8000/test", nil)
	reqMatch.Header.Set("X-Forwarded-For", clientIP)

	reqNoMatch, _ := http.NewRequest("GET", "http://localhost:8000/test", nil)
	reqNoMatch.Header.Set("X-Forwarded-For", "192.168.1.3")

	t.Run("IPMatches match", func(t *testing.T) {
		match := rule(reqMatch, reqMatch)
		assert.True(t, match)
	})

	t.Run("IPMatches no match", func(t *testing.T) {
		match := rule(reqNoMatch, reqNoMatch)
		assert.False(t, match)
	})
}

func TestHeaderRule(t *testing.T) {
	header := "X-My-Header"
	rule := rp.HasHeader(header)

	reqMatch, _ := http.NewRequest("GET", "http://localhost:8000/test", nil)
	reqMatch.Header.Set(header, "test-value")

	reqNoMatch, _ := http.NewRequest("GET", "http://localhost:8000/test", nil)

	t.Run("HasHeader match", func(t *testing.T) {
		match := rule(reqMatch, reqMatch)
		assert.True(t, match)
	})

	t.Run("HasHeader no match", func(t *testing.T) {
		match := rule(reqNoMatch, reqNoMatch)
		assert.False(t, match)
	})
}

func TestHeaderMatchesRule(t *testing.T) {
	header := "X-My-Header"
	value := "test-value"
	rule := rp.HeaderContains(header, value)

	reqMatch, _ := http.NewRequest("GET", "http://localhost:8000/test", nil)
	reqMatch.Header.Set(header, value)

	reqNoMatchValue, _ := http.NewRequest("GET", "http://localhost:8000/test", nil)
	reqNoMatchValue.Header.Set(header, "wrong-value")

	reqNoMatchHeader, _ := http.NewRequest("GET", "http://localhost:8000/test", nil)

	t.Run("HeaderContains match", func(t *testing.T) {
		match := rule(reqMatch, reqMatch)
		assert.True(t, match)
	})

	t.Run("HeaderContains no match value", func(t *testing.T) {
		match := rule(reqNoMatchValue, reqNoMatchValue)
		assert.False(t, match)
	})

	t.Run("HeaderContains no match header", func(t *testing.T) {
		match := rule(reqNoMatchHeader, reqNoMatchHeader)
		assert.False(t, match)
	})
}

func TestB(t *testing.T) {
	proxy := rp.New(
		rp.Select("https://myapi.com", rp.PathIsAt("/api")),
		rp.Select("https://exmaple.com", rp.Always()),
	)

	http.ListenAndServe(":8080", proxy)
}
