package admin

import (
	_ "embed"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"smg-backend/internal/middleware"
	"smg-backend/internal/models"
)

//go:embed static/admin.html
var adminHTML []byte

//go:embed static/classroom.html
var classroomHTML []byte

type AdminRouter struct {
	DB *gorm.DB
}

func RegisterRoutes(r *gin.Engine, db *gorm.DB) {
	ar := &AdminRouter{DB: db}

	// 1. Web Portal (public — serves embedded HTML)
	r.GET("/admin", ar.ServeAdminWeb)
	r.HEAD("/admin", ar.ServeAdminWeb)
	r.HEAD("/admin/*any", ar.ServeAdminWeb)
	r.GET("/admin/*any", ar.ServeAdminWeb)
	r.GET("/classroom", ar.ServeClassroomWeb)
	r.GET("/classroom/*any", ar.ServeClassroomWeb)
	r.GET("/learn/*any", ar.ServeClassroomWeb)

	// 2. Public admin login (no auth required)
	r.POST("/api/admin/login", ar.Login)

	// 3. Protected admin REST APIs (JWT + admin role required)
	adm := r.Group("/api/admin", middleware.AuthMiddleware(), middleware.AdminMiddleware())
	{
		adm.GET("/stats", ar.GetStats)

		// Courses & Curriculum Builder
		adm.GET("/courses", ar.ListCourses)
		adm.POST("/courses", ar.CreateCourse)
		adm.PUT("/courses/:id", ar.UpdateCourse)
		adm.DELETE("/courses/:id", ar.DeleteCourse)
		adm.POST("/courses/:id/modules", ar.AddModule)
		adm.PUT("/modules/:id", ar.UpdateModule)
		adm.DELETE("/modules/:id", ar.DeleteModule)
		adm.POST("/modules/:id/lessons", ar.AddLesson)
		adm.PUT("/lessons/:id", ar.UpdateLesson)
		adm.DELETE("/lessons/:id", ar.DeleteLesson)
		adm.GET("/courses/:id/curriculum", ar.GetCourseCurriculum)

		// Subscriptions & Payment Approvals
		adm.GET("/subscriptions", ar.ListSubscriptions)
		adm.POST("/subscriptions/:id/approve", ar.ApproveSubscription)
		adm.POST("/subscriptions/:id/reject", ar.RejectSubscription)
		adm.POST("/subscriptions/manual-enroll", ar.ManualEnroll)

		// Students / Users
		adm.GET("/users", ar.ListUsers)
		adm.POST("/users/:id/toggle-vip", ar.ToggleUserVip)

		// Community Moderation
		adm.GET("/community", ar.ListCommunity)
		adm.DELETE("/community/:id", ar.DeleteCommunityPost)
		adm.POST("/community/:id/solve", ar.MarkSolved)

		// Live Masterclasses
		adm.GET("/live-sessions", ar.ListLiveSessions)
		adm.POST("/live-sessions", ar.CreateLiveSession)
		adm.DELETE("/live-sessions/:id", ar.DeleteLiveSession)

		// Mobile App CMS
		adm.GET("/cms/banners", ar.GetBanners)
		adm.POST("/cms/banners", ar.UpdateBanners)
	}
}

func (ar *AdminRouter) ServeAdminWeb(c *gin.Context) {
	c.Data(http.StatusOK, "text/html; charset=utf-8", adminHTML)
}

func (ar *AdminRouter) ServeClassroomWeb(c *gin.Context) {
	c.Data(http.StatusOK, "text/html; charset=utf-8", classroomHTML)
}

type AdminLoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (ar *AdminRouter) Login(c *gin.Context) {
	var req AdminLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid login data"})
		return
	}

	email := strings.TrimSpace(strings.ToLower(req.Email))

	var user models.User
	if err := ar.DB.Where("LOWER(email) = ?", email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Invalid email or password"})
		return
	}

	if !strings.EqualFold(user.Role, "ADMIN") {
		c.JSON(http.StatusForbidden, gin.H{"success": false, "error": "Admin access required"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Invalid email or password"})
		return
	}

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "salen-mobile-guru-secret-key-2026"
	}

	claims := jwt.MapClaims{
		"userId": user.ID,
		"email":  user.Email,
		"role":   user.Role,
		"iat":    time.Now().Unix(),
		"exp":    time.Now().Add(12 * time.Hour).Unix(), // shorter-lived admin token
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(secret))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Could not issue token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"token":   tokenStr,
		"admin": gin.H{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
			"role":  user.Role,
		},
	})
}

func (ar *AdminRouter) GetStats(c *gin.Context) {
	var totalUsers int64
	ar.DB.Model(&models.User{}).Count(&totalUsers)

	var totalCourses int64
	ar.DB.Model(&models.Course{}).Count(&totalCourses)

	var totalSubs int64
	ar.DB.Model(&models.Subscription{}).Where("status = ?", "active").Count(&totalSubs)

	var totalPosts int64
	ar.DB.Model(&models.CommunityPost{}).Count(&totalPosts)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"stats": gin.H{
			"students":            totalUsers,
			"courses":             totalCourses,
			"subscriptions":       totalSubs,
			"communityPosts":      totalPosts,
			"monthlyRevenueNpr":   745000,
			"activePlatformUsers": 482,
		},
	})
}

func (ar *AdminRouter) ListCourses(c *gin.Context) {
	var courses []models.Course
	ar.DB.Preload("Modules.Lessons").Order("created_at desc").Find(&courses)
	c.JSON(http.StatusOK, gin.H{"success": true, "courses": courses, "data": courses})
}

func (ar *AdminRouter) CreateCourse(c *gin.Context) {
	var req models.Course
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	if req.ID == "" {
		req.ID = uuid.New().String()
	}
	if req.Slug == "" {
		req.Slug = strings.ToLower(strings.ReplaceAll(req.Title, " ", "-"))
	}
	req.CreatedAt = time.Now()
	req.UpdatedAt = time.Now()

	if err := ar.DB.Create(&req).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "course": req, "data": req})
}

func (ar *AdminRouter) DeleteCourse(c *gin.Context) {
	id := c.Param("id")
	if err := ar.DB.Where("id = ?", id).Delete(&models.Course{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Course deleted"})
}

type AddModuleRequest struct {
	Title    string `json:"title" binding:"required"`
	OrderNum int    `json:"orderNum"`
}

func (ar *AdminRouter) AddModule(c *gin.Context) {
	courseID := c.Param("id")
	var req AddModuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	module := models.CourseModule{
		ID:        uuid.New().String(),
		CourseID:  courseID,
		Title:     req.Title,
		OrderNum:  req.OrderNum,
		CreatedAt: time.Now(),
	}

	if err := ar.DB.Create(&module).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "module": module})
}

type AddLessonRequest struct {
	Title         string `json:"title" binding:"required"`
	Slug          string `json:"slug"`
	Duration      string `json:"duration"`
	VideoURL      string `json:"videoUrl"`
	StreamType    string `json:"streamType"`
	IsFreePreview bool   `json:"isFreePreview"`
	OrderNum      int    `json:"orderNum"`
}

func (ar *AdminRouter) AddLesson(c *gin.Context) {
	moduleID := c.Param("id")
	var req AddLessonRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	if req.Slug == "" {
		req.Slug = strings.ToLower(strings.ReplaceAll(req.Title, " ", "-"))
	}
	if req.StreamType == "" {
		req.StreamType = "mp4"
	}

	lesson := models.Lesson{
		ID:            uuid.New().String(),
		ModuleID:      moduleID,
		Title:         req.Title,
		Slug:          req.Slug,
		Duration:      req.Duration,
		VideoURL:      req.VideoURL,
		StreamType:    req.StreamType,
		IsFreePreview: req.IsFreePreview,
		HasVideo:      req.VideoURL != "",
		OrderNum:      req.OrderNum,
		CreatedAt:     time.Now(),
	}

	if err := ar.DB.Create(&lesson).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "lesson": lesson})
}

