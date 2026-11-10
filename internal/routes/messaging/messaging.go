package messaging

import (
	"net/http"
	"sort"
	"strings"
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

	// Chat threads — reading own threads
	r.GET("/messages/threads", mr.GetThreads)
	r.GET("/messages/threads/:threadId/messages", mr.GetMessages)

	// Sending a message
	r.POST("/messages/send", mr.SendMessage)

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
		userID = "aa7aa465-433c-49c8-9bb6-21eb94cd2a94"
	}

	var threads []models.ChatThread
	err := mr.DB.Preload("Messages", func(db *gorm.DB) *gorm.DB {
		return db.Order("created_at desc").Limit(1)
	}).Where("user_id = ? OR user_id = 'guest-tech' OR user_id = 'aa7aa465-433c-49c8-9bb6-21eb94cd2a94'", userID).
		Order("updated_at desc").Find(&threads).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	for i := range threads {
		threads[i].Name = threads[i].ParticipantName
		threads[i].UserName = threads[i].ParticipantName
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    threads,
	})
}

func (mr *MessagingRouter) GetMessages(c *gin.Context) {
	threadID := c.Param("threadId")

	var messages []models.ChatMessage
	mr.DB.Where("thread_id = ?", threadID).Order("created_at asc").Find(&messages)

	// Also check support_messages so threads are completely unified
	var sMsgs []models.SupportMessage
	mr.DB.Where("thread_id = ?", threadID).Order("created_at asc").Find(&sMsgs)

	seenMsgIds := make(map[string]bool)
	for _, m := range messages {
		seenMsgIds[m.ID] = true
	}

	for _, sm := range sMsgs {
		if !seenMsgIds[sm.ID] {
			messages = append(messages, models.ChatMessage{
				ID:            sm.ID,
				ThreadID:      sm.ThreadID,
				SenderID:      sm.SenderID,
				SenderName:    sm.SenderName,
				SenderRole:    sm.SenderRole,
				Message:       sm.Message,
				MessageText:   sm.Message,
				AttachmentURL: sm.AttachmentURL,
				IsRead:        sm.IsRead,
				CreatedAt:     sm.CreatedAt,
			})
			seenMsgIds[sm.ID] = true
		}
	}

	sort.Slice(messages, func(i, j int) bool {
		return messages[i].CreatedAt.Before(messages[j].CreatedAt)
	})

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    messages,
	})
}

type SendMessageRequest struct {
	ThreadID        string `json:"threadId"`
	UserID          string `json:"userId"`
	UserName        string `json:"userName"`
	ParticipantName string `json:"participantName"`
	SenderName      string `json:"senderName"`
	InstructorName  string `json:"instructorName"`
	SenderRole      string `json:"senderRole"` // student, instructor, admin
	Message         string `json:"message" binding:"required"`
	AttachmentURL   string `json:"attachmentUrl"`
}

