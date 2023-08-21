package rp_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/dbut2/reverse-proxy"
)

func TestBaseRule(t *testing.T) {
	n := time.Now()
	n.AddDate(0, 0, 1)

	rule := rp.BaseRule()

	req, _ := http.NewRequest("GET", "http://localhost:8000/test", nil)

	t.Run("BaseRule match", func(t *testing.T) {
		match := rule(req, nil)
		assert.True(t, match)
	})
}

func TestPathRule(t *testing.T) {
	path := "/test"
	rule := rp.PathRule(path)

	reqMatch, _ := http.NewRequest("GET", "http://localhost:8000/test/hello", nil)
	reqNoMatch, _ := http.NewRequest("GET", "http://localhost:8000/other/hello", nil)

	t.Run("PathRule match", func(t *testing.T) {
		match := rule(reqMatch, reqMatch)
		assert.True(t, match)
	})

	t.Run("PathRule no match", func(t *testing.T) {
		match := rule(reqNoMatch, reqNoMatch)
		assert.False(t, match)
	})
}

func TestIPRule(t *testing.T) {
	clientIP := "192.168.1.2"
	rule := rp.IPRule(clientIP)

	reqMatch, _ := http.NewRequest("GET", "http://localhost:8000/test", nil)
	reqMatch.Header.Set("X-Forwarded-For", clientIP)

	reqNoMatch, _ := http.NewRequest("GET", "http://localhost:8000/test", nil)
	reqNoMatch.Header.Set("X-Forwarded-For", "192.168.1.3")

	t.Run("IPRule match", func(t *testing.T) {
		match := rule(reqMatch, reqMatch)
		assert.True(t, match)
	})

	t.Run("IPRule no match", func(t *testing.T) {
		match := rule(reqNoMatch, reqNoMatch)
		assert.False(t, match)
	})
}

func TestHeaderRule(t *testing.T) {
	header := "X-My-Header"
	rule := rp.HeaderRule(header)

	reqMatch, _ := http.NewRequest("GET", "http://localhost:8000/test", nil)
	reqMatch.Header.Set(header, "test-value")

	reqNoMatch, _ := http.NewRequest("GET", "http://localhost:8000/test", nil)

	t.Run("HeaderRule match", func(t *testing.T) {
		match := rule(reqMatch, reqMatch)
		assert.True(t, match)
	})

	t.Run("HeaderRule no match", func(t *testing.T) {
		match := rule(reqNoMatch, reqNoMatch)
		assert.False(t, match)
	})
}

func TestHeaderMatchesRule(t *testing.T) {
	header := "X-My-Header"
	value := "test-value"
	rule := rp.HeaderMatchesRule(header, value)

	reqMatch, _ := http.NewRequest("GET", "http://localhost:8000/test", nil)
	reqMatch.Header.Set(header, value)

	reqNoMatchValue, _ := http.NewRequest("GET", "http://localhost:8000/test", nil)
	reqNoMatchValue.Header.Set(header, "wrong-value")

	reqNoMatchHeader, _ := http.NewRequest("GET", "http://localhost:8000/test", nil)

	t.Run("HeaderMatchesRule match", func(t *testing.T) {
		match := rule(reqMatch, reqMatch)
		assert.True(t, match)
	})

	t.Run("HeaderMatchesRule no match value", func(t *testing.T) {
		match := rule(reqNoMatchValue, reqNoMatchValue)
		assert.False(t, match)
	})

	t.Run("HeaderMatchesRule no match header", func(t *testing.T) {
		match := rule(reqNoMatchHeader, reqNoMatchHeader)
		assert.False(t, match)
	})
}
