// Package classroom provides interactive APIs for the video classroom:
//   - Personal timestamped notes  (GET/POST/PUT/DELETE)
//   - Lesson schematics & PDFs    (GET/POST/DELETE)
//   - Lesson Q&A discussions      (GET/POST, answers, upvote)
package classroom

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"smg-backend/internal/middleware"
	"smg-backend/internal/models"
)

type ClassroomRouter struct {
	DB *gorm.DB
}

func RegisterRoutes(r *gin.Engine, db *gorm.DB) {
	cr := &ClassroomRouter{DB: db}
	cr.seedSampleData()

	// ── Notes ─────────────────────────────────────────────────────────────
	r.GET("/api/lessons/:lessonId/notes", cr.GetNotes)
	r.POST("/api/lessons/:lessonId/notes", cr.CreateNote)
	r.PUT("/api/notes/:id", cr.UpdateNote)
	r.DELETE("/api/notes/:id", cr.DeleteNote)

	// ── Schematics ────────────────────────────────────────────────────────
	r.GET("/api/lessons/:lessonId/schematics", cr.GetSchematics)
	r.POST("/api/lessons/:lessonId/schematics", cr.CreateSchematic)
	r.DELETE("/api/schematics/:id", cr.DeleteSchematic)

	// ── Q&A ───────────────────────────────────────────────────────────────
	r.GET("/api/lessons/:lessonId/qa", cr.GetQA)
	r.POST("/api/lessons/:lessonId/qa", cr.PostQuestion)
	r.POST("/api/lessons/:lessonId/qa/:questionId/answers", cr.PostAnswer)
	r.POST("/api/lessons/:lessonId/qa/:questionId/upvote", cr.UpvoteQuestion)
	r.PATCH("/api/lessons/:lessonId/qa/:questionId/resolve", cr.ResolveQuestion)

	// ── Admin: Manage lesson resources ────────────────────────────────────
	adm := r.Group("/api/admin/classroom", middleware.AuthMiddleware(), middleware.AdminMiddleware())
	{
		// Schematics
		adm.GET("/schematics", cr.AdminListSchematics)
		adm.POST("/schematics", cr.AdminCreateSchematic)
		adm.PUT("/schematics/:id", cr.AdminUpdateSchematic)
		adm.DELETE("/schematics/:id", cr.DeleteSchematic)

		// Q&A moderation
		adm.GET("/qa", cr.AdminListQA)
		adm.DELETE("/qa/:id", cr.AdminDeleteQuestion)
		adm.DELETE("/qa/answers/:id", cr.AdminDeleteAnswer)
		adm.PATCH("/qa/:id/resolve", cr.AdminResolveQuestion)
	}
}

// ─── Notes ────────────────────────────────────────────────────────────────

// GET /api/lessons/:lessonId/notes?userId=&courseSlug=
func (cr *ClassroomRouter) GetNotes(c *gin.Context) {
	lessonID := c.Param("lessonId")
	userID := c.Query("userId")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "userId is required"})
		return
	}

	query := cr.DB.Where("lesson_id = ? AND user_id = ?", lessonID, userID)
	if slug := c.Query("courseSlug"); slug != "" {
		query = query.Where("course_slug = ?", slug)
	}

	var notes []models.UserVideoNote
	if err := query.Order("timestamp_seconds asc").Find(&notes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": notes})
}

