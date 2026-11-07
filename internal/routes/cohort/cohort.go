package cohort

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"smg-backend/internal/middleware"
	"smg-backend/internal/models"
)

type CohortRouter struct {
	DB *gorm.DB
}

func RegisterRoutes(r *gin.Engine, db *gorm.DB) {
	cr := &CohortRouter{DB: db}

	// Public — list cohorts (optionally filter by courseSlug)
	r.GET("/cohort", cr.GetCohorts)
	r.GET("/cohort/:id", cr.GetCohort)

	// Admin — protected
	adm := r.Group("/api/admin/cohorts", middleware.AuthMiddleware(), middleware.AdminMiddleware())
	{
		adm.GET("", cr.GetCohorts)
		adm.POST("", cr.CreateCohort)
		adm.PUT("/:id", cr.UpdateCohort)
		adm.DELETE("/:id", cr.DeleteCohort)
	}
}

// GET /cohort  or  GET /api/admin/cohorts
func (cr *CohortRouter) GetCohorts(c *gin.Context) {
	query := cr.DB.Model(&models.Cohort{})

	if slug := c.Query("courseSlug"); slug != "" {
		query = query.Where("course_slug = ?", slug)
	}
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	var cohorts []models.Cohort
	if err := query.Order("created_at desc").Find(&cohorts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to fetch cohorts"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": cohorts})
}

// GET /cohort/:id
func (cr *CohortRouter) GetCohort(c *gin.Context) {
	id := c.Param("id")
	var cohort models.Cohort
	if err := cr.DB.Where("id = ?", id).First(&cohort).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Cohort not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": cohort})
}

// POST /api/admin/cohorts
func (cr *CohortRouter) CreateCohort(c *gin.Context) {
	var cohort models.Cohort
	if err := c.ShouldBindJSON(&cohort); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	if cohort.ID == "" {
		cohort.ID = "cohort_" + uuid.New().String()[:8]
	}
	cohort.CreatedAt = time.Now()
	if cohort.Status == "" {
		cohort.Status = "upcoming"
	}
	if err := cr.DB.Create(&cohort).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": cohort})
}

// PUT /api/admin/cohorts/:id
func (cr *CohortRouter) UpdateCohort(c *gin.Context) {
	id := c.Param("id")
	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	if err := cr.DB.Model(&models.Cohort{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	var updated models.Cohort
	cr.DB.Where("id = ?", id).First(&updated)
	c.JSON(http.StatusOK, gin.H{"success": true, "data": updated})
}

// DELETE /api/admin/cohorts/:id
func (cr *CohortRouter) DeleteCohort(c *gin.Context) {
	id := c.Param("id")
	if err := cr.DB.Delete(&models.Cohort{}, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}
