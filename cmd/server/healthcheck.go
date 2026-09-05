package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/tempest-concorde/fw-app/internal/config"
)

var healthcheckCmd = &cobra.Command{
	Use:   "healthcheck",
	Short: "Probe the local /health endpoint (used by container HEALTHCHECK)",
	Long:  `Performs an authenticated TLS GET against https://<fqdn>:<port>/health and exits 0 if healthy, 1 otherwise.`,
	RunE:  runHealthcheck,
}

func runHealthcheck(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadHealthcheck(cfgFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "healthcheck: config error: %v\n", err)
		return err
	}

	fqdn, err := readFQDN(cfg.Server.FQDNMetaFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "healthcheck: %v\n", err)
		return err
	}

	port := cfg.Server.EffectivePort(cfg.TLS.Enabled)
	url := fmt.Sprintf("https://%s:%d/health", fqdn, port)

	client, err := healthClient(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "healthcheck: %v\n", err)
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "healthcheck: build request: %v\n", err)
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "healthcheck: %v\n", err)
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))

	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "healthcheck: status %d: %s\n", resp.StatusCode, strings.TrimSpace(string(body)))
		return fmt.Errorf("unhealthy status %d", resp.StatusCode)
	}

	var h struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal(body, &h); err != nil || h.Status != "ok" {
		fmt.Fprintf(os.Stderr, "healthcheck: unexpected response: %s\n", strings.TrimSpace(string(body)))
		return fmt.Errorf("health response not ok")
	}

	fmt.Println("healthcheck: ok")
	return nil
}

// readFQDN reads the Tailscale DNSName from the meta file written by the host.
func readFQDN(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read FQDN meta file %s: %w", path, err)
	}
	fqdn := strings.TrimSpace(string(data))
	if fqdn == "" {
		return "", fmt.Errorf("FQDN meta file %s is empty", path)
	}
	return fqdn, nil
}

// healthClient builds an HTTP client that verifies the served certificate
// against the app's own configured cert/key for the Tailscale FQDN. The leaf
// of the configured cert is used as the trust anchor (direct trust), and the
// TLS handshake additionally verifies that ServerName matches the FQDN. This
// is a self-check: it fails if the served cert differs from the configured
// one or is not valid for the FQDN.
func healthClient(cfg *config.Config) (*http.Client, error) {
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	if cfg.TLS.Enabled {
		cert, err := tls.LoadX509KeyPair(cfg.TLS.CertFile, cfg.TLS.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("load cert/key: %w", err)
		}
		pool := x509.NewCertPool()
		for _, der := range cert.Certificate {
			if c, err := x509.ParseCertificate(der); err == nil {
				pool.AddCert(c)
			}
		}
		tlsConfig.RootCAs = pool
	}

	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
		DialContext: (&net.Dialer{
			Timeout: 3 * time.Second,
		}).DialContext,
	}

	return &http.Client{Transport: transport}, nil
}
