package plans

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"smg-backend/internal/middleware"
	"smg-backend/internal/models"
)

type PlansRouter struct {
	DB *gorm.DB
}

func RegisterRoutes(r *gin.Engine, db *gorm.DB) {
	pr := &PlansRouter{DB: db}

	// Mobile Plans
	r.GET("/plans/mobile", pr.GetMobilePlans)
	// Chat Plans
	r.GET("/plans/chat", pr.GetChatPlans)
	// All plans (both types)
	r.GET("/plans", pr.GetAllPlans)
	r.GET("/api/v1/plans", pr.GetAllPlans)

	// Admin Plan Management
	adm := r.Group("/api/admin/plans", middleware.AuthMiddleware(), middleware.AdminMiddleware())
	{
		adm.GET("", pr.AdminListPlans)
		adm.POST("", pr.AdminCreatePlan)
		adm.PUT("/:id", pr.AdminUpdatePlan)
		adm.DELETE("/:id", pr.AdminDeletePlan)
	}
}

// GET /plans/mobile  — list all active mobile plans in order
func (pr *PlansRouter) GetMobilePlans(c *gin.Context) {
var plans []models.Plan
if err := pr.DB.Where("plan_type = ? AND is_active = ?", "MOBILE", true).
Order("order_num asc").Find(&plans).Error; err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to fetch mobile plans"})
return
}
c.JSON(http.StatusOK, gin.H{
"success": true,
"data":    formatPlans(plans),
})
}

// GET /plans/chat  — list all active chat plans in order
func (pr *PlansRouter) GetChatPlans(c *gin.Context) {
var plans []models.Plan
if err := pr.DB.Where("plan_type = ? AND is_active = ?", "CHAT", true).
Order("order_num asc").Find(&plans).Error; err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to fetch chat plans"})
return
}
c.JSON(http.StatusOK, gin.H{
"success": true,
"data":    formatPlans(plans),
})
}

// GET /plans?type=MOBILE|CHAT  — optional filter
func (pr *PlansRouter) GetAllPlans(c *gin.Context) {
planType := c.Query("type")
query := pr.DB.Where("is_active = ?", true)
if planType != "" {
query = query.Where("plan_type = ?", planType)
}
var plans []models.Plan
if err := query.Order("plan_type asc, order_num asc").Find(&plans).Error; err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to fetch plans"})
return
}
c.JSON(http.StatusOK, gin.H{
"success": true,
"data":    formatPlans(plans),
})
}

// formatPlans unmarshals the JSON features string back into a slice for the response.
func formatPlans(plans []models.Plan) []map[string]interface{} {
result := make([]map[string]interface{}, 0, len(plans))
for _, p := range plans {
var features []string
_ = json.Unmarshal([]byte(p.Features), &features)

item := map[string]interface{}{
"id":               p.ID,
"planType":         p.PlanType,
"name":             p.Name,
"durationLabel":    p.DurationLabel,
"durationMonths":   p.DurationMonths,
"price":            p.Price,
"originalPrice":    p.OriginalPrice,
"discountPercent":  p.DiscountPercent,
"currency":         p.Currency,
"currencySymbol":   p.CurrencySymbol,
"badge":            p.Badge,
"features":         features,
"responseTimeMins": p.ResponseTimeMins,
"isActive":         p.IsActive,
"order":            p.OrderNum,
}
result = append(result, item)
}
return result
}

func (pr *PlansRouter) AdminListPlans(c *gin.Context) {
	var plans []models.Plan
	if err := pr.DB.Order("plan_type asc, order_num asc").Find(&plans).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    formatPlans(plans),
		"plans":   plans,
	})
}

func (pr *PlansRouter) AdminCreatePlan(c *gin.Context) {
	var plan models.Plan
	if err := c.ShouldBindJSON(&plan); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	if plan.ID == "" {
		plan.ID = "plan_" + uuid.New().String()[:8]
	}
	if plan.Currency == "" {
		plan.Currency = "NPR"
	}
	if plan.CurrencySymbol == "" {
		plan.CurrencySymbol = "Rs."
	}
	plan.CreatedAt = time.Now()
	if err := pr.DB.Create(&plan).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": plan})
}

func (pr *PlansRouter) AdminUpdatePlan(c *gin.Context) {
	id := c.Param("id")
	var req models.Plan
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	var existing models.Plan
	if err := pr.DB.Where("id = ?", id).First(&existing).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Plan not found"})
		return
	}
	req.ID = existing.ID
	req.CreatedAt = existing.CreatedAt
	if err := pr.DB.Save(&req).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": req})
}

func (pr *PlansRouter) AdminDeletePlan(c *gin.Context) {
	id := c.Param("id")
	pr.DB.Delete(&models.Plan{}, "id = ?", id)
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Plan deleted"})
}
