package onboarding

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type OnboardingRouter struct {
	DB *gorm.DB
}

func RegisterRoutes(r *gin.Engine, db *gorm.DB) {
	obr := &OnboardingRouter{DB: db}

	ob := r.Group("/onboarding")
	{
		ob.POST("/create-org", obr.CreateOrg)
		ob.POST("/update-metadata", obr.UpdateMetadata)
		ob.POST("/welcome-email-pending", obr.WelcomeEmailPending)
	}
}

type CreateOrgRequest struct {
	Fullname string `json:"fullname"`
	OrgName  string `json:"orgName"`
	SiteName string `json:"siteName"`
}

func (obr *OnboardingRouter) CreateOrg(c *gin.Context) {
	var req CreateOrgRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data": gin.H{
			"id":       "smg-org-1",
			"name":     req.OrgName,
			"siteName": req.SiteName,
		},
	})
}

func (obr *OnboardingRouter) UpdateMetadata(c *gin.Context) {
	var body map[string]interface{}
	c.ShouldBindJSON(&body)
	c.JSON(http.StatusOK, gin.H{"success": true, "data": body})
}

func (obr *OnboardingRouter) WelcomeEmailPending(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"success": true, "data": nil})
}
