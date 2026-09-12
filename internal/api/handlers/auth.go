package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tempest-concorde/fw-app/internal/auth"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	githubAuth *auth.GitHubAuth
	jwtManager *auth.JWTManager
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(githubAuth *auth.GitHubAuth, jwtManager *auth.JWTManager) *AuthHandler {
	return &AuthHandler{
		githubAuth: githubAuth,
		jwtManager: jwtManager,
	}
}

// Login godoc
// @Summary Initiate GitHub OAuth login
// @Description Starts the GitHub OAuth flow by setting state cookie and redirecting to GitHub
// @Tags auth
// @Accept json
// @Produce json
// @Success 302 {string} string "Redirect to GitHub OAuth"
// @Router /auth/login [get]
func (h *AuthHandler) Login(c *gin.Context) {
	state, url := h.githubAuth.StartLogin()

	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		"oauth_state",
		state,
		600, // 10 minutes
		"/",
		"",
		true, // Always Secure - use HTTPS even in development
		true, // HttpOnly
	)

	c.Redirect(http.StatusFound, url)
}

// CallbackRequest represents the OAuth callback parameters
type CallbackRequest struct {
	Code  string `form:"code" binding:"required"`
	State string `form:"state" binding:"required"`
}

// Callback godoc
// @Summary Handle GitHub OAuth callback
// @Description Validates state, exchanges code for token, generates JWT, and sets session cookie
// @Tags auth
// @Accept json
// @Produce json
// @Param code query string true "OAuth authorization code"
// @Param state query string true "OAuth state parameter"
// @Success 302 {string} string "Redirect to /app"
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /auth/callback [get]
func (h *AuthHandler) Callback(c *gin.Context) {
	var req CallbackRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.Redirect(http.StatusFound, "/login?error=invalid_state")
		return
	}

	storedState, err := c.Cookie("oauth_state")
	if err != nil || storedState != req.State {
		c.Redirect(http.StatusFound, "/login?error=invalid_state")
		return
	}

	user, err := h.githubAuth.HandleCallback(c.Request.Context(), req.Code)
	if err != nil {
		c.Redirect(http.StatusFound, "/login?error="+authErrorCode(err))
		return
	}

	token, err := h.jwtManager.GenerateToken(user.Login, user.Login, user.ID)
	if err != nil {
		c.Redirect(http.StatusFound, "/login?error=auth_failed")
		return
	}

	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie(
		"fw-session",
		token,
		int(24*time.Hour.Seconds()), // 24 hours
		"/",
		"",
		true, // Secure
		true, // HttpOnly
	)

	// Clear state cookie
	c.SetCookie("oauth_state", "", -1, "/", "", true, true)

	c.Redirect(http.StatusFound, "/app")
}

// authErrorCode maps a GitHub auth error to a frontend-facing error code used
// by the /login?error=<code> failure screen.
func authErrorCode(err error) string {
	switch {
	case errors.Is(err, auth.ErrNotMember):
		return "not_member"
	case errors.Is(err, auth.ErrConfig):
		return "config_error"
	default:
		return "auth_failed"
	}
}

// Me godoc
// @Summary Current user
// @Description Returns the authenticated user's identity from the session
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]string
// @Router /auth/me [get]
func (h *AuthHandler) Me(c *gin.Context) {
	cookie, err := c.Cookie("fw-session")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{keyError: "unauthenticated"})
		return
	}

	claims, err := h.jwtManager.ValidateToken(cookie)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{keyError: "unauthenticated"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"login":   claims.Login,
		"user_id": claims.UserID,
	})
}

// Logout godoc
// @Summary Logout user
// @Description Clears the session cookie
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	c.SetCookie(
		"fw-session",
		"",
		-1,
		"/",
		"",
		true,
		true,
	)

	c.JSON(http.StatusOK, gin.H{"message": "logged out"})
}
