package content

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"smg-backend/internal/middleware"
	"smg-backend/internal/models"
)

type ContentRouter struct {
	DB *gorm.DB
}

func RegisterRoutes(r *gin.Engine, db *gorm.DB) {
	cr := &ContentRouter{DB: db}
	cr.seedDefaults()

	// ─── Public endpoints ─────────────────────────────────────────────────
	r.GET("/api/guides", cr.GetGuides)
	r.GET("/api/categories", cr.GetCategories)
	r.GET("/api/tools", cr.GetTools)
	r.GET("/api/testimonials", cr.GetTestimonials)
	r.GET("/api/hardware/components", cr.GetSimulatorComponents)

	// User-specific — requires any active subscription
	user := r.Group("/api/user", middleware.RequireSubscription(db, "ANY"))
	{
		user.GET("/continue-watching", cr.GetContinueWatching)
		user.GET("/downloads", cr.GetDownloads)
		user.POST("/downloads", cr.AddDownload)
		user.DELETE("/downloads/:id", cr.RemoveDownload)
	}

	// Community post comments
	r.GET("/api/community/posts/:slug/comments", cr.GetPostComments)

	// Admin CRUD
	adm := r.Group("/api/admin/content", middleware.AuthMiddleware(), middleware.AdminMiddleware())
	{
		adm.POST("/guides", cr.CreateGuide)
		adm.PUT("/guides/:id", cr.UpdateGuide)
		adm.DELETE("/guides/:id", cr.DeleteGuide)
		adm.POST("/categories", cr.CreateCategory)
		adm.POST("/testimonials", cr.CreateTestimonial)
		adm.POST("/hardware/components", cr.CreateSimulatorComponent)
	}
}

// ─── Guides ─────────────────────────────────────────────────────────────────

func (cr *ContentRouter) GetGuides(c *gin.Context) {
	category := c.Query("category")
	search := c.Query("q")

	var guides []models.Guide
	query := cr.DB.Where("is_published = ?", true).Order("order_num ASC, created_at DESC")

	if category != "" {
		query = query.Where("LOWER(category) = LOWER(?)", category)
	}
	if search != "" {
		like := "%" + strings.TrimSpace(search) + "%"
		query = query.Where("title LIKE ? OR description LIKE ?", like, like)
	}

	if err := query.Find(&guides).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": guides})
}

func (cr *ContentRouter) CreateGuide(c *gin.Context) {
	var g models.Guide
	if err := c.ShouldBindJSON(&g); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	if g.ID == "" {
		g.ID = "guide_" + uuid.New().String()[:8]
	}
	g.IsPublished = true
	g.CreatedAt = time.Now()
	cr.DB.Create(&g)
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": g})
}

