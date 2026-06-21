# Flight Wall Application - Multi-stage Go Build

# Build stage - Red Hat Hardened Go builder
FROM registry.access.redhat.com/hi/go:latest AS builder

WORKDIR /src

# Copy go modules manifests
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build with CGO disabled (using pure Go modernc.org/sqlite)
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /tmp/fw-app ./cmd/server

# Runtime stage - Red Hat Hardened static (for CGO_ENABLED=0 binaries)
FROM registry.access.redhat.com/hi/static:latest

# Metadata
LABEL org.opencontainers.image.title="Flight Wall Application"
LABEL org.opencontainers.image.description="Go application for Flight Wall LED display - REST API + LED control + embedded UI"
LABEL org.opencontainers.image.source="https://github.com/tempest-concorde/fw-app"
LABEL org.opencontainers.image.licenses="Apache-2.0"

# Copy binary from builder
COPY --from=builder /tmp/fw-app /usr/local/bin/fw-app

# Expose API port
EXPOSE 8080

# Volume for audit logs
VOLUME /var/log/fw-app

# Entrypoint
ENTRYPOINT ["/usr/local/bin/fw-app"]
