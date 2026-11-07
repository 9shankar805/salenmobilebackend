package course

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"smg-backend/internal/models"
)

type CourseRouter struct {
	DB *gorm.DB
}

func RegisterRoutes(r *gin.Engine, db *gorm.DB) {
	cr := &CourseRouter{DB: db}

	// 1. Organization & Public Course Listing
	r.GET("/organization/courses", cr.ListOrgCourses)
	r.POST("/organization/courses", cr.CreateCourse)
	r.GET("/organization/courses/public", cr.GetPublicCourses)

	// Public catalog aliases — GET /courses and GET /api/courses
	r.GET("/courses", cr.GetPublicCourses)
	r.GET("/api/courses", cr.GetPublicCourses)

	// 2. Course Root CRUD (SvelteKit & Mobile RPC)
	r.POST("/course", cr.CreateCourse)
	r.GET("/course/:id", cr.GetCourseDetail)
	r.PUT("/course/:id", cr.UpdateCourse)
	r.DELETE("/course/:id", cr.DeleteCourse)
	r.GET("/course/slug/:slug", cr.GetCourseDetail)
	r.POST("/course/:id/clone", cr.CloneCourse)
	r.POST("/course/:id/enroll", cr.EnrollCourse)

	// 3. Sections / Modules CRUD
	r.POST("/course/:id/section", cr.CreateSection)
	r.PUT("/course/:id/section/:sectionId", cr.UpdateSection)
	r.DELETE("/course/:id/section/:sectionId", cr.DeleteSection)
	r.POST("/course/:id/section/reorder", cr.ReorderSections)
	r.POST("/course/:id/section/promote-ungrouped", cr.PromoteUngrouped)

	// 4. Lessons CRUD
	r.GET("/course/:id/lesson", cr.ListLessons)
	r.POST("/course/:id/lesson", cr.CreateLesson)
	r.GET("/course/:id/lesson/:lessonId", cr.GetLessonDetail)
	r.PUT("/course/:id/lesson/:lessonId", cr.UpdateLesson)
	r.DELETE("/course/:id/lesson/:lessonId", cr.DeleteLesson)
	r.POST("/course/:id/lesson/reorder", cr.ReorderLessons)

	// 5. Watch Progress & Completion
	r.GET("/course/:id/lesson/:lessonId/watch-progress", cr.GetWatchProgress)
	r.PUT("/course/:id/lesson/:lessonId/watch-progress", cr.UpdateWatchProgress)
	r.PUT("/course/:id/lesson/:lessonId/completion", cr.UpdateLessonCompletion)

	// 6. Comments
	r.GET("/course/:id/lesson/:lessonId/comment", cr.GetComments)
	r.POST("/course/:id/lesson/:lessonId/comment", cr.AddComment)
	r.PUT("/course/:id/lesson/comment/:commentId", cr.UpdateComment)
	r.DELETE("/course/:id/lesson/comment/:commentId", cr.DeleteComment)

	// 7. Settings, Tags, AI Tutor, Gradebook & Analytics
	r.GET("/course/:id/tags", cr.GetCourseTags)
	r.PUT("/course/:id/tags", cr.UpdateCourseTags)
	r.GET("/course/:id/ai-tutor", cr.GetAiTutor)
	r.PUT("/course/:id/ai-tutor", cr.UpdateAiTutor)
	r.GET("/course/:id/curriculum", cr.GetCourseCurriculum)
	r.GET("/course/:id/analytics", cr.GetCourseAnalytics)
	r.GET("/course/:id/members", cr.GetCourseMembers)
	r.GET("/course/:id/exercise", cr.GetExercises)
	r.GET("/course/:id/exercise/:exerciseId", cr.GetExerciseDetail)
	r.GET("/org-site/course/:courseSlug/exercises", cr.GetExercises)

	// 8. Org-site Mobile App Player & Streaming
	r.GET("/org-site/course/:courseSlug", cr.GetCourseTree)
	r.GET("/org-site/course/:courseSlug/item/:itemSlug", cr.GetCourseLesson)
	r.GET("/org-site/course/:courseSlug/item/:itemSlug/stream", cr.GetVideoStream)
	r.POST("/org-site/course/:courseSlug/item/:itemSlug/hls-cookie", cr.MintHlsCookie)

	// 9. New REST v1 — Course Detail Screen
	// GET /api/v1/courses/:courseSlug/curriculum         — full module+lesson tree with quizzes
	// GET /api/v1/users/:userId/courses/:courseSlug/progress — per-course student progress + quiz grades
	// GET /api/v1/courses/:courseSlug/resources          — downloadable schematics & PDFs
	v1 := r.Group("/api/v1")
	{
		v1.GET("/courses/:courseSlug/curriculum", cr.GetCurriculumV1)
		v1.GET("/users/:userId/courses/:courseSlug/progress", cr.GetCourseProgressV1)
		v1.GET("/courses/:courseSlug/resources", cr.GetCourseResourcesV1)
	}
}

// ═══════════════════════════════════════════════════════════════════════════
// 1. PUBLIC COURSES (FLUTTER APP CATALOG)
// ═══════════════════════════════════════════════════════════════════════════

func (cr *CourseRouter) GetPublicCourses(c *gin.Context) {
	query := cr.DB.Model(&models.Course{}).Where("is_published = ?", true)

	search := strings.TrimSpace(c.Query("search"))
	if search != "" {
		like := "%" + strings.ToLower(search) + "%"
		query = query.Where("LOWER(title) LIKE ? OR LOWER(description) LIKE ? OR LOWER(tags) LIKE ?", like, like, like)
	}

	tags := strings.TrimSpace(c.Query("tags"))
	if tags != "" && tags != "All" {
		likeTag := "%" + strings.ToLower(tags) + "%"
		query = query.Where("LOWER(tags) LIKE ? OR LOWER(category) LIKE ?", likeTag, likeTag)
	}

	pricing := strings.ToLower(strings.TrimSpace(c.Query("pricing")))
	if pricing == "free" {
		query = query.Where("price = 0")
	} else if pricing == "paid" {
		query = query.Where("price > 0")
	}

	limitStr := c.DefaultQuery("limit", "20")
	limit, _ := strconv.Atoi(limitStr)
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	pageStr := c.DefaultQuery("page", "1")
	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * limit

	var courses []models.Course
	if err := query.Limit(limit).Offset(offset).Order("created_at desc").Find(&courses).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to query courses"})
		return
	}

	courseList := make([]map[string]interface{}, 0, len(courses))
	for _, crs := range courses {
		courseList = append(courseList, map[string]interface{}{
			"id":            crs.ID,
			"title":         crs.Title,
			"slug":          crs.Slug,
			"description":   crs.Description,
			"category":      crs.Category,
			"tags":          crs.Tags,
			"price":         crs.Price,
			"originalPrice": crs.OriginalPrice,
			"rating":        crs.Rating,
			"reviews":       crs.ReviewsCount,
			"lessons":       crs.LessonsCount,
			"duration":      crs.Duration,
			"level":         crs.Level,
			"imageAsset":    crs.ImageAsset,
			"instructor": map[string]interface{}{
				"name": crs.Instructor,
				"bio":  crs.InstructorBio,
			},
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"courses":        courseList,
			"hasMoreCourses": false,
			"total":          len(courseList),
			"page":           page,
			"limit":          limit,
			"totalPages":     1,
		},
		"courses": courseList,
	})
}