func (ar *AdminRouter) DeleteLesson(c *gin.Context) {
	id := c.Param("id")
	if err := ar.DB.Where("id = ?", id).Delete(&models.Lesson{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Lesson deleted"})
}

func (ar *AdminRouter) GetCourseCurriculum(c *gin.Context) {
	courseID := c.Param("id")
	var modules []models.CourseModule
	ar.DB.Preload("Lessons").Where("course_id = ?", courseID).Order("order_num asc").Find(&modules)
	c.JSON(http.StatusOK, gin.H{"success": true, "curriculum": modules, "data": modules})
}

func (ar *AdminRouter) ListSubscriptions(c *gin.Context) {
	var subs []models.Subscription
	ar.DB.Order("created_at desc").Find(&subs)
	c.JSON(http.StatusOK, gin.H{"success": true, "subscriptions": subs, "data": subs})
}

func (ar *AdminRouter) ApproveSubscription(c *gin.Context) {
	id := c.Param("id")
	if err := ar.DB.Model(&models.Subscription{}).Where("id = ?", id).Update("status", "active").Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Subscription approved"})
}

func (ar *AdminRouter) RejectSubscription(c *gin.Context) {
	id := c.Param("id")
	if err := ar.DB.Model(&models.Subscription{}).Where("id = ?", id).Update("status", "cancelled").Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Subscription cancelled"})
}

type ManualEnrollRequest struct {
	UserID     string `json:"userId" binding:"required"`
	CourseSlug string `json:"courseSlug" binding:"required"`
	PlanType   string `json:"planType"`
}

func (ar *AdminRouter) ManualEnroll(c *gin.Context) {
	var req ManualEnrollRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	if req.PlanType == "" {
		req.PlanType = "VIP"
	}

	sub := models.Subscription{
		ID:            uuid.New().String(),
		UserID:        req.UserID,
		PlanType:      req.PlanType,
		Status:        "active",
		StartDate:     time.Now(),
		EndDate:       time.Now().AddDate(1, 0, 0),
		PaymentMethod: "MANUAL_ADMIN",
		TransactionID: "ADM-" + uuid.New().String()[:8],
		Amount:        0,
		Currency:      "NPR",
		EnrolledSlugs: req.CourseSlug,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := ar.DB.Create(&sub).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "subscription": sub})
}

func (ar *AdminRouter) ListUsers(c *gin.Context) {
	var users []models.User
	ar.DB.Order("created_at desc").Find(&users)
	c.JSON(http.StatusOK, gin.H{"success": true, "users": users, "data": users})
}

func (ar *AdminRouter) ToggleUserVip(c *gin.Context) {
	id := c.Param("id")
	var user models.User
	if err := ar.DB.Where("id = ?", id).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "User not found"})
		return
	}

	user.IsVIP = !user.IsVIP
	ar.DB.Save(&user)
	c.JSON(http.StatusOK, gin.H{"success": true, "isVip": user.IsVIP})
}

func (ar *AdminRouter) ListCommunity(c *gin.Context) {
	var posts []models.CommunityPost
	ar.DB.Preload("Replies").Order("created_at desc").Find(&posts)
	c.JSON(http.StatusOK, gin.H{"success": true, "posts": posts, "data": posts})
}

func (ar *AdminRouter) DeleteCommunityPost(c *gin.Context) {
	id := c.Param("id")
	ar.DB.Where("post_id = ?", id).Delete(&models.CommunityReply{})
	ar.DB.Where("id = ?", id).Delete(&models.CommunityPost{})
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Post deleted"})
}

func (ar *AdminRouter) MarkSolved(c *gin.Context) {
	id := c.Param("id")
	ar.DB.Model(&models.CommunityPost{}).Where("id = ?", id).Update("is_solved", true)
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Post marked as solved"})
}

