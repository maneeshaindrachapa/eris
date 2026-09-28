package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestStringListAcceptsRepeatedAndCommaSeparatedValues(t *testing.T) {
	var values stringList
	if err := values.Set("openid, profile"); err != nil {
		t.Fatalf("Set() error = %v", err)
	}
	if err := values.Set("email"); err != nil {
		t.Fatalf("Set() error = %v", err)
	}
	if got, want := values.String(), "openid,profile,email"; got != want {
		t.Fatalf("String() = %q, want %q", got, want)
	}
}

func TestPromptForMissingFields(t *testing.T) {
	input := strings.NewReader("Example app\nexample-client\nhttps://example.com/callback,http://localhost:3000/callback\n\nauthorization_code\n")
	var output bytes.Buffer
	var name, clientID string
	var redirectURIs, scopes, grantTypes stringList

	err := promptForMissing(input, &output, &name, &clientID, &redirectURIs, &scopes, &grantTypes)
	if err != nil {
		t.Fatalf("promptForMissing() error = %v", err)
	}
	if name != "Example app" || clientID != "example-client" {
		t.Fatalf("name/clientID = %q/%q", name, clientID)
	}
	if got := redirectURIs.String(); got != "https://example.com/callback,http://localhost:3000/callback" {
		t.Fatalf("redirect URIs = %q", got)
	}
	if got := scopes.String(); got != "openid,profile,email" {
		t.Fatalf("default scopes = %q", got)
	}
	if got := grantTypes.String(); got != "authorization_code" {
		t.Fatalf("grant types = %q", got)
	}
	for _, label := range []string{"Client name", "Client ID", "Redirect URIs", "Scopes", "Grant types"} {
		if !strings.Contains(output.String(), label) {
			t.Fatalf("prompt output missing %q: %q", label, output.String())
		}
	}
}

func TestPromptForMissingSkipsProvidedFields(t *testing.T) {
	input := strings.NewReader("\n\n")
	var output bytes.Buffer
	name, clientID := "Existing app", "existing-client"
	redirectURIs := stringList{"https://example.com/callback"}
	var scopes, grantTypes stringList

	err := promptForMissing(input, &output, &name, &clientID, &redirectURIs, &scopes, &grantTypes)
	if err != nil {
		t.Fatalf("promptForMissing() error = %v", err)
	}
	if strings.Contains(output.String(), "Client name") || strings.Contains(output.String(), "Redirect URIs") {
		t.Fatalf("prompted for provided fields: %q", output.String())
	}
	if scopes.String() != "openid,profile,email" || grantTypes.String() != "authorization_code,refresh_token" {
		t.Fatalf("defaults = scopes %q, grants %q", scopes.String(), grantTypes.String())
	}
}

func TestValidateRedirectURI(t *testing.T) {
	for _, value := range []string{"http://localhost:3000/callback", "https://app.example.com/oauth/callback"} {
		if err := validateRedirectURI(value); err != nil {
			t.Fatalf("validateRedirectURI(%q) error = %v", value, err)
		}
	}
	for _, value := range []string{"", "/callback", "javascript:alert(1)", "ftp://example.com/callback"} {
		if err := validateRedirectURI(value); err == nil {
			t.Fatalf("validateRedirectURI(%q) error = nil", value)
		}
	}
}

func TestGeneratedIdentifiersUseExpectedEncoding(t *testing.T) {
	if value := randomHex(12); len(value) != 24 {
		t.Fatalf("randomHex() length = %d", len(value))
	}
	if value := randomSecret(32); len(value) != 43 || strings.ContainsAny(value, "+/=") {
		t.Fatalf("randomSecret() = %q", value)
	}
}
