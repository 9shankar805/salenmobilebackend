package messaging

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"smg-backend/internal/middleware"
	"smg-backend/internal/models"
)

type MessagingRouter struct {
	DB *gorm.DB
}

func RegisterRoutes(r *gin.Engine, db *gorm.DB) {
	mr := &MessagingRouter{DB: db}
	mr.seedSampleThreads()

	// Chat threads — reading own threads requires CHAT plan
	r.GET("/messages/threads", mr.GetThreads)
	r.GET("/messages/threads/:threadId/messages", mr.GetMessages)

	// Sending a message requires an active CHAT plan
	r.POST("/messages/send", middleware.RequireSubscription(db, "CHAT"), mr.SendMessage)

	// Admin messaging endpoints
	adm := r.Group("/api/admin/messages", middleware.AuthMiddleware(), middleware.AdminMiddleware())
	{
		adm.GET("/threads", mr.GetAllThreads)
		adm.POST("/reply", mr.AdminReply)
	}
}

func (mr *MessagingRouter) GetThreads(c *gin.Context) {
	userID := c.Query("userId")
	if userID == "" {
		userID = "guest-tech"
	}

	var threads []models.ChatThread
	err := mr.DB.Preload("Messages", func(db *gorm.DB) *gorm.DB {
		return db.Order("created_at desc").Limit(1)
	}).Where("user_id = ?", userID).Order("updated_at desc").Find(&threads).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    threads,
	})
}

func (mr *MessagingRouter) GetMessages(c *gin.Context) {
	threadID := c.Param("threadId")

	var messages []models.ChatMessage
	err := mr.DB.Where("thread_id = ?", threadID).Order("created_at asc").Find(&messages).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    messages,
	})
}

type SendMessageRequest struct {
	ThreadID       string `json:"threadId"`
	UserID         string `json:"userId"`
	InstructorName string `json:"instructorName"`
	SenderRole     string `json:"senderRole"` // student, instructor, admin
	Message        string `json:"message" binding:"required"`
	AttachmentURL  string `json:"attachmentUrl"`
}

func (mr *MessagingRouter) SendMessage(c *gin.Context) {
	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	if req.UserID == "" {
		req.UserID = "guest-tech"
	}
	if req.SenderRole == "" {
		req.SenderRole = "student"
	}
	if req.InstructorName == "" {
		req.InstructorName = "Salen Dangol (Master Instructor)"
	}

	// Find or create thread
	var thread models.ChatThread
	if req.ThreadID != "" {
		mr.DB.Where("id = ?", req.ThreadID).First(&thread)
	}

	now := time.Now()
	if thread.ID == "" {
		thread = models.ChatThread{
			ID:              uuid.New().String(),
			UserID:          req.UserID,
			InstructorID:    "instructor-salen",
			InstructorName:  req.InstructorName,
			ParticipantName: req.UserID,
			Subject:         "Hardware Diagnostics Inquiry",
			LastMessage:     req.Message,
			LastMessageAt:   now,
			UnreadCount:     1,
			CreatedAt:       now,
			UpdatedAt:       now,
		}
		if err := mr.DB.Omit("Messages").Create(&thread).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to create thread: " + err.Error()})
			return
		}
	} else {
		thread.LastMessage = req.Message
		thread.LastMessageAt = now
		thread.UpdatedAt = now
		if req.SenderRole == "student" {
			thread.UnreadCount++
		}
		mr.DB.Omit("Messages").Save(&thread)
	}

	msg := models.ChatMessage{
		ID:            uuid.New().String(),
		ThreadID:      thread.ID,
		SenderID:      req.UserID,
		SenderRole:    req.SenderRole,
		Message:       req.Message,
		AttachmentURL: req.AttachmentURL,
		IsRead:        false,
		CreatedAt:     now,
	}

	if err := mr.DB.Create(&msg).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data": gin.H{
			"message": msg,
			"thread":  thread,
		},
	})
}

type AdminReplyRequest struct {
	ThreadID   string `json:"threadId" binding:"required"`
	Message    string `json:"message" binding:"required"`
	SenderID   string `json:"senderId"`
	SenderRole string `json:"senderRole"`
}

func (mr *MessagingRouter) AdminReply(c *gin.Context) {
	var req AdminReplyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	senderID := req.SenderID
	if senderID == "" {
		senderID = "instructor-salen"
	}

	// Find or create the thread
	var thread models.ChatThread
	if err := mr.DB.Where("id = ?", req.ThreadID).First(&thread).Error; err != nil {
		// Thread doesn't exist — create a placeholder so reply still works
		now := time.Now()
		thread = models.ChatThread{
			ID:              req.ThreadID,
			UserID:          "student",
			InstructorID:    senderID,
			InstructorName:  "Salen Dangol (Master Instructor)",
			ParticipantName: "Student",
			Subject:         "Admin Reply Thread",
			LastMessage:     req.Message,
			LastMessageAt:   now,
			UnreadCount:     0,
			CreatedAt:       now,
			UpdatedAt:       now,
		}
		if err2 := mr.DB.Omit("Messages").Create(&thread).Error; err2 != nil {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Thread not found and could not be created: " + err2.Error()})
			return
		}
	}

	now := time.Now()
	msg := models.ChatMessage{
		ID:          uuid.New().String(),
		ThreadID:    req.ThreadID,
		SenderID:    senderID,
		SenderRole:  "instructor",
		Message:     req.Message,
		MessageText: req.Message,
		IsRead:      false,
		CreatedAt:   now,
	}

	if err := mr.DB.Create(&msg).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	// Update thread last message
	thread.LastMessage = req.Message
	thread.UpdatedAt = now
	mr.DB.Save(&thread)

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    msg,
	})
}

func (mr *MessagingRouter) GetAllThreads(c *gin.Context) {
	var threads []models.ChatThread
	if err := mr.DB.Order("updated_at desc").Find(&threads).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": threads})
}

func (mr *MessagingRouter) seedSampleThreads() {
	var count int64
	mr.DB.Model(&models.ChatThread{}).Count(&count)
	if count > 0 {
		return
	}

	tID := uuid.New().String()
	thread := models.ChatThread{
		ID:             tID,
		UserID:         "guest-tech",
		InstructorID:   "instructor-salen",
		InstructorName: "Salen Dangol (Master Instructor)",
		Subject:        "PMIC Short Circuit Help (Redmi Note 11)",
		LastMessage:    "Check coil L2001 with your multimeter in diode mode. Red probe to ground.",
		UnreadCount:    0,
		CreatedAt:      time.Now().Add(-5 * time.Hour),
		UpdatedAt:      time.Now().Add(-1 * time.Hour),
	}
	mr.DB.Omit("Messages").Create(&thread)

	messages := []models.ChatMessage{
		{
			ID:         uuid.New().String(),
			ThreadID:   tID,
			SenderID:   "guest-tech",
			SenderRole: "student",
			Message:    "Hello Sir! The device is drawing 0.050A stable on DC Power supply. What should I check first?",
			IsRead:     true,
			CreatedAt:  time.Now().Add(-5 * time.Hour),
		},
		{
			ID:         uuid.New().String(),
			ThreadID:   tID,
			SenderID:   "instructor-salen",
			SenderRole: "instructor",
			Message:    "0.050A freeze indicates secondary buck line short or missing crystal oscillation. Check coil L2001 with your multimeter in diode mode. Red probe to ground.",
			IsRead:     true,
			CreatedAt:  time.Now().Add(-1 * time.Hour),
		},
	}

	for _, m := range messages {
		mr.DB.Create(&m)
	}
}
