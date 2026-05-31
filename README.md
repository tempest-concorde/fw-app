# fw-app - Flight Wall Application

Go application for the Flight Wall LED display system - REST API, LED control, and embedded web UI.

## Features

- **LED Matrix Control** - WS2812B LED matrix via rpi_ws281x library (CGO bindings)
- **Flight Tracking** - Real-time flight data from OpenSky Network, AeroAPI, FlightWall CDN
- **Multiple Display Modes** - Nearby flights, track specific flight, images, text, test patterns
- **REST API** - Full API for remote control and monitoring
- **GitHub OAuth** - Authentication via GitHub organization membership
- **Embedded Web UI** - Svelte-based UI compiled and embedded via `//go:embed`
- **Prometheus Metrics** - `/metrics` endpoint for observability
- **SQLite Storage** - Lightweight database for settings and schedules

## Architecture

```
fw-app/
├── cmd/server/           # Entry point
├── internal/
│   ├── api/              # REST API (Gin framework)
│   ├── led/              # LED renderer + matrix mapping
│   ├── flight/           # Flight data fetching (OpenSky, AeroAPI, FlightWall CDN)
│   ├── storage/          # SQLite database
│   ├── auth/             # GitHub OAuth + org membership check
│   └── config/           # Configuration from env + secrets
├── ui/                   # Svelte web UI (embedded via go:embed)
└── Containerfile         # Multi-stage build (Red Hat Hardened Images)
```

## Display Modes

1. **Nearby Flights** - Fetch ADS-B data, show flight cards (airline, route, aircraft)
2. **Track Flight** - Lock onto specific callsign/flight number
3. **Image** - Upload and display PNG/JPEG images (scaled to LED resolution)
4. **Text** - Scrolling or static text with configurable color
5. **Test Patterns** - Diagnostic patterns (horizontal, vertical, diagonal stripes, RGB, checkerboard)

## LED Configuration

Configurable panel dimensions for testing:

| Phase | Config | Resolution | LEDs |
|-------|--------|-----------|------|
| 1x1 | 16x16, 1x1 | 16x16 | 256 |
| 2x2 | 16x16, 2x2 | 32x32 | 1,024 |
| 2x10 | 16x16, 10x2 | 160x32 | 5,120 |

Set via environment:
```bash
LED_TILE_W=16 LED_TILE_H=16 LED_TILES_X=10 LED_TILES_Y=2
```

## REST API

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/api/flights` | Yes | Current nearby flights |
| GET/POST | `/api/mode` | Yes | Get/set display mode |
| GET/POST | `/api/settings` | Yes | Settings (brightness, location, schedule) |
| GET/POST/DELETE | `/api/schedule[/:id]` | Yes | On/off schedules |
| POST | `/api/image` | Yes | Upload image (max 5MB) |
| POST | `/api/text` | Yes | Set text message |
| POST | `/api/test-pattern` | Yes | Trigger test pattern |
| GET | `/api/preview` | Yes | Current LED state as PNG |
| GET | `/metrics` | No | Prometheus metrics |
| GET | `/health` | No | Health check |

## Authentication

- **GitHub OAuth** - Organization membership check (`tempest-concorde`)
- **JWT Sessions** - 24h TTL, httpOnly/Secure/SameSite=Strict cookies
- **Tailnet-only** - Served exclusively on Tailscale VPN (no public exposure)

## Development

### Prerequisites

- Go 1.23+
- For LED testing: Raspberry Pi 4 with /dev/gpiomem access
- For local dev without hardware: LED renderer has simulator mode

### Build

```bash
make build
```

### Test

```bash
make test           # Unit tests
make test-integration  # Integration tests (requires hardware or simulator)
```

### Run Locally

```bash
make run
```

### Linting

```bash
make lint
make lint-fix
```

## Container Image

Built with **Red Hat Hardened Images**:
- Build stage: `registry.access.redhat.com/hi/go:latest`
- Runtime stage: `registry.access.redhat.com/hi/core-runtime:latest`

Published to: `ghcr.io/tempest-concorde/fw-app:latest`

## CI/CD

Uses reusable workflows from `tempest-concorde/fw-cicd`:

- **PR**: Unit tests, linting, coverage
- **Main push**: Semantic release → version tag
- **Tag push**: Build multi-arch, sign with cosign, SLSA attestation, push to GHCR
- **Nightly**: Extended integration tests

## License

Apache License 2.0 - see [LICENSE](LICENSE)

## Related Repositories

- [fw-os](https://github.com/tempest-concorde/fw-os) - OS layer (quadlet units, cert renewal)
- [fedora-bootc-pi](https://github.com/tempest-concorde/fedora-bootc-pi) - Platform base
- [fw-cicd](https://github.com/tempest-concorde/fw-cicd) - Shared CI/CD workflows
