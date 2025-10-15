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

const (
	dockerComposeCmd  = "docker-compose"
	dockerComposeFile = "ftest/docker-compose.yml"

	ServiceDefault  = "default"
	ServiceAuth     = "auth"
	ServiceReadonly = "readonly"
	ServiceSlugonly = "slugonly"
)

var (
	enabled   bool
	skipSetup bool
	env       *TestingEnvironment
)

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
	services    map[string]TestingEnvironmentService
}

type TestingEnvironmentService struct {
	Port     int
	Auth     string
	Readonly bool
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
		services: map[string]TestingEnvironmentService{
			ServiceDefault: {
				Port: 5000,
			},
			ServiceAuth: {
				Port: 5001,
				Auth: "user:pass",
			},
			ServiceReadonly: {
				Port:     5002,
				Readonly: true,
			},
			ServiceSlugonly: {
				Port: 5003,
			},
		},
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

	run(exec.Command(dockerComposeCmd, "-f", dockerComposeFile, "stop"))
	run(exec.Command(dockerComposeCmd, "-f", dockerComposeFile, "build"))
	run(exec.Command(dockerComposeCmd, "-f", dockerComposeFile, "up", "-d"))

	for svcName, svc := range env.services {
		run(exec.Command("docker", "cp", filepath.Join(env.projectRoot, "ftest"), fmt.Sprintf("ftest_shaas.%s_1:ftest", svcName)))

		log.Print("Waiting for server...")
		var err error
		for i := 0; i < 5; i++ {
			if _, err = http.Get(env.baseUrl(svc)); err == nil {
				return nil
			}
			time.Sleep(time.Second)
		}
		return fmt.Errorf("Server not responding: %s: \n%s", env.baseUrl(svc), err.Error())
	}

	return nil
}

func (env *TestingEnvironment) destroy() error {
	stop := exec.Command("docker-compose", "stop")
	stop.Dir = env.projectRoot
	stop.Stdout = os.Stdout
	stop.Stderr = os.Stderr
	return stop.Run()
}

func (env *TestingEnvironment) baseUrl(svc TestingEnvironmentService) string {
	auth := ""
	if svc.Auth != "" {
		auth = fmt.Sprintf("%s@", svc.Auth)
	}

	host := "localhost"
	if b2dUrlStr, ok := os.LookupEnv("DOCKER_HOST"); ok {
		b2dUrl, err := url.Parse(b2dUrlStr)
		if err != nil {
			panic(err)
		}
		host = strings.SplitAfter(b2dUrl.Host, ":")[0]
	}

	return fmt.Sprintf("http://%s%s:%d", auth, host, svc.Port)
}

func (env *TestingEnvironment) fixturesUrl(svc TestingEnvironmentService) string {
	return env.baseUrl(svc) + "/ftest/fixtures"
}
