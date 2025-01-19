package ftest

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHealthEndpoint(t *testing.T) {
	res, err := http.Get(env.baseUrl("auth") + "/health")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)

	// Test status=204
	uri, err := url.Parse(env.baseUrl("auth") + "/health")
	assert.Nil(t, err)
	q := uri.Query()
	q.Add("status", "204")
	uri.RawQuery = q.Encode()
	res, err = http.Get(uri.String())
	assert.Nil(t, err)
	assert.Equal(t, http.StatusNoContent, res.StatusCode)
}

func TestGetFile(t *testing.T) {
	res, err := http.Get(env.fixturesUrl("default") + "/a")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)

	body, err := ioutil.ReadAll(res.Body)
	assert.Nil(t, err)
	assert.Equal(t, "A\n", string(body))
}

func TestGetFile_NotFound(t *testing.T) {
	res, err := http.Get(env.fixturesUrl("default") + "/b")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusNotFound, res.StatusCode)
}

func TestGetDir(t *testing.T) {
	res, err := http.Get(env.fixturesUrl("default"))
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)

	body, err := ioutil.ReadAll(res.Body)
	assert.Nil(t, err)
	dir := &map[string]FileInfoDetails{}
	assert.Nil(t, json.Unmarshal(body, dir))

	a := (*dir)["a"]
	assert.NotNil(t, a)
	assert.Equal(t, "-", a.Type)
	assert.Equal(t, int64(2), a.Size)
	assert.Equal(t, 420, a.Perm)
}

func TestPostFile(t *testing.T) {
	res, err := http.Post(env.baseUrl("default")+"/usr/bin/factor", "", strings.NewReader("42"))
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)

	body, err := ioutil.ReadAll(res.Body)
	assert.Nil(t, err)
	assert.Equal(t, "42: 2 3 7\n", string(body))
}

func TestPostFile_NotFound(t *testing.T) {
	res, err := http.Post(env.fixturesUrl("default")+"/b", "", strings.NewReader(""))
	assert.Nil(t, err)
	assert.Equal(t, http.StatusNotFound, res.StatusCode)
}

func TestPostDir(t *testing.T) {
	res, err := http.Post(env.fixturesUrl("default"), "", strings.NewReader("pwd"))
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)

	body, err := ioutil.ReadAll(res.Body)
	assert.Nil(t, err)
	assert.Equal(t, strings.TrimPrefix(env.fixturesUrl("default"), env.baseUrl("default"))+"\n", string(body))
}

func TestReadonlyAllowsGet(t *testing.T) {
	res, err := http.Get(env.fixturesUrl("readonly"))
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)
}

func TestReadonlyForbidsNonGet(t *testing.T) {
	res, err := http.Post(env.fixturesUrl("readonly"), "", nil)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusMethodNotAllowed, res.StatusCode)
}

func TestBasicAuthAuthorized(t *testing.T) {
	uri := env.baseUrl("auth") + "/health"
	req, _ := http.NewRequest(http.MethodGet, uri, nil)
	req.SetBasicAuth("user", "pass")

	res, err := http.DefaultClient.Do(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)
}

func TestBasicAuthUnauthorizedMissingAuth(t *testing.T) {
	res, err := http.Get(env.baseUrl("auth") + "/health")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
}

func TestBasicAuthUnauthorizedWrongAuth(t *testing.T) {
	uri := env.baseUrl("auth") + "/health"
	req, _ := http.NewRequest(http.MethodGet, uri, nil)
	req.SetBasicAuth("wrong", "credentials")

	res, err := http.DefaultClient.Do(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
}

// FileInfoDetails contains basic stat + permission details about a file
type FileInfoDetails struct {
	Size    int64     `json:"size"`
	Type    string    `json:"type"`
	Perm    int       `json:"permission"`
	ModTime time.Time `json:"updated_at"`
}

func toFileInfoDetails(fi os.FileInfo) FileInfoDetails {
	return FileInfoDetails{
		Size:    fi.Size(),
		Type:    string(fi.Mode().String()[0]),
		Perm:    int(fi.Mode().Perm()),
		ModTime: fi.ModTime(),
	}
}
