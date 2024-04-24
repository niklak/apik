package apik

import (
	"context"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/niklak/apik/request"
)

func TestClient_FetchDiscard(t *testing.T) {
	// We need only the status code, so we discard the response body
	client := New(WithBaseUrl("https://httpbin.org"))

	resp, err := client.Fetch(
		request.NewRequest(context.Background(), "/get"),
		io.Discard,
	)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.Raw.StatusCode)

}

func TestClient_Fetch(t *testing.T) {
	client := New(WithBaseUrl("https://httpbin.org"))

	ctx := context.Background()
	req := request.NewRequest(ctx, "/get")

	var result string
	resp, err := client.Fetch(req, &result)
	assert.NoError(t, err)

	assert.Equal(t, 200, resp.Raw.StatusCode)

	assert.Contains(t, result, "https://httpbin.org/get")
}

func TestClient_JSON(t *testing.T) {

	type httpBinResponse struct {
		URL string `json:"url"`
	}

	client := New(WithBaseUrl("https://httpbin.org"))

	req := request.NewRequest(
		context.Background(),
		"/post",
		request.Method("POST"),
	)

	result := new(httpBinResponse)
	resp, err := client.JSON(req, result)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.Raw.StatusCode)
	assert.Equal(t, "https://httpbin.org/post", result.URL)
}

func TestClient_JSONResp(t *testing.T) {

	type httpBinResponse struct {
		URL string `json:"url"`
	}

	client := New(WithBaseUrl("https://httpbin.org"))

	req := request.NewRequest(
		context.Background(),
		"/post",
		request.Method("POST"),
	)

	result := new(httpBinResponse)
	resp, err := client.JSON(req, result)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.Raw.StatusCode)
	assert.Equal(t, "https://httpbin.org/post", result.URL)
}

func TestClient_QueryParams(t *testing.T) {

	type httpBinResponse struct {
		URL  string              `json:"url"`
		Args map[string][]string `json:"args"`
	}

	client := New(WithBaseUrl("https://httpbin.org"))

	req := request.NewRequest(
		context.Background(),
		"/get",
		request.AddParam("k", "v1"),
		request.AddParam("k", "v2"),
	)

	result := new(httpBinResponse)
	resp, err := client.JSON(req, result)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.Raw.StatusCode)

	expectedArgs := map[string][]string{
		"k": {"v1", "v2"},
	}
	assert.Equal(t, expectedArgs, result.Args)
	assert.Equal(t, "https://httpbin.org/get?k=v1&k=v2", result.URL)
}