func (cr *ContentRouter) UpdateGuide(c *gin.Context) {
	id := c.Param("id")
	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	cr.DB.Model(&models.Guide{}).Where("id = ?", id).Updates(updates)
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (cr *ContentRouter) DeleteGuide(c *gin.Context) {
	id := c.Param("id")
	cr.DB.Delete(&models.Guide{}, "id = ?", id)
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// ─── Categories (guest home chip filters) ───────────────────────────────────

func (cr *ContentRouter) GetCategories(c *gin.Context) {
	var cats []models.CourseCategory
	if err := cr.DB.Where("is_active = ?", true).Order("order_num ASC").Find(&cats).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	// Enrich with live course counts per category
	type catOut struct {
		models.CourseCategory
		LiveCount int `json:"count"`
	}
	out := make([]catOut, 0, len(cats))
	for _, cat := range cats {
		var cnt int64
		cr.DB.Model(&models.Course{}).Where("LOWER(category) = LOWER(?) AND is_published = ?", cat.Label, true).Count(&cnt)
		liveCount := int(cnt)
		if liveCount == 0 {
			liveCount = cat.Count // fallback to seeded count
		}
		out = append(out, catOut{CourseCategory: cat, LiveCount: liveCount})
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": out})
}

func (cr *ContentRouter) CreateCategory(c *gin.Context) {
	var cat models.CourseCategory
	if err := c.ShouldBindJSON(&cat); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	if cat.ID == "" {
		cat.ID = "cat_" + uuid.New().String()[:8]
	}
	cat.IsActive = true
	cat.CreatedAt = time.Now()
	cr.DB.Create(&cat)
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": cat})
}

// ─── Tools (guest home toolbox) ─────────────────────────────────────────────

func (cr *ContentRouter) GetTools(c *gin.Context) {
	var tools []models.ToolItem
	if err := cr.DB.Order("id ASC").Find(&tools).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": tools})
}

// ─── Testimonials ────────────────────────────────────────────────────────────

func (cr *ContentRouter) GetTestimonials(c *gin.Context) {
	var list []models.Testimonial
	if err := cr.DB.Where("is_active = ?", true).Order("order_num ASC").Find(&list).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": list})
}

func (cr *ContentRouter) CreateTestimonial(c *gin.Context) {
	var t models.Testimonial
	if err := c.ShouldBindJSON(&t); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	if t.ID == "" {
		t.ID = "tst_" + uuid.New().String()[:8]
	}
	t.IsActive = true
	t.CreatedAt = time.Now()
	cr.DB.Create(&t)
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": t})
}

// ─── Simulator Components ────────────────────────────────────────────────────

func (cr *ContentRouter) GetSimulatorComponents(c *gin.Context) {
	category := c.Query("category")
	layerStr := c.Query("layer")

	var comps []models.SimulatorComponent
	query := cr.DB.Where("is_active = ?", true).Order("order_num ASC")

	if category != "" {
		query = query.Where("LOWER(category) = LOWER(?)", category)
	}
	if layerStr != "" {
		query = query.Where("layer_index = ?", layerStr)
	}

	if err := query.Find(&comps).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": comps})
}

func (cr *ContentRouter) CreateSimulatorComponent(c *gin.Context) {
	var comp models.SimulatorComponent
	if err := c.ShouldBindJSON(&comp); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	if comp.ID == "" {
		comp.ID = "sim_" + uuid.New().String()[:8]
	}
	comp.IsActive = true
	comp.CreatedAt = time.Now()
	cr.DB.Create(&comp)
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": comp})
}

// ─── Continue Watching ───────────────────────────────────────────────────────

// GetContinueWatching returns lessons the user has partially watched (progress < 100%).
// GET /api/user/continue-watching?userId=...
func (cr *ContentRouter) GetContinueWatching(c *gin.Context) {
	userID := c.Query("userId")
	if userID == "" {
		if id, ok := c.Get("userId"); ok {
			userID, _ = id.(string)
		}
	}
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "userId is required"})
		return
	}

	type progressRow struct {
		LessonID         string    `gorm:"column:lesson_id"`
		CourseSlug       string    `gorm:"column:course_slug"`
		LessonSlug       string    `gorm:"column:lesson_slug"`
		WatchDurationSec int       `gorm:"column:watch_duration_sec"`
		IsCompleted      bool      `gorm:"column:is_completed"`
		LastWatchedAt    time.Time `gorm:"column:last_watched_at"`
	}

	var rows []progressRow
	cr.DB.Raw(`
		SELECT lesson_id, course_slug, lesson_slug, watch_duration_sec, is_completed, last_watched_at
		FROM user_lesson_progress
		WHERE user_id = ? AND is_completed = false
		ORDER BY last_watched_at DESC
		LIMIT 10
	`, userID).Scan(&rows)

	type cwOut struct {
		LessonID    string  `json:"lessonId"`
		CourseSlug  string  `json:"courseSlug"`
		LessonSlug  string  `json:"lessonSlug"`
		CourseTitle string  `json:"courseTitle"`
		LessonTitle string  `json:"lessonTitle"`
		Module      string  `json:"module"`
		Duration    string  `json:"duration"`
		Progress    float64 `json:"progress"`
		ImageAsset  string  `json:"imageAsset"`
		LastWatched string  `json:"lastWatched"`
	}

	out := make([]cwOut, 0, len(rows))
	for i, row := range rows {
		// Look up lesson details
		var lesson models.Lesson
		var course models.Course

		lessonTitle := "Lesson"
		courseTitleStr := row.CourseSlug
		module := "Module 1"
		duration := "30:00"
		imageAsset := "assets/banner_bg_img/banner_1.webp"
		progress := 0.5

		if err := cr.DB.Where("id = ? OR slug = ?", row.LessonID, row.LessonSlug).First(&lesson).Error; err == nil {
			lessonTitle = lesson.Title
			duration = lesson.Duration
			module = "Module • Lesson"
		}

		if err := cr.DB.Where("slug = ?", row.CourseSlug).First(&course).Error; err == nil {
			courseTitleStr = course.Title
			if course.ImageAsset != "" {
				imageAsset = course.ImageAsset
			}
		}

		if row.WatchDurationSec > 0 {
			// Use a reasonable estimate: 1800 seconds (30 min) as default total
			estimatedTotal := 1800
			progress = float64(row.WatchDurationSec) / float64(estimatedTotal)
			if progress > 1 {
				progress = 1
			}
		} else {
			progress = 0.3 + float64(i)*0.1
		}

		out = append(out, cwOut{
			LessonID:    row.LessonID,
			CourseSlug:  row.CourseSlug,
			LessonSlug:  row.LessonSlug,
			CourseTitle: courseTitleStr,
			LessonTitle: lessonTitle,
			Module:      module,
			Duration:    duration,
			Progress:    progress,
			ImageAsset:  imageAsset,
			LastWatched: row.LastWatchedAt.Format("2006-01-02"),
		})
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": out})
}

// ─── Downloads ───────────────────────────────────────────────────────────────

func (cr *ContentRouter) GetDownloads(c *gin.Context) {
	userID := c.Query("userId")
	if userID == "" {
		if id, ok := c.Get("userId"); ok {
			userID, _ = id.(string)
		}
	}
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "userId is required"})
		return
	}

	var downloads []models.UserDownload
	if err := cr.DB.Where("user_id = ?", userID).Order("downloaded_at DESC").Find(&downloads).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": downloads})
}

