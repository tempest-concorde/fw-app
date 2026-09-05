# Flight Wall Application - Multi-stage Go Build

# Build stage - Red Hat Hardened Go builder
# Pin to Go 1.27 stream tag + SHA256 digest (Go 1.27.0, resolves to `latest`)
FROM registry.access.redhat.com/hi/go:1.27@sha256:7f767edb96945cef41fdd678e67bb014ba00a9d19ceb2d9e56b33d4ae9a33b45 AS builder

WORKDIR /src

# Copy go modules manifests
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build with CGO disabled (using pure Go modernc.org/sqlite)
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /tmp/fw-app ./cmd/server

# Runtime stage - Red Hat Hardened static (for CGO_ENABLED=0 binaries)
# Pinned to SHA256 digest of the floating `latest` tag
FROM registry.access.redhat.com/hi/static:latest@sha256:41595122bb70793cd58c9e22f625b5c557e4459c43235cbca5c117d057a11424

# Metadata
LABEL org.opencontainers.image.title="Flight Wall Application"
LABEL org.opencontainers.image.description="Go application for Flight Wall LED display - REST API + LED control + embedded UI"
LABEL org.opencontainers.image.source="https://github.com/tempest-concorde/fw-app"
LABEL org.opencontainers.image.licenses="Apache-2.0"
LABEL org.opencontainers.image.vendor="tempest-concorde"

# Copy binary from builder
COPY --from=builder /tmp/fw-app /usr/local/bin/fw-app

# Expose API port
EXPOSE 8080

# Volume for audit logs
VOLUME /var/log/fw-app

# Health check — probes https://<fqdn>:8443/health with cert verification
HEALTHCHECK --interval=30s --timeout=10s --start-period=15s --retries=3 \
  CMD ["/usr/local/bin/fw-app", "healthcheck"]

# Entrypoint
ENTRYPOINT ["/usr/local/bin/fw-app"]
