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

func setupService(serviceName string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintf(w, "Hello from %s!", serviceName)
	}))
}

func TestReverseProxy(t *testing.T) {
	service1 := setupService("Service1")
	defer service1.Close()

	service2 := setupService("Service2")
	defer service2.Close()

	selections := []*rp.Selector{
		rp.Select(service1.URL, rp.PathIsAt("/service1")),
		rp.Select(service2.URL, rp.PathIsAt("/service2")),
	}

	proxy := httptest.NewServer(rp.New(selections...))
	defer proxy.Close()

	client := &http.Client{}

	t.Run("Request to Service1", func(t *testing.T) {
		resp, err := client.Get(proxy.URL + "/service1")
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()

		assert.Equal(t, "Hello from Service1!", string(body))
	})

	t.Run("Request to Service2", func(t *testing.T) {
		resp, err := client.Get(proxy.URL + "/service2")
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()

		assert.Equal(t, "Hello from Service2!", string(body))
	})

	t.Run("Request to unknown path", func(t *testing.T) {
		resp, err := client.Get(proxy.URL + "/unknown")
		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}
