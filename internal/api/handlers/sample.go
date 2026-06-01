package handlers

import (
	"context"
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tempest-concorde/fw-app/internal/storage/models"
	"github.com/uptrace/bun"
)

// SampleHandler handles CRUD operations for samples
type SampleHandler struct {
	db *bun.DB
}

// NewSampleHandler creates a new sample handler
func NewSampleHandler(db *bun.DB) *SampleHandler {
	return &SampleHandler{
		db: db,
	}
}

// CreateSampleRequest represents the request body for creating a sample
type CreateSampleRequest struct {
	Name        string `json:"name" binding:"required" example:"Sample 1"`
	Description string `json:"description" example:"This is a sample"`
}

// UpdateSampleRequest represents the request body for updating a sample
type UpdateSampleRequest struct {
	Name        string `json:"name" example:"Updated Sample"`
	Description string `json:"description" example:"Updated description"`
}

// ListSamples godoc
// @Summary List all samples
// @Description Returns a list of all samples
// @Tags samples
// @Accept json
// @Produce json
// @Security CookieAuth
// @Success 200 {array} models.Sample
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/samples [get]
func (h *SampleHandler) ListSamples(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var samples []models.Sample
	if err := h.db.NewSelect().Model(&samples).Scan(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch samples"})
		return
	}

	c.JSON(http.StatusOK, samples)
}

// CreateSample godoc
// @Summary Create a new sample
// @Description Creates a new sample with the provided data
// @Tags samples
// @Accept json
// @Produce json
// @Security CookieAuth
// @Param sample body CreateSampleRequest true "Sample data"
// @Success 201 {object} models.Sample
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/samples [post]
func (h *SampleHandler) CreateSample(c *gin.Context) {
	var req CreateSampleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	sample := &models.Sample{
		Name:        req.Name,
		Description: req.Description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if _, err := h.db.NewInsert().Model(sample).Exec(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create sample"})
		return
	}

	c.JSON(http.StatusCreated, sample)
}

// GetSample godoc
// @Summary Get sample by ID
// @Description Returns a single sample by ID
// @Tags samples
// @Accept json
// @Produce json
// @Security CookieAuth
// @Param id path string true "Sample ID"
// @Success 200 {object} models.Sample
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/samples/{id} [get]
func (h *SampleHandler) GetSample(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	sample := new(models.Sample)
	if err := h.db.NewSelect().Model(sample).Where("id = ?", id).Scan(ctx); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "sample not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch sample"})
		return
	}

	c.JSON(http.StatusOK, sample)
}

// UpdateSample godoc
// @Summary Update sample by ID
// @Description Updates an existing sample by ID
// @Tags samples
// @Accept json
// @Produce json
// @Security CookieAuth
// @Param id path string true "Sample ID"
// @Param sample body UpdateSampleRequest true "Sample data"
// @Success 200 {object} models.Sample
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/samples/{id} [put]
func (h *SampleHandler) UpdateSample(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID"})
		return
	}

	var req UpdateSampleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	sample := new(models.Sample)
	if err := h.db.NewSelect().Model(sample).Where("id = ?", id).Scan(ctx); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "sample not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch sample"})
		return
	}

	if req.Name != "" {
		sample.Name = req.Name
	}
	if req.Description != "" {
		sample.Description = req.Description
	}
	sample.UpdatedAt = time.Now()

	if _, err := h.db.NewUpdate().Model(sample).WherePK().Exec(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update sample"})
		return
	}

	c.JSON(http.StatusOK, sample)
}

// DeleteSample godoc
// @Summary Delete sample by ID
// @Description Deletes a sample by ID
// @Tags samples
// @Accept json
// @Produce json
// @Security CookieAuth
// @Param id path string true "Sample ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/samples/{id} [delete]
func (h *SampleHandler) DeleteSample(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	result, err := h.db.NewDelete().Model((*models.Sample)(nil)).Where("id = ?", id).Exec(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete sample"})
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to verify deletion"})
		return
	}

	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "sample not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "sample deleted"})
}
