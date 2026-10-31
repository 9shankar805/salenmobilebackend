package invite

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type InviteRouter struct {
	DB *gorm.DB
}

func RegisterRoutes(r *gin.Engine, db *gorm.DB) {
	ir := &InviteRouter{DB: db}

	inv := r.Group("/invite")
	{
		inv.GET("/organization/pending", ir.GetPendingOrganizationInvite)
		inv.GET("/organization/:token/preview", ir.PreviewOrgInvite)
		inv.POST("/organization/:token/accept", ir.AcceptOrgInvite)

		inv.GET("/student/:token", ir.PreviewStudentInvite)

		inv.GET("/link/:token/preview", ir.PreviewLinkInvite)
		inv.POST("/link/:token/accept", ir.AcceptLinkInvite)

		inv.GET("/r/:token/preview", ir.PreviewLinkInvite)
		inv.POST("/r/:token/accept", ir.AcceptLinkInvite)
	}
}

func (ir *InviteRouter) GetPendingOrganizationInvite(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    nil,
	})
}

func (ir *InviteRouter) PreviewOrgInvite(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"organization": gin.H{
				"id":   "smg-org-1",
				"name": "Salen Mobile Guru Academy",
			},
			"role": "member",
		},
	})
}

func (ir *InviteRouter) AcceptOrgInvite(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"joined": true,
		},
	})
}

func (ir *InviteRouter) PreviewStudentInvite(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"courseTitle": "Android Hardware Repairing Level 1",
			"valid":       true,
		},
	})
}

func (ir *InviteRouter) PreviewLinkInvite(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"valid": true,
		},
	})
}

func (ir *InviteRouter) AcceptLinkInvite(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"joined": true,
		},
	})
}
