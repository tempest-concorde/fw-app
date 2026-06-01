package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"
)

// Config holds server configuration
type Config struct {
	Host        string
	Port        int
	Handler     http.Handler
	TLSEnabled  bool
	TLSCertFile string
	TLSKeyFile  string
}

// Server wraps http.Server with TLS hot-reload capabilities.
type Server struct {
	httpServer *http.Server
	tlsEnabled bool
	certFile   string
	keyFile    string
	cert       atomic.Pointer[tls.Certificate]
}

// NewServer creates a new Server instance from the provided config.
// If TLS is enabled, TLSCertFile and TLSKeyFile must be non-empty.
func NewServer(cfg Config) (*Server, error) {
	if cfg.TLSEnabled && (cfg.TLSCertFile == "" || cfg.TLSKeyFile == "") {
		return nil, fmt.Errorf("TLS is enabled but cert_file or key_file is empty")
	}

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	s := &Server{
		httpServer: &http.Server{
			Addr:              addr,
			Handler:           cfg.Handler,
			ReadHeaderTimeout: 10 * time.Second,
			ReadTimeout:       30 * time.Second,
			WriteTimeout:      30 * time.Second,
			IdleTimeout:       120 * time.Second,
		},
		tlsEnabled: cfg.TLSEnabled,
		certFile:   cfg.TLSCertFile,
		keyFile:    cfg.TLSKeyFile,
	}

	if cfg.TLSEnabled {
		if err := s.loadCertificate(); err != nil {
			return nil, fmt.Errorf("failed to load initial certificate: %w", err)
		}

		s.httpServer.TLSConfig = &tls.Config{
			GetCertificate: s.getCertificate,
			MinVersion:     tls.VersionTLS12,
		}
	}

	return s, nil
}

// Start begins serving HTTP or HTTPS requests.
func (s *Server) Start() error {
	if s.tlsEnabled {
		// Use empty strings for cert/key since we handle certificates via GetCertificate
		return s.httpServer.ListenAndServeTLS("", "")
	}
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully shuts down the server.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

// ReloadCert reloads the TLS certificate from disk and atomically swaps it.
func (s *Server) ReloadCert() error {
	if !s.tlsEnabled {
		return fmt.Errorf("TLS is not enabled")
	}
	return s.loadCertificate()
}

// loadCertificate loads the certificate from disk and stores it atomically.
func (s *Server) loadCertificate() error {
	cert, err := tls.LoadX509KeyPair(s.certFile, s.keyFile)
	if err != nil {
		return fmt.Errorf("failed to load certificate: %w", err)
	}
	s.cert.Store(&cert)
	return nil
}

// getCertificate is the callback used by tls.Config to retrieve the current certificate.
func (s *Server) getCertificate(_ *tls.ClientHelloInfo) (*tls.Certificate, error) {
	cert := s.cert.Load()
	if cert == nil {
		return nil, fmt.Errorf("no certificate loaded")
	}
	return cert, nil
}
