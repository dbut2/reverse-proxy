# Reverse Proxy

This package provides a flexible, rule-based reverse proxy implementation capable of handling dynamic routing scenarios.

## Features

- Rule-based request matching
- Customizable request and response modification capabilities
- Simple API for setting up a reverse proxy in your web application

## Getting Started

**Installation**

To use the reverse proxy package in your Go projects, run the following command:

```sh
go get github.com/dbut2/reverse-proxy
```

Then, import the package in your Go code:

```go
import "github.com/dbut2/reverse-proxy"
```

**Defining Rules and Selectors**

Create rules to match requests based on different criteria such as path, method, headers, and more:

```go
alwaysMatchRule := rp.Always() // Matches any request
pathMatchRule := rp.PathIsAt("/api") // Matches requests with the path starting with "/api"
headerMatchRule := rp.HasHeader("X-My-Header") // Matches requests containing a specific header
```

Create selectors to associate rules with target URLs and optional modifications:

```go
apiSelector := rp.Select("https://backend-service/api", pathMatchRule)
headerSelector := rp.Select("https://another-service", headerMatchRule)
```

**Setting Up a Reverse Proxy**

Create an instance of a reverse proxy with previously defined selectors:

```go
proxy := rp.New(apiSelector, headerSelector)

// Configure options like OIDC token authorization for secure services
secureSelector := rp.Select("https://secure-service", alwaysMatchRule, rp.WithOIDC())
```

**Running the Reverse Proxy**

Use the reverse proxy as an HTTP handler in your application and start listening:

```go
http.ListenAndServe(":8080", proxy)
```

Refer to the example below for a complete setup.

## Example

Here's an example of setting up a reverse proxy that routes requests to different backend services based on request paths:

```go
package main

import (
	"net/http"

	"github.com/dbut2/reverse-proxy"
)

func main() {
	proxy := rp.New(
		rp.Select("https://service-1.example.com", rp.PathIsAt("/service1")),
		rp.Select("https://service-2.example.com", rp.PathIsAt("/service2")),
		rp.Select("https://service-3.example.com", rp.Always()),
	)

	http.ListenAndServe(":8080", proxy)
}
```

In the example above, requests to `/service1` and `/service2` are proxied to their respective backend services, while all other requests go to `service-3.example.com`.

## Documentation

For detailed usage and configuration options, review the provided GoDoc documentation or refer to individual functions and types in the source code.

## Contributing

Contributions are welcome! Please review the `CODE_OF_CONDUCT.md` and `CONTRIBUTING.md` files for guidelines on contributing to this package.

## License

This package is licensed under the MIT License. See the `LICENSE` file for details.