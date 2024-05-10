# APIK

[![Go Reference](https://pkg.go.dev/badge/github.com/niklak/apik.svg)](https://pkg.go.dev/github.com/niklak/apik)
[![Go](https://github.com/niklak/apik/actions/workflows/go.yml/badge.svg)](https://github.com/niklak/apik/actions/workflows/go.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/niklak/apik)](https://goreportcard.com/report/github.com/niklak/apik)
[![codecov](https://codecov.io/gh/niklak/apik/graph/badge.svg?token=Y8IB5RQ3OP)](https://codecov.io/gh/niklak/apik)

This package is a wrapper over the standard `net/http` package and is intended to simplify working with http requests.


## Examples

### Post request with form data

```go

package main

import (
	"context"
	"fmt"

	"github.com/niklak/apik"
	"github.com/niklak/apik/reqopt"
)

type httpBinResponse struct {
	URL     string              `json:"url"`
	Form    map[string][]string `json:"form"`
	Headers map[string]string   `json:"headers"`
}

func main() {

	// Creating a client with a base URL, that will be common for all requests
	client := apik.New(apik.WithBaseUrl("https://httpbin.org"))

	// Creating a POST request
	req := apik.NewRequest(
		context.Background(),
		"/post",
		reqopt.Method("POST"),
		reqopt.AddFormField("k", "v1"),
		reqopt.AddFormField("k", "v2"),
		reqopt.Header("User-Agent", "apik/0.1"),
	)

	result := new(httpBinResponse)
	resp, err := client.JSON(req, result)

	if err != nil {
		panic(err)
	}

	fmt.Printf("status code: %d\n", resp.Raw.StatusCode)

	fmt.Printf("response: %#v\n", result)
}

```

### JSON request

[sending json](./examples/httpbinjson/main.go)

### Building API wrapper

[api wrapper](./examples/httpbinapi/api.go)