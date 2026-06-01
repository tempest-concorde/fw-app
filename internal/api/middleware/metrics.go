package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// httpRequestDuration tracks HTTP request duration in seconds
	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path_template", "status"},
	)

	// httpRequestsTotal tracks total HTTP requests
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path_template", "status"},
	)
)

// Metrics returns a Gin middleware that collects Prometheus metrics for HTTP requests.
// It tracks request duration and total request count, labeled by method, path template, and status.
func Metrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(start).Seconds()

		// Get status and path template
		status := strconv.Itoa(c.Writer.Status())
		method := c.Request.Method
		pathTemplate := c.FullPath()

		// Use actual path if no route matched (404s, etc.)
		if pathTemplate == "" {
			pathTemplate = c.Request.URL.Path
		}

		// Record metrics
		httpRequestDuration.WithLabelValues(method, pathTemplate, status).Observe(duration)
		httpRequestsTotal.WithLabelValues(method, pathTemplate, status).Inc()
	}
}
