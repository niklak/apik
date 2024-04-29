package apik

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/niklak/apik/internal/proxy"
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

func TestClient_JSONResponse(t *testing.T) {

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

func TestClient_Files_UnsupportedBodyType(t *testing.T) {

	type httpBinResponse struct {
		URL string `json:"url"`
	}

	client := New(WithBaseUrl("https://httpbin.org"))

	req := request.NewRequest(
		context.Background(),
		"/post",
		request.Method("POST"),
		request.SetFileBody("file_0", "file_0.txt", 1),
	)

	result := new(httpBinResponse)
	_, err := client.JSON(req, result)
	assert.ErrorIs(t, err, request.ErrUnsupportedBodyType)

}

func TestClient_TraceProxy(t *testing.T) {

	// Start a test server that will act as a proxy
	testServer := httptest.NewServer(http.HandlerFunc(proxy.ProxyConnectHandler))

	defer testServer.Close()

	u, err := url.Parse(testServer.URL)
	assert.NoError(t, err)

	// Create an http client with a proxy

	httpClient := &http.Client{
		Transport: &http.Transport{Proxy: http.ProxyURL(u)},
		Timeout:   10 * time.Second,
	}

	// Using this http client for our apik client

	client := New(
		WithBaseUrl("https://httpbin.org"),
		WithTrace(),
		WithHttpClient(httpClient),
	)

	// This request will contain the trace information
	req := request.NewRequest(
		context.Background(),
		"/get",
		request.AddParam("k", "v"),
	)

	type httpBinResponse struct {
		URL string `json:"url"`
	}

	// in this case we require either a result or a response
	result := new(httpBinResponse)
	_, err = client.JSON(req, result)

	assert.NoError(t, err)

	// ensure that result has the expected value
	assert.Equal(t, result.URL, "https://httpbin.org/get?k=v")

	traceInfo := req.TraceInfo()
	assert.NotNil(t, traceInfo)

	address := strings.TrimPrefix(testServer.URL, "http://")

	// compare connect done address with the proxy server address
	assert.Equal(t, address, traceInfo.ConnectDone[0].Address)

}

func TestClient_ClientCookie(t *testing.T) {

	client := New(
		WithBaseUrl("https://httpbin.org"),
		WithCookies([]*http.Cookie{
			{Name: "k", Value: "v", Path: "/"},
		}),
	)

	req := request.NewRequest(
		context.Background(),
		"/cookies",
	)

	type httpBinResponse struct {
		Cookies map[string]string `json:"cookies"`
	}

	result := new(httpBinResponse)
	resp, err := client.JSON(req, result)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.Raw.StatusCode)

	expectedCookies := map[string]string{"k": "v"}
	assert.Equal(t, expectedCookies, result.Cookies)
}

func TestClient_RequestCookie(t *testing.T) {

	client := New(WithBaseUrl("https://httpbin.org"))

	req := request.NewRequest(
		context.Background(),
		"/cookies",
		request.AddCookie(&http.Cookie{Name: "k", Value: "v", Path: "/cookies"}),
	)

	type httpBinResponse struct {
		Cookies map[string]string `json:"cookies"`
	}

	result := new(httpBinResponse)
	resp, err := client.JSON(req, result)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.Raw.StatusCode)

	expectedCookies := map[string]string{"k": "v"}
	assert.Equal(t, expectedCookies, result.Cookies)
}

func TestClient_CookieIntersection(t *testing.T) {

	// This test demonstrates that request can contain cookies with the same name.
	// And this is a problem.
	// At first request cookies will be added to the request, and then jar cookies will be added.
	// Because request.AddCookie() add a cookie directly into the request header
	// and jar cookies (client.Jar.Cookies()) are added just before request is sent.

	type briefCookie struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	}

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		httpCookies := r.Cookies()
		cookies := make([]briefCookie, 0)

		for _, c := range httpCookies {
			cookies = append(cookies, briefCookie{Name: c.Name, Value: c.Value})
		}

		body, err := json.Marshal(cookies)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(body)
	}))

	defer testServer.Close()

	client := New(WithBaseUrl(testServer.URL), WithCookies([]*http.Cookie{
		{Name: "k", Value: "client-cookie", Path: "/"},
	}))

	req := request.NewRequest(
		context.Background(),
		"/cookies",
		request.AddCookie(&http.Cookie{Name: "k", Value: "request-cookie", Path: "/cookies"}),
	)

	var result []briefCookie
	_, err := client.JSON(req, &result)

	assert.NoError(t, err)

	expectedCookies := []briefCookie{{Name: "k", Value: "request-cookie"}, {Name: "k", Value: "client-cookie"}}
	assert.Equal(t, expectedCookies, result)

	// the next request sent without additional cookies

	req = request.NewRequest(
		context.Background(),
		"/cookies",
	)

	var resultNext []briefCookie
	_, err = client.JSON(req, &resultNext)
	assert.NoError(t, err)

	expectedCookies = []briefCookie{{Name: "k", Value: "request-cookie"}, {Name: "k", Value: "client-cookie"}}

	assert.Equal(t, expectedCookies, result)

}
