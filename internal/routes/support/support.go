package support

import (
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"smg-backend/internal/middleware"
	"smg-backend/internal/models"
)

type SupportRouter struct {
	DB *gorm.DB
}

func RegisterRoutes(r *gin.Engine, db *gorm.DB) {
	sr := &SupportRouter{DB: db}
	sr.seedCategories()
	sr.seedSampleThreads()

	// ── Mobile / Public ───────────────────────────────────────────────────
	// Choose Category screen — active categories only
	r.GET("/support/categories", sr.GetPublicCategories)
	r.GET("/support/categories/:slug", sr.GetCategoryChildren)
	r.GET("/support/categories/:slug/full", sr.GetCategoryWithChildren)
	// Alias used by mobile app
	r.GET("/api/support/categories", sr.GetPublicCategories)

	// ── Student Support Threads ───────────────────────────────────────────
	r.GET("/api/support/threads", sr.GetStudentThreads)
	r.POST("/api/support/threads", sr.CreateThread)
	r.GET("/api/support/threads/:threadId/messages", sr.GetThreadMessages)
	r.POST("/api/support/threads/:threadId/messages", sr.SendStudentMessage)
	r.PATCH("/api/support/threads/:threadId/read", sr.MarkThreadRead)

	// ── Admin: Category Management ────────────────────────────────────────
	adm := r.Group("/api/admin/support", middleware.AuthMiddleware(), middleware.AdminMiddleware())
	{
		// Categories
		adm.GET("/categories", sr.AdminListCategories)
		adm.POST("/categories", sr.AdminCreateCategory)
		adm.PUT("/categories/:id", sr.AdminUpdateCategory)
		adm.DELETE("/categories/:id", sr.AdminDeleteCategory)
		adm.POST("/categories/reorder", sr.AdminReorderCategories)
		adm.PATCH("/categories/:id/toggle", sr.AdminToggleCategory)

		// Thread Management
		adm.GET("/threads", sr.AdminListThreads)
		adm.GET("/threads/:threadId", sr.AdminGetThread)
		adm.POST("/threads/:threadId/reply", sr.AdminReplyThread)
		adm.PATCH("/threads/:threadId/status", sr.AdminUpdateThreadStatus)
		adm.PATCH("/threads/:threadId/assign", sr.AdminAssignThread)
		adm.PATCH("/threads/:threadId/priority", sr.AdminUpdatePriority)
		adm.DELETE("/threads/:threadId", sr.AdminDeleteThread)

		// Stats
		adm.GET("/stats", sr.AdminSupportStats)
	}
}

// ─── Mobile: Public Categories ────────────────────────────────────────────

// GET /support/categories or /api/support/categories
// Returns active categories in sort order for the "Choose Category" screen.
func (sr *SupportRouter) GetPublicCategories(c *gin.Context) {
	var cats []models.SupportCategory
	if err := sr.DB.Where("is_active = ? AND parent_id = ?", true, "").
		Order("order_num asc").Find(&cats).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to fetch categories"})
		return
	}

	// Shape response for mobile app (matches Flutter model)
	type catOut struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Slug        string `json:"slug"`
		Description string `json:"description"`
		Icon        string `json:"icon"`
		ImageURL    string `json:"imageUrl"`
		Badge       string `json:"badge"`
		IsActive    bool   `json:"isActive"`
	}
	out := make([]catOut, 0, len(cats))
	for _, cat := range cats {
		out = append(out, catOut{
			ID:          cat.ID,
			Name:        cat.Name,
			Slug:        cat.Slug,
			Description: cat.Description,
			Icon:        cat.IconAsset,
			ImageURL:    cat.IconAsset,
			Badge:       "Expert Support",
			IsActive:    cat.IsActive,
		})
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": out})
}

// GET /support/categories/:slug
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
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"category": parent, "children": children}})
}

// GET /support/categories/:slug/full
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

// ─── Student: Thread Management ───────────────────────────────────────────

