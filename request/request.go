package request

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type FileFieldData struct {
	Fieldname string
	Filename  string
	Source    string
	Body      any
}

func (f *FileFieldData) Write(w *multipart.Writer) (err error) {

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
	}
	return
}

type Request struct {
	Ctx     context.Context
	Method  string
	Header  http.Header
	Body    []byte
	Form    url.Values
	Params  url.Values
	Files   []*FileFieldData
	Cookies []*http.Cookie
	URL     *url.URL
}

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
			file.Write(writer)
		}

		for key, values := range r.Form {
			for _, value := range values {
				writer.WriteField(key, value)
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
	}

	req, err = http.NewRequestWithContext(r.Ctx, r.Method, r.URL.String(), body)
	if err != nil {
		return
	}

	if len(r.Header) > 0 {
		req.Header = r.Header
	}

	if len(multipartContentType) > 0 {
		req.Header.Set("Content-Type", multipartContentType)
	}
	for _, cookie := range r.Cookies {
		req.AddCookie(cookie)

	}
	return
}

type RequestOption func(*Request)

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

func Method(method string) RequestOption {
	return func(r *Request) {
		r.Method = method
	}
}

func Header(key, value string) RequestOption {
	return func(r *Request) {
		r.Header.Add(key, value)
	}
}

func Headers(header http.Header) RequestOption {
	return func(r *Request) {
		r.Header = header
	}
}

func AddParam(key, value string) RequestOption {
	return func(r *Request) {
		r.Params.Add(key, value)
	}
}

func SetParam(key, value string) RequestOption {
	return func(r *Request) {
		r.Params.Set(key, value)
	}
}

func SetParams(params url.Values) RequestOption {
	return func(r *Request) {
		r.Params = params
	}
}

func AddFormField(key, value string) RequestOption {
	return func(r *Request) {
		r.Form.Add(key, value)
	}
}

func SetFormField(key, value string) RequestOption {
	return func(r *Request) {
		r.Form.Set(key, value)
	}
}

func SetForm(form url.Values) RequestOption {
	return func(r *Request) {
		r.Form = form
	}
}

func SetBody(body []byte) RequestOption {
	return func(r *Request) {
		r.Body = body
	}
}

func SetFile(fieldname, source string) RequestOption {
	return func(r *Request) {
		r.Files = append(r.Files, &FileFieldData{
			Fieldname: fieldname,
			Source:    source,
		})
	}
}

func SetFileBody(fieldname, filename string, body any) RequestOption {
	return func(r *Request) {
		r.Files = append(r.Files, &FileFieldData{
			Fieldname: fieldname,
			Filename:  filename,
			Body:      body,
		})
	}
}
