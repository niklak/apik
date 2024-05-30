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
	"github.com/stretchr/testify/suite"

	"github.com/niklak/apik/internal/jar"
	"github.com/niklak/apik/internal/proxy"
	"github.com/niklak/apik/reqopt"
	"github.com/niklak/apik/request"
	"github.com/niklak/httpbulb"
)

type ClientSuite struct {
	suite.Suite
	testServer *httptest.Server
	client     *Client
}

func (s *ClientSuite) SetupSuite() {

	handleFunc := httpbulb.NewRouter()
	s.testServer = httptest.NewServer(handleFunc)

	s.client = New(
		WithBaseUrl(s.testServer.URL),
		WithTimeout(5*time.Second),
	)
}

func (s *ClientSuite) TearDownSuite() {
	s.testServer.Close()
}

func (s *ClientSuite) TestFetchDiscard() {
	resp, err := s.client.Fetch(
		request.NewRequest(context.Background(), "/get"),
		io.Discard,
	)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), 200, resp.Raw.StatusCode)
}

func (s *ClientSuite) TestFetchString() {

	var result string
	resp, err := s.client.Fetch(
		request.NewRequest(context.Background(), "/get"),
		&result,
	)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), 200, resp.Raw.StatusCode)
	apiURL, _ := url.JoinPath(s.testServer.URL, "/get")

	assert.Contains(s.T(), result, apiURL)

}

func (s *ClientSuite) TestFetchBytes() {

	var result []byte
	resp, err := s.client.Fetch(
		request.NewRequest(context.Background(), "/get"),
		&result,
	)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), 200, resp.Raw.StatusCode)
	apiURL, _ := url.JoinPath(s.testServer.URL, "/get")
	assert.Contains(s.T(), string(result), apiURL)

}

func (s *ClientSuite) TestJSONResponse() {

	type httpBinResponse struct {
		URL string `json:"url"`
	}

	req := request.NewRequest(
		context.Background(),
		"/post",
		reqopt.Method(http.MethodPost),
	)
	result := new(httpBinResponse)
	resp, err := s.client.JSON(req, result)

	assert.NoError(s.T(), err)
	assert.Equal(s.T(), 200, resp.Raw.StatusCode)
	apiURL, _ := url.JoinPath(s.testServer.URL, "/post")
	assert.Contains(s.T(), result.URL, apiURL)

}

func (s *ClientSuite) TestJSONUnableDecode() {

	req := request.NewRequest(
		context.Background(),
		"/post",
		reqopt.Method(http.MethodPost),
	)

	result := ""
	_, err := s.client.JSON(req, &result)
	// json: cannot unmarshal object into Go value of type string
	assert.Error(s.T(), err)
}

func (s *ClientSuite) TestAddParam() {

	type httpBinResponse struct {
		URL  string              `json:"url"`
		Args map[string][]string `json:"args"`
	}

	req := request.NewRequest(
		context.Background(),
		"/get",
		reqopt.AddParam("k", "v1"),
		reqopt.AddParam("k", "v2"),
	)

	result := new(httpBinResponse)
	resp, err := s.client.JSON(req, result)
	assert.NoError(s.T(), err)

	assert.Equal(s.T(), 200, resp.Raw.StatusCode)
	apiURL, _ := url.JoinPath(s.testServer.URL, "/get")
	assert.Contains(s.T(), result.URL, apiURL)

	expectedArgs := map[string][]string{
		"k": {"v1", "v2"},
	}
	assert.Equal(s.T(), expectedArgs, result.Args)

}

func (s *ClientSuite) TestSetParam() {

	type httpBinResponse struct {
		URL  string              `json:"url"`
		Args map[string][]string `json:"args"`
	}

	req := request.NewRequest(
		context.Background(),
		"/get",
		reqopt.SetParam("k", "v1"),
		reqopt.SetParam("k", "v2"),
	)

	result := new(httpBinResponse)
	resp, err := s.client.JSON(req, result)
	assert.NoError(s.T(), err)

	assert.Equal(s.T(), 200, resp.Raw.StatusCode)
	assert.Equal(s.T(), result.URL, s.testServer.URL+"/get?k=v2")

	expectedArgs := map[string][]string{
		"k": {"v2"},
	}
	assert.Equal(s.T(), expectedArgs, result.Args)

}