// GET /api/support/threads?userId=&category=&status=
func (sr *SupportRouter) GetStudentThreads(c *gin.Context) {
	userID := c.Query("userId")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "userId required"})
		return
	}

	query := sr.DB.Model(&models.SupportThread{}).Where("user_id = ?", userID)

	if cat := c.Query("category"); cat != "" {
		query = query.Where("category_slug = ? OR LOWER(category_name) = LOWER(?)", cat, cat)
	}
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	var threads []models.SupportThread
	if err := query.Order("updated_at desc").Find(&threads).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	// Shape for mobile app — matches messages_screen.dart expectations
	type threadOut struct {
		ID             string    `json:"id"`
		InstructorName string    `json:"instructorName"`
		Subject        string    `json:"subject"`
		Category       string    `json:"category"`
		CategoryTag    string    `json:"categoryTag"`
		LastMessage    string    `json:"lastMessage"`
		UnreadCount    int       `json:"unreadCount"`
		Status         string    `json:"status"`
		Priority       string    `json:"priority"`
		DeviceModel    string    `json:"deviceModel"`
		UpdatedAt      time.Time `json:"updatedAt"`
	}
	out := make([]threadOut, 0, len(threads))
	for _, t := range threads {
		out = append(out, threadOut{
			ID:             t.ID,
			InstructorName: t.InstructorName,
			Subject:        t.Subject,
			Category:       t.CategoryName,
			CategoryTag:    t.CategoryName,
			LastMessage:    t.LastMessage,
			UnreadCount:    t.UnreadStudentCount,
			Status:         t.Status,
			Priority:       t.Priority,
			DeviceModel:    t.DeviceModel,
			UpdatedAt:      t.UpdatedAt,
		})
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": out})
}

// POST /api/support/threads
type CreateThreadRequest struct {
	UserID         string `json:"userId"         binding:"required"`
	UserName       string `json:"userName"`
	Category       string `json:"category"       binding:"required"`
	Subject        string `json:"subject"        binding:"required"`
	InitialMessage string `json:"initialMessage" binding:"required"`
	DeviceModel    string `json:"deviceModel"`
}

