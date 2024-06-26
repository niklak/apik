package apik

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/net/publicsuffix"

	"github.com/niklak/apik/reqopt"
	"github.com/niklak/apik/request"
)

var defaultTimeout = time.Minute

// Request is an alias for request.Request
type Request = request.Request

// NewRequest is an alias for request.NewRequest
var NewRequest = request.NewRequest

// Response represents a wrapper around http.Response with the result of the request
type Response struct {
	Raw        *http.Response
	Result     any
	Request    *Request
	StatusCode int
}

type Client struct {
	c       *http.Client
	timeout time.Duration
	trace   bool
	logger  zerolog.Logger
	cookies []*http.Cookie
	header  http.Header
	baseURL *url.URL
	jar     http.CookieJar
}

// Do sends an http.Request built from Request and returns an http.Response
func (c *Client) Do(req *Request) (resp *http.Response, err error) {

	if c.baseURL != nil {
		req.URL = c.baseURL.ResolveReference(req.URL)
	}

	if c.trace {
		reqopt.Trace()(req)
	}

	for key, values := range c.header {
		if _, ok := req.Header[key]; !ok {
			req.Header[key] = values
		}
	}

	var rawReq *http.Request
	if rawReq, err = req.IntoHttpRequest(); err != nil {
		return
	}
	return c.c.Do(rawReq)
}

// Fetch sends an http.Request built from Request and returns a Response,
// containing the http.Response and the result of the request.
// The result can be a *string, a *[]byte or an io.Writer.
// If the result is nil, then result will be set as a *bytes.Buffer.
func (c *Client) Fetch(req *request.Request, result any) (resp *Response, err error) {
	var rawResp *http.Response
	if rawResp, err = c.Do(req); err != nil {
		return
	}

	defer rawResp.Body.Close()
	resp = &Response{Raw: rawResp, StatusCode: rawResp.StatusCode}

	if result == nil {
		result = new(bytes.Buffer)
	}

	switch v := result.(type) {
	case io.Writer:
		_, err = io.Copy(v, rawResp.Body)
	case *[]byte:
		*v, err = io.ReadAll(rawResp.Body)
	case *string:
		var b []byte
		b, err = io.ReadAll(rawResp.Body)
		*v = string(b)
	}
	resp.Result = result

	return
}

// JSON sends an http.Request built from Request and returns a Response,
// containing the http.Response and the result of the request.
// The result must be a pointer to entity that can be decoded from json body.
func (c *Client) JSON(req *request.Request, result any) (resp *Response, err error) {
	var rawResp *http.Response
	if rawResp, err = c.Do(req); err != nil {
		return
	}

	defer rawResp.Body.Close()
	resp = &Response{Raw: rawResp, StatusCode: rawResp.StatusCode}

	if result == nil {
		return
	}
	err = json.NewDecoder(rawResp.Body).Decode(result)
	if err != nil {
		return
	}
	resp.Result = result
	return
}

// New creates a new Client with the given options
func New(opts ...ClientOption) *Client {

	c := &Client{
		header: make(http.Header),
	}

	for _, opt := range opts {
		opt(c)
	}

	if c.c == nil {
		c.c = &http.Client{}
	}

	if c.timeout == 0 {
		c.timeout = defaultTimeout
		c.c.Timeout = c.timeout
	}

	if c.jar != nil {
		c.c.Jar = c.jar
	} else if c.c.Jar == nil {
		cookieJar, _ := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
		c.c.Jar = cookieJar
	}

	if c.cookies != nil {
		c.c.Jar.SetCookies(c.baseURL, c.cookies)
	}

	c.logger = log.With().Str("module", "apik").Str("component", "Client").Logger()

	return c
}

// ClientOption is a function that modifies a Client
type ClientOption func(*Client)

// WithHttpClient sets the http.Client to use
func WithHttpClient(hc *http.Client) ClientOption {
	return func(c *Client) {
		c.c = hc
	}
}

// WithTimeout sets the timeout for the http.Client
func WithTimeout(t time.Duration) ClientOption {
	return func(c *Client) {
		c.timeout = t
	}
}

// WithTrace enables tracing for the http.Request
func WithTrace() ClientOption {
	return func(c *Client) {
		c.trace = true
	}
}

// WithCookies sets the cookies for the http.Client
func WithCookies(cookies []*http.Cookie) ClientOption {
	return func(c *Client) {
		c.cookies = cookies
	}
}

// WithCookieJar sets the cookie jar for the http.Client
// Also CookieJar can be set with `WithHttpClient` option.
func WithCookieJar(jar http.CookieJar) ClientOption {
	return func(c *Client) {
		c.jar = jar
	}
}

// WithHeaders sets an http.Header for the http.Client
func WithHeaders(header http.Header) ClientOption {
	return func(c *Client) {
		c.header = header
	}
}

// WithHeader adds a key-value pair in the http.Header for the http.Client
func WithHeader(key, value string) ClientOption {
	return func(c *Client) {
		c.header.Add(key, value)
	}
}

// WithBaseUrl sets the base url for the http.Client
func WithBaseUrl(baseURL string) ClientOption {
	return func(c *Client) {
		u, _ := url.Parse(baseURL)
		c.baseURL = u
	}
}