// POST /api/lessons/:lessonId/notes
func (cr *ClassroomRouter) CreateNote(c *gin.Context) {
	lessonID := c.Param("lessonId")
	var body struct {
		UserID           string  `json:"userId"           binding:"required"`
		Text             string  `json:"text"             binding:"required"`
		TimestampSeconds float64 `json:"timestampSeconds"`
		CourseSlug       string  `json:"courseSlug"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	now := time.Now()
	note := models.UserVideoNote{
		ID:               uuid.New().String(),
		UserID:           body.UserID,
		LessonID:         lessonID,
		CourseSlug:       body.CourseSlug,
		TimestampSeconds: body.TimestampSeconds,
		Text:             body.Text,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	if err := cr.DB.Create(&note).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": note})
}

// PUT /api/notes/:id
func (cr *ClassroomRouter) UpdateNote(c *gin.Context) {
	id := c.Param("id")
	var body struct {
		Text             string  `json:"text"`
		TimestampSeconds float64 `json:"timestampSeconds"`
		UserID           string  `json:"userId"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	updates := map[string]interface{}{"updated_at": time.Now()}
	if body.Text != "" {
		updates["text"] = body.Text
	}
	if body.TimestampSeconds > 0 {
		updates["timestamp_seconds"] = body.TimestampSeconds
	}

	query := cr.DB.Model(&models.UserVideoNote{}).Where("id = ?", id)
	if body.UserID != "" {
		query = query.Where("user_id = ?", body.UserID) // ensure ownership
	}
	if err := query.Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	var updated models.UserVideoNote
	cr.DB.Where("id = ?", id).First(&updated)
	c.JSON(http.StatusOK, gin.H{"success": true, "data": updated})
}

// DELETE /api/notes/:id?userId=
func (cr *ClassroomRouter) DeleteNote(c *gin.Context) {
	id := c.Param("id")
	userID := c.Query("userId")

	query := cr.DB.Where("id = ?", id)
	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}
	if err := query.Delete(&models.UserVideoNote{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Note deleted successfully"})
}

// ─── Schematics ───────────────────────────────────────────────────────────

// GET /api/lessons/:lessonId/schematics?courseSlug=
func (cr *ClassroomRouter) GetSchematics(c *gin.Context) {
	lessonID := c.Param("lessonId")
	query := cr.DB.Where("lesson_id = ?", lessonID)
	if slug := c.Query("courseSlug"); slug != "" {
		query = query.Where("course_slug = ?", slug)
	}
	if cat := c.Query("category"); cat != "" {
		query = query.Where("category = ?", cat)
	}

	var schematics []models.LessonSchematic
	if err := query.Order("created_at asc").Find(&schematics).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": schematics})
}

// POST /api/lessons/:lessonId/schematics
func (cr *ClassroomRouter) CreateSchematic(c *gin.Context) {
	lessonID := c.Param("lessonId")
	var body struct {
		Title       string `json:"title"   binding:"required"`
		Description string `json:"description"`
		FileURL     string `json:"fileUrl" binding:"required"`
		FileType    string `json:"fileType"`
		FileSize    string `json:"fileSize"`
		TotalPages  int    `json:"totalPages"`
		Category    string `json:"category"`
		CourseSlug  string `json:"courseSlug"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	ft := body.FileType
	if ft == "" {
		ft = "pdf"
	}
	cat := body.Category
	if cat == "" {
		cat = "Motherboard Schematic"
	}
	pages := body.TotalPages
	if pages == 0 {
		pages = 1
	}

	s := models.LessonSchematic{
		ID:          uuid.New().String(),
		LessonID:    lessonID,
		CourseSlug:  body.CourseSlug,
		Title:       body.Title,
		Description: body.Description,
		FileURL:     body.FileURL,
		FileType:    ft,
		FileSize:    body.FileSize,
		TotalPages:  pages,
		Category:    cat,
		CreatedAt:   time.Now(),
	}
	if err := cr.DB.Create(&s).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": s})
}

// DELETE /api/schematics/:id  (shared by student and admin)
func (cr *ClassroomRouter) DeleteSchematic(c *gin.Context) {
	id := c.Param("id")
	cr.DB.Delete(&models.LessonSchematic{}, "id = ?", id)
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Schematic removed"})
}

// ─── Admin Schematic Management ───────────────────────────────────────────

func (cr *ClassroomRouter) AdminListSchematics(c *gin.Context) {
	query := cr.DB.Model(&models.LessonSchematic{})
	if lessonID := c.Query("lessonId"); lessonID != "" {
		query = query.Where("lesson_id = ?", lessonID)
	}
	if slug := c.Query("courseSlug"); slug != "" {
		query = query.Where("course_slug = ?", slug)
	}
	var list []models.LessonSchematic
	query.Order("created_at desc").Find(&list)
	c.JSON(http.StatusOK, gin.H{"success": true, "data": list})
}

func (cr *ClassroomRouter) AdminCreateSchematic(c *gin.Context) {
	var body struct {
		LessonID    string `json:"lessonId"  binding:"required"`
		CourseSlug  string `json:"courseSlug"`
		Title       string `json:"title"     binding:"required"`
		Description string `json:"description"`
		FileURL     string `json:"fileUrl"   binding:"required"`
		FileType    string `json:"fileType"`
		FileSize    string `json:"fileSize"`
		TotalPages  int    `json:"totalPages"`
		Category    string `json:"category"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	s := models.LessonSchematic{
		ID: uuid.New().String(), LessonID: body.LessonID, CourseSlug: body.CourseSlug,
		Title: body.Title, Description: body.Description, FileURL: body.FileURL,
		FileType: body.FileType, FileSize: body.FileSize, TotalPages: body.TotalPages,
		Category: body.Category, CreatedAt: time.Now(),
	}
	if s.FileType == "" { s.FileType = "pdf" }
	if s.Category == "" { s.Category = "Motherboard Schematic" }
	cr.DB.Create(&s)
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": s})
}

func (cr *ClassroomRouter) AdminUpdateSchematic(c *gin.Context) {
	id := c.Param("id")
	var updates map[string]interface{}
	c.ShouldBindJSON(&updates)
	cr.DB.Model(&models.LessonSchematic{}).Where("id = ?", id).Updates(updates)
	var updated models.LessonSchematic
	cr.DB.Where("id = ?", id).First(&updated)
	c.JSON(http.StatusOK, gin.H{"success": true, "data": updated})
}

// ─── Q&A ──────────────────────────────────────────────────────────────────

// GET /api/lessons/:lessonId/qa?courseSlug=&sort=newest|upvotes
func (cr *ClassroomRouter) GetQA(c *gin.Context) {
	lessonID := c.Param("lessonId")
	query := cr.DB.Where("lesson_id = ?", lessonID)
	if slug := c.Query("courseSlug"); slug != "" {
		query = query.Where("course_slug = ?", slug)
	}

	sort := c.DefaultQuery("sort", "newest")
	orderStr := "created_at desc"
	if sort == "upvotes" {
		orderStr = "upvotes desc, created_at desc"
	} else if sort == "unresolved" {
		query = query.Where("is_resolved = ?", false)
	}

	var questions []models.LessonQAQuestion
	if err := query.Preload("Answers", func(db *gorm.DB) *gorm.DB {
		return db.Order("created_at asc")
	}).Order(orderStr).Find(&questions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": questions})
}

// POST /api/lessons/:lessonId/qa
func (cr *ClassroomRouter) PostQuestion(c *gin.Context) {
	lessonID := c.Param("lessonId")
	var body struct {
		UserID           string  `json:"userId"           binding:"required"`
		AuthorName       string  `json:"authorName"       binding:"required"`
		AuthorRole       string  `json:"authorRole"`
		AuthorAvatar     string  `json:"authorAvatar"`
		Question         string  `json:"question"         binding:"required"`
		TimestampSeconds float64 `json:"timestampSeconds"`
		CourseSlug       string  `json:"courseSlug"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	role := body.AuthorRole
	if role == "" {
		role = "VIP Technician"
	}

	now := time.Now()
	q := models.LessonQAQuestion{
		ID:               uuid.New().String(),
		LessonID:         lessonID,
		CourseSlug:       body.CourseSlug,
		UserID:           body.UserID,
		AuthorName:       body.AuthorName,
		AuthorRole:       role,
		AuthorAvatar:     body.AuthorAvatar,
		TimestampSeconds: body.TimestampSeconds,
		Question:         body.Question,
		IsResolved:       false,
		Upvotes:          0,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	if err := cr.DB.Omit("Answers").Create(&q).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": q})
}

// POST /api/lessons/:lessonId/qa/:questionId/answers
func (cr *ClassroomRouter) PostAnswer(c *gin.Context) {
	questionID := c.Param("questionId")
	var body struct {
		UserID       string `json:"userId"       binding:"required"`
		AuthorName   string `json:"authorName"   binding:"required"`
		AuthorRole   string `json:"authorRole"`
		AuthorAvatar string `json:"authorAvatar"`
		Answer       string `json:"answer"       binding:"required"`
		IsInstructor bool   `json:"isInstructor"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	// Verify question exists
	var q models.LessonQAQuestion
	if err := cr.DB.Where("id = ?", questionID).First(&q).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Question not found"})
		return
	}

	role := body.AuthorRole
	if role == "" {
		if body.IsInstructor {
			role = "Instructor"
		} else {
			role = "Student"
		}
	}

	a := models.LessonQAAnswer{
		ID:           uuid.New().String(),
		QuestionID:   questionID,
		UserID:       body.UserID,
		AuthorName:   body.AuthorName,
		AuthorRole:   role,
		AuthorAvatar: body.AuthorAvatar,
		IsInstructor: body.IsInstructor,
		Answer:       body.Answer,
		CreatedAt:    time.Now(),
	}
	if err := cr.DB.Create(&a).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	// Auto-resolve if instructor answered
	if body.IsInstructor {
		cr.DB.Model(&models.LessonQAQuestion{}).Where("id = ?", questionID).
			Updates(map[string]interface{}{"is_resolved": true, "updated_at": time.Now()})
	}

	c.JSON(http.StatusCreated, gin.H{"success": true, "data": a})
}

// POST /api/lessons/:lessonId/qa/:questionId/upvote
func (cr *ClassroomRouter) UpvoteQuestion(c *gin.Context) {
	questionID := c.Param("questionId")
	cr.DB.Model(&models.LessonQAQuestion{}).Where("id = ?", questionID).
		UpdateColumn("upvotes", gorm.Expr("upvotes + 1"))
	var q models.LessonQAQuestion
	cr.DB.Where("id = ?", questionID).First(&q)
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"id": questionID, "upvotes": q.Upvotes}})
}

// PATCH /api/lessons/:lessonId/qa/:questionId/resolve
func (cr *ClassroomRouter) ResolveQuestion(c *gin.Context) {
	questionID := c.Param("questionId")
	cr.DB.Model(&models.LessonQAQuestion{}).Where("id = ?", questionID).
		Updates(map[string]interface{}{"is_resolved": true, "updated_at": time.Now()})
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"id": questionID, "isResolved": true}})
}