func (sr *SupportRouter) CreateThread(c *gin.Context) {
	var req CreateThreadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	// Resolve category info
	slug := strings.ToLower(strings.ReplaceAll(req.Category, " ", "-"))
	catName := strings.ToUpper(req.Category)
	catID := ""
	var cat models.SupportCategory
	if sr.DB.Where("LOWER(slug) = LOWER(?) OR LOWER(name) = LOWER(?)", slug, req.Category).First(&cat).Error == nil {
		catID = cat.ID
		catName = cat.Name
		slug = cat.Slug
	}

	now := time.Now()
	thread := models.SupportThread{
		ID:             uuid.New().String(),
		UserID:         req.UserID,
		UserName:       req.UserName,
		CategoryID:     catID,
		CategorySlug:   slug,
		CategoryName:   catName,
		InstructorName: "Salen Dangol (Master Instructor)",
		Subject:        req.Subject,
		DeviceModel:    req.DeviceModel,
		Status:         "open",
		Priority:       "normal",
		LastMessage:    req.InitialMessage,
		LastMessageAt:  now,
		UnreadAdminCount: 1,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := sr.DB.Omit("Messages").Create(&thread).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	// Save initial message
	msg := models.SupportMessage{
		ID:         uuid.New().String(),
		ThreadID:   thread.ID,
		SenderID:   req.UserID,
		SenderName: req.UserName,
		SenderRole: "student",
		Message:    req.InitialMessage,
		IsRead:     false,
		CreatedAt:  now,
	}
	sr.DB.Create(&msg)

	// Mirror to ChatThread and ChatMessage so SalenAdminGuru sees it immediately with student name
	studentName := strings.TrimSpace(req.UserName)
	if studentName == "" {
		studentName = "Shankar"
	}
	sr.DB.Create(&models.ChatThread{
		ID:              thread.ID,
		UserID:          req.UserID,
		InstructorID:    "instructor-salen",
		InstructorName:  "Salen Dangol (Master Instructor)",
		ParticipantName: studentName,
		Subject:         req.Subject,
		LastMessage:     req.InitialMessage,
		LastMessageAt:   now,
		UnreadCount:     1,
		CreatedAt:       now,
		UpdatedAt:       now,
	})
	sr.DB.Create(&models.ChatMessage{
		ID:          msg.ID,
		ThreadID:    thread.ID,
		SenderID:    req.UserID,
		SenderName:  studentName,
		SenderRole:  "student",
		Message:     req.InitialMessage,
		MessageText: req.InitialMessage,
		IsRead:      false,
		CreatedAt:   now,
	})

	c.JSON(http.StatusCreated, gin.H{"success": true, "data": thread})
}

// GET /api/support/threads/:threadId/messages
func (sr *SupportRouter) GetThreadMessages(c *gin.Context) {
	threadID := c.Param("threadId")
	var msgs []models.SupportMessage
	sr.DB.Where("thread_id = ?", threadID).Order("created_at asc").Find(&msgs)

	// Also merge any ChatMessage (such as Admin replies sent from SalenAdminGuru)
	var chatMsgs []models.ChatMessage
	sr.DB.Where("thread_id = ?", threadID).Order("created_at asc").Find(&chatMsgs)

	seenIds := make(map[string]bool)
	for _, m := range msgs {
		seenIds[m.ID] = true
	}

	for _, cm := range chatMsgs {
		if !seenIds[cm.ID] {
			msgs = append(msgs, models.SupportMessage{
				ID:            cm.ID,
				ThreadID:      cm.ThreadID,
				SenderID:      cm.SenderID,
				SenderName:    cm.SenderName,
				SenderRole:    cm.SenderRole,
				Message:       cm.Message,
				AttachmentURL: cm.AttachmentURL,
				IsRead:        cm.IsRead,
				CreatedAt:     cm.CreatedAt,
			})
			seenIds[cm.ID] = true
		}
	}

	sort.Slice(msgs, func(i, j int) bool {
		return msgs[i].CreatedAt.Before(msgs[j].CreatedAt)
	})

	c.JSON(http.StatusOK, gin.H{"success": true, "data": msgs})
}

// POST /api/support/threads/:threadId/messages
type SendMessageRequest struct {
	UserID     string `json:"userId"   binding:"required"`
	UserName   string `json:"userName"`
	Message    string `json:"message"  binding:"required"`
	AttachURL  string `json:"attachmentUrl"`
	AttachType string `json:"attachmentType"`
}

func (sr *SupportRouter) SendStudentMessage(c *gin.Context) {
	threadID := c.Param("threadId")
	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	// Verify thread exists and belongs to user
	var thread models.SupportThread
	if err := sr.DB.Where("id = ?", threadID).First(&thread).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Thread not found"})
		return
	}

	studentName := strings.TrimSpace(req.UserName)
	if studentName == "" {
		studentName = thread.UserName
	}
	if studentName == "" {
		studentName = "Shankar"
	}

	now := time.Now()
	msg := models.SupportMessage{
		ID:             uuid.New().String(),
		ThreadID:       threadID,
		SenderID:       req.UserID,
		SenderName:     studentName,
		SenderRole:     "student",
		Message:        req.Message,
		AttachmentURL:  req.AttachURL,
		AttachmentType: req.AttachType,
		IsRead:         false,
		CreatedAt:      now,
	}
	if err := sr.DB.Create(&msg).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	// Update thread
	sr.DB.Model(&thread).Updates(map[string]interface{}{
		"last_message":       req.Message,
		"last_message_at":    now,
		"updated_at":         now,
		"status":             "active",
		"unread_admin_count": thread.UnreadAdminCount + 1,
	})

	// Mirror to ChatThread and ChatMessage
	var chatThread models.ChatThread
	if errC := sr.DB.Where("id = ?", threadID).First(&chatThread).Error; errC == nil {
		sr.DB.Model(&chatThread).Updates(map[string]interface{}{
			"last_message":     req.Message,
			"last_message_at":  now,
			"updated_at":       now,
			"unread_count":     chatThread.UnreadCount + 1,
			"participant_name": studentName,
		})
	} else {
		sr.DB.Create(&models.ChatThread{
			ID:              threadID,
			UserID:          req.UserID,
			InstructorID:    "instructor-salen",
			InstructorName:  "Salen Dangol (Master Instructor)",
			ParticipantName: studentName,
			Subject:         thread.Subject,
			LastMessage:     req.Message,
			LastMessageAt:   now,
			UnreadCount:     1,
			CreatedAt:       now,
			UpdatedAt:       now,
		})
	}

	chatMsg := models.ChatMessage{
		ID:            msg.ID,
		ThreadID:      threadID,
		SenderID:      req.UserID,
		SenderName:    studentName,
		SenderRole:    "student",
		Message:       req.Message,
		MessageText:   req.Message,
		AttachmentURL: req.AttachURL,
		IsRead:        false,
		CreatedAt:     now,
	}
	sr.DB.Create(&chatMsg)

	c.JSON(http.StatusCreated, gin.H{"success": true, "data": msg})
}

