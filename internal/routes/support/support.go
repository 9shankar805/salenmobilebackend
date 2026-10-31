package support

import (
"net/http"

"github.com/gin-gonic/gin"
"gorm.io/gorm"

"smg-backend/internal/models"
)

type SupportRouter struct {
DB *gorm.DB
}

func RegisterRoutes(r *gin.Engine, db *gorm.DB) {
sr := &SupportRouter{DB: db}

// Support category tree  (Choose Category screen)
r.GET("/support/categories", sr.GetTopLevelCategories)
r.GET("/support/categories/:slug", sr.GetCategoryChildren)
r.GET("/support/categories/:slug/full", sr.GetCategoryWithChildren)
}

// GET /support/categories  — top-level categories (SOFTWARE, HARDWARE)
func (sr *SupportRouter) GetTopLevelCategories(c *gin.Context) {
var cats []models.SupportCategory
if err := sr.DB.Where("parent_id = ? AND is_active = ?", "", true).
Order("order_num asc").Find(&cats).Error; err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to fetch categories"})
return
}
c.JSON(http.StatusOK, gin.H{"success": true, "data": cats})
}

// GET /support/categories/:slug  — sub-categories for a given parent (e.g. /support/categories/software)
func (sr *SupportRouter) GetCategoryChildren(c *gin.Context) {
slug := c.Param("slug")

var parent models.SupportCategory
if err := sr.DB.Where("slug = ?", slug).First(&parent).Error; err != nil {
c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Category not found"})
return
}

var children []models.SupportCategory
sr.DB.Where("parent_id = ? AND is_active = ?", parent.ID, true).
Order("order_num asc").Find(&children)

c.JSON(http.StatusOK, gin.H{
"success": true,
"data": gin.H{
"category": parent,
"children": children,
},
})
}

// GET /support/categories/:slug/full  — category + children in one call
func (sr *SupportRouter) GetCategoryWithChildren(c *gin.Context) {
slug := c.Param("slug")

var cat models.SupportCategory
if err := sr.DB.Preload("Children", func(db *gorm.DB) *gorm.DB {
return db.Where("is_active = ?", true).Order("order_num asc")
}).Where("slug = ?", slug).First(&cat).Error; err != nil {
c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Category not found"})
return
}

c.JSON(http.StatusOK, gin.H{"success": true, "data": cat})
}
