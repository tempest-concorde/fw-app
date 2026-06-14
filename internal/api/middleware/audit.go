package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tempest-concorde/fw-app/internal/audit"
)

// Audit returns a Gin middleware that creates audit trail entries for state-changing requests.
// It only audits POST, PUT, and DELETE requests, extracting user information from the context
// (set by the Auth middleware) and writing a CloudEvents audit event.
func Audit(writer *audit.Writer) gin.HandlerFunc {
	return func(c *gin.Context) {
		if writer == nil {
			c.Next()
			return
		}

		// Only audit state-changing methods
		method := c.Request.Method
		if method != http.MethodPost && method != http.MethodPut && method != http.MethodDelete {
			c.Next()
			return
		}

		// Extract user from context (set by auth middleware)
		userValue, exists := c.Get("user")
		username := "unknown"
		if exists {
			if user, ok := userValue.(string); ok {
				username = user
			}
		}

		// Capture request body if present
		var requestBody interface{}
		if c.Request.Body != nil {
			bodyBytes, err := io.ReadAll(c.Request.Body)
			if err == nil && len(bodyBytes) > 0 {
				// Restore the body for the actual handler
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

				// Try to parse as JSON for audit
				var jsonBody map[string]interface{}
				if json.Unmarshal(bodyBytes, &jsonBody) == nil {
					requestBody = jsonBody
				} else {
					requestBody = string(bodyBytes)
				}
			}
		}

		// Process request
		c.Next()

		// Get response status
		status := c.Writer.Status()

		// Build audit event data
		auditData := map[string]interface{}{
			"method": method,
			"path":   c.Request.URL.Path,
			"status": status,
		}

		if requestBody != nil {
			auditData["request"] = requestBody
		}

		// Build extensions
		extensions := map[string]string{
			"fwuser":   username,
			"fwaction": method,
			"fwstatus": http.StatusText(status),
		}

		// Determine event type based on status
		eventType := "com.tempest-concorde.fw-app.api.request"
		if status >= 400 {
			eventType = "com.tempest-concorde.fw-app.api.request.error"
		}

		// Write audit event (don't block request on audit failure)
		if err := writer.WriteEvent(
			context.Background(),
			eventType,
			"fw-app/api",
			c.Request.URL.Path,
			auditData,
			extensions,
		); err != nil {
			// Log error but don't fail the request
			slog.Error("failed to write audit event", "error", err)
		}
	}
}
