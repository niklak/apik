package httpbinapi

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/niklak/apik"
	"github.com/niklak/apik/reqopt"
	"github.com/niklak/apik/request"
)

const apiBaseURL = "https://httpbin.org/"

const (
	getEndpoint    = "/get"
	postEndpoint   = "/post"
	putEndpoint    = "/put"
	deleteEndpoint = "/delete"
	patchEndpoint  = "/patch"
)

type MethodsResponse struct {
	Args    map[string]string `json:"args"`
	Data    string            `json:"data"`
	Files   map[string]string `json:"files"`
	Form    map[string]string `json:"form"`
	Headers map[string]string `json:"headers"`
	JSON    interface{}       `json:"json"`
	Origin  string            `json:"origin"`
	URL     string            `json:"url"`
}

type HttpBinApi struct {
	c *apik.Client
}

func (c *HttpBinApi) methodsRequest(endpoint string, opts ...request.RequestOption) (res *MethodsResponse, err error) {
	req := apik.NewRequest(context.Background(), endpoint, opts...)
	res = new(MethodsResponse)
	_, err = c.c.JSON(req, res)
	return
}

func (c *HttpBinApi) Get(limit int, offset int) (res *MethodsResponse, err error) {

	return c.methodsRequest(
		getEndpoint,
		reqopt.AddParam("limit", strconv.Itoa(limit)),
		reqopt.AddParam("offset", strconv.Itoa(offset)),
	)
}

func (c *HttpBinApi) Post(address string, comment string) (res *MethodsResponse, err error) {
	return c.methodsRequest(
		postEndpoint,
		reqopt.Method(http.MethodPost),
		reqopt.SetJSON(map[string]string{"address": address, "comment": comment}),
	)
}

func (c *HttpBinApi) Put(form url.Values) (res *MethodsResponse, err error) {
	return c.methodsRequest(
		putEndpoint,
		reqopt.Method(http.MethodPut),
		reqopt.SetForm(form),
	)
}

func (c *HttpBinApi) Delete() (res *MethodsResponse, err error) {
	return c.methodsRequest(deleteEndpoint, reqopt.Method(http.MethodDelete))
}

func (c *HttpBinApi) Patch(key, value string) (res *MethodsResponse, err error) {
	return c.methodsRequest(
		patchEndpoint,
		reqopt.Method(http.MethodPatch),
		reqopt.AddFormField(key, value),
	)
}

func New() *HttpBinApi {
	return &HttpBinApi{
		c: apik.New(apik.WithBaseUrl(apiBaseURL)),
	}
}