func (ar *AdminRouter) ListLiveSessions(c *gin.Context) {
	var sessions []models.LiveSession
	ar.DB.Order("created_at desc").Find(&sessions)
	c.JSON(http.StatusOK, gin.H{"success": true, "sessions": sessions, "data": sessions})
}

func (ar *AdminRouter) CreateLiveSession(c *gin.Context) {
	var session models.LiveSession
	if err := c.ShouldBindJSON(&session); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	session.ID = uuid.New().String()
	session.CreatedAt = time.Now()
	if err := ar.DB.Create(&session).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "session": session, "data": session})
}

func (ar *AdminRouter) DeleteLiveSession(c *gin.Context) {
	id := c.Param("id")
	ar.DB.Where("id = ?", id).Delete(&models.LiveSession{})
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Live session deleted"})
}

func (ar *AdminRouter) GetBanners(c *gin.Context) {
	var screen models.AppScreenConfig
	err := ar.DB.Where("site_name = ? AND screen_key = ?", "smg", "guest_home").First(&screen).Error
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"banners": []string{
				"assets/banner_bg_img/banner_1.webp",
				"assets/banner_bg_img/banner_2.webp",
				"assets/banner_bg_img/banner_3.webp",
				"assets/banner_bg_img/banner_4.webp",
				"assets/banner_bg_img/banner_5.webp",
			},
		})
		return
	}

	var data map[string]interface{}
	json.Unmarshal([]byte(screen.ConfigJSON), &data)
	c.JSON(http.StatusOK, gin.H{"success": true, "banners": data["banners"]})
}

type UpdateBannersRequest struct {
	Banners []string `json:"banners" binding:"required"`
}

func (ar *AdminRouter) UpdateBanners(c *gin.Context) {
	var req UpdateBannersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	payload := map[string]interface{}{
		"banners":     req.Banners,
		"bannerImage": req.Banners[0],
	}
	bytes, _ := json.Marshal(payload)

	ar.DB.Model(&models.AppScreenConfig{}).
		Where("site_name = ? AND screen_key = ?", "smg", "guest_home").
		Update("config_json", string(bytes))

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Banners updated successfully"})
}

func (ar *AdminRouter) ListMedia(c *gin.Context) {
	var media []models.MediaAsset
	ar.DB.Order("created_at desc").Find(&media)
	c.JSON(http.StatusOK, gin.H{"success": true, "media": media, "data": media})
}

func (ar *AdminRouter) CreateMedia(c *gin.Context) {
	var req models.MediaAsset
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	req.ID = uuid.New().String()
	req.CreatedAt = time.Now()
	if err := ar.DB.Create(&req).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "media": req, "data": req})
}

func (ar *AdminRouter) DeleteMedia(c *gin.Context) {
	id := c.Param("id")
	ar.DB.Where("id = ?", id).Delete(&models.MediaAsset{})
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Media deleted"})
}

// ── UpdateCourse ──────────────────────────────────────────────────────────────