// ═══════════════════════════════════════════════════════════════════════════
// 2. COURSE DETAILS & SVELTEKIT SHAPE (GET, CREATE, UPDATE, DELETE)
// ═══════════════════════════════════════════════════════════════════════════

func (cr *CourseRouter) formatCourseForDashboard(course *models.Course) map[string]interface{} {
	sections := make([]map[string]interface{}, 0, len(course.Modules))
	for _, m := range course.Modules {
		items := make([]map[string]interface{}, 0, len(m.Lessons))
		for _, l := range m.Lessons {
			items = append(items, map[string]interface{}{
				"id":            l.ID,
				"title":         l.Title,
				"slug":          l.Slug,
				"type":          "lesson",
				"duration":      l.Duration,
				"videoUrl":      l.VideoURL,
				"streamType":    l.StreamType,
				"isFreePreview": l.IsFreePreview,
				"order":         l.OrderNum,
				"moduleId":      m.ID,
			})
		}
		sections = append(sections, map[string]interface{}{
			"id":       m.ID,
			"courseId": course.ID,
			"title":    m.Title,
			"order":    m.OrderNum,
			"items":    items,
		})
	}

	return map[string]interface{}{
		"id":            course.ID,
		"title":         course.Title,
		"slug":          course.Slug,
		"description":   course.Description,
		"category":      course.Category,
		"tags":          course.Tags,
		"price":         course.Price,
		"originalPrice": course.OriginalPrice,
		"rating":        course.Rating,
		"reviews":       course.ReviewsCount,
		"lessons":       course.LessonsCount,
		"duration":      course.Duration,
		"level":         course.Level,
		"imageAsset":    course.ImageAsset,
		"isPublished":   course.IsPublished,
		"type":          "DEFAULT",
		"createdAt":     course.CreatedAt,
		"updatedAt":     course.UpdatedAt,
		"content": map[string]interface{}{
			"grouped":  true,
			"sections": sections,
			"items":    []interface{}{},
		},
		"group": map[string]interface{}{
			"id":       "group-" + course.ID,
			"tutors":   []interface{}{},
			"students": []interface{}{},
			"people": []interface{}{
				map[string]interface{}{
					"profileId": "admin-user-1",
					"role":      "OWNER",
					"roleId":    1,
				},
			},
		},
		"modules": sections,
	}
}

func (cr *CourseRouter) GetCourseDetail(c *gin.Context) {
	idOrSlug := c.Param("id")
	if idOrSlug == "" {
		idOrSlug = c.Param("slug")
	}

	var course models.Course
	err := cr.DB.Preload("Modules.Lessons", func(db *gorm.DB) *gorm.DB {
		return db.Order("order_num asc")
	}).Where("id = ? OR slug = ?", idOrSlug, idOrSlug).First(&course).Error

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Course not found: " + idOrSlug})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    cr.formatCourseForDashboard(&course),
	})
}

func (cr *CourseRouter) ListOrgCourses(c *gin.Context) {
	var courses []models.Course
	cr.DB.Preload("Modules.Lessons").Order("created_at desc").Find(&courses)

	list := make([]map[string]interface{}, 0, len(courses))
	for _, crs := range courses {
		list = append(list, cr.formatCourseForDashboard(&crs))
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    list,
		"courses": list,
	})
}

