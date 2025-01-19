package ftest

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type TestingEnvironment struct {
	services map[string]ServiceConfig
}

type ServiceConfig struct {
	Name     string
	Port     int
	Auth     string
	Readonly bool
	Cmd      *exec.Cmd
	BaseURL  string
}

var env *TestingEnvironment

func TestMain(m *testing.M) {
	// Compile the shaas binary
	binaryPath, err := compileShaasBinary()
	if err != nil {
		log.Fatalf("Failed to compile shaas binary: %v", err)
	}

	// Initialize and start services
	env, err = NewTestingEnvironment(binaryPath)
	if err != nil {
		log.Fatalf("Failed to initialize testing environment: %v", err)
	}
	defer env.Shutdown()

	// Wait for services to initialize
	time.Sleep(1 * time.Second)

	// Run tests
	status := m.Run()

	// Cleanup and exit
	os.Exit(status)
}

func compileShaasBinary() (string, error) {
	// Adjust the binary output and source path
	binaryPath := filepath.Join("..", "shaas")
	sourcePath := filepath.Join("..", "shaas.go")

	// Compile the shaas binary
	cmd := exec.Command("go", "build", "-o", binaryPath, sourcePath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to compile shaas: %w", err)
	}
	return binaryPath, nil
}

func NewTestingEnvironment(binaryPath string) (*TestingEnvironment, error) {
	services := map[string]ServiceConfig{
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
	}

	env := &TestingEnvironment{services: services}

	// Start each service
	for name, config := range services {
		cmd, err := startShaasService(binaryPath, config.Port, config.Auth, config.Readonly)
		if err != nil {
			env.Shutdown() // Stop any running services
			return nil, fmt.Errorf("failed to start %s service: %w", name, err)
		}
		config.Cmd = cmd
		services[name] = config
	}

	return env, nil
}

func stopExistingService(port int) {
	cmd := exec.Command("lsof", "-ti", fmt.Sprintf(":%d", port))
	output, err := cmd.Output()
	if err == nil && len(output) > 0 {
		pid := strings.TrimSpace(string(output))
		exec.Command("kill", "-9", pid).Run()
	}
}

func startShaasService(binaryPath string, port int, auth string, readonly bool) (*exec.Cmd, error) {
	stopExistingService(port)

	cmd := exec.Command(binaryPath, "--port", fmt.Sprintf("%d", port))
	cmd.Env = os.Environ()
	if auth != "" {
		cmd.Env = append(cmd.Env, fmt.Sprintf("BASIC_AUTH=%s", auth))
	}
	if readonly {
		cmd.Env = append(cmd.Env, "READ_ONLY=1")
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start shaas service on port %d: %w", port, err)
	}
	return cmd, nil
}

func (env *TestingEnvironment) Shutdown() {
	for name, config := range env.services {
		if config.Cmd != nil {
			if err := config.Cmd.Process.Kill(); err != nil {
				log.Printf("Failed to stop %s service: %v", name, err)
			}
		}
	}
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
