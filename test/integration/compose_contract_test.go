package integration_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestComposeRuntimeBoundaries(t *testing.T) {
	repositoryRoot := repositoryRoot(t)
	compose, err := os.ReadFile(filepath.Join(repositoryRoot, "docker-compose.yaml"))
	if err != nil {
		t.Fatalf("read docker compose file: %v", err)
	}

	composeText := string(compose)
	for _, serviceName := range []string{
		"  notegic-client-gateway:",
		"  notegic-core:",
		"  notegic-realtime-gateway:",
		"  notegic-durable-job:",
		"  notegic-email:",
		"  notegic-yjs-worker:",
	} {
		if !strings.Contains(composeText, serviceName) {
			t.Errorf("docker compose is missing runtime service %q", strings.TrimSpace(serviceName))
		}
	}

	if !strings.Contains(composeText, "/healthz") {
		t.Error("docker compose is missing the health status route")
	}

	for _, environmentVariable := range []string{
		"YJS_DB_HOST:",
		"YJS_DB_PORT:",
		"YJS_DB_USER:",
		"YJS_DB_PASSWORD:",
		"YJS_DB_NAME:",
	} {
		if !strings.Contains(composeText, environmentVariable) {
			t.Errorf("docker compose is missing Yjs database environment variable %q", environmentVariable)
		}
	}

	initCompose, err := os.ReadFile(filepath.Join(repositoryRoot, "docker-compose.init.yaml"))
	if err != nil {
		t.Fatalf("read init docker compose file: %v", err)
	}
	if !strings.Contains(string(initCompose), "notegic-core-database-init:") {
		t.Error("init docker compose is missing the Core database initialization service")
	}
}

func repositoryRoot(t *testing.T) string {
	t.Helper()

	workingDirectory, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}

	return filepath.Clean(filepath.Join(workingDirectory, "..", ".."))
}
