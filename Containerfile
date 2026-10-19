# Flight Wall Application - Multi-stage Go Build

# Frontend build stage - React + PatternFly single-page app (ephemeral, not shipped)
# Red Hat certified ubi9/nodejs-20, digest-pinned (T011; fw-gsd
# specs/004-fedora-version-alignment FR-011→FR-012: no hardened `hi/node` base
# exists, so the certified tier is the constitution-compliant substitute).
# UBI nodejs images expect /opt/app-root/src as the app workdir and run as a
# non-root user (uid 1001) by default — no USER root, no chown plumbing.
FROM registry.access.redhat.com/ubi9/nodejs-20:9.7-1778648167@sha256:74cc7b1d13592b1e425074f434b90e470ab209da85fd1fdb8e6e9e4cabaec51a AS webbuild
WORKDIR /opt/app-root/src
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

# Build stage - Red Hat Hardened Go builder
# Pin to Go 1.27 stream tag + SHA256 digest (Go 1.27.0, resolves to `latest`)
FROM registry.access.redhat.com/hi/go:1.27@sha256:4194d758108275e600477b9a72974b63407f6cb98dc26a97329a21f26b12bbb4 AS builder

WORKDIR /src

# Copy go modules manifests
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Embed the built frontend into the Go binary via //go:embed
# Path matches the webbuild stage workdir (/opt/app-root/src) on ubi9/nodejs-20.
COPY --from=webbuild /opt/app-root/src/dist ./web/dist

# Build with CGO disabled (using pure Go modernc.org/sqlite)
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /tmp/fw-app ./cmd/server

# Runtime stage - Red Hat Hardened static (for CGO_ENABLED=0 binaries)
# Pinned to SHA256 digest of the floating `latest` tag
FROM registry.access.redhat.com/hi/static:latest@sha256:f0ee5d3625df237937e7fbf242c8dca3d7c1f5584a059f1410c108b9ff767dc3

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