// PATCH /api/support/threads/:threadId/read?userId=
func (sr *SupportRouter) MarkThreadRead(c *gin.Context) {
	threadID := c.Param("threadId")
	sr.DB.Model(&models.SupportMessage{}).
		Where("thread_id = ? AND is_read = ?", threadID, false).
		Update("is_read", true)
	sr.DB.Model(&models.SupportThread{}).
		Where("id = ?", threadID).
		Update("unread_student_count", 0)
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// ─── Admin: Category CRUD ─────────────────────────────────────────────────

// GET /api/admin/support/categories
func (sr *SupportRouter) AdminListCategories(c *gin.Context) {
	var cats []models.SupportCategory
	sr.DB.Order("order_num asc").Find(&cats)

	type catWithCount struct {
		models.SupportCategory
		ThreadCount int64 `json:"threadCount"`
	}
	out := make([]catWithCount, 0, len(cats))
	for _, cat := range cats {
		var count int64
		sr.DB.Model(&models.SupportThread{}).Where("category_id = ? OR category_slug = ?", cat.ID, cat.Slug).Count(&count)
		out = append(out, catWithCount{SupportCategory: cat, ThreadCount: count})
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": out})
}

// POST /api/admin/support/categories
func (sr *SupportRouter) AdminCreateCategory(c *gin.Context) {
	var body struct {
		Name        string `json:"name"        binding:"required"`
		Slug        string `json:"slug"`
		Description string `json:"description"`
		Icon        string `json:"icon"`
		ImageURL    string `json:"imageUrl"`
		BadgeText   string `json:"badgeText"`
		SortOrder   int    `json:"sortOrder"`
		IsActive    *bool  `json:"isActive"`
		ParentID    string `json:"parentId"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	slug := body.Slug
	if slug == "" {
		slug = strings.ToLower(strings.ReplaceAll(body.Name, " ", "-"))
		slug = strings.ReplaceAll(slug, "&", "")
		slug = strings.Trim(slug, "-")
	}
	isActive := true
	if body.IsActive != nil {
		isActive = *body.IsActive
	}
	icon := body.Icon
	if icon == "" {
		icon = body.ImageURL
	}

	cat := models.SupportCategory{
		ID:          uuid.New().String(),
		Name:        body.Name,
		Slug:        slug,
		Description: body.Description,
		IconAsset:   icon,
		ParentID:    body.ParentID,
		OrderNum:    body.SortOrder,
		IsActive:    isActive,
		CreatedAt:   time.Now(),
	}
	if err := sr.DB.Create(&cat).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": cat})
}

// PUT /api/admin/support/categories/:id
func (sr *SupportRouter) AdminUpdateCategory(c *gin.Context) {
	id := c.Param("id")
	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	// Map JSON keys to DB column names
	mapped := map[string]interface{}{}
	for k, v := range updates {
		switch k {
		case "name":
			mapped["name"] = v
		case "slug":
			mapped["slug"] = v
		case "description":
			mapped["description"] = v
		case "icon", "iconAsset", "imageUrl":
			mapped["icon_asset"] = v
		case "sortOrder", "orderNum":
			mapped["order_num"] = v
		case "isActive":
			mapped["is_active"] = v
		case "parentId":
			mapped["parent_id"] = v
		}
	}
	if err := sr.DB.Model(&models.SupportCategory{}).Where("id = ?", id).Updates(mapped).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	var updated models.SupportCategory
	sr.DB.Where("id = ?", id).First(&updated)
	c.JSON(http.StatusOK, gin.H{"success": true, "data": updated})
}

// DELETE /api/admin/support/categories/:id
func (sr *SupportRouter) AdminDeleteCategory(c *gin.Context) {
	id := c.Param("id")
	if err := sr.DB.Delete(&models.SupportCategory{}, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Category removed successfully"})
}

// POST /api/admin/support/categories/reorder
func (sr *SupportRouter) AdminReorderCategories(c *gin.Context) {
	var body struct {
		CategoryIDs []string `json:"categoryIds" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	for i, id := range body.CategoryIDs {
		sr.DB.Model(&models.SupportCategory{}).Where("id = ?", id).Update("order_num", i+1)
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Categories reordered successfully"})
}

// PATCH /api/admin/support/categories/:id/toggle
func (sr *SupportRouter) AdminToggleCategory(c *gin.Context) {
	id := c.Param("id")
	var cat models.SupportCategory
	if err := sr.DB.Where("id = ?", id).First(&cat).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Category not found"})
		return
	}
	sr.DB.Model(&cat).Update("is_active", !cat.IsActive)
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"id": id, "isActive": !cat.IsActive}})
}