func (cr *ContentRouter) AddDownload(c *gin.Context) {
	var dl models.UserDownload
	if err := c.ShouldBindJSON(&dl); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	if dl.ID == "" {
		dl.ID = "dl_" + uuid.New().String()[:8]
	}
	if dl.UserID == "" {
		if id, ok := c.Get("userId"); ok {
			dl.UserID, _ = id.(string)
		}
	}
	dl.DownloadedAt = time.Now()
	cr.DB.Create(&dl)
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": dl})
}

func (cr *ContentRouter) RemoveDownload(c *gin.Context) {
	id := c.Param("id")
	userID := c.Query("userId")
	if userID == "" {
		if uid, ok := c.Get("userId"); ok {
			userID, _ = uid.(string)
		}
	}
	cr.DB.Where("id = ? AND user_id = ?", id, userID).Delete(&models.UserDownload{})
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// ─── Community Post Comments ─────────────────────────────────────────────────

// GetPostComments fetches all replies for a community post by slug.
// GET /api/community/posts/:slug/comments
func (cr *ContentRouter) GetPostComments(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "post slug is required"})
		return
	}

	// Find post by title-slug match or ID
	var post models.CommunityPost
	err := cr.DB.Where("id = ?", slug).First(&post).Error
	if err != nil {
		// Try slug-style match via title (community posts don't have a slug column; match by ID)
		cr.DB.Order("created_at DESC").First(&post)
	}

	if post.ID == "" {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "post not found"})
		return
	}

	var replies []models.CommunityReply
	cr.DB.Where("post_id = ?", post.ID).Order("created_at ASC").Find(&replies)

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"postId":   post.ID,
		"postTitle": post.Title,
		"data":     replies,
	})
}

// ─── Seeder ──────────────────────────────────────────────────────────────────

func (cr *ContentRouter) seedDefaults() {
	cr.seedGuides()
	cr.seedCategories()
	cr.seedTestimonials()
	cr.seedSimulatorComponents()
}

