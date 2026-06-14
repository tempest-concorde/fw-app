package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tempest-concorde/fw-app/internal/auth"
)

// Auth returns a Gin middleware that validates JWT tokens from the "fw-session" cookie.
// It sets the authenticated user claims in the Gin context and returns 401 Unauthorized
// if the token is missing or invalid.
func Auth(jwtMgr *auth.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract cookie
		cookie, err := c.Cookie("fw-session")
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "missing authentication cookie",
			})
			c.Abort()
			return
		}

		// Validate token
		claims, err := jwtMgr.ValidateToken(cookie)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid authentication token",
			})
			c.Abort()
			return
		}

		// Store claims in context for handlers to use
		c.Set("user", claims.Subject)
		c.Set("login", claims.Login)
		c.Set("user_id", claims.UserID)
		c.Next()
	}
}
