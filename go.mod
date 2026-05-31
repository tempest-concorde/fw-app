module github.com/tempest-concorde/fw-app

go 1.23

require (
	// Web framework
	github.com/gin-gonic/gin v1.10.0

	// LED control
	github.com/rpi-ws281x/rpi-ws281x-go v1.0.8

	// Database
	github.com/mattn/go-sqlite3 v1.14.24

	// Auth
	github.com/golang-jwt/jwt/v5 v5.2.1

	// Image processing
	github.com/fogleman/gg v1.3.0
	golang.org/x/image v0.23.0

	// HTTP client
	golang.org/x/oauth2 v0.24.0

	// Prometheus metrics
	github.com/prometheus/client_golang v1.20.5

	// Testing
	github.com/stretchr/testify v1.10.0
)