func (cr *CourseRouter) CreateCourse(c *gin.Context) {
	var input struct {
		Title          string  `json:"title" binding:"required"`
		Description    string  `json:"description"`
		Type           string  `json:"type"`
		Price          float64 `json:"price"`
		Category       string  `json:"category"`
		Tags           string  `json:"tags"`
		OrganizationID string  `json:"organizationId"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	courseID := "c-" + uuid.New().String()[:8]
	slug := strings.ToLower(strings.ReplaceAll(input.Title, " ", "-"))

	newCourse := models.Course{
		ID:            courseID,
		Title:         input.Title,
		Slug:          slug,
		Description:   input.Description,
		Category:      input.Category,
		Tags:          input.Tags,
		Price:         input.Price,
		OriginalPrice: input.Price * 1.5,
		Rating:        4.9,
		ReviewsCount:  1,
		LessonsCount:  0,
		Duration:      "0h 0m",
		Level:         "All Levels",
		ImageAsset:    "assets/banner_bg_img/banner_1.webp",
		Instructor:    "Salen Dangol",
		InstructorBio: "Master Hardware & Micro-Soldering Engineer",
		IsPublished:   true,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if newCourse.Category == "" {
		newCourse.Category = "Hardware"
	}

	if err := cr.DB.Create(&newCourse).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	formatted := cr.formatCourseForDashboard(&newCourse)
	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data": gin.H{
			"course": formatted,
		},
	})
}

func (cr *CourseRouter) UpdateCourse(c *gin.Context) {
	courseID := c.Param("id")

	var course models.Course
	if err := cr.DB.Where("id = ? OR slug = ?", courseID, courseID).First(&course).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Course not found"})
		return
	}

	var updateData map[string]interface{}
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	if title, ok := updateData["title"].(string); ok && title != "" {
		course.Title = title
	}
	if desc, ok := updateData["description"].(string); ok {
		course.Description = desc
	}
	if price, ok := updateData["price"].(float64); ok {
		course.Price = price
	}
	if isPub, ok := updateData["isPublished"].(bool); ok {
		course.IsPublished = isPub
	}
	if tags, ok := updateData["tags"].(string); ok {
		course.Tags = tags
	}
	course.UpdatedAt = time.Now()

	cr.DB.Save(&course)
	cr.DB.Preload("Modules.Lessons").Where("id = ?", course.ID).First(&course)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    cr.formatCourseForDashboard(&course),
	})
}

func (cr *CourseRouter) DeleteCourse(c *gin.Context) {
	courseID := c.Param("id")
	cr.DB.Where("id = ?", courseID).Delete(&models.Course{})
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"id": courseID}})
}

func (cr *CourseRouter) CloneCourse(c *gin.Context) {
	courseID := c.Param("id")
	var course models.Course
	if err := cr.DB.Preload("Modules.Lessons").Where("id = ?", courseID).First(&course).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Course not found"})
		return
	}

	clonedID := "c-" + uuid.New().String()[:8]
	course.ID = clonedID
	course.Title = course.Title + " (Copy)"
	course.Slug = course.Slug + "-copy"
	course.CreatedAt = time.Now()
	course.UpdatedAt = time.Now()

	cr.DB.Create(&course)
	c.JSON(http.StatusOK, gin.H{"success": true, "course": course})
}

func (cr *CourseRouter) EnrollCourse(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Enrolled successfully"})
}

// ═══════════════════════════════════════════════════════════════════════════
// 3. SECTIONS (MODULES) MANAGEMENT
// ═══════════════════════════════════════════════════════════════════════════

type CreateSectionInput struct {
	Title string `json:"title" binding:"required"`
	Order int    `json:"order"`
}

func (cr *CourseRouter) CreateSection(c *gin.Context) {
	courseID := c.Param("id")

	var input CreateSectionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	section := models.Module{
		ID:          "sec-" + uuid.New().String()[:8],
		CourseID:    courseID,
		Title:       input.Title,
		OrderNum:    input.Order,
		LessonCount: 0,
		CreatedAt:   time.Now(),
	}

	if err := cr.DB.Create(&section).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data": gin.H{
			"id":       section.ID,
			"courseId": section.CourseID,
			"title":    section.Title,
			"order":    section.OrderNum,
			"items":    []interface{}{},
		},
	})
}

func (cr *CourseRouter) UpdateSection(c *gin.Context) {
	sectionID := c.Param("sectionId")

	var section models.Module
	if err := cr.DB.Where("id = ?", sectionID).First(&section).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Section not found"})
		return
	}

	var input struct {
		Title string `json:"title"`
		Order int    `json:"order"`
	}
	c.ShouldBindJSON(&input)

	if input.Title != "" {
		section.Title = input.Title
	}
	if input.Order > 0 {
		section.OrderNum = input.Order
	}
	cr.DB.Save(&section)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"id":       section.ID,
			"courseId": section.CourseID,
			"title":    section.Title,
			"order":    section.OrderNum,
		},
	})
}

func (cr *CourseRouter) DeleteSection(c *gin.Context) {
	sectionID := c.Param("sectionId")
	cr.DB.Where("module_id = ?", sectionID).Delete(&models.Lesson{})
	cr.DB.Where("id = ?", sectionID).Delete(&models.Module{})
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"id": sectionID}})
}

func (cr *CourseRouter) ReorderSections(c *gin.Context) {
	var input struct {
		Order []string `json:"order"`
	}
	c.ShouldBindJSON(&input)

	for idx, sID := range input.Order {
		cr.DB.Model(&models.Module{}).Where("id = ?", sID).Update("order_num", idx+1)
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (cr *CourseRouter) PromoteUngrouped(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// ═══════════════════════════════════════════════════════════════════════════
// 4. LESSONS MANAGEMENT (CRUD)
// ═══════════════════════════════════════════════════════════════════════════

func (cr *CourseRouter) ListLessons(c *gin.Context) {
	courseID := c.Param("id")

	var modules []models.Module
	cr.DB.Preload("Lessons").Where("course_id = ?", courseID).Find(&modules)

	lessonsList := make([]models.Lesson, 0)
	for _, m := range modules {
		lessonsList = append(lessonsList, m.Lessons...)
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": lessonsList})
}

type CreateLessonInput struct {
	Title         string `json:"title" binding:"required"`
	SectionID     string `json:"sectionId"`
	ModuleID      string `json:"moduleId"`
	Duration      string `json:"duration"`
	VideoURL      string `json:"videoUrl"`
	StreamType    string `json:"streamType"`
	IsFreePreview bool   `json:"isFreePreview"`
	Order         int    `json:"order"`
}

func (cr *CourseRouter) CreateLesson(c *gin.Context) {
	courseID := c.Param("id")

	var input CreateLessonInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	modID := input.SectionID
	if modID == "" {
		modID = input.ModuleID
	}

	// If no module specified, attach to first module or create one
	if modID == "" {
		var firstMod models.Module
		if err := cr.DB.Where("course_id = ?", courseID).First(&firstMod).Error; err == nil {
			modID = firstMod.ID
		} else {
			modID = "sec-" + uuid.New().String()[:8]
			cr.DB.Create(&models.Module{
				ID:        modID,
				CourseID:  courseID,
				Title:     "Module 1: Core Training",
				OrderNum:  1,
				CreatedAt: time.Now(),
			})
		}
	}

	slug := strings.ToLower(strings.ReplaceAll(input.Title, " ", "-"))
	lesson := models.Lesson{
		ID:            "les-" + uuid.New().String()[:8],
		ModuleID:      modID,
		CourseSlug:    courseID,
		Title:         input.Title,
		Slug:          slug,
		Duration:      input.Duration,
		VideoURL:      input.VideoURL,
		StreamType:    input.StreamType,
		IsFreePreview: input.IsFreePreview,
		HasVideo:      input.VideoURL != "",
		OrderNum:      input.Order,
		CreatedAt:     time.Now(),
	}
	if lesson.StreamType == "" {
		lesson.StreamType = "MP4"
	}
	if lesson.Duration == "" {
		lesson.Duration = "15m"
	}

	if err := cr.DB.Create(&lesson).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	// Increment module count
	cr.DB.Model(&models.Module{}).Where("id = ?", modID).UpdateColumn("lesson_count", gorm.Expr("lesson_count + 1"))
	cr.DB.Model(&models.Course{}).Where("id = ?", courseID).UpdateColumn("lessons_count", gorm.Expr("lessons_count + 1"))

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data": gin.H{
			"id":            lesson.ID,
			"title":         lesson.Title,
			"slug":          lesson.Slug,
			"moduleId":      lesson.ModuleID,
			"duration":      lesson.Duration,
			"videoUrl":      lesson.VideoURL,
			"streamType":    lesson.StreamType,
			"isFreePreview": lesson.IsFreePreview,
			"order":         lesson.OrderNum,
		},
	})
}

func (cr *CourseRouter) GetLessonDetail(c *gin.Context) {
	lessonID := c.Param("lessonId")

	var lesson models.Lesson
	if err := cr.DB.Where("id = ? OR slug = ?", lessonID, lessonID).First(&lesson).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Lesson not found"})
		return
	}

	ytID := extractYouTubeID(lesson.VideoURL)
	embedURL := toYouTubeEmbed(lesson.VideoURL)

	var exercises []models.Exercise
	cr.DB.Where("lesson_id = ? OR lesson_id = ?", lesson.ID, lesson.Slug).Find(&exercises)

	var exerciseData interface{} = nil
	if lesson.ExerciseTitle != "" {
		exerciseData = gin.H{
			"title":         lesson.ExerciseTitle,
			"description":   lesson.ExerciseDescription,
			"type":          lesson.ExerciseType,
			"estimatedMins": lesson.ExerciseEstimatedMins,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"id":              lesson.ID,
			"title":           lesson.Title,
			"slug":            lesson.Slug,
			"moduleId":        lesson.ModuleID,
			"courseSlug":      lesson.CourseSlug,
			"duration":        lesson.Duration,
			"videoUrl":        lesson.VideoURL,
			"embedUrl":        embedURL,
			"youtubeId":       ytID,
			"streamType":      lesson.StreamType,
			"isFreePreview":   lesson.IsFreePreview,
			"isQuiz":          lesson.IsQuiz,
			"hasVideo":        lesson.HasVideo,
			"order":           lesson.OrderNum,
			"description":     lesson.Description,
			"content":         lesson.Content,
			"keyPoints":       lesson.KeyPoints,
			"toolsRequired":   lesson.ToolsRequired,
			"exercise":        exerciseData,
			"exercises":       exercises,
			"lessonLanguages": []interface{}{},
		},
	})
}

func (cr *CourseRouter) UpdateLesson(c *gin.Context) {
	lessonID := c.Param("lessonId")

	var lesson models.Lesson
	if err := cr.DB.Where("id = ?", lessonID).First(&lesson).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Lesson not found"})
		return
	}

	var updateData map[string]interface{}
	c.ShouldBindJSON(&updateData)

	if title, ok := updateData["title"].(string); ok && title != "" {
		lesson.Title = title
	}
	if duration, ok := updateData["duration"].(string); ok {
		lesson.Duration = duration
	}
	if vURL, ok := updateData["videoUrl"].(string); ok {
		lesson.VideoURL = vURL
		lesson.HasVideo = vURL != ""
	}
	if isFree, ok := updateData["isFreePreview"].(bool); ok {
		lesson.IsFreePreview = isFree
	}
	cr.DB.Save(&lesson)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"id":            lesson.ID,
			"title":         lesson.Title,
			"slug":          lesson.Slug,
			"duration":      lesson.Duration,
			"videoUrl":      lesson.VideoURL,
			"isFreePreview": lesson.IsFreePreview,
		},
	})
}

func (cr *CourseRouter) DeleteLesson(c *gin.Context) {
	lessonID := c.Param("lessonId")
	cr.DB.Where("id = ?", lessonID).Delete(&models.Lesson{})
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"id": lessonID}})
}

func (cr *CourseRouter) ReorderLessons(c *gin.Context) {
	var input struct {
		Order []string `json:"order"`
	}
	c.ShouldBindJSON(&input)

	for idx, lID := range input.Order {
		cr.DB.Model(&models.Lesson{}).Where("id = ?", lID).Update("order_num", idx+1)
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// ═══════════════════════════════════════════════════════════════════════════
// 5. PROGRESS, COMMENTS, TAGS, AI-TUTOR
// ═══════════════════════════════════════════════════════════════════════════

type LessonProgressPayload struct {
	UserID         string `json:"userId"`
	WatchedSeconds int    `json:"watchedSeconds"`
	IsCompleted    bool   `json:"isCompleted"`
}

func (cr *CourseRouter) GetWatchProgress(c *gin.Context) {
	lessonID := c.Param("lessonId")
	userID := c.DefaultQuery("userId", "guest-tech")

	var prog models.UserLessonProgress
	err := cr.DB.Where("user_id = ? AND (lesson_id = ? OR lesson_slug = ?)", userID, lessonID, lessonID).First(&prog).Error
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"lessonId":       lessonID,
				"watchedSeconds": 0,
				"isCompleted":    false,
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"lessonId":       lessonID,
			"watchedSeconds": prog.WatchDurationSec,
			"isCompleted":    prog.IsCompleted,
		},
	})
}

func (cr *CourseRouter) UpdateWatchProgress(c *gin.Context) {
	lessonID := c.Param("lessonId")
	var payload LessonProgressPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	if payload.UserID == "" {
		payload.UserID = "guest-tech"
	}

	var lesson models.Lesson
	_ = cr.DB.Where("id = ? OR slug = ?", lessonID, lessonID).First(&lesson).Error

	var prog models.UserLessonProgress
	err := cr.DB.Where("user_id = ? AND (lesson_id = ? OR lesson_slug = ?)", payload.UserID, lessonID, lessonID).First(&prog).Error

	now := time.Now()
	if err != nil {
		prog = models.UserLessonProgress{
			ID:               uuid.New().String(),
			UserID:           payload.UserID,
			LessonID:         lessonID,
			CourseSlug:       lesson.CourseSlug,
			LessonSlug:       lesson.Slug,
			IsCompleted:      payload.IsCompleted,
			WatchDurationSec: payload.WatchedSeconds,
			LastWatchedAt:    now,
		}
		if payload.IsCompleted {
			prog.CompletedAt = &now
		}
		cr.DB.Create(&prog)
	} else {
		if payload.WatchedSeconds > prog.WatchDurationSec {
			prog.WatchDurationSec = payload.WatchedSeconds
		}
		if payload.IsCompleted {
			prog.IsCompleted = true
			if prog.CompletedAt == nil {
				prog.CompletedAt = &now
			}
		}
		prog.LastWatchedAt = now
		cr.DB.Save(&prog)
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": prog})
}

func (cr *CourseRouter) UpdateLessonCompletion(c *gin.Context) {
	lessonID := c.Param("lessonId")
	var payload LessonProgressPayload
	c.ShouldBindJSON(&payload)
	if payload.UserID == "" {
		payload.UserID = "guest-tech"
	}

	var lesson models.Lesson
	_ = cr.DB.Where("id = ? OR slug = ?", lessonID, lessonID).First(&lesson).Error

	var prog models.UserLessonProgress
	err := cr.DB.Where("user_id = ? AND (lesson_id = ? OR lesson_slug = ?)", payload.UserID, lessonID, lessonID).First(&prog).Error

	now := time.Now()
	if err != nil {
		prog = models.UserLessonProgress{
			ID:               uuid.New().String(),
			UserID:           payload.UserID,
			LessonID:         lessonID,
			CourseSlug:       lesson.CourseSlug,
			LessonSlug:       lesson.Slug,
			IsCompleted:      true,
			WatchDurationSec: payload.WatchedSeconds,
			LastWatchedAt:    now,
			CompletedAt:      &now,
		}
		cr.DB.Create(&prog)
	} else {
		prog.IsCompleted = true
		if prog.CompletedAt == nil {
			prog.CompletedAt = &now
		}
		prog.LastWatchedAt = now
		if payload.WatchedSeconds > prog.WatchDurationSec {
			prog.WatchDurationSec = payload.WatchedSeconds
		}
		cr.DB.Save(&prog)
	}

	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"message":     "Lesson marked completed! Next lesson in timeline unlocked.",
		"isCompleted": true,
		"data":        prog,
	})
}

func (cr *CourseRouter) GetComments(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"items": []interface{}{},
		},
	})
}

func (cr *CourseRouter) AddComment(c *gin.Context) {
	var body map[string]interface{}
	c.ShouldBindJSON(&body)
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": body})
}

func (cr *CourseRouter) UpdateComment(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (cr *CourseRouter) DeleteComment(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (cr *CourseRouter) GetCourseTags(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": []string{"Hardware", "Micro-Soldering", "PMIC"},
	})
}

func (cr *CourseRouter) UpdateCourseTags(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (cr *CourseRouter) GetAiTutor(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"enabled":     true,
			"persona":     "Salen Dangol AI Hardware Engineer",
			"instruction": "Explain chip-level diagnostics with diode test values and schematics references.",
		},
	})
}

func (cr *CourseRouter) UpdateAiTutor(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (cr *CourseRouter) GetCourseCurriculum(c *gin.Context) {
	courseID := c.Param("id")
	var modules []models.Module
	cr.DB.Preload("Lessons").Where("course_id = ?", courseID).Order("order_num asc").Find(&modules)
	c.JSON(http.StatusOK, gin.H{"success": true, "data": modules})
}

func (cr *CourseRouter) GetCourseAnalytics(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"views":       420,
			"completions": 68,
		},
	})
}

func (cr *CourseRouter) GetCourseMembers(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"success": true, "data": []interface{}{}})
}

func (cr *CourseRouter) GetGradebook(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"success": true, "data": []interface{}{}})
}

func (cr *CourseRouter) GetExercises(c *gin.Context) {
	courseID := c.Param("id")
	if courseID == "" {
		courseID = c.Param("courseSlug")
	}

	var exercises []models.Exercise
	query := cr.DB.Order("order_num asc")
	if courseID != "" {
		query = query.Where("course_id = ? OR course_slug = ?", courseID, courseID)
	}
	if modID := c.Query("moduleId"); modID != "" {
		query = query.Where("module_id = ?", modID)
	}
	if lesID := c.Query("lessonId"); lesID != "" {
		query = query.Where("lesson_id = ?", lesID)
	}

	query.Find(&exercises)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    exercises,
		"count":   len(exercises),
	})
}

func (cr *CourseRouter) GetExerciseDetail(c *gin.Context) {
	exerciseID := c.Param("exerciseId")

	var exercise models.Exercise
	if err := cr.DB.Where("id = ? OR slug = ?", exerciseID, exerciseID).First(&exercise).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Exercise not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    exercise,
	})
}

// ═══════════════════════════════════════════════════════════════════════════
// 6. ORG-SITE PLAYER & STREAMING (FOR FLUTTER VIDEO PLAYER)
// ═══════════════════════════════════════════════════════════════════════════

// Helper to extract YouTube video ID from standard or embed URL
func extractYouTubeID(videoURL string) string {
	if videoURL == "" {
		return ""
	}
	if strings.Contains(videoURL, "v=") {
		parts := strings.Split(videoURL, "v=")
		if len(parts) > 1 {
			id := parts[1]
			if idx := strings.Index(id, "&"); idx != -1 {
				id = id[:idx]
			}
			return id
		}
	}
	if strings.Contains(videoURL, "youtu.be/") {
		parts := strings.Split(videoURL, "youtu.be/")
		if len(parts) > 1 {
			id := parts[1]
			if idx := strings.Index(id, "?"); idx != -1 {
				id = id[:idx]
			}
			return id
		}
	}
	if strings.Contains(videoURL, "embed/") {
		parts := strings.Split(videoURL, "embed/")
		if len(parts) > 1 {
			id := parts[1]
			if idx := strings.Index(id, "?"); idx != -1 {
				id = id[:idx]
			}
			return id
		}
	}
	return ""
}

func toYouTubeEmbed(videoURL string) string {
	ytID := extractYouTubeID(videoURL)
	if ytID != "" {
		return "https://www.youtube.com/embed/" + ytID
	}
	return videoURL
}

func (cr *CourseRouter) GetCourseTree(c *gin.Context) {
	slug := c.Param("courseSlug")
	userId := c.Query("userId")
	if userId == "" {
		userId = "guest-tech"
	}

	var course models.Course
	err := cr.DB.Preload("Modules.Lessons", func(db *gorm.DB) *gorm.DB {
		return db.Order("order_num asc")
	}).Where("slug = ? OR id = ?", slug, slug).First(&course).Error

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Course not found: " + slug,
		})
		return
	}

	// 1. Fetch user progress for this course to evaluate timeline
	var userProgress []models.UserLessonProgress
	cr.DB.Where("user_id = ? AND course_slug = ?", userId, course.Slug).Find(&userProgress)

	completedMap := make(map[string]bool)
	watchedSecMap := make(map[string]int)
	for _, p := range userProgress {
		if p.IsCompleted {
			completedMap[p.LessonID] = true
		}
		watchedSecMap[p.LessonID] = p.WatchDurationSec
	}

	// Check if user is admin or instructor (bypass locking)
	isAdminOrInstructor := false
	var user models.User
	if err := cr.DB.Where("id = ?", userId).First(&user).Error; err == nil {
		if user.Role == "ADMIN" || user.Role == "INSTRUCTOR" {
			isAdminOrInstructor = true
		}
	}

	// Flatten all lessons into linear chronological timeline order
	type timelineEntry struct {
		Lesson *models.Lesson
	}
	var timeline []timelineEntry
	for _, m := range course.Modules {
		for i := range m.Lessons {
			timeline = append(timeline, timelineEntry{
				Lesson: &m.Lessons[i],
			})
		}
	}

	totalTimelineLessons := len(timeline)
	completedCount := 0
	currentActiveLessonID := ""

	type lessonTimelineStat struct {
		TimelineOrder int
		IsCompleted   bool
		IsLocked      bool
		CanAccess     bool
		WatchedSec    int
		PrevLessonID  string
		NextLessonID  string
	}
	statMap := make(map[string]lessonTimelineStat)

	for idx, item := range timeline {
		l := item.Lesson
		isCompleted := completedMap[l.ID]
		if isCompleted {
			completedCount++
		}

		isLocked := false
		if isAdminOrInstructor {
			isLocked = false
		} else if idx == 0 {
			// First lesson in course timeline is ALWAYS unlocked
			isLocked = false
		} else {
			// Lesson K is unlocked only if Lesson K-1 was completed OR it's a free preview
			prevItem := timeline[idx-1]
			prevCompleted := completedMap[prevItem.Lesson.ID]
			if prevCompleted || l.IsFreePreview {
				isLocked = false
			} else {
				isLocked = true
			}
		}

		if !isCompleted && !isLocked && currentActiveLessonID == "" {
			currentActiveLessonID = l.ID
		}

		prevID := ""
		if idx > 0 {
			prevID = timeline[idx-1].Lesson.ID
		}
		nextID := ""
		if idx < len(timeline)-1 {
			nextID = timeline[idx+1].Lesson.ID
		}

		statMap[l.ID] = lessonTimelineStat{
			TimelineOrder: idx + 1,
			IsCompleted:   isCompleted,
			IsLocked:      isLocked,
			CanAccess:     !isLocked,
			WatchedSec:    watchedSecMap[l.ID],
			PrevLessonID:  prevID,
			NextLessonID:  nextID,
		}
	}

	if currentActiveLessonID == "" && len(timeline) > 0 {
		currentActiveLessonID = timeline[0].Lesson.ID
	}

	// 2. Fetch schematics
	var schematics []models.Schematic
	cr.DB.Where("course_slug = ?", course.Slug).Find(&schematics)

	// 3. Format modules with timeline metadata and YouTube embeds
	modulesList := make([]map[string]interface{}, 0, len(course.Modules))
	for _, m := range course.Modules {
		lessonsList := make([]map[string]interface{}, 0, len(m.Lessons))
		for _, l := range m.Lessons {
			stat := statMap[l.ID]
			ytID := extractYouTubeID(l.VideoURL)
			embedURL := toYouTubeEmbed(l.VideoURL)

			lessonsList = append(lessonsList, map[string]interface{}{
				"id":               l.ID,
				"title":            l.Title,
				"slug":             l.Slug,
				"duration":         l.Duration,
				"videoUrl":         l.VideoURL,
				"embedUrl":         embedURL,
				"youtubeId":        ytID,
				"streamType":       l.StreamType,
				"isFreePreview":    l.IsFreePreview,
				"isQuiz":           l.IsQuiz,
				"hasVideo":         l.HasVideo,
				"order":            l.OrderNum,
				"timelineOrder":    stat.TimelineOrder,
				"isCompleted":      stat.IsCompleted,
				"isLocked":         stat.IsLocked,
				"canAccess":        stat.CanAccess,
				"watchedSeconds":   stat.WatchedSec,
				"previousLessonId": stat.PrevLessonID,
				"nextLessonId":     stat.NextLessonID,
				"description":      l.Description,
				"hasExercise":      l.ExerciseTitle != "",
				"exerciseTitle":    l.ExerciseTitle,
				"exerciseType":     l.ExerciseType,
			})
		}

		modulesList = append(modulesList, map[string]interface{}{
			"id":          m.ID,
			"title":       m.Title,
			"lessonCount": len(m.Lessons),
			"lessons":     lessonsList,
		})
	}

	schematicList := make([]map[string]interface{}, 0, len(schematics))
	for _, s := range schematics {
		schematicList = append(schematicList, map[string]interface{}{
			"id":          s.ID,
			"title":       s.Title,
			"category":    s.Category,
			"fileName":    s.FileName,
			"downloadUrl": s.DownloadURL,
			"fileSize":    s.FileSize,
			"totalPages":  s.TotalPages,
		})
	}

	var courseExercises []models.Exercise
	cr.DB.Where("course_slug = ? OR course_id = ?", course.Slug, course.ID).Order("order_num asc").Find(&courseExercises)

	overallPercent := 0
	if totalTimelineLessons > 0 {
		overallPercent = (completedCount * 100) / totalTimelineLessons
	}

	courseData := map[string]interface{}{
		"id":            course.ID,
		"title":         course.Title,
		"slug":          course.Slug,
		"description":   course.Description,
		"price":         course.Price,
		"originalPrice": course.OriginalPrice,
		"rating":        course.Rating,
		"reviews":       course.ReviewsCount,
		"lessons":       course.LessonsCount,
		"duration":      course.Duration,
		"level":         course.Level,
		"imageAsset":    course.ImageAsset,
		"instructor": map[string]interface{}{
			"name": course.Instructor,
			"bio":  course.InstructorBio,
		},
		"modules":    modulesList,
		"schematics": schematicList,
		"exercises":  courseExercises,
		"timeline": map[string]interface{}{
			"totalLessons":          totalTimelineLessons,
			"completedLessons":      completedCount,
			"progressPercent":       overallPercent,
			"currentActiveLessonId": currentActiveLessonID,
			"isFullyCompleted":      completedCount >= totalTimelineLessons && totalTimelineLessons > 0,
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    courseData,
	})
}

func (cr *CourseRouter) GetCourseLesson(c *gin.Context) {
	courseSlug := c.Param("courseSlug")
	itemSlug := c.Param("itemSlug")
	userId := c.Query("userId")
	if userId == "" {
		userId = "guest-tech"
	}

	var lesson models.Lesson
	err := cr.DB.Where("course_slug = ? AND (slug = ? OR id = ?)", courseSlug, itemSlug, itemSlug).First(&lesson).Error
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Lesson not found: " + itemSlug + " in course " + courseSlug,
		})
		return
	}

	var prog models.UserLessonProgress
	_ = cr.DB.Where("user_id = ? AND lesson_id = ?", userId, lesson.ID).First(&prog).Error

	ytID := extractYouTubeID(lesson.VideoURL)
	embedURL := toYouTubeEmbed(lesson.VideoURL)

	var exercises []models.Exercise
	cr.DB.Where("(course_slug = ? OR course_id = ?) AND (lesson_id = ? OR lesson_id = ?)", courseSlug, courseSlug, lesson.ID, lesson.Slug).Find(&exercises)

	var exerciseData interface{} = nil
	if lesson.ExerciseTitle != "" {
		exerciseData = gin.H{
			"title":         lesson.ExerciseTitle,
			"description":   lesson.ExerciseDescription,
			"type":          lesson.ExerciseType,
			"estimatedMins": lesson.ExerciseEstimatedMins,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": map[string]interface{}{
			"id":             lesson.ID,
			"title":          lesson.Title,
			"slug":           lesson.Slug,
			"duration":       lesson.Duration,
			"videoUrl":       lesson.VideoURL,
			"embedUrl":       embedURL,
			"youtubeId":      ytID,
			"streamType":     lesson.StreamType,
			"isFreePreview":  lesson.IsFreePreview,
			"isQuiz":         lesson.IsQuiz,
			"hasVideo":       lesson.HasVideo,
			"isCompleted":    prog.IsCompleted,
			"watchedSeconds": prog.WatchDurationSec,
			"description":    lesson.Description,
			"content":        lesson.Content,
			"keyPoints":      lesson.KeyPoints,
			"toolsRequired":  lesson.ToolsRequired,
			"exercise":       exerciseData,
			"exercises":      exercises,
		},
	})
}

func (cr *CourseRouter) GetVideoStream(c *gin.Context) {
	courseSlug := c.Param("courseSlug")
	itemSlug := c.Param("itemSlug")
	userId := c.Query("userId")

	var lesson models.Lesson
	err := cr.DB.Where("course_slug = ? AND (slug = ? OR id = ?)", courseSlug, itemSlug, itemSlug).First(&lesson).Error
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Lesson not found",
		})
		return
	}

	ytID := extractYouTubeID(lesson.VideoURL)
	embedURL := toYouTubeEmbed(lesson.VideoURL)

	// Timeline sequence check: if user is not admin and lesson is not free preview,
	// enforce that previous lesson was completed!
	if userId != "" && !lesson.IsFreePreview {
		isAdmin := false
		var user models.User
		if err := cr.DB.Where("id = ?", userId).First(&user).Error; err == nil && (user.Role == "ADMIN" || user.Role == "INSTRUCTOR") {
			isAdmin = true
		}

		if !isAdmin {
			var prevLesson models.Lesson
			err := cr.DB.Where("course_slug = ? AND order_num < ?", courseSlug, lesson.OrderNum).Order("order_num desc").First(&prevLesson).Error
			if err == nil {
				var prevProg models.UserLessonProgress
				err := cr.DB.Where("user_id = ? AND lesson_id = ? AND is_completed = ?", userId, prevLesson.ID, true).First(&prevProg).Error
				if err != nil {
					c.JSON(http.StatusForbidden, gin.H{
						"success":             false,
						"isLocked":            true,
						"code":                "TIMELINE_LOCKED",
						"error":               "This lesson is locked in the timeline. Please complete the previous lesson first.",
						"previousLessonId":    prevLesson.ID,
						"previousLessonTitle": prevLesson.Title,
					})
					return
				}
			}
		}
	}

	// If subscriber or free preview or lesson has video:
	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"streamUrl":  lesson.VideoURL,
		"embedUrl":   embedURL,
		"youtubeId":  ytID,
		"streamType": lesson.StreamType,
	})
}

func (cr *CourseRouter) MintHlsCookie(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"token":     "hls-jwt-token-" + strconv.FormatInt(time.Now().Unix(), 10),
		"expiresAt": time.Now().Add(24 * time.Hour).Format(time.RFC3339),
	})
}

// ═══════════════════════════════════════════════════════════════════════════
// 9. NEW V1 — COURSE DETAIL SCREEN ENDPOINTS
// ═══════════════════════════════════════════════════════════════════════════

// GET /api/v1/courses/:courseSlug/curriculum
//
// Returns the complete module + lesson tree for the course detail screen.
// Quiz lessons include their linked quizId so the app can navigate to the
// quiz-taking screen with the real quizId pre-loaded.
func (cr *CourseRouter) GetCurriculumV1(c *gin.Context) {
	slug := c.Param("courseSlug")

	// Load course with modules and lessons ordered correctly
	var course models.Course
	if err := cr.DB.Preload("Modules.Lessons", func(db *gorm.DB) *gorm.DB {
		return db.Order("order_num asc")
	}).Where("slug = ? OR id = ?", slug, slug).First(&course).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Course not found: " + slug})
		return
	}

	// Pre-load all quizzes for this course, keyed by module_id so we can attach
	// quizId to the quiz lesson row without a per-lesson query.
	type quizRow struct {
		ID       string `gorm:"column:id"`
		ModuleID string `gorm:"column:module_id"`
	}
	var quizRows []quizRow
	cr.DB.Table("quizzes").Select("id, module_id").
		Where("course_slug = ?", course.Slug).Scan(&quizRows)
	moduleQuizMap := make(map[string]string) // moduleID → quizID
	for _, q := range quizRows {
		moduleQuizMap[q.ModuleID] = q.ID
	}

	// Build the response shape
	type lessonOut struct {
		ID            string `json:"id"`
		Slug          string `json:"slug"`
		Title         string `json:"title"`
		Duration      string `json:"duration"`
		VideoURL      string `json:"videoUrl,omitempty"`
		StreamType    string `json:"streamType,omitempty"`
		YoutubeID     string `json:"youtubeId,omitempty"`
		IsFreePreview bool   `json:"isFreePreview"`
		IsQuiz        bool   `json:"isQuiz"`
		QuizID        string `json:"quizId,omitempty"`
		OrderNum      int    `json:"order"`
	}
	type moduleOut struct {
		ID          string      `json:"id"`
		Title       string      `json:"title"`
		LessonCount int         `json:"lessonCount"`
		OrderNum    int         `json:"order"`
		Lessons     []lessonOut `json:"lessons"`
	}

	totalLessons := 0
	totalQuizzes := 0
	modulesOut := make([]moduleOut, 0, len(course.Modules))

	for _, m := range course.Modules {
		lessons := make([]lessonOut, 0, len(m.Lessons))
		for _, l := range m.Lessons {
			ytID := extractYouTubeID(l.VideoURL)
			quizID := ""
			if l.IsQuiz {
				// First try module-level quiz lookup, then course-level fallback
				if qid, ok := moduleQuizMap[m.ID]; ok {
					quizID = qid
				} else {
					// Fallback: any quiz for this course
					var qFallback struct{ ID string }
					cr.DB.Table("quizzes").Select("id").
						Where("course_slug = ?", course.Slug).
						Limit(1).Scan(&qFallback)
					quizID = qFallback.ID
				}
				totalQuizzes++
			}
			totalLessons++
			lessons = append(lessons, lessonOut{
				ID:            l.ID,
				Slug:          l.Slug,
				Title:         l.Title,
				Duration:      l.Duration,
				VideoURL:      l.VideoURL,
				StreamType:    l.StreamType,
				YoutubeID:     ytID,
				IsFreePreview: l.IsFreePreview,
				IsQuiz:        l.IsQuiz,
				QuizID:        quizID,
				OrderNum:      l.OrderNum,
			})
		}
		modulesOut = append(modulesOut, moduleOut{
			ID:          m.ID,
			Title:       m.Title,
			LessonCount: len(lessons),
			OrderNum:    m.OrderNum,
			Lessons:     lessons,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"courseId":     course.ID,
			"courseSlug":   course.Slug,
			"courseTitle":  course.Title,
			"totalLessons": totalLessons,
			"totalQuizzes": totalQuizzes,
			"modules":      modulesOut,
		},
	})
}

// GET /api/v1/users/:userId/courses/:courseSlug/progress
//
// Returns the student's completed lessons, passed quizzes, and overall
// progress percentage for the specified course.
func (cr *CourseRouter) GetCourseProgressV1(c *gin.Context) {
	userID := c.Param("userId")
	if userID == "" {
		userID = c.Query("userId")
	}
	var u struct{ ID string }
	if cr.DB.Table("users").Select("id").Where("id = ? OR LOWER(email) = LOWER(?)", userID, userID).Scan(&u).Error == nil && u.ID != "" {
		userID = u.ID
	}
	courseSlug := c.Param("courseSlug")

	// Count total lessons in the course
	var totalLessons int64
	cr.DB.Table("lessons").Where("course_slug = ?", courseSlug).Count(&totalLessons)
	if totalLessons == 0 {
		// Fallback: use lessons_count from courses table
		var crs struct{ LessonsCount int }
		cr.DB.Table("courses").Select("lessons_count").Where("slug = ?", courseSlug).Scan(&crs)
		totalLessons = int64(crs.LessonsCount)
	}

	// Fetch all completed progress rows for this user+course
	var progressRows []models.UserLessonProgress
	cr.DB.Where("user_id = ? AND course_slug = ? AND is_completed = ?", userID, courseSlug, true).
		Order("last_watched_at desc").
		Find(&progressRows)

	completedSlugs := make([]string, 0, len(progressRows))
	lastWatchedSlug := ""
	for i, p := range progressRows {
		completedSlugs = append(completedSlugs, p.LessonSlug)
		if i == 0 {
			lastWatchedSlug = p.LessonSlug
		}
	}

	// Find last-watched lesson (including in-progress, not just completed)
	var lastWatched struct{ LessonSlug string }
	cr.DB.Table("user_lesson_progress").
		Select("lesson_slug").
		Where("user_id = ? AND course_slug = ?", userID, courseSlug).
		Order("last_watched_at desc").
		Limit(1).
		Scan(&lastWatched)
	if lastWatched.LessonSlug != "" {
		lastWatchedSlug = lastWatched.LessonSlug
	}

	// Fetch passed quiz IDs from quiz_results
	type quizResultRow struct {
		QuizID string `gorm:"column:quiz_id"`
	}
	var passedQuizRows []quizResultRow
	cr.DB.Table("quiz_results").
		Select("quiz_id").
		Where("user_id = ? AND course_slug = ? AND is_passed = ?", userID, courseSlug, true).
		Find(&passedQuizRows)

	passedQuizIDs := make([]string, 0, len(passedQuizRows))
	for _, r := range passedQuizRows {
		passedQuizIDs = append(passedQuizIDs, r.QuizID)
	}

	// Compute progress percent
	completedCount := int64(len(completedSlugs))
	progressPct := 0.0
	if totalLessons > 0 {
		progressPct = (float64(completedCount) / float64(totalLessons)) * 100.0
	}
	if progressPct > 100 {
		progressPct = 100
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"completedLessonSlugs": completedSlugs,
			"completedCount":       completedCount,
			"totalLessons":         totalLessons,
			"passedQuizIds":        passedQuizIDs,
			"quizzesPassedCount":   len(passedQuizIDs),
			"lastWatchedLessonSlug": lastWatchedSlug,
			"progressPercent":      progressPct,
		},
	})
}

// GET /api/v1/courses/:courseSlug/resources
//
// Returns downloadable schematics and PDF guides for a course.
// Source: schematics table, filtered by course_slug.
func (cr *CourseRouter) GetCourseResourcesV1(c *gin.Context) {
	slug := c.Param("courseSlug")

	var schematics []models.Schematic
	if err := cr.DB.Where("course_slug = ?", slug).Order("id asc").Find(&schematics).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	type resourceOut struct {
		ID          string `json:"id"`
		Title       string `json:"title"`
		Category    string `json:"category"`
		FileType    string `json:"fileType"`
		FileSize    string `json:"fileSize"`
		TotalPages  int    `json:"totalPages"`
		DownloadURL string `json:"downloadUrl"`
	}

	result := make([]resourceOut, 0, len(schematics))
	for _, s := range schematics {
		// Ensure download URL is absolute for clients that don't know the base URL
		dlURL := s.DownloadURL
		if len(dlURL) > 0 && dlURL[0] == '/' {
			dlURL = "https://salenmobileguru.nshlms.com" + dlURL
		}
		result = append(result, resourceOut{
			ID:          s.ID,
			Title:       s.Title,
			Category:    s.Category,
			FileType:    "PDF",
			FileSize:    s.FileSize,
			TotalPages:  s.TotalPages,
			DownloadURL: dlURL,
		})
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}
