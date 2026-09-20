package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadFromEnvFile(t *testing.T) {
	clearConfigEnvironment(t)
	path := writeEnvFile(t, `
STORAGE_BACKEND=postgres
POSTGRES_DSN=postgres://file-value
SERVER_ADDRESS=:9090
ISSUER_URL=https://issuer.example.com
`)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.StorageBackend != "postgres" {
		t.Fatalf("StorageBackend = %q", cfg.StorageBackend)
	}
	if cfg.PostgresDSN != "postgres://file-value" {
		t.Fatalf("PostgresDSN = %q", cfg.PostgresDSN)
	}
	if cfg.ServerAddress != ":9090" {
		t.Fatalf("ServerAddress = %q", cfg.ServerAddress)
	}
	if cfg.IssuerURL != "https://issuer.example.com" {
		t.Fatalf("IssuerURL = %q", cfg.IssuerURL)
	}
}

func TestEnvironmentOverridesEnvFile(t *testing.T) {
	clearConfigEnvironment(t)
	t.Setenv("STORAGE_BACKEND", "memory")
	t.Setenv("SERVER_ADDRESS", ":7070")
	path := writeEnvFile(t, "STORAGE_BACKEND=postgres\nPOSTGRES_DSN=postgres://file-value\n")

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.StorageBackend != "memory" {
		t.Fatalf("StorageBackend = %q", cfg.StorageBackend)
	}
	if cfg.ServerAddress != ":7070" {
		t.Fatalf("ServerAddress = %q", cfg.ServerAddress)
	}
	if cfg.IssuerURL != "http://localhost:8080" {
		t.Fatalf("IssuerURL default = %q", cfg.IssuerURL)
	}
}

func TestLoadRejectsPostgresWithoutDSN(t *testing.T) {
	clearConfigEnvironment(t)
	t.Setenv("POSTGRES_DSN", "")
	path := writeEnvFile(t, "STORAGE_BACKEND=postgres\n")

	_, err := Load(path)
	if err == nil || !strings.Contains(err.Error(), "POSTGRES_DSN is required") {
		t.Fatalf("Load() error = %v", err)
	}
}

func TestLoadRejectsUnknownBackend(t *testing.T) {
	clearConfigEnvironment(t)
	path := writeEnvFile(t, "STORAGE_BACKEND=unknown\n")

	_, err := Load(path)
	if err == nil || !strings.Contains(err.Error(), "unsupported STORAGE_BACKEND") {
		t.Fatalf("Load() error = %v", err)
	}
}

func TestLoadRequiresEnvFile(t *testing.T) {
	clearConfigEnvironment(t)
	_, err := Load(filepath.Join(t.TempDir(), "missing.env"))
	if err == nil || !strings.Contains(err.Error(), "read config file") {
		t.Fatalf("Load() error = %v", err)
	}
}

func clearConfigEnvironment(t *testing.T) {
	t.Helper()
	for _, key := range []string{"STORAGE_BACKEND", "POSTGRES_DSN", "SERVER_ADDRESS", "ISSUER_URL"} {
		value, exists := os.LookupEnv(key)
		if err := os.Unsetenv(key); err != nil {
			t.Fatalf("unset %s: %v", key, err)
		}
		t.Cleanup(func() {
			if exists {
				_ = os.Setenv(key, value)
				return
			}
			_ = os.Unsetenv(key)
		})
	}
}

func writeEnvFile(t *testing.T, contents string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("write env file: %v", err)
	}
	return path
}
