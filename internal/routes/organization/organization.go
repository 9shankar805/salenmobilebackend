package organization

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type OrganizationRouter struct {
	DB *gorm.DB
}

func RegisterRoutes(r *gin.Engine, db *gorm.DB) {
	orgR := &OrganizationRouter{DB: db}

	// Organization compatibility endpoints
	org := r.Group("/organization")
	{
		org.GET("", orgR.GetOrganization)
		org.GET("/first", orgR.GetOrganization)
		org.GET("/tags", orgR.GetTags)
		org.GET("/tags/public", orgR.GetTags)
		org.GET("/assets", orgR.GetAssets)
		org.GET("/assets/storage", orgR.GetAssetsStorage)
		org.GET("/automation/keys", orgR.GetAutomationKeys)
		org.GET("/app-screens", orgR.GetAppScreens)
		org.GET("/widgets", orgR.GetWidgets)
	}
}

func (or *OrganizationRouter) getOrgData() gin.H {
	return gin.H{
		"id":         "smg-org-1",
		"name":       "Salen Mobile Guru",
		"siteName":   "smg",
		"avatar_url": "https://images.unsplash.com/photo-1581092918056-0c4c3acd3789?w=200",
		"role":       "owner",
		"roleId":     1,
		"created_at": "2026-01-01T00:00:00Z",
		"landingpage": gin.H{
			"heading":     "Salen Mobile Guru Academy",
			"description": "Master Professional Mobile Hardware, Software & Micro-Soldering Engineering",
		},
	}
}

func (or *OrganizationRouter) GetOrganization(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    []gin.H{or.getOrgData()},
	})
}

func (or *OrganizationRouter) GetTags(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": []string{
			"Hardware",
			"Software",
			"Micro-Soldering",
			"PMIC Diagnostics",
			"Schematics",
			"eMMC & UFS",
			"iPhone FaceID & Display",
		},
	})
}

func (or *OrganizationRouter) GetAssets(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": []gin.H{
			{
				"id":   "asset-1",
				"name": "banner_1.webp",
				"url":  "/static/assets/banner_1.webp",
				"size": 1048576,
			},
		},
	})
}

func (or *OrganizationRouter) GetAssetsStorage(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"usedBytes":  52428800,
			"totalBytes": 10737418240,
			"plan":       "Unlimited VIP",
		},
	})
}

func (or *OrganizationRouter) GetAutomationKeys(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"success": true, "data": []gin.H{}})
}

func (or *OrganizationRouter) GetAppScreens(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"success": true, "data": []gin.H{}})
}

func (or *OrganizationRouter) GetWidgets(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"success": true, "data": []gin.H{}})
}

func (or *OrganizationRouter) GetSsoOrg(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"enabled":  false,
			"enforced": false,
		},
	})
}

func (or *OrganizationRouter) GetSsoDiscover(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"ssoAvailable": false,
		},
	})
}

func (or *OrganizationRouter) GetPendingInvites(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    nil,
	})
}