func (s *ClientSuite) TestSetParams() {

	type httpBinResponse struct {
		URL  string              `json:"url"`
		Args map[string][]string `json:"args"`
	}

	req := request.NewRequest(
		context.Background(),
		"/get",
		reqopt.SetParams(url.Values{"k": {"v1", "v2"}}),
	)

	result := new(httpBinResponse)
	resp, err := s.client.JSON(req, result)
	assert.NoError(s.T(), err)

	assert.Equal(s.T(), 200, resp.Raw.StatusCode)
	assert.Equal(s.T(), result.URL, s.testServer.URL+"/get?k=v1&k=v2")

	expectedArgs := map[string][]string{
		"k": {"v1", "v2"},
	}
	assert.Equal(s.T(), expectedArgs, result.Args)

}

func (s *ClientSuite) TestFiles() {

	type httpBinResponse struct {
		URL   string              `json:"url"`
		Files map[string][]string `json:"files"`
		Form  map[string][]string `json:"form"`
	}

	req := request.NewRequest(
		context.Background(),
		"/post",
		reqopt.Method("POST"),
		reqopt.SetFileBody("file_0", "file_0.txt", []byte("test content")),
		reqopt.SetFileBody("file_1", "file_1.txt", "test content"),
		reqopt.SetFileBody("file_2", "file_2.txt", strings.NewReader("test content")),
		reqopt.SetFile("file_3", "test/data/test.txt"),
		reqopt.AddFormField("k", "v"),
	)

	result := new(httpBinResponse)
	resp, err := s.client.JSON(req, result)

	assert.NoError(s.T(), err)
	assert.Equal(s.T(), 200, resp.Raw.StatusCode)

	expectedFiles := map[string][]string{
		"file_0": {"test content"},
		"file_1": {"test content"},
		"file_2": {"test content"},
		"file_3": {"test content"},
	}
	assert.Equal(s.T(), expectedFiles, result.Files)

	expectedForm := map[string][]string{"k": {"v"}}
	assert.Equal(s.T(), expectedForm, result.Form)

}

func (s *ClientSuite) TestFileError() {

	type httpBinResponse struct {
		URL   string              `json:"url"`
		Files map[string][]string `json:"files"`
	}

	req := request.NewRequest(
		context.Background(),
		"/post",
		reqopt.Method("POST"),
		reqopt.SetFile("file_3", "test/data/non-existing.txt"),
	)

	result := new(httpBinResponse)
	_, err := s.client.JSON(req, result)

	assert.Error(s.T(), err)

}

func (s *ClientSuite) TestFilesUnsupportedBodyType() {

	type httpBinResponse struct {
		URL string `json:"url"`
	}

	req := request.NewRequest(
		context.Background(),
		"/post",
		reqopt.Method("POST"),
		reqopt.SetFileBody("file_0", "file_0.txt", 1),
	)

	result := new(httpBinResponse)
	_, err := s.client.JSON(req, result)

	assert.ErrorIs(s.T(), err, request.ErrUnsupportedBodyType)

}

func (s *ClientSuite) TestAddFormField() {

	type httpBinResponse struct {
		URL  string              `json:"url"`
		Form map[string][]string `json:"form"`
	}

	req := request.NewRequest(
		context.Background(),
		"/post",
		reqopt.Method("POST"),
		reqopt.AddFormField("k", "v1"),
		reqopt.AddFormField("k", "v2"),
	)

	result := new(httpBinResponse)
	resp, err := s.client.JSON(req, result)

	assert.NoError(s.T(), err)
	assert.Equal(s.T(), 200, resp.Raw.StatusCode)

	expectedForm := map[string][]string{"k": {"v1", "v2"}}
	assert.Equal(s.T(), expectedForm, result.Form)

}

func (s *ClientSuite) TestSetFormField() {

	type httpBinResponse struct {
		URL  string              `json:"url"`
		Form map[string][]string `json:"form"`
	}

	req := request.NewRequest(
		context.Background(),
		"/post",
		reqopt.Method("POST"),
		reqopt.SetFormField("k", "v1"),
		// SetFormField will overwrite the previous value
		reqopt.SetFormField("k", "v2"),
	)

	result := new(httpBinResponse)
	resp, err := s.client.JSON(req, result)

	assert.NoError(s.T(), err)
	assert.Equal(s.T(), 200, resp.Raw.StatusCode)

	expectedForm := map[string][]string{"k": {"v2"}}
	assert.Equal(s.T(), expectedForm, result.Form)

}

