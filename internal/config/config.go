package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	Server   ServerConfig
	Auth     AuthConfig
	Database DatabaseConfig
	TLS      TLSConfig
	Audit    AuditConfig
	Swagger  SwaggerConfig
}

// ServerConfig holds HTTP server settings
type ServerConfig struct {
	Host string
	Port int
}

// AuthConfig holds authentication settings
type AuthConfig struct {
	GitHubClientID     string        `mapstructure:"github_client_id"`
	GitHubClientSecret string        `mapstructure:"github_client_secret"`
	GitHubOrg          string        `mapstructure:"github_org"`
	JWTSecret          string        `mapstructure:"jwt_secret"`
	SessionMaxAge      time.Duration `mapstructure:"session_max_age"`
}

// DatabaseConfig holds database settings
type DatabaseConfig struct {
	Path string
}

// TLSConfig holds TLS certificate settings
type TLSConfig struct {
	Enabled  bool
	CertFile string `mapstructure:"cert_file"`
	KeyFile  string `mapstructure:"key_file"`
}

// AuditConfig holds audit trail settings
type AuditConfig struct {
	Enabled bool
	LogPath string `mapstructure:"log_path"`
}

// SwaggerConfig holds Swagger UI settings
type SwaggerConfig struct {
	Enabled bool
}

// Load loads configuration from environment variables, config file, and defaults
func Load(cfgFile string) (*Config, error) {
	v := viper.New()

	// Set defaults
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 8080)
	v.SetDefault("auth.session_max_age", 24*time.Hour)
	v.SetDefault("database.path", "./fw-app.db")
	v.SetDefault("tls.enabled", true)
	v.SetDefault("tls.cert_file", "/run/secrets/tls.crt")
	v.SetDefault("tls.key_file", "/run/secrets/tls.key")
	v.SetDefault("audit.enabled", true)
	v.SetDefault("audit.log_path", "/var/log/fw-app/audit.jsonl")
	v.SetDefault("swagger.enabled", false)

	// Environment variables (FW_ prefix)
	v.SetEnvPrefix("FW")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Explicitly bind env vars for nested keys
	_ = v.BindEnv("auth.github_client_id")
	_ = v.BindEnv("auth.github_client_secret")
	_ = v.BindEnv("auth.github_org")
	_ = v.BindEnv("auth.jwt_secret")
	_ = v.BindEnv("auth.session_max_age")
	_ = v.BindEnv("server.host")
	_ = v.BindEnv("server.port")
	_ = v.BindEnv("database.path")
	_ = v.BindEnv("tls.enabled")
	_ = v.BindEnv("tls.cert_file")
	_ = v.BindEnv("tls.key_file")
	_ = v.BindEnv("audit.enabled")
	_ = v.BindEnv("audit.log_path")
	_ = v.BindEnv("swagger.enabled")

	// Config file
	if cfgFile != "" {
		v.SetConfigFile(cfgFile)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath("/etc/fw-app")
		v.AddConfigPath(".")
	}

	// Read config file if it exists (don't error if missing)
	_ = v.ReadInConfig()

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate required fields
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Validate checks that required configuration is present
func (c *Config) Validate() error {
	if c.Auth.GitHubClientID == "" {
		return fmt.Errorf("auth.github_client_id is required (set FW_AUTH_GITHUB_CLIENT_ID)")
	}
	if c.Auth.GitHubClientSecret == "" {
		return fmt.Errorf("auth.github_client_secret is required (set FW_AUTH_GITHUB_CLIENT_SECRET)")
	}
	if c.Auth.GitHubOrg == "" {
		return fmt.Errorf("auth.github_org is required (set FW_AUTH_GITHUB_ORG)")
	}
	if c.Auth.JWTSecret == "" {
		return fmt.Errorf("auth.jwt_secret is required (set FW_AUTH_JWT_SECRET)")
	}
	if c.TLS.Enabled && (c.TLS.CertFile == "" || c.TLS.KeyFile == "") {
		return fmt.Errorf("tls.cert_file and tls.key_file are required when tls.enabled is true")
	}
	return nil
}
