package config

import (
	"os"
	"testing"
	"time"
)

func setEnv(t *testing.T, key, value string) {
	t.Helper()
	t.Setenv(key, value)
}

func setRequiredEnvVars(t *testing.T) {
	t.Helper()
	setEnv(t, "FW_AUTH_GITHUB_CLIENT_ID", "test-client-id")
	setEnv(t, "FW_AUTH_GITHUB_CLIENT_SECRET", "test-secret")
	setEnv(t, "FW_AUTH_GITHUB_ORG", "test-org")
	setEnv(t, "FW_AUTH_JWT_SECRET", "test-jwt-secret")
	setEnv(t, "FW_TLS_ENABLED", "false")
}

func TestLoad_EnvVarsPopulateAuthConfig(t *testing.T) {
	setRequiredEnvVars(t)

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if cfg.Auth.GitHubClientID != "test-client-id" {
		t.Errorf("GitHubClientID = %q, want %q", cfg.Auth.GitHubClientID, "test-client-id")
	}
	if cfg.Auth.GitHubClientSecret != "test-secret" {
		t.Errorf("GitHubClientSecret = %q, want %q", cfg.Auth.GitHubClientSecret, "test-secret")
	}
	if cfg.Auth.GitHubOrg != "test-org" {
		t.Errorf("GitHubOrg = %q, want %q", cfg.Auth.GitHubOrg, "test-org")
	}
	if cfg.Auth.JWTSecret != "test-jwt-secret" {
		t.Errorf("JWTSecret = %q, want %q", cfg.Auth.JWTSecret, "test-jwt-secret")
	}
}

func TestLoad_Defaults(t *testing.T) {
	setRequiredEnvVars(t)

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if cfg.Server.Host != "0.0.0.0" {
		t.Errorf("Server.Host = %q, want %q", cfg.Server.Host, "0.0.0.0")
	}
	if cfg.Server.Port != 8080 {
		t.Errorf("Server.Port = %d, want %d", cfg.Server.Port, 8080)
	}
	if cfg.Database.Path != "./fw-app.db" {
		t.Errorf("Database.Path = %q, want %q", cfg.Database.Path, "./fw-app.db")
	}
	if cfg.Auth.SessionMaxAge != 24*time.Hour {
		t.Errorf("Auth.SessionMaxAge = %v, want %v", cfg.Auth.SessionMaxAge, 24*time.Hour)
	}
	if !cfg.Audit.Enabled {
		t.Error("Audit.Enabled should default to true")
	}
	if cfg.Audit.LogPath != "/var/log/fw-app/audit.jsonl" {
		t.Errorf("Audit.LogPath = %q, want %q", cfg.Audit.LogPath, "/var/log/fw-app/audit.jsonl")
	}
}

func TestLoad_TLSEnabledByDefault(t *testing.T) {
	setEnv(t, "FW_AUTH_GITHUB_CLIENT_ID", "id")
	setEnv(t, "FW_AUTH_GITHUB_CLIENT_SECRET", "secret")
	setEnv(t, "FW_AUTH_GITHUB_ORG", "org")
	setEnv(t, "FW_AUTH_JWT_SECRET", "jwt")

	// TLS defaults to enabled, and default cert paths exist but files won't be there.
	// Validation should pass since cert paths are non-empty strings.
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if !cfg.TLS.Enabled {
		t.Error("TLS.Enabled should default to true")
	}
	if cfg.TLS.CertFile != "/run/secrets/tls.crt" {
		t.Errorf("TLS.CertFile = %q, want %q", cfg.TLS.CertFile, "/run/secrets/tls.crt")
	}
	if cfg.TLS.KeyFile != "/run/secrets/tls.key" {
		t.Errorf("TLS.KeyFile = %q, want %q", cfg.TLS.KeyFile, "/run/secrets/tls.key")
	}
}

func TestLoad_TLSDisabledViaEnv(t *testing.T) {
	setRequiredEnvVars(t)

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if cfg.TLS.Enabled {
		t.Error("TLS.Enabled should be false when FW_TLS_ENABLED=false")
	}
}