// ─── Admin Q&A Moderation ─────────────────────────────────────────────────

func (cr *ClassroomRouter) AdminListQA(c *gin.Context) {
	query := cr.DB.Model(&models.LessonQAQuestion{})
	if lessonID := c.Query("lessonId"); lessonID != "" {
		query = query.Where("lesson_id = ?", lessonID)
	}
	if slug := c.Query("courseSlug"); slug != "" {
		query = query.Where("course_slug = ?", slug)
	}
	if resolved := c.Query("resolved"); resolved != "" {
		query = query.Where("is_resolved = ?", resolved == "true")
	}
	var questions []models.LessonQAQuestion
	query.Preload("Answers").Order("created_at desc").Find(&questions)
	c.JSON(http.StatusOK, gin.H{"success": true, "data": questions})
}

func (cr *ClassroomRouter) AdminDeleteQuestion(c *gin.Context) {
	id := c.Param("id")
	cr.DB.Where("question_id = ?", id).Delete(&models.LessonQAAnswer{})
	cr.DB.Delete(&models.LessonQAQuestion{}, "id = ?", id)
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (cr *ClassroomRouter) AdminDeleteAnswer(c *gin.Context) {
	id := c.Param("id")
	cr.DB.Delete(&models.LessonQAAnswer{}, "id = ?", id)
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (cr *ClassroomRouter) AdminResolveQuestion(c *gin.Context) {
	id := c.Param("id")
	cr.DB.Model(&models.LessonQAQuestion{}).Where("id = ?", id).
		Updates(map[string]interface{}{"is_resolved": true, "updated_at": time.Now()})
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// ─── Seed ─────────────────────────────────────────────────────────────────

func (cr *ClassroomRouter) seedSampleData() {
	// Seed schematics
	var schCount int64
	cr.DB.Model(&models.LessonSchematic{}).Count(&schCount)
	if schCount == 0 {
		schematics := []models.LessonSchematic{
			{
				ID: uuid.New().String(), LessonID: "lesson_1", CourseSlug: "android-hardware-repairing-level-1",
				Title: "Qualcomm PM6150 PMIC Schematic & PCB Trace",
				Description: "High-resolution circuit diagram of charging and buck regulator rails.",
				FileURL: "https://pub-de6175c42cb241aab9f5c23b24292596.r2.dev/schematics/pdf/pm6150.pdf",
				FileType: "pdf", FileSize: "3.8 MB", TotalPages: 16,
				Category: "Motherboard Schematic", CreatedAt: time.Now(),
			},
			{
				ID: uuid.New().String(), LessonID: "lesson_1", CourseSlug: "android-hardware-repairing-level-1",
				Title: "Snapdragon 888 CPU Power Rail Diagram",
				Description: "VDD_CORE, VDD_APC, VREG_BOB supply chain trace map.",
				FileURL: "https://pub-de6175c42cb241aab9f5c23b24292596.r2.dev/schematics/pdf/snapdragon888_layout.pdf",
				FileType: "pdf", FileSize: "5.2 MB", TotalPages: 24,
				Category: "CPU Power Section", CreatedAt: time.Now(),
			},
			{
				ID: uuid.New().String(), LessonID: "lesson_2", CourseSlug: "android-hardware-repairing-level-1",
				Title: "UFS 2.2 / 3.1 ISP Jumper Trace & Pinout",
				Description: "ISP clock, CMD, DQ0-DQ3 resistor values for Easy J-Tag and UFi Box.",
				FileURL: "https://pub-de6175c42cb241aab9f5c23b24292596.r2.dev/schematics/pdf/bga254_ufs.pdf",
				FileType: "pdf", FileSize: "2.1 MB", TotalPages: 6,
				Category: "ISP Pinout Solution", CreatedAt: time.Now(),
			},
		}
		for _, s := range schematics {
			cr.DB.Create(&s)
		}
	}

	// Seed Q&A
	var qaCount int64
	cr.DB.Model(&models.LessonQAQuestion{}).Count(&qaCount)
	if qaCount == 0 {
		now := time.Now()
		qID := uuid.New().String()
		q := models.LessonQAQuestion{
			ID: qID, LessonID: "lesson_1", CourseSlug: "android-hardware-repairing-level-1",
			UserID: "usr_vip_99", AuthorName: "Sunil Shrestha", AuthorRole: "VIP Hardware Technician",
			TimestampSeconds: 165.0,
			Question: "If VPH_PWR reads 3.7V on cold battery, should we replace OVP IC first or PM6150?",
			IsResolved: true, Upvotes: 7,
			CreatedAt: now.Add(-48 * time.Hour), UpdatedAt: now.Add(-46 * time.Hour),
		}
		cr.DB.Omit("Answers").Create(&q)

		ans := models.LessonQAAnswer{
			ID: uuid.New().String(), QuestionID: qID,
			UserID: "inst_amit", AuthorName: "Er. Amit Verma",
			AuthorRole: "Senior Hardware Instructor", IsInstructor: true,
			Answer: "Check diode mode value on VBUS first. If OVP output is 5.0V, the fault is inside PM6150 BUCK coil switching. Replace BUCK inductor before the IC.",
			CreatedAt: now.Add(-46 * time.Hour),
		}
		cr.DB.Create(&ans)

		// Seed a note
		var noteCount int64
		cr.DB.Model(&models.UserVideoNote{}).Count(&noteCount)
		if noteCount == 0 {
			note := models.UserVideoNote{
				ID: uuid.New().String(), UserID: "usr_1", LessonID: "lesson_1",
				CourseSlug: "android-hardware-repairing-level-1",
				TimestampSeconds: 85.0,
				Text: "PMIC BUCK capacitor C201 must read between 1.1V and 1.8V. Anything below = replace capacitor.",
				CreatedAt: now, UpdatedAt: now,
			}
			cr.DB.Create(&note)
		}
	}
}

// suppress unused import warning
var _ = strings.ToLower