func (mr *MessagingRouter) SendMessage(c *gin.Context) {
	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	if req.UserID == "" {
		req.UserID = "aa7aa465-433c-49c8-9bb6-21eb94cd2a94"
	}
	if req.SenderRole == "" {
		req.SenderRole = "student"
	}
	if req.InstructorName == "" {
		req.InstructorName = "Salen Dangol (Master Instructor)"
	}

	// Resolve student real name
	studentName := strings.TrimSpace(req.UserName)
	if studentName == "" {
		studentName = strings.TrimSpace(req.ParticipantName)
	}
	if studentName == "" {
		studentName = strings.TrimSpace(req.SenderName)
	}
	if studentName == "" || studentName == req.UserID || strings.ToLower(studentName) == "student" || strings.ToLower(studentName) == "guest-tech" {
		var u models.User
		if err := mr.DB.Where("id = ? OR LOWER(email) = LOWER(?)", req.UserID, req.UserID).First(&u).Error; err == nil && u.Name != "" {
			studentName = u.Name
		} else if strings.Contains(req.UserID, "@") {
			parts := strings.Split(req.UserID, "@")
			studentName = parts[0]
		} else if req.UserID == "aa7aa465-433c-49c8-9bb6-21eb94cd2a94" || req.UserID == "guest-tech" {
			studentName = "Shankar"
		} else {
			studentName = "Shankar"
		}
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
			ParticipantName: studentName,
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
		if studentName != "" && (thread.ParticipantName == "" || thread.ParticipantName == thread.UserID || thread.ParticipantName == "guest-tech" || thread.ParticipantName == "Student") {
			thread.ParticipantName = studentName
		}
		if req.SenderRole == "student" {
			thread.UnreadCount++
		}
		mr.DB.Omit("Messages").Save(&thread)
	}

	msg := models.ChatMessage{
		ID:            uuid.New().String(),
		ThreadID:      thread.ID,
		SenderID:      req.UserID,
		SenderName:    studentName,
		SenderRole:    req.SenderRole,
		Message:       req.Message,
		MessageText:   req.Message,
		AttachmentURL: req.AttachmentURL,
		IsRead:        false,
		CreatedAt:     now,
	}

	if err := mr.DB.Create(&msg).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	// Also mirror to support_threads / support_messages if present or needed
	var sThread models.SupportThread
	if err := mr.DB.Where("id = ?", thread.ID).First(&sThread).Error; err == nil {
		sMsg := models.SupportMessage{
			ID:            uuid.New().String(),
			ThreadID:      thread.ID,
			SenderID:      req.UserID,
			SenderName:    studentName,
			SenderRole:    "student",
			Message:       req.Message,
			AttachmentURL: req.AttachmentURL,
			IsRead:        false,
			CreatedAt:     now,
		}
		mr.DB.Create(&sMsg)
		mr.DB.Model(&sThread).Updates(map[string]interface{}{
			"last_message":       req.Message,
			"last_message_at":    now,
			"updated_at":         now,
			"unread_admin_count": sThread.UnreadAdminCount + 1,
		})
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
		// Thread doesn't exist in chat_threads — check if it exists in support_threads
		var s models.SupportThread
		now := time.Now()
		participantName := "Shankar"
		if errS := mr.DB.Where("id = ?", req.ThreadID).First(&s).Error; errS == nil {
			if s.UserName != "" {
				participantName = s.UserName
			}
		}

		thread = models.ChatThread{
			ID:              req.ThreadID,
			UserID:          "aa7aa465-433c-49c8-9bb6-21eb94cd2a94",
			InstructorID:    senderID,
			InstructorName:  "Salen Dangol (Master Instructor)",
			ParticipantName: participantName,
			Subject:         "Hardware Diagnostics Inquiry",
			LastMessage:     req.Message,
			LastMessageAt:   now,
			UnreadCount:     0,
			CreatedAt:       now,
			UpdatedAt:       now,
		}
		mr.DB.Omit("Messages").Create(&thread)
	}

	now := time.Now()
	msg := models.ChatMessage{
		ID:          uuid.New().String(),
		ThreadID:    req.ThreadID,
		SenderID:    senderID,
		SenderName:  "Salen Dangol (Master Instructor)",
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
	thread.UnreadCount = 0
	mr.DB.Save(&thread)

	// Also mirror reply to support_messages if support thread exists
	var sThread models.SupportThread
	if err := mr.DB.Where("id = ?", req.ThreadID).First(&sThread).Error; err == nil {
		sMsg := models.SupportMessage{
			ID:         uuid.New().String(),
			ThreadID:   req.ThreadID,
			SenderID:   senderID,
			SenderName: "Salen Dangol (Master Instructor)",
			SenderRole: "instructor",
			Message:    req.Message,
			IsRead:     false,
			CreatedAt:  now,
		}
		mr.DB.Create(&sMsg)
		mr.DB.Model(&sThread).Updates(map[string]interface{}{
			"last_message":         req.Message,
			"last_message_at":      now,
			"updated_at":           now,
			"unread_student_count": sThread.UnreadStudentCount + 1,
			"unread_admin_count":   0,
		})
	}

	c.JSON(http.StatusOK, gin.H{
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

	// Also load support_threads to bridge legacy or parallel support tickets
	var supportThreads []models.SupportThread
	mr.DB.Order("updated_at desc").Find(&supportThreads)

	existingThreadIDs := make(map[string]bool)
	for i := range threads {
		existingThreadIDs[threads[i].ID] = true
	}

	for _, st := range supportThreads {
		if !existingThreadIDs[st.ID] {
			userName := st.UserName
			if userName == "" {
				userName = "Shankar"
			}
			threads = append(threads, models.ChatThread{
				ID:              st.ID,
				UserID:          st.UserID,
				InstructorID:    st.InstructorID,
				InstructorName:  st.InstructorName,
				ParticipantName: userName,
				Subject:         st.Subject,
				LastMessage:     st.LastMessage,
				LastMessageAt:   st.LastMessageAt,
				UnreadCount:     st.UnreadAdminCount,
				CreatedAt:       st.CreatedAt,
				UpdatedAt:       st.UpdatedAt,
			})
			existingThreadIDs[st.ID] = true
		}
	}

	// Cache user details to populate real names and email
	userMap := make(map[string]models.User)
	var allUsers []models.User
	mr.DB.Find(&allUsers)
	for _, u := range allUsers {
		userMap[u.ID] = u
		userMap[strings.ToLower(u.Email)] = u
	}

	for i := range threads {
		name := strings.TrimSpace(threads[i].ParticipantName)
		if name == "" || name == threads[i].UserID || strings.ToLower(name) == "student" || strings.ToLower(name) == "guest-tech" || (len(name) > 30 && strings.Contains(name, "-")) {
			if u, ok := userMap[threads[i].UserID]; ok && u.Name != "" {
				name = u.Name
			} else if u, ok := userMap[strings.ToLower(threads[i].UserID)]; ok && u.Name != "" {
				name = u.Name
			} else if strings.Contains(threads[i].UserID, "@") {
				name = strings.Split(threads[i].UserID, "@")[0]
			} else if threads[i].UserID == "aa7aa465-433c-49c8-9bb6-21eb94cd2a94" || threads[i].UserID == "guest-tech" {
				name = "Shankar"
			} else if threads[i].UserID == "u_ramesh" {
				name = "Ramesh KC"
			} else if threads[i].UserID == "u_suresh" || threads[i].UserID == "u_suresh_t" {
				name = "Suresh Tamang"
			} else if threads[i].UserID == "u_anil" {
				name = "Anil Kumar"
			} else if threads[i].UserID == "u_rohit" || threads[i].UserID == "u_rohit_k" {
				name = "Rohit Kumar"
			} else {
				name = "Shankar"
			}
			threads[i].ParticipantName = name
			mr.DB.Model(&models.ChatThread{}).Where("id = ?", threads[i].ID).Update("participant_name", name)
		}

		threads[i].Name = threads[i].ParticipantName
		threads[i].UserName = threads[i].ParticipantName
		if u, ok := userMap[threads[i].UserID]; ok {
			threads[i].UserEmail = u.Email
			threads[i].UserPhone = u.Phone
			threads[i].IsVIP = u.IsVIP || u.IsSubscribed || strings.ToLower(threads[i].ParticipantName) == "shankar"
		} else {
			threads[i].IsVIP = strings.ToLower(threads[i].ParticipantName) == "shankar"
		}
		if threads[i].IsVIP {
			threads[i].RegID = "SMG-TECH-2024-2885"
		}
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": threads})
}

func (mr *MessagingRouter) seedSampleThreads() {
	var count int64
	mr.DB.Model(&models.ChatThread{}).Count(&count)
	if count > 0 {
		// Update any existing thread whose participant_name is generic
		mr.DB.Model(&models.ChatThread{}).Where("participant_name = '' OR participant_name = 'guest-tech' OR participant_name = 'Student' OR participant_name = 'aa7aa465-433c-49c8-9bb6-21eb94cd2a94'").
			Update("participant_name", "Shankar")
		mr.DB.Model(&models.ChatThread{}).Where("user_id = 'u_ramesh'").Update("participant_name", "Ramesh KC")
		return
	}

	tID := uuid.New().String()
	thread := models.ChatThread{
		ID:              tID,
		UserID:          "aa7aa465-433c-49c8-9bb6-21eb94cd2a94",
		InstructorID:    "instructor-salen",
		InstructorName:  "Salen Dangol (Master Instructor)",
		ParticipantName: "Shankar",
		Subject:         "PMIC Short Circuit Help (Redmi Note 11)",
		LastMessage:     "Check coil L2001 with your multimeter in diode mode. Red probe to ground.",
		UnreadCount:     0,
		CreatedAt:       time.Now().Add(-5 * time.Hour),
		UpdatedAt:       time.Now().Add(-1 * time.Hour),
	}
	mr.DB.Omit("Messages").Create(&thread)

	messages := []models.ChatMessage{
		{
			ID:          uuid.New().String(),
			ThreadID:    tID,
			SenderID:    "aa7aa465-433c-49c8-9bb6-21eb94cd2a94",
			SenderName:  "Shankar",
			SenderRole:  "student",
			Message:     "Hello Sir! The device is drawing 0.050A stable on DC Power supply. What should I check first?",
			MessageText: "Hello Sir! The device is drawing 0.050A stable on DC Power supply. What should I check first?",
			IsRead:      true,
			CreatedAt:   time.Now().Add(-5 * time.Hour),
		},
		{
			ID:          uuid.New().String(),
			ThreadID:    tID,
			SenderID:    "instructor-salen",
			SenderName:  "Salen Dangol (Master Instructor)",
			SenderRole:  "instructor",
			Message:     "0.050A freeze indicates secondary buck line short or missing crystal oscillation. Check coil L2001 with your multimeter in diode mode. Red probe to ground.",
			MessageText: "0.050A freeze indicates secondary buck line short or missing crystal oscillation. Check coil L2001 with your multimeter in diode mode. Red probe to ground.",
			IsRead:      true,
			CreatedAt:   time.Now().Add(-1 * time.Hour),
		},
	}

	for _, m := range messages {
		mr.DB.Create(&m)
	}
}

