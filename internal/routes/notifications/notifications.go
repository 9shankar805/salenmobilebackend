package notifications

import (
	"context"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"smg-backend/internal/fcm"
	"smg-backend/internal/middleware"
	"smg-backend/internal/models"
)

// resetAndInit resets the FCM singleton and re-initialises with new credentials.
// This allows hot-reloading Firebase config from the admin panel without a restart.
var fcmInitMu sync.Mutex

func resetAndInit() {
	fcmInitMu.Lock()
	defer fcmInitMu.Unlock()
	// Re-call Init() — because fcm uses sync.Once internally we need to
	// reset the package state via the exported Reset helper.
	fcm.Reset()
	if err := fcm.Init(); err != nil {
		log.Printf("⚠️  FCM re-init: %v", err)
	}
}

type NotificationsRouter struct {
	DB *gorm.DB
}

func RegisterRoutes(r *gin.Engine, db *gorm.DB) {
	nr := &NotificationsRouter{DB: db}
	nr.seedSampleNotifications()

	// ── Flutter App APIs ──────────────────────────────────────────────────
	r.POST("/notifications/token", nr.RegisterDeviceToken)
	r.GET("/notifications", nr.GetUserNotifications)
	r.POST("/notifications/:id/read", nr.MarkAsRead)  // backward compat
	r.PATCH("/notifications/:id/read", nr.MarkAsRead) // Flutter uses PATCH
	r.POST("/notifications/mark-all-read", nr.MarkAllRead)

	// ── Admin APIs ────────────────────────────────────────────────────────
	adm := r.Group("/api/admin/notifications", middleware.AuthMiddleware(), middleware.AdminMiddleware())
	{
		adm.GET("", nr.GetAllNotifications)
		adm.POST("", nr.BroadcastNotification)
		adm.POST("/broadcast", nr.BroadcastNotification)
		adm.POST("/send-user", nr.SendToUser)   // push to a single user
		adm.POST("/send-topic", nr.SendToTopic) // push to an FCM topic
		adm.DELETE("/:id", nr.DeleteNotification)
		adm.GET("/fcm-status", nr.FCMStatus)         // check FCM health
		adm.POST("/fcm-config", nr.SaveFCMConfig)    // save credentials from admin panel
		adm.POST("/fcm-validate", nr.ValidateFCMCredentials) // validate JSON before saving
	}
}

// ─── Device Token Registration ────────────────────────────────────────────

type DeviceTokenRequest struct {
	UserID   string `json:"userId"`
	Token    string `json:"token" binding:"required"`
	FCMToken string `json:"fcmToken"` // Flutter app may send under either key
	Platform string `json:"platform"` // android, ios, web
}

// POST /notifications/token
// Called by the Flutter app on startup to register / refresh the FCM token.
// Also subscribes the device to the "all-users" FCM topic for broadcasts.
func (nr *NotificationsRouter) RegisterDeviceToken(c *gin.Context) {
	var req DeviceTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	// Accept token from either field name
	token := req.Token
	if token == "" {
		token = req.FCMToken
	}
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "token or fcmToken is required"})
		return
	}

	if req.Platform == "" {
		req.Platform = "android"
	}
	if req.UserID == "" {
		req.UserID = "guest-tech"
	}

	now := time.Now()
	var existing models.DeviceToken
	err := nr.DB.Where("fcm_token = ? OR token = ?", token, token).First(&existing).Error

	if err != nil {
		// New token — insert
		rec := models.DeviceToken{
			ID:          uuid.New().String(),
			UserID:      req.UserID,
			Token:       token,
			FCMToken:    token,
			Platform:    req.Platform,
			UpdatedAt:   now,
			LastUpdated: now,
		}
		if dbErr := nr.DB.Create(&rec).Error; dbErr != nil {
			log.Printf("⚠️  FCM token save error: %v", dbErr)
		}
	} else {
		// Existing token — update user and timestamp
		existing.UserID = req.UserID
		existing.Token = token
		existing.FCMToken = token
		existing.Platform = req.Platform
		existing.UpdatedAt = now
		existing.LastUpdated = now
		nr.DB.Save(&existing)
	}

	// Subscribe token to the "all-users" FCM topic so broadcasts work without
	// iterating every token individually.
	if fcm.IsReady() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if subErr := fcm.SubscribeToTopic(ctx, []string{token}, "all-users"); subErr != nil {
			log.Printf("⚠️  FCM topic subscription error: %v", subErr)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"message":  "Device token registered",
		"fcmReady": fcm.IsReady(),
	})
}

