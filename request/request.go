package request

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// FileField represents a file field data for one file
type FileField struct {
	// Fieldname is the name of the field
	Fieldname string
	// Filename is the name of the file
	Filename string
	// Source is the path to the file
	Source string
	// Body is the content of the file
	Body any
}

// Write writes the file field to the multipart writer
func (f *FileField) Write(w *multipart.Writer) (err error) {

	if f.Source != "" {
		body, err := os.ReadFile(f.Source)
		if err != nil {
			return err
		}
		f.Body = body

		f.Filename = filepath.Base(f.Source)
	}

	part, err := w.CreateFormFile(f.Fieldname, f.Filename)
	if err != nil {
		return
	}
	switch v := f.Body.(type) {
	case string:
		_, err = part.Write([]byte(v))
	case []byte:
		_, err = part.Write(v)
	case io.Reader:
		_, err = io.Copy(part, v)
	default:
		err = fmt.Errorf("%w: %T", ErrUnsupportedBodyType, f.Body)
	}
	return
}

// Request represents a  wrapper around http.Request
type Request struct {
	// Ctx is the context of the request
	Ctx context.Context
	// Method is the HTTP method. Default is GET
	Method string
	// Header is the HTTP headers
	Header http.Header
	// Body is the raw request body
	Body []byte
	// Form is the form data that will be encoded as application/x-www-form-urlencoded
	Form url.Values
	// Params is the query parameters
	Params url.Values
	// Files represents the files that will be sent in the request's body as multipart/form-data
	Files []*FileField
	// Cookies is the cookies that will be sent in the request
	Cookies []*http.Cookie
	// URL is the URL of the request
	URL       *url.URL
	trace     bool
	traceInfo *TraceInfo
}

// TraceInfo represents the trace information of the request. Available only if the request is traced.
func (r *Request) TraceInfo() *TraceInfo {
	return r.traceInfo
}

// IntoHttpRequest converts the request to http.Request
func (r *Request) IntoHttpRequest() (req *http.Request, err error) {

	if len(r.Params) > 0 {
		r.URL.RawQuery = r.Params.Encode()
	}

	var body io.Reader
	var multipartContentType string

	if len(r.Files) > 0 {
		buf := new(bytes.Buffer)
		writer := multipart.NewWriter(buf)
		for _, file := range r.Files {
			if err = file.Write(writer); err != nil {
				return
			}
		}

		for key, values := range r.Form {
			for _, value := range values {
				if err = writer.WriteField(key, value); err != nil {
					return
				}
			}
		}
		if err = writer.Close(); err != nil {
			return
		}
		body = buf
		multipartContentType = writer.FormDataContentType()

	} else if len(r.Body) > 0 {
		body = bytes.NewReader(r.Body)
	} else if len(r.Form) > 0 {
		body = strings.NewReader(r.Form.Encode())
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	req, err = http.NewRequestWithContext(r.Ctx, r.Method, r.URL.String(), body)
	if err != nil {
		return
	}

	if r.trace {
		info, ctx := createTraceContext(req.Context())
		r.traceInfo = info
		req = req.WithContext(ctx)
	}

	req.Header = r.Header

	if ct := req.Header.Get("Content-Type"); ct == "" && r.Method == http.MethodPost {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	if len(multipartContentType) > 0 {
		req.Header.Set("Content-Type", multipartContentType)
	}
	for _, cookie := range r.Cookies {
		req.AddCookie(cookie)

	}
	return
}

// NewRequest creates a new wrapped request with options
func NewRequest(ctx context.Context, dstURL string, opts ...RequestOption) *Request {

	parsedURL, _ := url.Parse(dstURL)

	r := &Request{
		Ctx:    ctx,
		URL:    parsedURL,
		Header: make(http.Header),
		Form:   make(url.Values),
		Params: make(url.Values),
	}

	for _, opt := range opts {
		opt(r)
	}

	return r
}

// RequestOption is a function that modifies the request
type RequestOption func(*Request)

// Method sets the HTTP method of the request
func Method(method string) RequestOption {
	return func(r *Request) {
		r.Method = method
	}
}

// Header adds one HTTP header
func Header(key, value string) RequestOption {
	return func(r *Request) {
		r.Header.Add(key, value)
	}
}

// Headers sets the HTTP header
func Headers(header http.Header) RequestOption {
	return func(r *Request) {
		r.Header = header
	}
}

// AddParam adds a query parameter
func AddParam(key, value string) RequestOption {
	return func(r *Request) {
		r.Params.Add(key, value)
	}
}

// SetParam sets the query parameter
func SetParam(key, value string) RequestOption {
	return func(r *Request) {
		r.Params.Set(key, value)
	}
}

// SetParams sets the query parameters
func SetParams(params url.Values) RequestOption {
	return func(r *Request) {
		r.Params = params
	}
}

// AddFormField adds a form field
func AddFormField(key, value string) RequestOption {
	return func(r *Request) {
		r.Form.Add(key, value)
	}
}

// SetFormField sets a form field
func SetFormField(key, value string) RequestOption {
	return func(r *Request) {
		r.Form.Set(key, value)
	}
}

// SetForm sets the form data
func SetForm(form url.Values) RequestOption {
	return func(r *Request) {
		r.Form = form
	}
}

// SetBody sets the raw request body
func SetBody(body []byte) RequestOption {
	return func(r *Request) {
		r.Body = body
	}
}

// SetFile sets a file field
func SetFile(fieldname, source string) RequestOption {
	return func(r *Request) {
		r.Files = append(r.Files, &FileField{
			Fieldname: fieldname,
			Source:    source,
		})
	}
}

// SetFileBody sets a file field with body
func SetFileBody(fieldname, filename string, body any) RequestOption {
	return func(r *Request) {
		r.Files = append(r.Files, &FileField{
			Fieldname: fieldname,
			Filename:  filename,
			Body:      body,
		})
	}
}

// Trace enables tracing for the request
func Trace() RequestOption {
	return func(r *Request) {
		r.trace = true
	}
}

// AddCookie adds a cookie
func AddCookie(cookie *http.Cookie) RequestOption {
	return func(r *Request) {
		r.Cookies = append(r.Cookies, cookie)
	}
}

// SetCookies sets the cookies
func SetCookies(cookies []*http.Cookie) RequestOption {
	return func(r *Request) {
		r.Cookies = cookies
	}
}

// reqopt
