package cohort

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CohortRouter struct {
	DB *gorm.DB
}

func RegisterRoutes(r *gin.Engine, db *gorm.DB) {
	cr := &CohortRouter{DB: db}

	r.GET("/cohort", cr.GetCohorts)
	r.POST("/cohort", cr.CreateCohort)
}

func (cr *CohortRouter) GetCohorts(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": []gin.H{
			{
				"id":         "cohort-2026-b1",
				"name":       "Master Hardware Batch 2026-A",
				"courseId":   "android-hardware-repairing-level-1",
				"totalSeats": 50,
				"enrolled":   42,
				"startDate":  "2026-02-01",
				"endDate":    "2026-05-01",
				"status":     "active",
			},
		},
	})
}

func (cr *CohortRouter) CreateCohort(c *gin.Context) {
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	body["id"] = "cohort-new"
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": body})
}