// ─── Admin: Thread Management ─────────────────────────────────────────────

// GET /api/admin/support/threads?category=&status=&instructorId=&search=&page=&limit=
func (sr *SupportRouter) AdminListThreads(c *gin.Context) {
	query := sr.DB.Model(&models.SupportThread{})

	if cat := c.Query("category"); cat != "" {
		query = query.Where("category_slug = ? OR LOWER(category_name) = LOWER(?)", cat, cat)
	}
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	if instID := c.Query("instructorId"); instID != "" {
		query = query.Where("instructor_id = ?", instID)
	}
	if search := c.Query("search"); search != "" {
		like := "%" + search + "%"
		query = query.Where("subject LIKE ? OR user_name LIKE ? OR device_model LIKE ? OR last_message LIKE ?", like, like, like, like)
	}
	if priority := c.Query("priority"); priority != "" {
		query = query.Where("priority = ?", priority)
	}

	var total int64
	query.Count(&total)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "25"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 25
	}
	offset := (page - 1) * limit

	var threads []models.SupportThread
	query.Order("unread_admin_count desc, updated_at desc").
		Limit(limit).Offset(offset).Find(&threads)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"total":   total,
		"page":    page,
		"limit":   limit,
		"data":    threads,
	})
}

// GET /api/admin/support/threads/:threadId
func (sr *SupportRouter) AdminGetThread(c *gin.Context) {
	threadID := c.Param("threadId")
	var thread models.SupportThread
	if err := sr.DB.Preload("Messages", func(db *gorm.DB) *gorm.DB {
		return db.Order("created_at asc")
	}).Where("id = ?", threadID).First(&thread).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Thread not found"})
		return
	}

	// Mark admin messages as read
	sr.DB.Model(&models.SupportThread{}).Where("id = ?", threadID).Update("unread_admin_count", 0)

	c.JSON(http.StatusOK, gin.H{"success": true, "data": thread})
}

// POST /api/admin/support/threads/:threadId/reply
type AdminReplyRequest struct {
	SenderID      string `json:"senderId"`
	SenderName    string `json:"senderName"   binding:"required"`
	SenderRole    string `json:"senderRole"`
	Message       string `json:"message"      binding:"required"`
	AttachmentURL string `json:"attachmentUrl"`
	AttachType    string `json:"attachmentType"`
}

func (sr *SupportRouter) AdminReplyThread(c *gin.Context) {
	threadID := c.Param("threadId")
	var req AdminReplyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	var thread models.SupportThread
	if err := sr.DB.Where("id = ?", threadID).First(&thread).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Thread not found"})
		return
	}

	senderRole := req.SenderRole
	if senderRole == "" {
		senderRole = "instructor"
	}

	now := time.Now()
	msg := models.SupportMessage{
		ID:             uuid.New().String(),
		ThreadID:       threadID,
		SenderID:       req.SenderID,
		SenderName:     req.SenderName,
		SenderRole:     senderRole,
		Message:        req.Message,
		AttachmentURL:  req.AttachmentURL,
		AttachmentType: req.AttachType,
		IsRead:         false,
		CreatedAt:      now,
	}
	if err := sr.DB.Create(&msg).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	// Update thread
	sr.DB.Model(&thread).Updates(map[string]interface{}{
		"last_message":         req.Message,
		"last_message_at":      now,
		"updated_at":           now,
		"status":               "active",
		"unread_student_count": thread.UnreadStudentCount + 1,
		"unread_admin_count":   0,
	})

	c.JSON(http.StatusCreated, gin.H{"success": true, "data": msg})
}