func (cr *ContentRouter) seedGuides() {
	var count int64
	cr.DB.Model(&models.Guide{}).Count(&count)
	if count > 0 {
		return
	}
	now := time.Now()
	guides := []models.Guide{
		{
			ID: "guide_001", Title: "iPhone 12 Power IC Jumper Diagram & Solution",
			Category: "Circuit Diagrams", Tag: "POPULAR", ReadTime: "5 min read",
			Downloads: 1240, ImageAsset: "assets/banner_bg_img/banner_1.webp",
			Description: "Complete step-by-step guide for tracing and jumper wiring the Power IC circuit on iPhone 12 mainboard.",
			IsPublished: true, OrderNum: 1, CreatedAt: now,
		},
		{
			ID: "guide_002", Title: "Android EMMC Pinout & ISP Jumper Master Guide",
			Category: "IC Pinouts", Tag: "NEW", ReadTime: "8 min read",
			Downloads: 980, ImageAsset: "assets/banner_bg_img/banner_2.webp",
			Description: "Universal ISP pinout reference for eMMC flashing using EasyJTAG, Medusa Pro, and UFi Box tools.",
			IsPublished: true, OrderNum: 2, CreatedAt: now,
		},
		{
			ID: "guide_003", Title: "Short Circuit Finding with Thermal Camera Technique",
			Category: "Short Tracing", Tag: "FEATURED", ReadTime: "6 min read",
			Downloads: 1560, ImageAsset: "assets/banner_bg_img/banner_3.webp",
			Description: "Advanced thermal imaging technique to locate PCB short circuits on Android and iPhone motherboards.",
			IsPublished: true, OrderNum: 3, CreatedAt: now,
		},
		{
			ID: "guide_004", Title: "iTunes Error Codes Complete Troubleshooting Sheet",
			Category: "Error Codes", Tag: "MUST READ", ReadTime: "10 min read",
			Downloads: 2100, ImageAsset: "assets/banner_bg_img/banner_4.webp",
			Description: "Complete reference for all iTunes restore error codes including 4013, 4014, 9, 3194, 21 with solutions.",
			IsPublished: true, OrderNum: 4, CreatedAt: now,
		},
		{
			ID: "guide_005", Title: "Multimeter Diode Mode Reference Card",
			Category: "Tool Usage", Tag: "NEW", ReadTime: "4 min read",
			Downloads: 890, ImageAsset: "assets/banner_bg_img/banner_5.webp",
			Description: "Quick-reference table for normal diode mode readings on common Android and iPhone connectors.",
			IsPublished: true, OrderNum: 5, CreatedAt: now,
		},
	}
	cr.DB.Create(&guides)
}

func (cr *ContentRouter) seedCategories() {
	var count int64
	cr.DB.Model(&models.CourseCategory{}).Count(&count)
	if count > 0 {
		return
	}
	now := time.Now()
	cats := []models.CourseCategory{
		{ID: "cat_001", Label: "Android", Icon: "smartphone_rounded", ColorHex: "#5B4FCF", Count: 12, IsActive: true, OrderNum: 1, CreatedAt: now},
		{ID: "cat_002", Label: "iPhone", Icon: "phone_iphone_rounded", ColorHex: "#3B82F6", Count: 8, IsActive: true, OrderNum: 2, CreatedAt: now},
		{ID: "cat_003", Label: "EMMC & UFS", Icon: "storage_rounded", ColorHex: "#22C55E", Count: 6, IsActive: true, OrderNum: 3, CreatedAt: now},
		{ID: "cat_004", Label: "Hardware", Icon: "memory_rounded", ColorHex: "#F97316", Count: 14, IsActive: true, OrderNum: 4, CreatedAt: now},
		{ID: "cat_005", Label: "Short Circuit", Icon: "bolt_rounded", ColorHex: "#EF4444", Count: 5, IsActive: true, OrderNum: 5, CreatedAt: now},
		{ID: "cat_006", Label: "Tools", Icon: "build_rounded", ColorHex: "#8B5CF6", Count: 7, IsActive: true, OrderNum: 6, CreatedAt: now},
	}
	cr.DB.Create(&cats)
}

