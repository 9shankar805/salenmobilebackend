package widgets

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type WidgetsRouter struct {
	DB *gorm.DB
}

func RegisterRoutes(r *gin.Engine, db *gorm.DB) {
	wr := &WidgetsRouter{DB: db}

	r.GET("/widgets", wr.GetWidgets)
	r.GET("/widgets/:id", wr.GetWidgetByID)
}

func (wr *WidgetsRouter) GetWidgets(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": []gin.H{
			{
				"id":    "w-multimeter",
				"type":  "hardware_tool",
				"title": "Interactive Diode Multimeter",
			},
			{
				"id":    "w-jumper-guide",
				"type":  "hardware_tool",
				"title": "Jumper Micro-Soldering Tracing",
			},
		},
	})
}

func (wr *WidgetsRouter) GetWidgetByID(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"id":    id,
			"type":  "hardware_tool",
			"title": "Interactive Hardware Guide",
		},
	})
}