func (ar *AdminRouter) UpdateCourse(c *gin.Context) {
	id := c.Param("id")
	var course models.Course
	if err := ar.DB.Where("id = ? OR slug = ?", id, id).First(&course).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Course not found"})
		return
	}

	var req struct {
		Title         string  `json:"title"`
		Description   string  `json:"description"`
		Category      string  `json:"category"`
		Tags          string  `json:"tags"`
		Price         float64 `json:"price"`
		OriginalPrice float64 `json:"originalPrice"`
		Level         string  `json:"level"`
		Duration      string  `json:"duration"`
		ImageAsset    string  `json:"imageAsset"`
		Instructor    string  `json:"instructorName"`
		InstructorBio string  `json:"instructorBio"`
		IsPublished   *bool   `json:"isPublished"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	if req.Title != "" {
		course.Title = req.Title
	}
	if req.Description != "" {
		course.Description = req.Description
	}
	if req.Category != "" {
		course.Category = req.Category
	}
	if req.Tags != "" {
		course.Tags = req.Tags
	}
	if req.Price > 0 {
		course.Price = req.Price
	}
	if req.OriginalPrice > 0 {
		course.OriginalPrice = req.OriginalPrice
	}
	if req.Level != "" {
		course.Level = req.Level
	}
	if req.Duration != "" {
		course.Duration = req.Duration
	}
	if req.ImageAsset != "" {
		course.ImageAsset = req.ImageAsset
	}
	if req.Instructor != "" {
		course.Instructor = req.Instructor
	}
	if req.InstructorBio != "" {
		course.InstructorBio = req.InstructorBio
	}
	if req.IsPublished != nil {
		course.IsPublished = *req.IsPublished
	}
	course.UpdatedAt = time.Now()

	ar.DB.Save(&course)
	c.JSON(http.StatusOK, gin.H{"success": true, "course": course, "data": course})
}

// ── UpdateModule ──────────────────────────────────────────────────────────────

func (ar *AdminRouter) UpdateModule(c *gin.Context) {
	id := c.Param("id")
	var module models.CourseModule
	if err := ar.DB.Where("id = ?", id).First(&module).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Module not found"})
		return
	}

	var req struct {
		Title    string `json:"title"`
		OrderNum int    `json:"orderNum"`
	}
	c.ShouldBindJSON(&req)

	if req.Title != "" {
		module.Title = req.Title
	}
	if req.OrderNum > 0 {
		module.OrderNum = req.OrderNum
	}
	ar.DB.Save(&module)
	c.JSON(http.StatusOK, gin.H{"success": true, "module": module})
}

func (ar *AdminRouter) DeleteModule(c *gin.Context) {
	id := c.Param("id")
	// Delete all lessons in this module first
	ar.DB.Where("module_id = ?", id).Delete(&models.Lesson{})
	ar.DB.Where("id = ?", id).Delete(&models.CourseModule{})
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Module and its lessons deleted"})
}

// ── UpdateLesson ──────────────────────────────────────────────────────────────

func (ar *AdminRouter) UpdateLesson(c *gin.Context) {
	id := c.Param("id")
	var lesson models.Lesson
	if err := ar.DB.Where("id = ?", id).First(&lesson).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Lesson not found"})
		return
	}

	var req struct {
		Title         string `json:"title"`
		Slug          string `json:"slug"`
		Duration      string `json:"duration"`
		VideoURL      string `json:"videoUrl"`
		StreamType    string `json:"streamType"`
		IsFreePreview *bool  `json:"isFreePreview"`
		IsQuiz        *bool  `json:"isQuiz"`
		OrderNum      int    `json:"orderNum"`
		Description   string `json:"description"`
		CourseSlug    string `json:"courseSlug"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	if req.Title != "" {
		lesson.Title = req.Title
	}
	if req.Slug != "" {
		lesson.Slug = req.Slug
	}
	if req.Duration != "" {
		lesson.Duration = req.Duration
	}
	if req.VideoURL != "" {
		lesson.VideoURL = req.VideoURL
		lesson.HasVideo = true
	}
	if req.StreamType != "" {
		lesson.StreamType = req.StreamType
	}
	if req.IsFreePreview != nil {
		lesson.IsFreePreview = *req.IsFreePreview
	}
	if req.IsQuiz != nil {
		lesson.IsQuiz = *req.IsQuiz
	}
	if req.OrderNum > 0 {
		lesson.OrderNum = req.OrderNum
	}
	if req.Description != "" {
		lesson.Description = req.Description
	}
	if req.CourseSlug != "" {
		lesson.CourseSlug = req.CourseSlug
	}

	ar.DB.Save(&lesson)
	c.JSON(http.StatusOK, gin.H{"success": true, "lesson": lesson, "data": lesson})
}