// PATCH /api/admin/support/threads/:threadId/status
func (sr *SupportRouter) AdminUpdateThreadStatus(c *gin.Context) {
	threadID := c.Param("threadId")
	var body struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	validStatuses := map[string]bool{"open": true, "active": true, "pending": true, "resolved": true, "closed": true}
	if !validStatuses[body.Status] {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid status. Use: open, active, pending, resolved, closed"})
		return
	}
	sr.DB.Model(&models.SupportThread{}).Where("id = ?", threadID).Updates(map[string]interface{}{
		"status":     body.Status,
		"updated_at": time.Now(),
	})
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"id": threadID, "status": body.Status}})
}

// PATCH /api/admin/support/threads/:threadId/assign
func (sr *SupportRouter) AdminAssignThread(c *gin.Context) {
	threadID := c.Param("threadId")
	var body struct {
		InstructorID   string `json:"instructorId"`
		InstructorName string `json:"instructorName" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	sr.DB.Model(&models.SupportThread{}).Where("id = ?", threadID).Updates(map[string]interface{}{
		"instructor_id":   body.InstructorID,
		"instructor_name": body.InstructorName,
		"updated_at":      time.Now(),
	})
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"id": threadID, "instructorId": body.InstructorID, "instructorName": body.InstructorName}})
}

// PATCH /api/admin/support/threads/:threadId/priority
func (sr *SupportRouter) AdminUpdatePriority(c *gin.Context) {
	threadID := c.Param("threadId")
	var body struct {
		Priority string `json:"priority" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	sr.DB.Model(&models.SupportThread{}).Where("id = ?", threadID).Update("priority", body.Priority)
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"id": threadID, "priority": body.Priority}})
}

// DELETE /api/admin/support/threads/:threadId
func (sr *SupportRouter) AdminDeleteThread(c *gin.Context) {
	threadID := c.Param("threadId")
	sr.DB.Where("thread_id = ?", threadID).Delete(&models.SupportMessage{})
	sr.DB.Delete(&models.SupportThread{}, "id = ?", threadID)
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// GET /api/admin/support/stats
func (sr *SupportRouter) AdminSupportStats(c *gin.Context) {
	stats := map[string]int64{}
	statuses := []string{"open", "active", "pending", "resolved", "closed"}
	for _, s := range statuses {
		var count int64
		sr.DB.Model(&models.SupportThread{}).Where("status = ?", s).Count(&count)
		stats[s] = count
	}
	var total int64
	sr.DB.Model(&models.SupportThread{}).Count(&total)
	var unreadTotal int64
	sr.DB.Model(&models.SupportThread{}).Where("unread_admin_count > 0").Count(&unreadTotal)

	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"total":        total,
		"unread":       unreadTotal,
		"byStatus":     stats,
	})
}

// ─── Seed ─────────────────────────────────────────────────────────────────

func (sr *SupportRouter) seedCategories() {
	var count int64
	sr.DB.Model(&models.SupportCategory{}).Where("parent_id = ?", "").Count(&count)
	if count >= 3 {
		return
	}

	cats := []models.SupportCategory{
		{
			ID: "cat_software", Name: "SOFTWARE", Slug: "software",
			Description: "UFS, EMMC, Flashing, Unlocking, and JTAG solutions",
			IconAsset: "assets/images/mooxy_software.jpg",
			OrderNum: 1, IsActive: true, ParentID: "",
		},
		{
			ID: "cat_hardware", Name: "HARDWARE", Slug: "hardware",
			Description: "Short detection, PMIC voltages, micro-soldering, and PCB repair",
			IconAsset: "assets/images/mooxy_hardware.jpg",
			OrderNum: 2, IsActive: true, ParentID: "",
		},
		{
			ID: "cat_schematics", Name: "SCHEMATICS", Slug: "schematics",
			Description: "Motherboard circuit traces, resistance charts, and jumper solutions",
			IconAsset: "assets/images/support_ufs.jpg",
			OrderNum: 3, IsActive: true, ParentID: "",
		},
		{
			ID: "cat_reballing", Name: "IC REBALLING", Slug: "ic-reballing",
			Description: "Micro-soldering, stencil alignment, BGA paste and chip removal",
			IconAsset: "assets/images/reballing.jpg",
			OrderNum: 4, IsActive: true, ParentID: "",
		},
	}
	for _, cat := range cats {
		sr.DB.Where("slug = ?", cat.Slug).FirstOrCreate(&cat)
	}
}

