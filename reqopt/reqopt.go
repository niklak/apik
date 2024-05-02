package reqopt

import (
	"net/http"
	"net/url"

	"github.com/niklak/apik/request"
)

// Method sets the HTTP method of the request
func Method(method string) request.RequestOption {
	return func(r *request.Request) {
		r.Method = method
	}
}

// Header adds one HTTP header
func Header(key, value string) request.RequestOption {
	return func(r *request.Request) {
		r.Header.Add(key, value)
	}
}

// Headers sets the HTTP header
func Headers(header http.Header) request.RequestOption {
	return func(r *request.Request) {
		r.Header = header
	}
}

// AddParam adds a query parameter
func AddParam(key, value string) request.RequestOption {
	return func(r *request.Request) {
		r.Params.Add(key, value)
	}
}

// SetParam sets the query parameter
func SetParam(key, value string) request.RequestOption {
	return func(r *request.Request) {
		r.Params.Set(key, value)
	}
}

// SetParams sets the query parameters
func SetParams(params url.Values) request.RequestOption {
	return func(r *request.Request) {
		r.Params = params
	}
}

// AddFormField adds a form field
func AddFormField(key, value string) request.RequestOption {
	return func(r *request.Request) {
		r.Form.Add(key, value)
	}
}

// SetFormField sets a form field
func SetFormField(key, value string) request.RequestOption {
	return func(r *request.Request) {
		r.Form.Set(key, value)
	}
}

// SetForm sets the form data
func SetForm(form url.Values) request.RequestOption {
	return func(r *request.Request) {
		r.Form = form
	}
}

// SetBody sets the raw request body
func SetBody(body []byte) request.RequestOption {
	return func(r *request.Request) {
		r.Body = body
	}
}

// SetFile sets a file field
func SetFile(fieldname, source string) request.RequestOption {
	return func(r *request.Request) {
		r.Files = append(r.Files, &request.FileField{
			Fieldname: fieldname,
			Source:    source,
		})
	}
}

// SetFileBody sets a file field with body
func SetFileBody(fieldname, filename string, body any) request.RequestOption {
	return func(r *request.Request) {
		r.Files = append(r.Files, &request.FileField{
			Fieldname: fieldname,
			Filename:  filename,
			Body:      body,
		})
	}
}

// Trace enables tracing for the request
func Trace() request.RequestOption {
	return func(r *request.Request) {
		r.Trace = true
	}
}

// AddCookie adds a cookie
func AddCookie(cookie *http.Cookie) request.RequestOption {
	return func(r *request.Request) {
		r.Cookies = append(r.Cookies, cookie)
	}
}

// SetCookies sets the cookies
func SetCookies(cookies []*http.Cookie) request.RequestOption {
	return func(r *request.Request) {
		r.Cookies = cookies
	}
}

// SetJSON sets the URL of the request
func SetJSON(entity any) request.RequestOption {
	return func(r *request.Request) {
		r.JSON = entity
	}
}
