package apik

import (
	"context"
	"io"
	"strings"
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

func TestClient_Files(t *testing.T) {

	type httpBinResponse struct {
		URL   string            `json:"url"`
		Files map[string]string `json:"files"`
		Form  map[string]string `json:"form"`
	}

	client := New(WithBaseUrl("https://httpbin.org"))

	req := request.NewRequest(
		context.Background(),
		"/post",
		request.Method("POST"),
		request.SetFileBody("file_0", "file_0.txt", []byte("test content")),
		request.SetFileBody("file_1", "file_1.txt", "test content"),
		request.SetFileBody("file_2", "file_2.txt", strings.NewReader("test content")),
		request.SetFile("file_3", "test_data/test.txt"),
		request.AddFormField("k", "v"),
	)

	result := new(httpBinResponse)
	resp, err := client.JSON(req, result)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.Raw.StatusCode)

	expectedFiles := map[string]string{
		"file_0": "test content",
		"file_1": "test content",
		"file_2": "test content",
		"file_3": "test content",
	}
	assert.Equal(t, expectedFiles, result.Files)

	expectedForm := map[string]string{"k": "v"}
	assert.Equal(t, expectedForm, result.Form)
}
