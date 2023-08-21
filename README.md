# Reverse Proxy

The reverse proxy package provides a wrapper around `httputil.ReverseProxy`, adding the ability to create flexible, rule based routing

## Usage

Import reverse-proxy
```shell
go get github.com/dbut2/reverse-proxy
```

In your Go code, add
```go
import "github.com/dbut2/reverse-proxy"
```

Creating a reverse proxy with selectors
```go
proxy := rp.New(
    rp.Select("https://myapi.com", rp.PathIsAt("/api")),
    rp.Select("https://example.com", rp.Always()),
)
```

Using selector options
```go
proxy := rp.New(
    rp.Select("https://private-cloud-run-instance.com", rp.IPMatches("12.34.56.78"), rp.WithOIDC()),
    rp.Select("https://example.com", rp.Always()),
)
```

Listen and serve
```go
proxy := rp.New(
    rp.Select("https://myapi.com", rp.PathIsAt("/api")),
    rp.Select("https://example.com", rp.Always()),
)

http.ListenAndServe(":8080", proxy)
```

### Example

```go
package main

import (
	"net/http"

	"github.com/dbut2/reverse-proxy"
)

func main() {
	proxy := rp.New(
		rp.Select("https://myapi.com", rp.PathIsAt("/api")),
		rp.Select("https://example.com", rp.Always()),
	)

	http.ListenAndServe(":8080", proxy)
}
```

