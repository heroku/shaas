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
