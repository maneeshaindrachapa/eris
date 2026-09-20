package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	eriscrypto "github.com/maneeshaindrachapa/eris/internal/adapter/crypto"
	httpadapter "github.com/maneeshaindrachapa/eris/internal/adapter/http"
	"github.com/maneeshaindrachapa/eris/internal/adapter/repository/memory"
	"github.com/maneeshaindrachapa/eris/internal/adapter/repository/postgres"
	"github.com/maneeshaindrachapa/eris/internal/application"
	"github.com/maneeshaindrachapa/eris/internal/config"
	"github.com/maneeshaindrachapa/eris/internal/domain"
)

func main() {
	cfg, err := config.Load(".env")
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// ---- driven adapters -----
	var users domain.UserRepository
	var clients domain.ClientRepository
	var codes domain.AuthCodeStore
	var refresh domain.RefreshTokenStore
	var sessions domain.SessionStore
	var consents domain.ConsentStore
	var keystore domain.KeyStore

	switch cfg.StorageBackend {
	case "memory":
		users = memory.NewUserRepository()
		clients = memory.NewClientRepository()
		codes = memory.NewAuthCodeStore()
		refresh = memory.NewRefreshTokenStore()
		sessions = memory.NewSessionStore()
		consents = memory.NewConsentStore()
		keystore = eriscrypto.NewMemoryKeyStore()
		log.Println("using in-memory storage")
	case "postgres":
		pool, err := pgxpool.New(ctx, cfg.PostgresDSN)
		if err != nil {
			log.Fatalf("postgres pool: %v", err)
		}
		defer pool.Close()

		pingCtx, cancelPing := context.WithTimeout(ctx, 5*time.Second)
		defer cancelPing()
		if err := pool.Ping(pingCtx); err != nil {
			log.Fatalf("postgres connection: %v", err)
		}

		users = postgres.NewUserRepository(pool)
		clients = postgres.NewClientRepository(pool)
		codes = postgres.NewAuthCodeStore(pool)
		refresh = postgres.NewRefreshTokenStore(pool)
		sessions = postgres.NewSessionStore(pool)
		consents = postgres.NewConsentStore(pool)
		keystore = eriscrypto.NewPostgresKeyStore(pool)
		log.Println("connected to postgres")
	}

	issuer, err := eriscrypto.NewJWTIssuer(ctx, keystore, cfg.IssuerURL)
	if err != nil {
		log.Fatalf("token issuer init: %v", err)
	}

	// ---- application layer, wired against those interfaces ----
	authSvc := application.NewAuthorizeService(clients, users, codes, sessions, consents)
	tokenSvc := application.NewTokenService(clients, codes, refresh, issuer)

	// ---- driving adapter ----
	e := httpadapter.NewServer(authSvc, tokenSvc)
	go func() {
		log.Printf("eris listening on %s", cfg.ServerAddress)
		if err := e.Start(cfg.ServerAddress); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = e.Shutdown(shutdownCtx)
}