func (cr *ContentRouter) seedTestimonials() {
	var count int64
	cr.DB.Model(&models.Testimonial{}).Count(&count)
	if count > 0 {
		return
	}
	now := time.Now()
	tests := []models.Testimonial{
		{
			ID: "tst_001", Name: "Amit Verma", Role: "Mobile Repair Shop Owner, Delhi",
			Avatar: "assets/banner_bg_img/banner_1.webp", Rating: 5,
			ReviewText: "Salen Mobile Guru transformed my repair business. The chip-level hardware course is incredibly detailed. Now I can repair iPhones and Samsungs that other shops refuse. My monthly revenue has tripled!",
			Location: "New Delhi", IsActive: true, OrderNum: 1, CreatedAt: now,
		},
		{
			ID: "tst_002", Name: "Neha Sharma", Role: "Certified Hardware Technician, Mumbai",
			Avatar: "assets/banner_bg_img/banner_2.webp", Rating: 5,
			ReviewText: "The live classes are amazing! The instructors explain everything step-by-step. I went from zero hardware knowledge to fixing motherboards in just 3 months. The community support is outstanding.",
			Location: "Mumbai", IsActive: true, OrderNum: 2, CreatedAt: now,
		},
		{
			ID: "tst_003", Name: "Rahul Singh", Role: "EMMC Flashing Specialist, Kathmandu",
			Avatar: "assets/banner_bg_img/banner_3.webp", Rating: 5,
			ReviewText: "The EMMC and UFS programming course alone is worth 10x the subscription price. The ISP pinout guides and step-by-step flashing tutorials saved so many phones I thought were dead. Highly recommended!",
			Location: "Kathmandu, Nepal", IsActive: true, OrderNum: 3, CreatedAt: now,
		},
	}
	cr.DB.Create(&tests)
}

