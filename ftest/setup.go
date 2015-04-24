package ftest

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"time"
)

const baseUrl = "http://localhost:5000"
const appUrl  = baseUrl + "/go/src/app"
const fixturesUrl  = appUrl + "/ftest/fixtures"

var enabled bool
var skipSetup bool

func init() {
	flag.BoolVar(&enabled, "ftest", false, "enable functional tests")
	flag.BoolVar(&skipSetup, "ftest-skip-setup", false, "skip environment setup")
	flag.Parse()
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
	stop := exec.Command("docker-compose", "stop")
	stop.Dir = env.projectRoot
	stop.Stdout = os.Stdout
	stop.Stderr = os.Stderr
	if err := stop.Run(); err != nil {
		return err
	}
	build := exec.Command("docker-compose", "build")
	build.Dir = env.projectRoot
	build.Stdout = os.Stdout
	build.Stderr = os.Stderr
	if err := build.Run(); err != nil {
		return err
	}
	start := exec.Command("docker-compose", "up", "-d")
	start.Dir = env.projectRoot
	start.Stdout = os.Stdout
	start.Stderr = os.Stderr
	start.Run()

	log.Print("Waiting for server...")
	for i := 0; i < 5; i++ {
		if _, err := http.Get(baseUrl); err == nil {
			return nil
		}
		time.Sleep(time.Second)
	}
	return fmt.Errorf("Server not responding")
}

func (env *TestingEnvironment) Destroy() error {
	stop := exec.Command("docker-compose", "stop")
	stop.Dir = env.projectRoot
	stop.Stdout = os.Stdout
	stop.Stderr = os.Stderr
	return stop.Run()
}
