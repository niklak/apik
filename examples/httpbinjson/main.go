package main

import (
	"context"

	"github.com/niklak/apik"
	"github.com/niklak/apik/reqopt"
	"github.com/niklak/apik/request"
)

func main() {
	type httpBinResponse struct {
		JSON map[string]interface{} `json:"json"`
	}

	client := apik.New(apik.WithBaseUrl("https://httpbin.org"))

	req := request.NewRequest(
		context.Background(),
		"/post",
		reqopt.Method("POST"),
		reqopt.SetJSON(map[string]interface{}{"k": "v"}),
	)

	result := new(httpBinResponse)
	resp, err := client.JSON(req, result)

	if err != nil {
		panic(err)
	}

	if resp.Raw.StatusCode != 200 {
		panic("unexpected status code")

	}

	expectedJSON := map[string]interface{}{"k": "v"}

	if result.JSON["k"] != expectedJSON["k"] {
		panic("unexpected response")
	}
}