// ─── In-App Notification Feed ─────────────────────────────────────────────

// GET /notifications?userId=<id>
func (nr *NotificationsRouter) GetUserNotifications(c *gin.Context) {
	userID := c.Query("userId")
	if userID == "" {
		userID = "guest-tech"
	}

	var list []models.InAppNotification
	if err := nr.DB.Where("user_id = ? OR user_id = 'all'", userID).
		Order("created_at desc").Limit(50).Find(&list).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	unread := 0
	for _, n := range list {
		if !n.IsRead {
			unread++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"data":        list,
		"unreadCount": unread,
	})
}

// POST /notifications/:id/read  or  PATCH /notifications/:id/read
func (nr *NotificationsRouter) MarkAsRead(c *gin.Context) {
	id := c.Param("id")
	if err := nr.DB.Model(&models.InAppNotification{}).
		Where("id = ?", id).Update("is_read", true).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// POST /notifications/mark-all-read?userId=<id>
func (nr *NotificationsRouter) MarkAllRead(c *gin.Context) {
	userID := c.Query("userId")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "userId required"})
		return
	}
	nr.DB.Model(&models.InAppNotification{}).
		Where("user_id = ? OR user_id = 'all'", userID).
		Update("is_read", true)
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// ─── Admin: Broadcast to all users ────────────────────────────────────────

type BroadcastRequest struct {
	Title     string `json:"title"     binding:"required"`
	Body      string `json:"body"      binding:"required"`
	Type      string `json:"type"`      // alert, course_update, live_reminder, payment, announcement
	DeepLink  string `json:"deepLink"`  // in-app route e.g. /live/sessions
	ImageURL  string `json:"imageUrl"`  // optional rich-push banner image
	UseTopic  bool   `json:"useTopic"`  // true = FCM topic fan-out (fast), false = token batch
}

// POST /api/admin/notifications  or  POST /api/admin/notifications/broadcast
// Saves the notification to DB and dispatches a real FCM push to every device.
func (nr *NotificationsRouter) BroadcastNotification(c *gin.Context) {
	var req BroadcastRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	if req.Type == "" {
		req.Type = "announcement"
	}

	// 1. Persist in-app notification record
	notif := models.InAppNotification{
		ID:        uuid.New().String(),
		UserID:    "all",
		Title:     req.Title,
		Body:      req.Body,
		Type:      req.Type,
		DeepLink:  req.DeepLink,
		ActionURL: req.DeepLink,
		IsRead:    false,
		CreatedAt: time.Now(),
	}
	if err := nr.DB.Create(&notif).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	// 2. Count devices
	var deviceCount int64
	nr.DB.Model(&models.DeviceToken{}).Count(&deviceCount)

	// 3. Dispatch FCM push
	fcmResult := dispatchBroadcast(nr.DB, req)

	c.JSON(http.StatusCreated, gin.H{
		"success":         true,
		"data":            notif,
		"devicesTargeted": deviceCount,
		"fcm":             fcmResult,
	})
}

