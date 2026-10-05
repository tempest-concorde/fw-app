# Flight Wall Application - Multi-stage Go Build

# Frontend build stage - React + PatternFly single-page app (ephemeral, not shipped)
FROM docker.io/library/node:26-alpine AS webbuild
WORKDIR /web
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

# Build stage - Red Hat Hardened Go builder
# Pin to Go 1.27 stream tag + SHA256 digest (Go 1.27.0, resolves to `latest`)
FROM registry.access.redhat.com/hi/go:1.27@sha256:1973bd3bd3c7d3d875c45683ddfe03144599437a29b783f4c8311b6480d5a059 AS builder

WORKDIR /src

# Copy go modules manifests
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Embed the built frontend into the Go binary via //go:embed
COPY --from=webbuild /web/dist ./web/dist

# Build with CGO disabled (using pure Go modernc.org/sqlite)
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /tmp/fw-app ./cmd/server

# Runtime stage - Red Hat Hardened static (for CGO_ENABLED=0 binaries)
# Pinned to SHA256 digest of the floating `latest` tag
FROM registry.access.redhat.com/hi/static:latest@sha256:08d039e8b4f70c0b22118acfff9ada93fc7f9d349e5c5ac991fbdf6b876eea91

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