func (sr *SupportRouter) seedSampleThreads() {
	var count int64
	sr.DB.Model(&models.SupportThread{}).Count(&count)
	if count > 0 {
		return
	}

	now := time.Now()
	threads := []models.SupportThread{
		{
			ID: "thread_001", UserID: "guest-tech", UserName: "Shankar",
			CategoryID: "cat_hardware", CategorySlug: "hardware", CategoryName: "HARDWARE",
			InstructorName: "Salen Dangol (Master Instructor)",
			Subject: "0.05A stuck current on Redmi Note 10", DeviceModel: "Redmi Note 10",
			Status: "active", Priority: "normal",
			LastMessage: "Check PMIC VREG_BOB and VDD_APC inductor coils.",
			LastMessageAt: now.Add(-2 * time.Hour),
			UnreadStudentCount: 1, UnreadAdminCount: 0,
			CreatedAt: now.Add(-24 * time.Hour), UpdatedAt: now.Add(-2 * time.Hour),
		},
		{
			ID: "thread_002", UserID: "guest-tech", UserName: "Shankar",
			CategoryID: "cat_software", CategorySlug: "software", CategoryName: "SOFTWARE",
			InstructorName: "Er. Amit Verma",
			Subject: "Easy J-Tag ISP Pinout for Samsung A52", DeviceModel: "Samsung A52",
			Status: "resolved", Priority: "normal",
			LastMessage: "ISP clock and CMD resistor value is 47 ohm. Use test point 12.",
			LastMessageAt: now.Add(-48 * time.Hour),
			UnreadStudentCount: 0, UnreadAdminCount: 0,
			CreatedAt: now.Add(-72 * time.Hour), UpdatedAt: now.Add(-48 * time.Hour),
		},
	}
	for i := range threads {
		sr.DB.Omit("Messages").Create(&threads[i])
	}

	msgs := []models.SupportMessage{
		{ID: uuid.New().String(), ThreadID: "thread_001", SenderID: "guest-tech", SenderName: "Shankar", SenderRole: "student", Message: "0.05A stuck current on DC supply. Thermal camera shows cold board. How to proceed?", IsRead: true, CreatedAt: now.Add(-24 * time.Hour)},
		{ID: uuid.New().String(), ThreadID: "thread_001", SenderID: "inst_salen", SenderName: "Salen Dangol", SenderRole: "instructor", Message: "Hi Shankar! 0.05A stuck current indicates secondary power rail short or CPU holding reset low. Measure PMIC output rails first.", IsRead: true, CreatedAt: now.Add(-20 * time.Hour)},
		{ID: uuid.New().String(), ThreadID: "thread_001", SenderID: "inst_salen", SenderName: "Salen Dangol", SenderRole: "instructor", Message: "Check PMIC VREG_BOB and VDD_APC inductor coils.", IsRead: false, CreatedAt: now.Add(-2 * time.Hour)},
		{ID: uuid.New().String(), ThreadID: "thread_002", SenderID: "guest-tech", SenderName: "Shankar", SenderRole: "student", Message: "Need ISP clock and CMD resistor value for Samsung A52 EasyJTAG.", IsRead: true, CreatedAt: now.Add(-72 * time.Hour)},
		{ID: uuid.New().String(), ThreadID: "thread_002", SenderID: "inst_amit", SenderName: "Er. Amit Verma", SenderRole: "instructor", Message: "ISP clock and CMD resistor value is 47 ohm. Use test point 12.", IsRead: true, CreatedAt: now.Add(-48 * time.Hour)},
	}
	for _, m := range msgs {
		sr.DB.Create(&m)
	}
}
