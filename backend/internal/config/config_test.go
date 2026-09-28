package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadTOML(t *testing.T) {
	path := writeConfig(t, `environment = "test"
[backend]
storage_backend = "postgres"
server_address = ":9090"
issuer_url = "https://issuer.example.com"
[frontend]
origin = "https://admin.example.com"
public_url = "https://admin.example.com"
[postgres]
dsn = "postgres://file-value"
`)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Backend.ServerAddress != ":9090" {
		t.Fatalf("ServerAddress = %q", cfg.Backend.ServerAddress)
	}
	if cfg.Frontend.Origin != "https://admin.example.com" {
		t.Fatalf("Frontend origin = %q", cfg.Frontend.Origin)
	}
	if cfg.Postgres.DSN != "postgres://file-value" {
		t.Fatalf("Postgres DSN = %q", cfg.Postgres.DSN)
	}
}

func TestLoadRejectsPostgresWithoutDSN(t *testing.T) {
	path := writeConfig(t, `[backend]
storage_backend = "postgres"
server_address = ":8080"
issuer_url = "http://localhost:8080"
[frontend]
origin = "http://localhost:3000"
public_url = "http://localhost:3000"
`)
	_, err := Load(path)
	if err == nil || !strings.Contains(err.Error(), "postgres.dsn") {
		t.Fatalf("Load() error = %v", err)
	}
}

func TestPathUsesExplicitConfig(t *testing.T) {
	t.Setenv("ERIS_CONFIG", "/tmp/eris-test.toml")
	if got := Path(); got != "/tmp/eris-test.toml" {
		t.Fatalf("Path() = %q", got)
	}
}

func writeConfig(t *testing.T, contents string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}
