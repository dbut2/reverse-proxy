package rp_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/dbut2/reverse-proxy"
)

// TestBaseRule tests the Always() rule, which should always match regardless of the HTTP request.
func TestBaseRule(t *testing.T) {
	rule := rp.Always()

	req, _ := http.NewRequest("GET", "http://localhost:8000/test", nil)

	t.Run("Always match", func(t *testing.T) {
		match := rule.Matcher(req)
		assert.True(t, match)
	})
}

// TestPathRule tests the PathIsAt() rule by verifying if an HTTP request matches based on the request path.
func TestPathRule(t *testing.T) {
	path := "/test"
	rule := rp.PathIsAt(path)

	reqMatch, _ := http.NewRequest("GET", "http://localhost:8000/test/hello", nil)
	reqNoMatch, _ := http.NewRequest("GET", "http://localhost:8000/other/hello", nil)

	t.Run("PathIsAt match", func(t *testing.T) {
		match := rule.Matcher(reqMatch)
		assert.True(t, match)
	})

	t.Run("PathIsAt no match", func(t *testing.T) {
		match := rule.Matcher(reqNoMatch)
		assert.False(t, match)
	})
}

// TestIPRule tests the IPMatches() rule by verifying if an HTTP request matches based on the client IP (X-Forwarded-For header).
func TestIPRule(t *testing.T) {
	clientIP := "192.168.1.2"
	rule := rp.IPMatches(clientIP)

	reqMatch, _ := http.NewRequest("GET", "http://localhost:8000/test", nil)
	reqMatch.Header.Set("X-Forwarded-For", clientIP)

	reqNoMatch, _ := http.NewRequest("GET", "http://localhost:8000/test", nil)
	reqNoMatch.Header.Set("X-Forwarded-For", "192.168.1.3")

	t.Run("IPMatches match", func(t *testing.T) {
		match := rule.Matcher(reqMatch)
		assert.True(t, match)
	})

	t.Run("IPMatches no match", func(t *testing.T) {
		match := rule.Matcher(reqNoMatch)
		assert.False(t, match)
	})
}

// TestHeaderRule tests the HasHeader() rule by verifying if an HTTP request matches based on the presence of a specific request header.
func TestHeaderRule(t *testing.T) {
	header := "X-My-Header"
	rule := rp.HasHeader(header)

	reqMatch, _ := http.NewRequest("GET", "http://localhost:8000/test", nil)
	reqMatch.Header.Set(header, "test-value")

	reqNoMatch, _ := http.NewRequest("GET", "http://localhost:8000/test", nil)

	t.Run("HasHeader match", func(t *testing.T) {
		match := rule.Matcher(reqMatch)
		assert.True(t, match)
	})

	t.Run("HasHeader no match", func(t *testing.T) {
		match := rule.Matcher(reqNoMatch)
		assert.False(t, match)
	})
}

// TestHeaderMatchesRule tests the HeaderContains() rule by checking if an HTTP request matches based on a specific header-value pair.
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
		match := rule.Matcher(reqMatch)
		assert.True(t, match)
	})

	t.Run("HeaderContains no match value", func(t *testing.T) {
		match := rule.Matcher(reqNoMatchValue)
		assert.False(t, match)
	})

	t.Run("HeaderContains no match header", func(t *testing.T) {
		match := rule.Matcher(reqNoMatchHeader)
		assert.False(t, match)
	})
}
