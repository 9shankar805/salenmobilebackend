package calculator

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"smg-backend/internal/middleware"
	"smg-backend/internal/models"
)

type CalculatorRouter struct {
	DB *gorm.DB
}

func RegisterRoutes(r *gin.Engine, db *gorm.DB) {
	c := &CalculatorRouter{DB: db}
	c.seedDefaults()

	// Pricing matrix is free to view (attracts users)
	r.GET("/calculator/pricing-matrix", c.GetPricingMatrix)

	// Work orders require MOBILE plan to create/track repairs
	calcAuth := r.Group("/calculator", middleware.RequireSubscription(db, "MOBILE"))
	{
		calcAuth.GET("/work-orders", c.GetWorkOrders)
		calcAuth.POST("/work-orders", c.CreateWorkOrder)
		calcAuth.PATCH("/work-orders/:id/status", c.UpdateWorkOrderStatus)
	}

	// Admin calculator endpoints
	admCalc := r.Group("/api/admin/calculator", middleware.AuthMiddleware(), middleware.AdminMiddleware())
	{
		admCalc.POST("/pricing-matrix", c.CreatePricingItem)
		admCalc.DELETE("/pricing-matrix/:id", c.DeletePricingItem)
		admCalc.GET("/work-orders", c.GetAllWorkOrders)
	}
}

func (cr *CalculatorRouter) GetPricingMatrix(c *gin.Context) {
	brand := c.Query("brand")
	modelName := c.Query("model")

	var list []models.RepairPriceGuide
	query := cr.DB.Order("brand ASC, model_name ASC")

	if brand != "" {
		query = query.Where("LOWER(brand) = LOWER(?)", brand)
	}
	if modelName != "" {
		query = query.Where("LOWER(model_name) LIKE LOWER(?)", "%"+modelName+"%")
	}

	if err := query.Find(&list).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": list})
}

func (cr *CalculatorRouter) CreatePricingItem(c *gin.Context) {
	var item models.RepairPriceGuide
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	if item.ID == "" {
		item.ID = "pg_" + uuid.New().String()[:8]
	}
	if item.TotalPrice == 0 {
		item.TotalPrice = item.PartCost + item.LaborCharge
	}

	if err := cr.DB.Create(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": item})
}

func (cr *CalculatorRouter) DeletePricingItem(c *gin.Context) {
	id := c.Param("id")
	cr.DB.Delete(&models.RepairPriceGuide{}, "id = ?", id)
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Price item deleted"})
}

func (cr *CalculatorRouter) GetWorkOrders(c *gin.Context) {
	techID := c.Query("technicianId")
	var orders []models.WorkOrder

	query := cr.DB.Order("created_at DESC")
	if techID != "" {
		query = query.Where("technician_id = ?", techID)
	}

	if err := query.Find(&orders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": orders})
}

func (cr *CalculatorRouter) GetAllWorkOrders(c *gin.Context) {
	var orders []models.WorkOrder
	cr.DB.Order("created_at DESC").Find(&orders)
	c.JSON(http.StatusOK, gin.H{"success": true, "data": orders})
}

func (cr *CalculatorRouter) CreateWorkOrder(c *gin.Context) {
	var order models.WorkOrder
	if err := c.ShouldBindJSON(&order); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	if order.ID == "" {
		order.ID = "WO-" + time.Now().Format("20060102") + "-" + uuid.New().String()[:4]
	}
	if order.Status == "" {
		order.Status = "PENDING"
	}
	order.CreatedAt = time.Now()
	order.UpdatedAt = time.Now()

	if err := cr.DB.Create(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": order})
}

func (cr *CalculatorRouter) UpdateWorkOrderStatus(c *gin.Context) {
	id := c.Param("id")
	var body struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	if err := cr.DB.Model(&models.WorkOrder{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":     body.Status,
		"updated_at": time.Now(),
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Work order status updated"})
}

func (cr *CalculatorRouter) seedDefaults() {
	var count int64
	cr.DB.Model(&models.RepairPriceGuide{}).Count(&count)
	if count > 0 {
		return
	}

	defaults := []models.RepairPriceGuide{
		{ID: "pg1", Brand: "Apple", ModelName: "iPhone 13", RepairType: "Original OLED Display", PartCost: 8500, LaborCharge: 1500, TotalPrice: 10000},
		{ID: "pg2", Brand: "Apple", ModelName: "iPhone 13", RepairType: "Battery Replacement (Original)", PartCost: 2800, LaborCharge: 1000, TotalPrice: 3800},
		{ID: "pg3", Brand: "Apple", ModelName: "iPhone 13", RepairType: "Charging Port Flex Cable", PartCost: 1400, LaborCharge: 800, TotalPrice: 2200},
		{ID: "pg4", Brand: "Xiaomi", ModelName: "Redmi Note 11", RepairType: "AMOLED Folder Assembly", PartCost: 3200, LaborCharge: 800, TotalPrice: 4000},
		{ID: "pg5", Brand: "Xiaomi", ModelName: "Redmi Note 11", RepairType: "PMIC Power IC Reballing", PartCost: 400, LaborCharge: 1600, TotalPrice: 2000},
		{ID: "pg6", Brand: "Samsung", ModelName: "Galaxy A52", RepairType: "Super AMOLED Frame Set", PartCost: 4800, LaborCharge: 1200, TotalPrice: 6000},
	}
	cr.DB.Create(&defaults)
}