// dispatchBroadcast sends the actual FCM push.
// Prefers topic-based fan-out when useTopic=true (or when device count > 100).
// Falls back to multicast batches for smaller audiences or explicit token mode.
func dispatchBroadcast(db *gorm.DB, req BroadcastRequest) map[string]interface{} {
	if !fcm.IsReady() {
		return map[string]interface{}{
			"status":  "skipped",
			"reason":  "FCM not configured — add FIREBASE_CREDENTIALS_FILE or FIREBASE_CREDENTIALS_JSON to .env",
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	notif := fcm.Notification{
		Title:    req.Title,
		Body:     req.Body,
		ImageURL: req.ImageURL,
	}
	data := fcm.Data{
		"type":     req.Type,
		"deepLink": req.DeepLink,
		"notifId":  uuid.New().String(),
	}

	// Count devices to choose strategy
	var count int64
	db.Model(&models.DeviceToken{}).Count(&count)

	// Use topic fan-out if explicitly requested or device count > 100
	if req.UseTopic || count > 100 {
		msgID, err := fcm.SendToTopic(ctx, "all-users", notif, data)
		if err != nil {
			log.Printf("❌ FCM topic broadcast error: %v", err)
			return map[string]interface{}{"status": "error", "error": err.Error()}
		}
		log.Printf("✅ FCM topic broadcast sent — msgID: %s, devices: %d", msgID, count)
		return map[string]interface{}{
			"status":   "sent",
			"method":   "topic",
			"topic":    "all-users",
			"msgId":    msgID,
			"devices":  count,
		}
	}

	// Token-by-token multicast (batches of 500)
	var tokens []models.DeviceToken
	db.Select("fcm_token, token").Find(&tokens)

	fcmTokens := make([]string, 0, len(tokens))
	for _, t := range tokens {
		tok := t.FCMToken
		if tok == "" {
			tok = t.Token
		}
		if tok != "" {
			fcmTokens = append(fcmTokens, tok)
		}
	}

	if len(fcmTokens) == 0 {
		return map[string]interface{}{"status": "skipped", "reason": "no registered devices"}
	}

	totalSuccess, totalFail := 0, 0
	var invalid []string

	// Send in batches of 500 (FCM multicast limit)
	for i := 0; i < len(fcmTokens); i += 500 {
		end := i + 500
		if end > len(fcmTokens) {
			end = len(fcmTokens)
		}
		batch := fcmTokens[i:end]

		result, err := fcm.SendMulticast(ctx, batch, notif, data)
		if err != nil {
			log.Printf("❌ FCM multicast batch [%d-%d] error: %v", i, end, err)
			totalFail += len(batch)
			continue
		}
		totalSuccess += result.SuccessCount
		totalFail += result.FailureCount
		invalid = append(invalid, result.InvalidTokens...)
	}

	// Clean up invalid / unregistered tokens from DB
	if len(invalid) > 0 {
		log.Printf("🧹 FCM: removing %d invalid tokens from DB", len(invalid))
		db.Where("fcm_token IN ? OR token IN ?", invalid, invalid).
			Delete(&models.DeviceToken{})
	}

	log.Printf("✅ FCM multicast done — success: %d, failed: %d, invalid cleaned: %d",
		totalSuccess, totalFail, len(invalid))

	return map[string]interface{}{
		"status":         "sent",
		"method":         "multicast",
		"successCount":   totalSuccess,
		"failureCount":   totalFail,
		"invalidCleaned": len(invalid),
	}
}

// ─── Admin: Push to single user ───────────────────────────────────────────

type SendUserRequest struct {
	UserID   string `json:"userId"   binding:"required"`
	Title    string `json:"title"    binding:"required"`
	Body     string `json:"body"     binding:"required"`
	Type     string `json:"type"`
	DeepLink string `json:"deepLink"`
	ImageURL string `json:"imageUrl"`
}

// POST /api/admin/notifications/send-user
// Sends a push notification to all devices registered under a specific user ID.
func (nr *NotificationsRouter) SendToUser(c *gin.Context) {
	var req SendUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	if req.Type == "" {
		req.Type = "personal"
	}

	// Save in-app notification for this specific user
	notifRecord := models.InAppNotification{
		ID:        uuid.New().String(),
		UserID:    req.UserID,
		Title:     req.Title,
		Body:      req.Body,
		Type:      req.Type,
		DeepLink:  req.DeepLink,
		ActionURL: req.DeepLink,
		IsRead:    false,
		CreatedAt: time.Now(),
	}
	nr.DB.Create(&notifRecord)

	if !fcm.IsReady() {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    notifRecord,
			"fcm":     map[string]interface{}{"status": "skipped", "reason": "FCM not configured"},
		})
		return
	}

	// Fetch all FCM tokens for this user
	var tokens []models.DeviceToken
	nr.DB.Where("user_id = ?", req.UserID).Find(&tokens)

	if len(tokens) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    notifRecord,
			"fcm":     map[string]interface{}{"status": "skipped", "reason": "no devices found for user"},
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	notif := fcm.Notification{Title: req.Title, Body: req.Body, ImageURL: req.ImageURL}
	data := fcm.Data{
		"type":     req.Type,
		"deepLink": req.DeepLink,
		"userId":   req.UserID,
		"notifId":  notifRecord.ID,
	}

	successCount, failCount := 0, 0
	for _, t := range tokens {
		tok := t.FCMToken
		if tok == "" {
			tok = t.Token
		}
		if tok == "" {
			continue
		}
		if _, err := fcm.SendToToken(ctx, tok, notif, data); err != nil {
			log.Printf("⚠️  FCM send to user %s token error: %v", req.UserID, err)
			failCount++
		} else {
			successCount++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    notifRecord,
		"fcm": map[string]interface{}{
			"status":       "sent",
			"successCount": successCount,
			"failureCount": failCount,
		},
	})
}

