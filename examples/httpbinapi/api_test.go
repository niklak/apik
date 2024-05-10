package httpbinapi

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestApiGet(t *testing.T) {

	cli := New()
	res, err := cli.Get(10, 20)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "https://httpbin.org/get?limit=10&offset=20", res.URL)
	assert.Equal(t, "10", res.Args["limit"])
}

func TestApiPost(t *testing.T) {

	cli := New()
	res, err := cli.Post("address", "comment")
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "https://httpbin.org/post", res.URL)

	obj := res.JSON.(map[string]interface{})
	address := obj["address"].(string)
	comment := obj["comment"].(string)
	assert.Equal(t, "address", address)
	assert.Equal(t, "comment", comment)
}

func TestApiPut(t *testing.T) {

	cli := New()
	res, err := cli.Put(url.Values{"key": []string{"value"}})
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "https://httpbin.org/put", res.URL)

	assert.Equal(t, "value", res.Form["key"])
}

func TestApiDelete(t *testing.T) {

	cli := New()
	res, err := cli.Delete()
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "https://httpbin.org/delete", res.URL)
}

func TestApiPatch(t *testing.T) {

	cli := New()
	res, err := cli.Patch("key", "value")
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "https://httpbin.org/patch", res.URL)

	assert.Equal(t, "value", res.Form["key"])
}