func (s *ClientSuite) TestSetForm() {

	type httpBinResponse struct {
		Form map[string][]string `json:"form"`
	}

	req := request.NewRequest(
		context.Background(),
		"/post",
		reqopt.Method("POST"),
		reqopt.SetForm(url.Values{"k": {"v1", "v2"}}),
	)

	result := new(httpBinResponse)
	resp, err := s.client.JSON(req, result)

	assert.NoError(s.T(), err)
	assert.Equal(s.T(), 200, resp.Raw.StatusCode)

	expectedForm := map[string][]string{"k": {"v1", "v2"}}
	assert.Equal(s.T(), expectedForm, result.Form)

}

func (s *ClientSuite) TestBody() {

	type httpBinResponse struct {
		Data string `json:"data"`
	}

	req := request.NewRequest(
		context.Background(),
		"/post",
		reqopt.Method("POST"),
		reqopt.SetBody([]byte("test")),
	)

	result := new(httpBinResponse)
	resp, err := s.client.JSON(req, result)

	assert.NoError(s.T(), err)
	assert.Equal(s.T(), 200, resp.Raw.StatusCode)

	assert.Equal(s.T(), "test", result.Data)

}

func (s *ClientSuite) TestSetHeaders() {

	type httpBinResponse struct {
		Headers map[string][]string `json:"headers"`
	}

	req := request.NewRequest(
		context.Background(),
		"/get",
		reqopt.Headers(http.Header{"X-Test": {"Test"}}),
	)

	result := new(httpBinResponse)
	resp, err := s.client.JSON(req, result)

	assert.NoError(s.T(), err)
	assert.Equal(s.T(), 200, resp.Raw.StatusCode)

	assert.Equal(s.T(), "Test", result.Headers["X-Test"][0])
}

func (s *ClientSuite) TestNoContext() {

	req := request.NewRequest(nil, "/get")

	_, err := s.client.Fetch(req, nil)

	assert.Error(s.T(), err)

}

func (s *ClientSuite) TestClientCookie() {

	client := New(
		WithBaseUrl(s.testServer.URL),
		WithCookies([]*http.Cookie{
			{Name: "k", Value: "v", Path: "/"},
		}),
	)

	req := request.NewRequest(
		context.Background(),
		"/cookies",
	)

	type httpBinResponse struct {
		Cookies map[string][]string `json:"cookies"`
	}

	result := new(httpBinResponse)
	resp, err := client.JSON(req, result)

	assert.NoError(s.T(), err)
	assert.Equal(s.T(), 200, resp.Raw.StatusCode)

	expectedCookies := map[string][]string{"k": {"v"}}
	assert.Equal(s.T(), expectedCookies, result.Cookies)

}

func (s *ClientSuite) TestRequestAddCookie() {

	req := request.NewRequest(
		context.Background(),
		"/cookies",
		reqopt.AddCookie(&http.Cookie{Name: "k", Value: "v", Path: "/cookies"}),
	)

	type httpBinResponse struct {
		Cookies map[string][]string `json:"cookies"`
	}

	result := new(httpBinResponse)
	_, err := s.client.JSON(req, result)

	assert.NoError(s.T(), err)

	expectedCookies := map[string][]string{"k": {"v"}}
	assert.Equal(s.T(), expectedCookies, result.Cookies)

}

func (s *ClientSuite) TestRequestSetCookie() {

	req := request.NewRequest(
		context.Background(),
		"/cookies",
		reqopt.SetCookies(
			[]*http.Cookie{
				{Name: "k1", Value: "v1", Path: "/cookies"},
				{Name: "k2", Value: "v2", Path: "/cookies"},
			},
		),
	)

	type httpBinResponse struct {
		Cookies map[string][]string `json:"cookies"`
	}

	result := new(httpBinResponse)
	_, err := s.client.JSON(req, result)

	assert.NoError(s.T(), err)

	expectedCookies := map[string][]string{"k1": {"v1"}, "k2": {"v2"}}
	assert.Equal(s.T(), expectedCookies, result.Cookies)
}