// ─── Admin: Push to FCM topic ─────────────────────────────────────────────

type SendTopicRequest struct {
	Topic    string `json:"topic"    binding:"required"` // e.g. "all-users", "vip-users"
	Title    string `json:"title"    binding:"required"`
	Body     string `json:"body"     binding:"required"`
	Type     string `json:"type"`
	DeepLink string `json:"deepLink"`
	ImageURL string `json:"imageUrl"`
}

// POST /api/admin/notifications/send-topic
func (nr *NotificationsRouter) SendToTopic(c *gin.Context) {
	var req SendTopicRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	if !fcm.IsReady() {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"success": false,
			"error":   "FCM not configured — add FIREBASE_CREDENTIALS_FILE or FIREBASE_CREDENTIALS_JSON to .env",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	notif := fcm.Notification{Title: req.Title, Body: req.Body, ImageURL: req.ImageURL}
	data := fcm.Data{
		"type":     req.Type,
		"deepLink": req.DeepLink,
	}

	msgID, err := fcm.SendToTopic(ctx, req.Topic, notif, data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"msgId":   msgID,
		"topic":   req.Topic,
	})
}

// ─── Admin: FCM Health & Config ───────────────────────────────────────────

// GET /api/admin/notifications/fcm-status
func (nr *NotificationsRouter) FCMStatus(c *gin.Context) {
	var deviceCount int64
	nr.DB.Model(&models.DeviceToken{}).Count(&deviceCount)

	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"fcm":          fcm.StatusMap(),
		"deviceTokens": deviceCount,
	})
}

type FCMConfigRequest struct {
	CredentialsJSON string `json:"credentialsJson"` // raw service account JSON
	CredentialsFile string `json:"credentialsFile"` // path to file (alternative)
	ProjectID       string `json:"projectId"`
}

// POST /api/admin/notifications/fcm-config
// Saves Firebase credentials to .env and re-initialises FCM.
// The admin can paste the service account JSON directly in the admin panel.
func (nr *NotificationsRouter) SaveFCMConfig(c *gin.Context) {
	var req FCMConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	if req.CredentialsJSON == "" && req.CredentialsFile == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "credentialsJson or credentialsFile required"})
		return
	}

	if req.CredentialsJSON != "" {
		// Validate before saving
		if err := fcm.ValidateCredentialsJSON(req.CredentialsJSON); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid credentials: " + err.Error()})
			return
		}
		// Write to a file so it persists across restarts
		credPath := "/var/www/salenmobilebackend/firebase-credentials.json"
		if err := os.WriteFile(credPath, []byte(req.CredentialsJSON), 0600); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to save credentials file: " + err.Error()})
			return
		}
		// Set env var so Init() picks it up on re-init
		os.Setenv("FIREBASE_CREDENTIALS_FILE", credPath)
		if req.ProjectID != "" {
			os.Setenv("FIREBASE_PROJECT_ID", req.ProjectID)
		}
	} else {
		os.Setenv("FIREBASE_CREDENTIALS_FILE", req.CredentialsFile)
	}

	// Re-initialise FCM with new credentials (reset the once guard)
	resetAndInit()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Firebase credentials saved and FCM re-initialised",
		"fcm":     fcm.StatusMap(),
	})
}

// POST /api/admin/notifications/fcm-validate
// Validates a service account JSON without saving it.
func (nr *NotificationsRouter) ValidateFCMCredentials(c *gin.Context) {
	var body struct {
		CredentialsJSON string `json:"credentialsJson" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	if err := fcm.ValidateCredentialsJSON(body.CredentialsJSON); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "valid": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "valid": true, "message": "Service account JSON is valid"})
}

// ─── Admin: List all & Delete ─────────────────────────────────────────────

func (nr *NotificationsRouter) GetAllNotifications(c *gin.Context) {
	var list []models.InAppNotification
	if err := nr.DB.Order("created_at desc").Find(&list).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": list})
}

func (nr *NotificationsRouter) DeleteNotification(c *gin.Context) {
	id := c.Param("id")
	nr.DB.Delete(&models.InAppNotification{}, "id = ?", id)
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// ─── Seed ─────────────────────────────────────────────────────────────────

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
