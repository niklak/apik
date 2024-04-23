package apik

import (
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"

	"github.com/niklak/apik/request"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/net/publicsuffix"
)

var defaultTimeout = time.Minute

type Client struct {
	c       *http.Client
	timeout time.Duration
	trace   bool
	logger  zerolog.Logger
	cookies []*http.Cookie
	header  http.Header
	baseURL *url.URL
}

func (c *Client) Do(req *request.Request) (res *http.Response, err error) {

	if c.baseURL != nil {
		req.URL = c.baseURL.ResolveReference(req.URL)
	}

	var rawReq *http.Request
	if rawReq, err = req.IntoHttpRequest(); err != nil {
		return

	}
	return c.c.Do(rawReq)
}

func New(opts ...ClientOption) *Client {
	cookieJar, _ := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	c := &Client{
		header: make(http.Header),
	}

	for _, opt := range opts {
		opt(c)
	}

	if c.c == nil {
		c.c = &http.Client{
			Jar: cookieJar,
		}
	}

	if c.timeout == 0 {
		c.timeout = defaultTimeout
		c.c.Timeout = c.timeout
	}

	c.logger = log.With().Str("module", "apik").Str("component", "Client").Logger()

	return c
}

type ClientOption func(*Client)

func WithHttpClient(hc *http.Client) ClientOption {
	return func(c *Client) {
		c.c = hc
	}
}

func WithTimeout(t time.Duration) ClientOption {
	return func(c *Client) {
		c.timeout = t
	}
}

func WithTrace() ClientOption {
	return func(c *Client) {
		c.trace = true
	}
}

func WithCookies(cookies []*http.Cookie) ClientOption {
	return func(c *Client) {
		c.cookies = cookies
		if c.c.Jar != nil {
			c.c.Jar.SetCookies(nil, cookies)
		}
	}
}

func WithCookieJar(jar http.CookieJar) ClientOption {
	return func(c *Client) {
		c.c.Jar = jar
	}

}

func WithHeaders(header http.Header) ClientOption {
	return func(c *Client) {
		c.header = header
	}
}

func WithHeader(key, value string) ClientOption {
	return func(c *Client) {
		c.header.Add(key, value)
	}
}

func WithBaseUrl(baseURL string) ClientOption {
	return func(c *Client) {
		u, _ := url.Parse(baseURL)
		c.baseURL = u
	}
}
