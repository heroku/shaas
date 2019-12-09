package ftest

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

var enabled bool
var skipSetup bool
var env *TestingEnvironment

func TestMain(m *testing.M) {
	flag.BoolVar(&enabled, "ftest", false, "enable functional tests")
	flag.BoolVar(&skipSetup, "ftest-skip-setup", false, "skip environment setup")
	flag.Parse()

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

type TestingEnvironment struct {
	projectRoot string
}

func New(skipCreate bool) (*TestingEnvironment, error) {
	if _, err := exec.LookPath("docker-compose"); err != nil {
		log.Fatal("docker-compose can not be found in $PATH. Is it installed?\n" +
			"https://docs.docker.com/compose/#installation-and-set-up")
		return nil, err
	}
	if _, err := exec.LookPath("docker"); err != nil {
		log.Fatal("docker can not be found in $PATH. Is it installed?")
		return nil, err
	}
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	env := &TestingEnvironment{
		projectRoot: path.Clean(filepath.Join(wd, "..")),
	}
	if skipCreate {
		return env, nil
	}
	return env, env.create()
}

func (env *TestingEnvironment) create() error {
	run := func(cmd *exec.Cmd) {
		cmd.Dir = env.projectRoot
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		log.Printf("ftest fn=create at=run cmd=%q", cmd.String())
		if err := cmd.Run(); err != nil {
			panic(err)
		}
	}

	run(exec.Command("docker-compose", "stop"))
	run(exec.Command("docker-compose", "build"))
	run(exec.Command("docker-compose", "up", "-d"))
	run(exec.Command("docker", "cp", filepath.Join(env.projectRoot, "ftest"), "shaas_shaas_1:ftest"))

	log.Print("Waiting for server...")
	var err error
	for i := 0; i < 5; i++ {
		if _, err = http.Get(env.baseUrl()); err == nil {
			return nil
		}
		time.Sleep(time.Second)
	}
	return fmt.Errorf("Server not responding: %s: \n%s", env.baseUrl(), err.Error())
}

func (env *TestingEnvironment) destroy() error {
	stop := exec.Command("docker-compose", "stop")
	stop.Dir = env.projectRoot
	stop.Stdout = os.Stdout
	stop.Stderr = os.Stderr
	return stop.Run()
}

func (env *TestingEnvironment) baseUrl() string {
	b2dUrlStr := os.Getenv("DOCKER_HOST")
	if b2dUrlStr == "" {
		return "http://localhost:5000"
	}

	b2dUrl, err := url.Parse(b2dUrlStr)
	if err != nil {
		panic(err)
	}

	return "http://" + strings.SplitAfter(b2dUrl.Host, ":")[0] + "5000"
}

func (env *TestingEnvironment) fixturesUrl() string {
	return env.baseUrl() + "/ftest/fixtures"
}
