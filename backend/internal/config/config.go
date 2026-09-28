package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/maneeshaindrachapa/zenith"
)

type Config struct {
	Environment string `zenith:"environment"`
	Backend     struct {
		StorageBackend string `zenith:"storage_backend"`
		ServerAddress  string `zenith:"server_address"`
		IssuerURL      string `zenith:"issuer_url"`
	} `zenith:"backend"`
	Frontend struct {
		Origin      string `zenith:"origin"`
		PublicURL   string `zenith:"public_url"`
		DefaultPath string `zenith:"default_callback_path"`
	} `zenith:"frontend"`
	Postgres struct {
		DSN       string `zenith:"dsn"`
		Container string `zenith:"container"`
		Image     string `zenith:"image"`
		User      string `zenith:"user"`
		Password  string `zenith:"password"`
		Database  string `zenith:"database"`
		Hostname  string `zenith:"hostname"`
		Port      int    `zenith:"port"`
		Volume    string `zenith:"volume"`
	} `zenith:"postgres"`
	Tools struct {
		MigrateImage string `zenith:"migrate_image"`
	} `zenith:"tools"`
}

func Path() string {
	if path := os.Getenv("ERIS_CONFIG"); path != "" {
		return path
	}
	for _, candidate := range []string{"config.toml", filepath.Join("..", "config.toml")} {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	return "config.toml"
}

func Load(path string) (Config, error) {
	var cfg Config
	if err := zenith.Load(path, &cfg, zenith.WithStrictMapping()); err != nil {
		return Config{}, fmt.Errorf("load config file %q: %w", path, err)
	}
	return cfg, nil
}

func (c Config) Validate() error {
	switch c.Backend.StorageBackend {
	case "memory":
	case "postgres":
		if c.Postgres.DSN == "" {
			return fmt.Errorf("postgres.dsn is required when backend.storage_backend is postgres")
		}
	default:
		return fmt.Errorf("unsupported backend.storage_backend %q: use memory or postgres", c.Backend.StorageBackend)
	}
	if c.Backend.ServerAddress == "" || c.Backend.IssuerURL == "" {
		return fmt.Errorf("backend.server_address and backend.issuer_url are required")
	}
	if c.Frontend.Origin == "" || c.Frontend.PublicURL == "" {
		return fmt.Errorf("frontend.origin and frontend.public_url are required")
	}
	return nil
}
