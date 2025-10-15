package ftest

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/heroku/shaas/pkg"
)

func TestHealthEndpoint(t *testing.T) {
	uri, err := url.Parse(env.baseUrl(env.services[ServiceAuth]) + "/health")
	assert.Nil(t, err)
	uri.User = nil

	res, err := http.Get(uri.String())
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)

	q := uri.Query()
	q.Add("status", "204")
	uri.RawQuery = q.Encode()
	res, err = http.Get(uri.String())
	assert.Nil(t, err)
	assert.Equal(t, http.StatusNoContent, res.StatusCode)
}

func TestGetFile(t *testing.T) {
	res, err := http.Get(env.fixturesUrl(env.services[ServiceDefault]) + "/a")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)

	body, err := ioutil.ReadAll(res.Body)
	assert.Nil(t, err)
	assert.Equal(t, "A\n", string(body))
}

func TestGetFile_NotFound(t *testing.T) {
	res, err := http.Get(env.fixturesUrl(env.services[ServiceDefault]) + "/b")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusNotFound, res.StatusCode)
}

func TestGetDir(t *testing.T) {
	res, err := http.Get(env.fixturesUrl(env.services[ServiceDefault]))
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)

	body, err := ioutil.ReadAll(res.Body)
	assert.Nil(t, err)
	dir := &map[string]pkg.FileInfoDetails{}
	assert.Nil(t, json.Unmarshal(body, dir))

	a := (*dir)["a"]
	assert.NotNil(t, a)
	assert.Equal(t, "-", a.Type)
	assert.Equal(t, int64(2), a.Size)
	assert.Equal(t, 420, a.Perm)
}

func TestPostFile(t *testing.T) {
	res, err := http.Post(env.baseUrl(env.services[ServiceDefault])+"/usr/bin/factor", "", strings.NewReader("42"))
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)

	body, err := ioutil.ReadAll(res.Body)
	assert.Nil(t, err)
	assert.Equal(t, "42: 2 3 7\n", string(body))
}

func TestPostFile_NotFound(t *testing.T) {
	res, err := http.Post(env.fixturesUrl(env.services[ServiceDefault])+"/b", "", strings.NewReader(""))
	assert.Nil(t, err)
	assert.Equal(t, http.StatusNotFound, res.StatusCode)
}

func TestPostDir(t *testing.T) {
	res, err := http.Post(env.fixturesUrl(env.services[ServiceDefault]), "", strings.NewReader("pwd"))
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)

	body, err := ioutil.ReadAll(res.Body)
	assert.Nil(t, err)
	assert.Equal(t, strings.TrimPrefix(env.fixturesUrl(env.services[ServiceDefault]), env.baseUrl(env.services[ServiceDefault]))+"\n", string(body))
}

func TestReadonlyAllowsGet(t *testing.T) {
	res, err := http.Get(env.fixturesUrl(env.services[ServiceReadonly]))
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)
}

func TestReadonlyForbidsNonGet(t *testing.T) {
	res, err := http.Post(env.fixturesUrl(env.services[ServiceReadonly]), "", nil)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusMethodNotAllowed, res.StatusCode)
}

func TestBasicAuthAuthorized(t *testing.T) {
	uri, err := url.Parse(env.fixturesUrl(env.services[ServiceAuth]))
	assert.Nil(t, err)
	assert.NotNil(t, uri.User)

	res, err := http.Get(uri.String())
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)
}

func TestBasicAuthUnauthorizedMissingAuth(t *testing.T) {
	uri, err := url.Parse(env.fixturesUrl(env.services[ServiceAuth]))
	assert.Nil(t, err)
	uri.User = nil

	res, err := http.Get(uri.String())
	assert.Nil(t, err)
	assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
}

func TestBasicAuthUnauthorizedWrongAuth(t *testing.T) {
	uri, err := url.Parse(env.fixturesUrl(env.services[ServiceAuth]))
	assert.Nil(t, err)
	uri.User = url.UserPassword("not", "correct")

	res, err := http.Get(uri.String())
	assert.Nil(t, err)
	assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
}

func TestSlugonlyAllowsSlugTgz(t *testing.T) {
	res, err := http.Get(env.baseUrl(env.services[ServiceSlugonly]) + "/app/slug.tgz")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)
}

func TestSlugonlyAllowsChecksum(t *testing.T) {
	res, err := http.Get(env.baseUrl(env.services[ServiceSlugonly]) + "/app/slug.tgz.sha256")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)
}

func TestSlugonlyForbidsOtherFiles(t *testing.T) {
	res, err := http.Get(env.fixturesUrl(env.services[ServiceSlugonly]) + "/a")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusForbidden, res.StatusCode)
}

func TestSlugonlyForbidsPost(t *testing.T) {
	res, err := http.Post(env.baseUrl(env.services[ServiceSlugonly])+"/app/slug.tgz", "", nil)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusForbidden, res.StatusCode)
}
