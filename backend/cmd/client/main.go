package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"github.com/maneeshaindrachapa/eris/internal/config"
)

type stringList []string

func (values *stringList) String() string { return strings.Join(*values, ",") }

func (values *stringList) Set(value string) error {
	for _, item := range strings.Split(value, ",") {
		if item = strings.TrimSpace(item); item != "" {
			*values = append(*values, item)
		}
	}
	return nil
}

type output struct {
	ClientID          string   `json:"client_id"`
	ClientSecret      string   `json:"client_secret"`
	Name              string   `json:"name"`
	Type              string   `json:"type"`
	RedirectURIs      []string `json:"redirect_uris"`
	AllowedScopes     []string `json:"allowed_scopes"`
	AllowedGrantTypes []string `json:"allowed_grant_types"`
}

func main() {
	log.SetFlags(0)
	var redirectURIs, scopes, grantTypes stringList
	name := flag.String("name", "", "client display name (required)")
	clientID := flag.String("id", "", "client ID (generated when omitted)")
	flag.Var(&redirectURIs, "redirect-uri", "allowed redirect URI; repeat or use comma-separated values")
	flag.Var(&scopes, "scope", "allowed scope; repeat or use comma-separated values")
	flag.Var(&grantTypes, "grant-type", "allowed grant type; repeat or use comma-separated values")
	flag.Parse()
	if stdinInfo, err := os.Stdin.Stat(); err == nil && stdinInfo.Mode()&os.ModeCharDevice != 0 {
		if err := promptForMissing(os.Stdin, os.Stdout, name, clientID, &redirectURIs, &scopes, &grantTypes); err != nil {
			log.Fatalf("create-client: read input: %v", err)
		}
	}

	if strings.TrimSpace(*name) == "" {
		log.Fatal("create-client: client name is required (use -name or run interactively)")
	}
	if len(redirectURIs) == 0 {
		log.Fatal("create-client: at least one redirect URI is required (use -redirect-uri or run interactively)")
	}
	for _, redirectURI := range redirectURIs {
		if err := validateRedirectURI(redirectURI); err != nil {
			log.Fatalf("create-client: %v", err)
		}
	}
	if len(scopes) == 0 {
		scopes = stringList{"openid", "profile", "email"}
	}
	if len(grantTypes) == 0 {
		grantTypes = stringList{"authorization_code", "refresh_token"}
	}
	if *clientID == "" {
		*clientID = "client-" + randomHex(12)
	}

	clientSecret := randomSecret(32)
	secretHash, err := bcrypt.GenerateFromPassword([]byte(clientSecret), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("create-client: hash secret: %v", err)
	}
	cfg, err := config.Load(config.Path())
	if err != nil {
		log.Fatalf("create-client: config: %v", err)
	}
	if cfg.Postgres.DSN == "" {
		log.Fatal("create-client: postgres.dsn is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, cfg.Postgres.DSN)
	if err != nil {
		log.Fatalf("create-client: postgres pool: %v", err)
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("create-client: postgres connection: %v", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO clients (id, name, type, client_secret_hash, redirect_uris, allowed_scopes, allowed_grant_types)
		VALUES ($1, $2, 'confidential', $3, $4, $5, $6)`,
		*clientID, strings.TrimSpace(*name), string(secretHash), []string(redirectURIs), []string(scopes), []string(grantTypes))
	if err != nil {
		var postgresErr *pgconn.PgError
		if errors.As(err, &postgresErr) && postgresErr.Code == "23505" {
			log.Fatalf("create-client: client ID %q already exists", *clientID)
		}
		log.Fatalf("create-client: insert client: %v", err)
	}

	result := output{
		ClientID: *clientID, ClientSecret: clientSecret, Name: strings.TrimSpace(*name), Type: "confidential",
		RedirectURIs: redirectURIs, AllowedScopes: scopes, AllowedGrantTypes: grantTypes,
	}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(result); err != nil {
		log.Fatalf("create-client: encode result: %v", err)
	}
}

func promptForMissing(
	input io.Reader,
	output io.Writer,
	name, clientID *string,
	redirectURIs, scopes, grantTypes *stringList,
) error {
	reader := bufio.NewReader(input)
	var err error
	if strings.TrimSpace(*name) == "" {
		*name, err = prompt(reader, output, "Client name", "")
		if err != nil {
			return err
		}
	}
	if *clientID == "" {
		*clientID, err = prompt(reader, output, "Client ID (blank generates one)", "")
		if err != nil {
			return err
		}
	}
	if len(*redirectURIs) == 0 {
		value, readErr := prompt(reader, output, "Redirect URIs (comma-separated)", "")
		if readErr != nil {
			return readErr
		}
		_ = redirectURIs.Set(value)
	}
	if len(*scopes) == 0 {
		value, readErr := prompt(reader, output, "Scopes", "openid,profile,email")
		if readErr != nil {
			return readErr
		}
		_ = scopes.Set(value)
	}
	if len(*grantTypes) == 0 {
		value, readErr := prompt(reader, output, "Grant types", "authorization_code,refresh_token")
		if readErr != nil {
			return readErr
		}
		_ = grantTypes.Set(value)
	}
	return nil
}

func prompt(reader *bufio.Reader, output io.Writer, label, defaultValue string) (string, error) {
	if defaultValue == "" {
		_, _ = fmt.Fprintf(output, "%s: ", label)
	} else {
		_, _ = fmt.Fprintf(output, "%s [%s]: ", label, defaultValue)
	}
	value, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return defaultValue, nil
	}
	return value, nil
}

func validateRedirectURI(value string) error {
	parsed, err := url.ParseRequestURI(value)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return fmt.Errorf("invalid redirect URI %q: use an absolute http or https URI", value)
	}
	return nil
}

func randomHex(size int) string {
	value := make([]byte, size)
	if _, err := rand.Read(value); err != nil {
		log.Fatalf("create-client: generate client ID: %v", err)
	}
	return hex.EncodeToString(value)
}

func randomSecret(size int) string {
	value := make([]byte, size)
	if _, err := rand.Read(value); err != nil {
		log.Fatalf("create-client: generate client secret: %v", err)
	}
	return base64.RawURLEncoding.EncodeToString(value)
}
