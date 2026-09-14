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
FROM registry.access.redhat.com/hi/go:1.27@sha256:666e77358fcda912f3251bc1651dc1908ea1d6e49c0eab43b0aacef272d187dc AS builder

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
FROM registry.access.redhat.com/hi/static:latest@sha256:20f419d12511f96524d9b9bb092ef5066d6bacee7ed45d7c528bef62f6d48f74

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
