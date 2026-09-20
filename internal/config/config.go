package config

import (
	"fmt"
	"os"

	"github.com/maneeshaindrachapa/zenith"
)

type Config struct {
	StorageBackend string `env:"STORAGE_BACKEND" default:"postgres"`
	PostgresDSN    string `env:"POSTGRES_DSN"`
	ServerAddress  string `env:"SERVER_ADDRESS" default:":8080"`
	IssuerURL      string `env:"ISSUER_URL" default:"http://localhost:8080"`
}

func Load(path string) (Config, error) {
	type fileConfig Config
	var cfg Config
	if err := zenith.Load(path, (*fileConfig)(&cfg)); err != nil {
		return Config{}, fmt.Errorf("read config file %q: %w", path, err)
	}

	overrides := []struct {
		key    string
		target *string
	}{
		{key: "STORAGE_BACKEND", target: &cfg.StorageBackend},
		{key: "POSTGRES_DSN", target: &cfg.PostgresDSN},
		{key: "SERVER_ADDRESS", target: &cfg.ServerAddress},
		{key: "ISSUER_URL", target: &cfg.IssuerURL},
	}
	for _, override := range overrides {
		if value, ok := os.LookupEnv(override.key); ok {
			*override.target = value
		}
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (c Config) Validate() error {
	switch c.StorageBackend {
	case "memory":
	case "postgres":
		if c.PostgresDSN == "" {
			return fmt.Errorf("POSTGRES_DSN is required when STORAGE_BACKEND=postgres")
		}
	default:
		return fmt.Errorf("unsupported STORAGE_BACKEND %q: use memory or postgres", c.StorageBackend)
	}

	if c.ServerAddress == "" {
		return fmt.Errorf("SERVER_ADDRESS is required")
	}
	if c.IssuerURL == "" {
		return fmt.Errorf("ISSUER_URL is required")
	}
	return nil
}
