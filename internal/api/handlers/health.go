package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

// HealthHandler handles health check endpoints
type HealthHandler struct {
	db *bun.DB
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(db *bun.DB) *HealthHandler {
	return &HealthHandler{
		db: db,
	}
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status string `json:"status" example:"ok"`
}

// Health godoc
// @Summary Health check
// @Description Returns the current health status of the service
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} HealthResponse
// @Router /health [get]
func (h *HealthHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, HealthResponse{
		Status: "ok",
	})
}

// ReadinessResponse represents the readiness check response
type ReadinessResponse struct {
	Status string `json:"status" example:"ok"`
	DBUp   bool   `json:"db_up" example:"true"`
}

// Readiness godoc
// @Summary Readiness check
// @Description Returns the readiness status including database connectivity
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} ReadinessResponse
// @Failure 503 {object} ReadinessResponse
// @Router /readyz [get]
func (h *HealthHandler) Readiness(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	dbUp := true
	if err := h.db.PingContext(ctx); err != nil {
		dbUp = false
		c.JSON(http.StatusServiceUnavailable, ReadinessResponse{
			Status: "unavailable",
			DBUp:   dbUp,
		})
		return
	}

	c.JSON(http.StatusOK, ReadinessResponse{
		Status: "ok",
		DBUp:   dbUp,
	})
}
