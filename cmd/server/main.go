package main

import (
	"os"

	"github.com/spf13/cobra"
)

// @title Flight Wall API
// @version 1.0
// @description REST API for Flight Wall LED display system
// @host localhost:8080
// @BasePath /api

// @tag.name auth
// @tag.description Authentication endpoints (GitHub OAuth, JWT sessions)

// @tag.name flights
// @tag.description Flight data retrieval and tracking

// @tag.name settings
// @tag.description System configuration and preferences

// @tag.name health
// @tag.description Health and readiness checks

var (
	cfgFile   string
	logLevel  string
	logFormat string
)

var rootCmd = &cobra.Command{
	Use:   "fw-app",
	Short: "Flight Wall Application Server",
	Long:  `REST API server for Flight Wall LED display system with GitHub OAuth, flight tracking, and embedded web UI.`,
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: /etc/fw-app/config.yaml or ./config.yaml)")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().StringVar(&logFormat, "log-format", "json", "log format (json, text)")

	rootCmd.AddCommand(serveCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
