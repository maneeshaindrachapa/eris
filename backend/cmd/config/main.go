package main

import (
	"fmt"
	"log"
	"os"

	"github.com/maneeshaindrachapa/eris/internal/config"
)

func main() {
	cfg, err := config.Load(config.Path())
	if err != nil {
		log.Fatal(err)
	}
	for _, key := range os.Args[1:] {
		value, ok := valueFor(cfg, key)
		if !ok {
			log.Fatalf("unknown configuration key %q", key)
		}
		fmt.Println(value)
	}
}

func valueFor(cfg config.Config, key string) (any, bool) {
	switch key {
	case "postgres.container":
		return cfg.Postgres.Container, true
	case "postgres.image":
		return cfg.Postgres.Image, true
	case "postgres.user":
		return cfg.Postgres.User, true
	case "postgres.password":
		return cfg.Postgres.Password, true
	case "postgres.database":
		return cfg.Postgres.Database, true
	case "postgres.hostname":
		return cfg.Postgres.Hostname, true
	case "postgres.port":
		return cfg.Postgres.Port, true
	case "postgres.volume":
		return cfg.Postgres.Volume, true
	case "tools.migrate_image":
		return cfg.Tools.MigrateImage, true
	default:
		return nil, false
	}
}
