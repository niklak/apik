package request

import (
	"bytes"
	"context"
	"encoding/json"
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
	URL *url.URL
	// Trace is a flag that indicates if the request should be traced
	Trace bool
	// JSON is a entity to be sent as JSON
	JSON      any
	traceInfo *TraceInfo
}

// TraceInfo represents the trace information of the request. Available only if the request is traced.
func (r *Request) TraceInfo() *TraceInfo {
	return r.traceInfo
}

func (r *Request) writeMultiPartFormData() (body io.Reader, err error) {
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
	r.Header.Set("Content-Type", writer.FormDataContentType())
	return
}

func (r *Request) writeJSON() (body io.Reader, err error) {
	buf := new(bytes.Buffer)
	err = json.NewEncoder(buf).Encode(r.JSON)
	if err != nil {
		return
	}
	body = buf
	r.Header.Set("Content-Type", "application/json")
	return
}

func (r *Request) writeForm() (body io.Reader) {
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return strings.NewReader(r.Form.Encode())
}

// IntoHttpRequest converts the request to http.Request
func (r *Request) IntoHttpRequest() (req *http.Request, err error) {

	if len(r.Params) > 0 {
		r.URL.RawQuery = r.Params.Encode()
	}

	var body io.Reader

	if len(r.Files) > 0 {
		body, err = r.writeMultiPartFormData()
	} else if len(r.Form) > 0 {
		body = r.writeForm()
	} else if r.JSON != nil {
		body, err = r.writeJSON()
	} else if len(r.Body) > 0 {
		body = bytes.NewReader(r.Body)
	}

	if err != nil {
		return
	}

	req, err = http.NewRequestWithContext(r.Ctx, r.Method, r.URL.String(), body)
	if err != nil {
		return
	}

	if r.Trace {
		info, ctx := createTraceContext(req.Context())
		r.traceInfo = info
		req = req.WithContext(ctx)
	}

	req.Header = r.Header

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
