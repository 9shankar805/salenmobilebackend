package home

import (
"net/http"

"github.com/gin-gonic/gin"
"gorm.io/gorm"

"smg-backend/internal/models"
)

type HomeRouter struct {
DB *gorm.DB
}

func RegisterRoutes(r *gin.Engine, db *gorm.DB) {
hr := &HomeRouter{DB: db}

// Home screen data (PDF Container grid + Repair Toolkit + Highlights)
r.GET("/home", hr.GetHomeData)
r.GET("/home/pdf-categories", hr.GetPdfCategories)
r.GET("/home/toolkit", hr.GetRepairToolkit)
r.GET("/home/highlights", hr.GetHighlights)
}

// GET /home  — all sections needed for the app Home screen in one call
func (hr *HomeRouter) GetHomeData(c *gin.Context) {
var pdfCats []models.PdfCategory
hr.DB.Where("is_active = ?", true).Order("order_num asc").Find(&pdfCats)

var toolkitItems []models.ToolItem
hr.DB.Where("category = ?", "Repair Toolkit").Find(&toolkitItems)

var highlights []models.Highlight
hr.DB.Where("is_active = ?", true).Order("order_num asc").Find(&highlights)

c.JSON(http.StatusOK, gin.H{
"success": true,
"data": gin.H{
"pdfContainer": gin.H{
"title":      "Pdf Container",
"subtitle":   "6 categories",
"categories": pdfCats,
},
"repairToolkit": gin.H{
"title":    "Repair Toolkit",
"subtitle": "Tools for the bench",
"items":    toolkitItems,
},
"highlights": gin.H{
"title":    "Highlights",
"subtitle": "Offers & announcements",
"items":    highlights,
},
},
})
}

// GET /home/pdf-categories
func (hr *HomeRouter) GetPdfCategories(c *gin.Context) {
var cats []models.PdfCategory
hr.DB.Where("is_active = ?", true).Order("order_num asc").Find(&cats)
c.JSON(http.StatusOK, gin.H{"success": true, "data": cats})
}

// GET /home/toolkit
func (hr *HomeRouter) GetRepairToolkit(c *gin.Context) {
var items []models.ToolItem
hr.DB.Where("category = ?", "Repair Toolkit").Find(&items)
c.JSON(http.StatusOK, gin.H{"success": true, "data": items})
}

// GET /home/highlights
func (hr *HomeRouter) GetHighlights(c *gin.Context) {
var highlights []models.Highlight
hr.DB.Where("is_active = ?", true).Order("order_num asc").Find(&highlights)
c.JSON(http.StatusOK, gin.H{"success": true, "data": highlights})
}
