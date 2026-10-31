package notifications

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"smg-backend/internal/middleware"
	"smg-backend/internal/models"
)

type NotificationsRouter struct {
	DB *gorm.DB
}

func RegisterRoutes(r *gin.Engine, db *gorm.DB) {
	nr := &NotificationsRouter{DB: db}
	nr.seedSampleNotifications()

	// Notification APIs for Flutter App
	r.POST("/notifications/token", nr.RegisterDeviceToken)
	r.GET("/notifications", nr.GetUserNotifications)
	r.POST("/notifications/:id/read", nr.MarkAsRead)  // kept for backward compat
	r.PATCH("/notifications/:id/read", nr.MarkAsRead) // Flutter app uses PATCH

	// Admin Notification APIs
	adm := r.Group("/api/admin/notifications", middleware.AuthMiddleware(), middleware.AdminMiddleware())
	{
		adm.GET("", nr.GetAllNotifications)
		adm.POST("/broadcast", nr.BroadcastNotification)
	}
}

type DeviceTokenRequest struct {
	UserID   string `json:"userId"`
	Token    string `json:"token" binding:"required"`
	Platform string `json:"platform"` // android, ios, web
}

func (nr *NotificationsRouter) RegisterDeviceToken(c *gin.Context) {
	var req DeviceTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	if req.Platform == "" {
		req.Platform = "android"
	}
	if req.UserID == "" {
		req.UserID = "guest-tech"
	}

	var existing models.DeviceToken
	err := nr.DB.Where("token = ?", req.Token).First(&existing).Error

	now := time.Now()
	if err != nil {
		// New token
		tokenRec := models.DeviceToken{
			ID:        uuid.New().String(),
			UserID:    req.UserID,
			Token:     req.Token,
			Platform:  req.Platform,
			UpdatedAt: now,
		}
		nr.DB.Create(&tokenRec)
	} else {
		// Update existing token with current user and timestamp
		existing.UserID = req.UserID
		existing.Platform = req.Platform
		existing.UpdatedAt = now
		nr.DB.Save(&existing)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "FCM device token registered successfully",
	})
}

func (nr *NotificationsRouter) GetUserNotifications(c *gin.Context) {
	userID := c.Query("userId")
	if userID == "" {
		userID = "guest-tech"
	}

	var list []models.InAppNotification
	err := nr.DB.Where("user_id = ? OR user_id = 'all'", userID).Order("created_at desc").Limit(50).Find(&list).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	unreadCount := 0
	for _, n := range list {
		if !n.IsRead {
			unreadCount++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"data":        list,
		"unreadCount": unreadCount,
	})
}

func (nr *NotificationsRouter) MarkAsRead(c *gin.Context) {
	id := c.Param("id")
	if err := nr.DB.Model(&models.InAppNotification{}).Where("id = ?", id).Update("is_read", true).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

type BroadcastRequest struct {
	Title    string `json:"title" binding:"required"`
	Body     string `json:"body" binding:"required"`
	Type     string `json:"type"` // alert, course_update, live_reminder, payment
	DeepLink string `json:"deepLink"`
}

func (nr *NotificationsRouter) BroadcastNotification(c *gin.Context) {
	var req BroadcastRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	if req.Type == "" {
		req.Type = "announcement"
	}

	notif := models.InAppNotification{
		ID:        uuid.New().String(),
		UserID:    "all",
		Title:     req.Title,
		Body:      req.Body,
		Type:      req.Type,
		DeepLink:  req.DeepLink,
		IsRead:    false,
		CreatedAt: time.Now(),
	}

	if err := nr.DB.Create(&notif).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	// Count targeted devices
	var deviceCount int64
	nr.DB.Model(&models.DeviceToken{}).Count(&deviceCount)

	c.JSON(http.StatusCreated, gin.H{
		"success":         true,
		"data":            notif,
		"broadcastStatus": "queued_fcm",
		"devicesTargeted": deviceCount,
	})
}

func (nr *NotificationsRouter) GetAllNotifications(c *gin.Context) {
	var list []models.InAppNotification
	if err := nr.DB.Order("created_at desc").Find(&list).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": list})
}

func (nr *NotificationsRouter) seedSampleNotifications() {
	var count int64
	nr.DB.Model(&models.InAppNotification{}).Count(&count)
	if count > 0 {
		return
	}

	samples := []models.InAppNotification{
		{
			ID:        uuid.New().String(),
			UserID:    "all",
			Title:     "🔥 New 4K PMIC Diagnostic Schematics Uploaded!",
			Body:      "Check Borneo and ZXW section for Redmi Note 12 4G power sequence lines.",
			Type:      "course_update",
			DeepLink:  "/tools/schematics",
			IsRead:    false,
			CreatedAt: time.Now().Add(-2 * time.Hour),
		},
		{
			ID:        uuid.New().String(),
			UserID:    "all",
			Title:     "🔴 Sunday Live Masterclass Alert",
			Body:      "Salen Dangol will demonstrate live double-decker CPU reballing at 8:00 PM.",
			Type:      "live_reminder",
			DeepLink:  "/live/sessions",
			IsRead:    false,
			CreatedAt: time.Now().Add(-24 * time.Hour),
		},
		{
			ID:        uuid.New().String(),
			UserID:    "all",
			Title:     "🎉 VIP Technician Access Granted",
			Body:      "Welcome to Salen Mobile Guru Academy! Unlimited access to hardware schematics.",
			Type:      "announcement",
			DeepLink:  "/member/home",
			IsRead:    true,
			CreatedAt: time.Now().Add(-48 * time.Hour),
		},
	}

	for _, s := range samples {
		nr.DB.Create(&s)
	}
}