func TestValidate_TLSEnabledWithEmptyCertFails(t *testing.T) {
	cfg := &Config{
		Auth: AuthConfig{
			GitHubClientID:     "id",
			GitHubClientSecret: "secret",
			GitHubOrg:          "org",
			JWTSecret:          "jwt",
		},
		TLS: TLSConfig{
			Enabled:  true,
			CertFile: "",
			KeyFile:  "",
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected validation error when TLS enabled with empty cert paths")
	}
	if !contains(err.Error(), "cert_file") {
		t.Errorf("error = %q, want it to mention cert_file", err.Error())
	}
}

func TestLoad_EnvOverridesDefaults(t *testing.T) {
	setRequiredEnvVars(t)
	setEnv(t, "FW_SERVER_PORT", "9090")
	setEnv(t, "FW_SERVER_HOST", "127.0.0.1")
	setEnv(t, "FW_DATABASE_PATH", "/tmp/test.db")
	setEnv(t, "FW_AUDIT_LOG_PATH", "/tmp/audit.jsonl")

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if cfg.Server.Port != 9090 {
		t.Errorf("Server.Port = %d, want %d", cfg.Server.Port, 9090)
	}
	if cfg.Server.Host != "127.0.0.1" {
		t.Errorf("Server.Host = %q, want %q", cfg.Server.Host, "127.0.0.1")
	}
	if cfg.Database.Path != "/tmp/test.db" {
		t.Errorf("Database.Path = %q, want %q", cfg.Database.Path, "/tmp/test.db")
	}
	if cfg.Audit.LogPath != "/tmp/audit.jsonl" {
		t.Errorf("Audit.LogPath = %q, want %q", cfg.Audit.LogPath, "/tmp/audit.jsonl")
	}
}

func TestLoad_MissingRequiredFieldsFail(t *testing.T) {
	tests := []struct {
		name    string
		skipVar string
		wantErr string
	}{
		{"missing client ID", "FW_AUTH_GITHUB_CLIENT_ID", "github_client_id"},
		{"missing client secret", "FW_AUTH_GITHUB_CLIENT_SECRET", "github_client_secret"},
		{"missing org", "FW_AUTH_GITHUB_ORG", "github_org"},
		{"missing JWT secret", "FW_AUTH_JWT_SECRET", "jwt_secret"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set all required vars
			setEnv(t, "FW_TLS_ENABLED", "false")
			setEnv(t, "FW_AUTH_GITHUB_CLIENT_ID", "id")
			setEnv(t, "FW_AUTH_GITHUB_CLIENT_SECRET", "secret")
			setEnv(t, "FW_AUTH_GITHUB_ORG", "org")
			setEnv(t, "FW_AUTH_JWT_SECRET", "jwt")

			// Unset the one we're testing
			os.Unsetenv(tt.skipVar)

			_, err := Load("")
			if err == nil {
				t.Fatalf("expected error when %s is missing", tt.skipVar)
			}
			if !contains(err.Error(), tt.wantErr) {
				t.Errorf("error = %q, want it to contain %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestLoad_ConfigFile(t *testing.T) {
	// Write a temporary config file
	content := `
server:
  host: "10.0.0.1"
  port: 3000
auth:
  github_client_id: "file-client-id"
  github_client_secret: "file-secret"
  github_org: "file-org"
  jwt_secret: "file-jwt"
tls:
  enabled: false
`
	tmpFile, err := os.CreateTemp(t.TempDir(), "config-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	cfg, err := Load(tmpFile.Name())
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if cfg.Server.Host != "10.0.0.1" {
		t.Errorf("Server.Host = %q, want %q", cfg.Server.Host, "10.0.0.1")
	}
	if cfg.Server.Port != 3000 {
		t.Errorf("Server.Port = %d, want %d", cfg.Server.Port, 3000)
	}
	if cfg.Auth.GitHubClientID != "file-client-id" {
		t.Errorf("GitHubClientID = %q, want %q", cfg.Auth.GitHubClientID, "file-client-id")
	}
}

func TestLoad_EnvOverridesConfigFile(t *testing.T) {
	content := `
auth:
  github_client_id: "file-id"
  github_client_secret: "file-secret"
  github_org: "file-org"
  jwt_secret: "file-jwt"
tls:
  enabled: false
`
	tmpFile, err := os.CreateTemp(t.TempDir(), "config-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	setEnv(t, "FW_AUTH_GITHUB_CLIENT_ID", "env-id")

	cfg, err := Load(tmpFile.Name())
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if cfg.Auth.GitHubClientID != "env-id" {
		t.Errorf("GitHubClientID = %q, want %q (env should override file)", cfg.Auth.GitHubClientID, "env-id")
	}
	if cfg.Auth.GitHubOrg != "file-org" {
		t.Errorf("GitHubOrg = %q, want %q (should come from file)", cfg.Auth.GitHubOrg, "file-org")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || s != "" && containsImpl(s, substr))
}

func containsImpl(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