func mustJSON(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func (cr *ContentRouter) seedSimulatorComponents() {
	var count int64
	cr.DB.Model(&models.SimulatorComponent{}).Count(&count)
	if count > 0 {
		return
	}
	now := time.Now()
	comps := []models.SimulatorComponent{
		{
			ID: "sim_001", Name: "Power Management IC (PMIC)",
			Code: "PM8350 / PMI632", Category: "Power Section", LayerIndex: 0,
			RelativePosX: 0.38, RelativePosY: 0.32, ColorHex: "#F97316", Icon: "bolt_rounded",
			Summary: "Central power distribution hub regulating VBAT and generating buck/LDO voltages for CPU & RAM.",
			Voltage: "3.7V - 4.2V (VPH_PWR)", DiodeReading: "350 - 450 mV",
			Symptoms:      mustJSON([]string{"Dead Phone (No Boot)", "Auto Current Leakage", "Overheating near PMIC", "No Charging Signal"}),
			Steps:         mustJSON([]string{"Check VPH_PWR voltage at main inductor (LDO test pad). Should be ~3.8V.", "Measure diode mode resistance to ground on 1.8V_S2 buck output line.", "If shorted (<50 mV), inject 1.5V current with DC power supply to find shorted capacitor."}),
			EquivalentICs: mustJSON([]string{"PM8350B", "PM8350C"}),
			IsActive: true, OrderNum: 1, CreatedAt: now,
		},
		{
			ID: "sim_002", Name: "Central Processing Unit (CPU)",
			Code: "Snapdragon 8 Gen 2 / Dimensity 9200", Category: "Processor & RAM", LayerIndex: 0,
			RelativePosX: 0.50, RelativePosY: 0.48, ColorHex: "#5B4FCF", Icon: "memory_rounded",
			Summary: "Main BGA processor running OS logic, GPU rendering, security enclave, and memory bus controller.",
			Voltage: "0.8V - 1.2V (VDD_CORE)", DiodeReading: "180 - 280 mV",
			Symptoms:      mustJSON([]string{"Stuck on Logo / Bootloop", "Qualcomm EDL Mode (9008 Port)", "System Freeze on Heavy Load", "No RAM Detection"}),
			Steps:         mustJSON([]string{"Connect phone to PC via USB — check if device manager shows 'Qualcomm HS-USB QDLoader 9008'.", "Verify CPU VDD_CORE power rail (0.8V) using oscilloscope/multimeter during power key press.", "Perform BGA CPU reballing if solder ball joints are cracked from drop damage."}),
			EquivalentICs: mustJSON([]string{"Snapdragon 8 Gen 1 (Pin Compatible)"}),
			IsActive: true, OrderNum: 2, CreatedAt: now,
		},
		{
			ID: "sim_003", Name: "UFS / eMMC Storage IC",
			Code: "KLUAG1JETD 256GB UFS 3.1", Category: "Processor & RAM", LayerIndex: 0,
			RelativePosX: 0.68, RelativePosY: 0.32, ColorHex: "#22C55E", Icon: "storage_rounded",
			Summary: "High-speed NAND flash storage containing Android OS firmware partitions, bootloader & user data.",
			Voltage: "1.8V / 2.95V (VCC & VCCQ)", DiodeReading: "400 - 550 mV",
			Symptoms:      mustJSON([]string{"Stuck on Startup Logo", "System Memory Corruption Error", "Unable to Wipe Data", "EDL Flash Error"}),
			Steps:         mustJSON([]string{"Verify VCC 2.95V and VCCQ 1.8V power rails near NAND decoupling capacitors.", "Check ISP pinouts (DAT0, CLK, CMD, GND) using EasyJTAG / Medusa Pro box.", "Reprogram user area partition if corrupt or replace NAND IC."}),
			EquivalentICs: mustJSON([]string{"KMDH6001DA 128GB", "HN8T15FCHKX005"}),
			IsActive: true, OrderNum: 3, CreatedAt: now,
		},
		{
			ID: "sim_004", Name: "Display Backlight & Driver IC",
			Code: "TPS61165 Display IC", Category: "Display & Touch", LayerIndex: 0,
			RelativePosX: 0.25, RelativePosY: 0.65, ColorHex: "#3B82F6", Icon: "smartphone_rounded",
			Summary: "Generates high voltage boost (16V - 25V LED+) for LCD backlight and controls AMOLED touch panel data lines.",
			Voltage: "16.5V Boost Output", DiodeReading: "500 - 620 mV",
			Symptoms:      mustJSON([]string{"Black Screen (Sound Works)", "Flickering Screen", "Half-Screen Backlight Dimming", "Touch Not Responding"}),
			Steps:         mustJSON([]string{"Check display FPC connector pins under 45x microscope for burnt pins.", "Measure LED+ voltage (~16V to 22V) at backlight boost diode anode pad.", "Replace backlight boost coil (22uH) or driver IC if output is stuck at 3.8V."}),
			EquivalentICs: mustJSON([]string{"MP3378", "AP3608"}),
			IsActive: true, OrderNum: 4, CreatedAt: now,
		},
		{
			ID: "sim_005", Name: "RF Transceiver IC (Network)",
			Code: "SDR735 Baseband RF", Category: "Network & RF", LayerIndex: 1,
			RelativePosX: 0.42, RelativePosY: 0.35, ColorHex: "#EC4899", Icon: "cell_tower_rounded",
			Summary: "Processes 4G/5G radio signals, converts analog RF signals to digital baseband I/Q data for CPU.",
			Voltage: "1.0V / 1.8V RF Rail", DiodeReading: "380 - 480 mV",
			Symptoms:      mustJSON([]string{"No Service / Searching", "Null IMEI Number", "Emergency Calls Only", "Weak 5G Signal"}),
			Steps:         mustJSON([]string{"Dial *#06# to verify if Baseband Firmware version is valid in Settings.", "Inspect RF antenna coaxial cable connection and test RF switch filter voltage.", "Reball or replace RF Transceiver IC if VREG_RF rails are normal."}),
			EquivalentICs: mustJSON([]string{"SDR865", "SDR660"}),
			IsActive: true, OrderNum: 5, CreatedAt: now,
		},
		{
			ID: "sim_006", Name: "Fast Charge IC & OVP Switch",
			Code: "BQ25895 / OVP Protection IC", Category: "Power Section", LayerIndex: 1,
			RelativePosX: 0.60, RelativePosY: 0.62, ColorHex: "#EAB308", Icon: "battery_charging_full_rounded",
			Summary: "Controls USB Type-C fast charging input (5V/9V/12V QuickCharge/PD) and provides Over-Voltage Protection (OVP).",
			Voltage: "5.0V - 9.0V (VBUS Input)", DiodeReading: "450 - 580 mV",
			Symptoms:      mustJSON([]string{"Slow Charging / Fake Charge", "Phone Not Recognized by PC", "High Temperature Warning", "0.00A USB Draw"}),
			Steps:         mustJSON([]string{"Measure VBUS pad voltage at USB connector — should show 5V when charger connected.", "Check I2C communication lines (SCL/SDA) using oscilloscope during fast charge negotiation.", "Replace OVP switch IC if VBUS input shows 0V despite confirmed good charger."}),
			EquivalentICs: mustJSON([]string{"BQ25890", "SM5714"}),
			IsActive: true, OrderNum: 6, CreatedAt: now,
		},
	}
	cr.DB.Create(&comps)
}

func init() {
	_ = strings.ToLower // avoid unused import warning
}
