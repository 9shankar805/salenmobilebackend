package sso

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type SsoRouter struct {
	DB *gorm.DB
}

func RegisterRoutes(r *gin.Engine, db *gorm.DB) {
	sr := &SsoRouter{DB: db}

	sso := r.Group("/sso")
	{
		sso.GET("/discover", sr.DiscoverSSO)
		sso.GET("/org/:orgId", sr.GetOrgSSO)
	}
}

func (sr *SsoRouter) DiscoverSSO(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"ssoAvailable": false,
		},
	})
}

func (sr *SsoRouter) GetOrgSSO(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"enabled":  false,
			"enforced": false,
		},
	})
}
