package ftest

import (
	"fmt"
	"os"
	"testing"
)

type TestingEnvironment struct {
	services map[string]ServiceConfig
}

type ServiceConfig struct {
	Name     string
	Port     int
	Auth     string
	Readonly bool
	BaseURL  string
}

var env *TestingEnvironment

func TestMain(m *testing.M) {
	// Initialize environment with predefined services
	env = &TestingEnvironment{
		services: map[string]ServiceConfig{
			"default": {
				Name:    "default",
				Port:    5003,
				BaseURL: "http://localhost:5003",
			},
			"auth": {
				Name:    "auth",
				Port:    5001,
				Auth:    "user:pass",
				BaseURL: "http://localhost:5001",
			},
			"readonly": {
				Name:     "readonly",
				Port:     5002,
				Readonly: true,
				BaseURL:  "http://localhost:5002",
			},
		},
	}

	// Run tests
	status := m.Run()

	// Exit with the test status
	os.Exit(status)
}

func (env *TestingEnvironment) baseUrl(service string) string {
	return env.services[service].BaseURL
}

func (env *TestingEnvironment) fixturesUrl(service string) string {
	pwd, err := os.Getwd()
	if err != nil {
		panic(fmt.Sprintf("Failed to get current working directory: %v", err))
	}

	return env.baseUrl(service) + pwd + "/fixtures"
}
