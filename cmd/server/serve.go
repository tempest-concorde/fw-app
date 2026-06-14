package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/tempest-concorde/fw-app/internal/api"
	"github.com/tempest-concorde/fw-app/internal/audit"
	"github.com/tempest-concorde/fw-app/internal/auth"
	"github.com/tempest-concorde/fw-app/internal/config"
	"github.com/tempest-concorde/fw-app/internal/server"
	"github.com/tempest-concorde/fw-app/internal/storage"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the HTTP server",
	Long:  `Start the Flight Wall API HTTP server with authentication, database, and LED control.`,
	Run:   runServe,
}

func runServe(cmd *cobra.Command, args []string) {
	// Load configuration
	cfg, err := config.Load(cfgFile)
	if err != nil {
		slog.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	// Setup structured logging
	var logHandler slog.Handler
	var logLevelVar slog.LevelVar

	switch logLevel {
	case "debug":
		logLevelVar.Set(slog.LevelDebug)
	case "info":
		logLevelVar.Set(slog.LevelInfo)
	case "warn":
		logLevelVar.Set(slog.LevelWarn)
	case "error":
		logLevelVar.Set(slog.LevelError)
	default:
		fmt.Fprintf(os.Stderr, "invalid log level: %s (must be debug, info, warn, or error)\n", logLevel)
		os.Exit(1)
	}

	if logFormat == "json" {
		logHandler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: &logLevelVar,
		})
	} else if logFormat == "text" {
		logHandler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: &logLevelVar,
		})
	} else {
		fmt.Fprintf(os.Stderr, "invalid log format: %s (must be json or text)\n", logFormat)
		os.Exit(1)
	}

	logger := slog.New(logHandler)
	slog.SetDefault(logger)

	logger.Info("starting Flight Wall server",
		"version", "0.1.0",
		"log_level", logLevel,
		"log_format", logFormat,
	)

	// Open database
	db, err := storage.New(storage.Config{
		Path:        cfg.Database.Path,
		Development: logLevel == "debug",
	})
	if err != nil {
		logger.Error("failed to open database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Run migrations
	ctx := context.Background()
	if err := db.Init(ctx); err != nil {
		logger.Error("failed to run database migrations", "error", err)
		os.Exit(1)
	}

	// Create audit writer
	var auditWriter *audit.Writer
	if cfg.Audit.Enabled {
		auditWriter, err = audit.NewWriter(cfg.Audit.LogPath)
		if err != nil {
			logger.Error("failed to create audit writer", "error", err)
			os.Exit(1)
		}
		defer auditWriter.Close()
		logger.Info("audit logging enabled", "path", cfg.Audit.LogPath)
	}

	// Initialize GitHub auth
	githubAuth := auth.NewGitHubAuth(
		cfg.Auth.GitHubClientID,
		cfg.Auth.GitHubClientSecret,
		cfg.Auth.GitHubOrg,
	)

	// Initialize JWT manager
	jwtManager := auth.NewJWTManager(
		cfg.Auth.JWTSecret,
		cfg.Auth.SessionMaxAge,
	)

	// Create router
	router := api.NewRouter(api.RouterConfig{
		DB:             db,
		GitHubAuth:     githubAuth,
		JWTManager:     jwtManager,
		AuditWriter:    auditWriter,
		SwaggerEnabled: cfg.Swagger.Enabled,
	})

	// Create HTTP server
	srv, err := server.NewServer(server.Config{
		Host:        cfg.Server.Host,
		Port:        cfg.Server.Port,
		Handler:     router,
		TLSEnabled:  cfg.TLS.Enabled,
		TLSCertFile: cfg.TLS.CertFile,
		TLSKeyFile:  cfg.TLS.KeyFile,
	})
	if err != nil {
		logger.Error("failed to create server", "error", err)
		os.Exit(1)
	}

	// Start server in goroutine
	serverErrors := make(chan error, 1)
	go func() {
		logger.Info("starting HTTP server",
			"host", cfg.Server.Host,
			"port", cfg.Server.Port,
			"tls_enabled", cfg.TLS.Enabled,
		)
		serverErrors <- srv.Start()
	}()

	// Setup signal handling for graceful shutdown
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Wait for shutdown signal or server error
	select {
	case err := <-serverErrors:
		logger.Error("server error", "error", err)
		os.Exit(1)
	case <-ctx.Done():
		logger.Info("shutdown signal received, starting graceful shutdown")
	}

	// Graceful shutdown with 10 second timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("server shutdown failed", "error", err)
		os.Exit(1)
	}

	logger.Info("server shutdown complete")
}
