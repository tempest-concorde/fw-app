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

	fqdn := cfg.Server.FQDN

	port := cfg.Server.EffectivePort(cfg.TLS.Enabled)

	client, err := healthClient(cfg, port)
	if err != nil {
		fmt.Fprintf(os.Stderr, "healthcheck: %v\n", err)
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Use the FQDN in the URL solely to drive hostname (ServerName) verification
	// of the served certificate. The dialer below always connects to the
	// loopback publish address, since the container is only exposed on
	// 127.0.0.1:443 -> 8443 (not on the tailnet IP).
	url := fmt.Sprintf("https://%s:%d/health", fqdn, port)

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

// healthClient builds an HTTP client that verifies the served certificate
// against the app's own configured cert/key for the Tailscale FQDN. The leaf
// of the configured cert is used as the trust anchor (direct trust), and the
// TLS handshake additionally verifies that ServerName matches the FQDN (the
// FQDN comes from the request URL host). This is a self-check: it fails if
// the served cert differs from the configured one or is not valid for the
// FQDN.
//
// The dialer always connects to 127.0.0.1:<port> (the loopback publish address)
// even though the URL host is the FQDN — the FQDN only drives ServerName
// verification, since the container is not reachable on its tailnet IP.
func healthClient(cfg *config.Config, port int) (*http.Client, error) {
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
		ForceAttemptHTTP2:   false,
		TLSClientConfig:     tlsConfig,
		TLSHandshakeTimeout: 3 * time.Second,
		DialContext: func(ctx context.Context, network, _ string) (net.Conn, error) {
			// Dial loopback regardless of the URL host (which is the FQDN).
			addr := fmt.Sprintf("127.0.0.1:%d", port)
			var d net.Dialer
			return d.DialContext(ctx, network, addr)
		},
	}

	return &http.Client{Transport: transport}, nil
}
