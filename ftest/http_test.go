package ftest

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"os"
	"testing"

	. "github.com/heroku/shaas"
	"strings"
)

var env *TestingEnvironment

func TestMain(m *testing.M) {
	var (
		status int
		err    error
	)
	if !enabled {
		fmt.Fprintln(os.Stderr, "WARNING: functional tests are not enabled")
		os.Exit(0)
	}
	env, err = New(skipSetup)
	if err != nil {
		panic(err)
	}
	defer func() {
		if !skipSetup {
			if err := env.destroy(); err != nil {
				panic(err)
			}
		}
		os.Exit(status)
	}()
	status = m.Run()
}

func TestGetFile(t *testing.T) {
	res, err := http.Get(env.fixturesUrl() + "/a")
	assert.Nil(t, err)
	assert.Equal(t, res.StatusCode, http.StatusOK)

	body, err := ioutil.ReadAll(res.Body)
	assert.Nil(t, err)
	assert.Equal(t, string(body), "A\n")
}

func TestGetFile_NotFound(t *testing.T) {
	res, err := http.Get(env.fixturesUrl() + "/b")
	assert.Nil(t, err)
	assert.Equal(t, res.StatusCode, http.StatusNotFound)
}

func TestGetDir(t *testing.T) {
	res, err := http.Get(env.fixturesUrl())
	assert.Nil(t, err)
	assert.Equal(t, res.StatusCode, http.StatusOK)

	body, err := ioutil.ReadAll(res.Body)
	assert.Nil(t, err)
	dir := &map[string]FileInfoDetails{}
	assert.Nil(t, json.Unmarshal(body, dir))

	a := (*dir)["a"]
	assert.NotNil(t, a)
	assert.Equal(t, a.Type, "-")
	assert.Equal(t, a.Size, int64(2))
	assert.Equal(t, a.Perm, 420)
}

func TestPostFile(t *testing.T) {
	res, err := http.Post(env.baseUrl()+"/usr/bin/factor", "", strings.NewReader("42"))
	assert.Nil(t, err)
	assert.Equal(t, res.StatusCode, http.StatusOK)

	body, err := ioutil.ReadAll(res.Body)
	assert.Nil(t, err)
	assert.Equal(t, string(body), "42: 2 3 7\n")
}

func TestPostFile_NotFound(t *testing.T) {
	res, err := http.Post(env.fixturesUrl()+"/b", "", strings.NewReader(""))
	assert.Nil(t, err)
	assert.Equal(t, res.StatusCode, http.StatusNotFound)
}

func TestPostDir(t *testing.T) {
	res, err := http.Post(env.fixturesUrl(), "", strings.NewReader("pwd"))
	assert.Nil(t, err)
	assert.Equal(t, res.StatusCode, http.StatusOK)

	body, err := ioutil.ReadAll(res.Body)
	assert.Nil(t, err)
	assert.Equal(t, string(body), strings.TrimPrefix(env.fixturesUrl(), env.baseUrl())+"\n")
}