func (s *ClientSuite) TestCookieIntersection() {

	// This test demonstrates that request can contain cookies with the same name.
	// And this is a problem.
	// At first, the request cookies will be added to the request  and then jar cookies will be added.
	// This happens because `request.AddCookie()` adds a cookie directly into the request header
	// while jar cookies (`client.Jar.Cookies()`) will be added just before sending the request.

	type httpBinResponse struct {
		Cookies map[string][]string `json:"cookies"`
	}

	client := New(WithBaseUrl(s.testServer.URL), WithCookies([]*http.Cookie{
		{Name: "k", Value: "client-cookie", Path: "/"},
	}))

	req := request.NewRequest(
		context.Background(),
		"/cookies",
		reqopt.AddCookie(&http.Cookie{Name: "k", Value: "request-cookie", Path: "/cookies"}),
	)

	result := new(httpBinResponse)
	_, err := client.JSON(req, &result)

	assert.NoError(s.T(), err)

	expectedCookies := map[string][]string{"k": {"request-cookie", "client-cookie"}}

	assert.Equal(s.T(), expectedCookies, result.Cookies)

	// the next request sent without additional cookies

	req = request.NewRequest(
		context.Background(),
		"/cookies",
	)

	resultNext := new(httpBinResponse)
	_, err = client.JSON(req, &resultNext)
	assert.NoError(s.T(), err)

	expectedCookies = map[string][]string{"k": {"request-cookie", "client-cookie"}}

	assert.Equal(s.T(), expectedCookies, result.Cookies)
}

func (s *ClientSuite) TestManualCookieJar() {

	//Must behave like a standard cookie jar and provide logging

	// Setting your own cookie jar manually
	cj := jar.New()
	client := New(
		WithBaseUrl("https://httpbin.org"),
		WithCookieJar(cj),
	)

	req := request.NewRequest(
		context.Background(),
		"/cookies",
		reqopt.SetCookies(
			[]*http.Cookie{
				{Name: "k1", Value: "v1", Path: "/cookies"},
				{Name: "k2", Value: "v2", Path: "/cookies"},
			},
		),
	)

	type httpBinResponse struct {
		Cookies map[string]string `json:"cookies"`
	}

	result := new(httpBinResponse)
	_, err := client.JSON(req, result)

	assert.NoError(s.T(), err)

	expectedCookies := map[string]string{"k1": "v1", "k2": "v2"}
	assert.Equal(s.T(), expectedCookies, result.Cookies)

}

func (s *ClientSuite) TestSendJSON() {

	type httpBinResponse struct {
		JSON map[string]interface{} `json:"json"`
	}

	client := New(WithBaseUrl("https://httpbin.org"))

	req := request.NewRequest(
		context.Background(),
		"/post",
		reqopt.Method("POST"),
		reqopt.SetJSON(map[string]interface{}{"k": "v"}),
	)

	result := new(httpBinResponse)
	resp, err := client.JSON(req, result)

	assert.NoError(s.T(), err)
	assert.Equal(s.T(), 200, resp.Raw.StatusCode)

	expectedJSON := map[string]interface{}{"k": "v"}
	assert.Equal(s.T(), expectedJSON, result.JSON)

}

func TestClientSuite(t *testing.T) {
	suite.Run(t, new(ClientSuite))
}

func TestClient_CookieIntersection(t *testing.T) {

	// This test demonstrates that request can contain cookies with the same name.
	// And this is a problem.
	// At first, the request cookies will be added to the request  and then jar cookies will be added.
	// This happens because `request.AddCookie()` adds a cookie directly into the request header
	// while jar cookies (`client.Jar.Cookies()`) will be added just before sending the request.

	// This is another example of the same problem but with a different approach.
	// It demonstrates the cookie list order.

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
		reqopt.AddCookie(&http.Cookie{Name: "k", Value: "request-cookie", Path: "/cookies"}),
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

func TestClient_TraceProxy(t *testing.T) {

	//zerolog.SetGlobalLevel(zerolog.InfoLevel)

	// Start a test server that will act as a proxy
	testServer := httptest.NewServer(http.HandlerFunc(proxy.HttpProxyConnectHandler))

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
		WithHeader("User-Agent", "apik/0.1"),
	)

	// This request will contain the trace information
	req := request.NewRequest(
		context.Background(),
		"/get",
		reqopt.AddParam("k", "v"),
		reqopt.Header("User-Agent", "apik/0.1"),
	)

	type httpBinResponse struct {
		URL     string            `json:"url"`
		Headers map[string]string `json:"headers"`
	}

	// in this case we require either a result or a response
	result := new(httpBinResponse)
	_, err = client.JSON(req, result)

	assert.NoError(t, err)

	// ensure that result has the expected value
	assert.Equal(t, result.URL, "https://httpbin.org/get?k=v")

	subsetHeaders := map[string]string{"User-Agent": "apik/0.1"}
	assert.Subset(t, result.Headers, subsetHeaders)

	traceInfo := req.TraceInfo()
	assert.NotNil(t, traceInfo)

	address := strings.TrimPrefix(testServer.URL, "http://")

	// compare connect done address with the proxy server address
	assert.Equal(t, address, traceInfo.ConnectDone[0].Address)

}
