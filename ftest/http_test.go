package ftest

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
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
