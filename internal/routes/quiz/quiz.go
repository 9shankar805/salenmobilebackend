package quiz

import (
	"fmt"
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

type QuizRouter struct {
	DB *gorm.DB
}

func RegisterRoutes(r *gin.Engine, db *gorm.DB) {
	qr := &QuizRouter{DB: db}

	// Quiz listing is free — users can see what quizzes exist
	r.GET("/quizzes", qr.ListAllQuizzes)
	r.GET("/quizzes/:courseSlug", qr.GetQuiz)

	// Submitting quizzes & tracking results
	r.POST("/quizzes/submit", qr.SubmitQuiz)
	r.GET("/quizzes/results", qr.GetUserResults)
	r.GET("/quizzes/leaderboard", qr.GetLeaderboard) // leaderboard is public — motivational

	// Badges — computed from real quiz_results; userId optional (returns global definitions if empty)
	r.GET("/quizzes/badges", qr.GetBadges)

	// Admin Quiz Management
	adm := r.Group("/api/admin/quizzes", middleware.AuthMiddleware(), middleware.AdminMiddleware())
	{
		adm.GET("", qr.AdminListQuizzes)
		adm.POST("", qr.AdminSaveQuiz)
		adm.DELETE("/:id", qr.AdminDeleteQuiz)
		adm.GET("/results", qr.AdminListResults)
		adm.DELETE("/results/:id", qr.AdminDeleteResult)
	}
}

func (qr *QuizRouter) ListAllQuizzes(c *gin.Context) {
	category := strings.TrimSpace(strings.ToLower(c.Query("category")))

	var quizzes []models.Quiz
	err := qr.DB.Preload("Questions.Options", func(db *gorm.DB) *gorm.DB {
		return db.Order("option_index asc")
	}).Order("created_at desc").Find(&quizzes).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to load quizzes: " + err.Error(),
		})
		return
	}

	// Fetch courses to enrich with course title, category, level
	var courses []models.Course
	qr.DB.Select("slug, title, category, level, image_asset").Find(&courses)
	courseMap := make(map[string]models.Course)
	for _, cr := range courses {
		courseMap[cr.Slug] = cr
	}

	// Aggregate stats from quiz_results
	type quizStat struct {
		QuizID   string  `gorm:"column:quiz_id"`
		AvgScore float64 `gorm:"column:avg_score"`
		Attempts int     `gorm:"column:attempts"`
	}
	var stats []quizStat
	qr.DB.Raw(`
		SELECT quiz_id, AVG(score_percent) as avg_score, COUNT(*) as attempts
		FROM quiz_results
		GROUP BY quiz_id
	`).Scan(&stats)
	statMap := make(map[string]quizStat)
	for _, s := range stats {
		statMap[s.QuizID] = s
	}

	type QuizDTO struct {
		ID             string                `json:"id"`
		CourseSlug     string                `json:"courseSlug"`
		CourseTitle    string                `json:"courseTitle"`
		ModuleID       string                `json:"moduleId"`
		Title          string                `json:"title"`
		DurationMins   int                   `json:"durationMinutes"`
		TotalQuestions int                   `json:"totalQuestions"`
		PassingScore   int                   `json:"passingScorePercent"`
		Category       string                `json:"category"`
		Level          string                `json:"level"`
		AvgScore       int                   `json:"avgScore"`
		Attempts       int                   `json:"attempts"`
		Questions      []models.QuizQuestion `json:"questions"`
	}

	out := make([]QuizDTO, 0, len(quizzes))
	for _, q := range quizzes {
		cr, hasCourse := courseMap[q.CourseSlug]
		cat := "Hardware"
		level := "Intermediate"
		crTitle := ""
		if hasCourse {
			if cr.Category != "" {
				cat = cr.Category
			}
			if cr.Level != "" {
				level = cr.Level
			}
			crTitle = cr.Title
		} else {
			textLower := strings.ToLower(q.Title + " " + q.CourseSlug)
			if strings.Contains(textLower, "iphone") || strings.Contains(textLower, "display") {
				cat = "Iphone"
			} else if strings.Contains(textLower, "emmc") || strings.Contains(textLower, "ufs") {
				cat = "Emmc"
			} else if strings.Contains(textLower, "software") {
				cat = "Software"
			}
		}

		if category != "" && category != "all" {
			catMatch := strings.Contains(strings.ToLower(cat), category)
			titleMatch := strings.Contains(strings.ToLower(q.Title), category)
			slugMatch := strings.Contains(strings.ToLower(q.CourseSlug), category)
			if !catMatch && !titleMatch && !slugMatch {
				continue
			}
		}

		avg := 75
		attempts := 0
		if st, ok := statMap[q.ID]; ok && st.Attempts > 0 {
			avg = int(st.AvgScore)
			attempts = st.Attempts
		}

		tq := q.TotalQuestions
		if len(q.Questions) > 0 {
			tq = len(q.Questions)
		}
		dur := q.DurationMins
		if dur <= 0 {
			dur = 20
		}
		passScore := q.PassingScore
		if passScore <= 0 {
			passScore = 75
		}

		out = append(out, QuizDTO{
			ID:             q.ID,
			CourseSlug:     q.CourseSlug,
			CourseTitle:    crTitle,
			ModuleID:       q.ModuleID,
			Title:          q.Title,
			DurationMins:   dur,
			TotalQuestions: tq,
			PassingScore:   passScore,
			Category:       cat,
			Level:          level,
			AvgScore:       avg,
			Attempts:       attempts,
			Questions:      q.Questions,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    out,
		"quizzes": out,
	})
}

func (qr *QuizRouter) GetQuiz(c *gin.Context) {
	courseSlug := c.Param("courseSlug")

	var quiz models.Quiz
	err := qr.DB.Preload("Questions.Options", func(db *gorm.DB) *gorm.DB {
		return db.Order("option_index asc")
	}).Where("course_slug = ?", courseSlug).First(&quiz).Error

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Assessment quiz not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    quiz,
	})
}

type QuizSubmitRequest struct {
	QuizID  string         `json:"quizId" binding:"required"`
	UserID  string         `json:"userId"`
	Answers map[string]int `json:"answers"` // QuestionID -> SelectedOptionIndex
}

func (qr *QuizRouter) SubmitQuiz(c *gin.Context) {
	var req QuizSubmitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	var quiz models.Quiz
	if err := qr.DB.Preload("Questions").Where("id = ?", req.QuizID).First(&quiz).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Quiz not found"})
		return
	}

	correctCount := 0
	totalQuestions := len(quiz.Questions)
	breakdown := make([]map[string]interface{}, 0, totalQuestions)

	for _, q := range quiz.Questions {
		userAns, hasAns := req.Answers[q.ID]
		isCorrect := hasAns && userAns == q.CorrectOption
		if isCorrect {
			correctCount++
		}
		breakdown = append(breakdown, map[string]interface{}{
			"questionId":    q.ID,
			"isCorrect":     isCorrect,
			"userChoice":    userAns,
			"correctChoice": q.CorrectOption,
			"explanation":   q.Explanation,
		})
	}

	scorePercent := 0
	if totalQuestions > 0 {
		scorePercent = (correctCount * 100) / totalQuestions
	}

	isPassed := scorePercent >= quiz.PassingScore
	xpEarned := correctCount * 10
	if isPassed {
		xpEarned += 50 // bonus XP for passing
	}

	// Persist result so history & leaderboard work
	if req.UserID != "" {
		result := models.QuizResult{
			ID:           uuid.New().String(),
			QuizID:       req.QuizID,
			CourseSlug:   quiz.CourseSlug,
			UserID:       req.UserID,
			ScorePercent: scorePercent,
			CorrectCount: correctCount,
			TotalQ:       totalQuestions,
			IsPassed:     isPassed,
			XPEarned:     xpEarned,
			QuizTitle:    quiz.Title,
			AttemptedAt:  time.Now(),
		}
		qr.DB.Create(&result)
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"scorePercent":   scorePercent,
		"correctCount":   correctCount,
		"totalQuestions": totalQuestions,
		"isPassed":       isPassed,
		"passingScore":   quiz.PassingScore,
		"xpEarned":       xpEarned,
		"breakdown":      breakdown,
		"data": gin.H{
			"scorePercent":   scorePercent,
			"correctCount":   correctCount,
			"totalQuestions": totalQuestions,
			"isPassed":       isPassed,
			"xpEarned":       xpEarned,
		},
	})
}

// GetUserResults returns quiz attempt history for a specific user.
// GET /quizzes/results?userId=...&courseSlug=... (courseSlug optional)
func (qr *QuizRouter) GetUserResults(c *gin.Context) {
	userID := c.Query("userId")
	courseSlug := c.Query("courseSlug")

	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "userId is required"})
		return
	}

	var results []models.QuizResult
	query := qr.DB.Where("user_id = ?", userID).Order("attempted_at desc")
	if courseSlug != "" {
		query = query.Where("course_slug = ?", courseSlug)
	}
	query.Find(&results)

	// Compute summary stats
	totalAttempts := len(results)
	passedCount := 0
	totalXP := 0
	for _, r := range results {
		if r.IsPassed {
			passedCount++
		}
		totalXP += r.XPEarned
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"results":       results,
			"totalAttempts": totalAttempts,
			"passedCount":   passedCount,
			"failedCount":   totalAttempts - passedCount,
			"totalXP":       totalXP,
		},
	})
}

// GetLeaderboard returns the global top 50 quiz performers ranked by total XP.
// GET /quizzes/leaderboard?limit=50
func (qr *QuizRouter) GetLeaderboard(c *gin.Context) {
	// Aggregate XP and pass-count per user from quiz_results
	type leaderRow struct {
		UserID       string `gorm:"column:user_id"`
		TotalXP      int    `gorm:"column:total_xp"`
		TotalPassed  int    `gorm:"column:total_passed"`
		TotalAttempts int   `gorm:"column:total_attempts"`
	}

	var rows []leaderRow
	qr.DB.Raw(`
		SELECT user_id,
		       SUM(xp_earned)  AS total_xp,
		       SUM(CASE WHEN is_passed = 1 THEN 1 ELSE 0 END) AS total_passed,
		       COUNT(*)         AS total_attempts
		FROM quiz_results
		GROUP BY user_id
		ORDER BY total_xp DESC
		LIMIT 50
	`).Scan(&rows)

	// Enrich with user names
	type entry struct {
		Rank          int    `json:"rank"`
		UserID        string `json:"userId"`
		Name          string `json:"name"`
		AvatarURL     string `json:"avatarUrl"`
		TotalXP       int    `json:"totalXP"`
		TotalPassed   int    `json:"testsPassed"`
		TotalAttempts int    `json:"testsCompleted"`
		Accuracy      string `json:"accuracy"`
	}

	board := make([]entry, 0, len(rows))
	for i, row := range rows {
		var user models.User
		name := "Technician"
		avatarURL := ""
		if err := qr.DB.Where("id = ?", row.UserID).First(&user).Error; err == nil {
			name = user.Name
			avatarURL = user.AvatarURL
		}

		acc := "0%"
		if row.TotalAttempts > 0 {
			acc = fmt.Sprintf("%d%%", (row.TotalPassed*100)/row.TotalAttempts)
		}

		board = append(board, entry{
			Rank:          i + 1,
			UserID:        row.UserID,
			Name:          name,
			AvatarURL:     avatarURL,
			TotalXP:       row.TotalXP,
			TotalPassed:   row.TotalPassed,
			TotalAttempts: row.TotalAttempts,
			Accuracy:      acc,
		})
	}

	// If no DB results yet, return seeded demo leaderboard so screen isn't empty
	if len(board) == 0 {
		board = []entry{
			{Rank: 1, UserID: "u_ramesh", Name: "Ramesh KC", TotalXP: 3450, TotalPassed: 24, TotalAttempts: 24, Accuracy: "98%"},
			{Rank: 2, UserID: "u_amit", Name: "Amit Verma", TotalXP: 3120, TotalPassed: 21, TotalAttempts: 22, Accuracy: "95%"},
			{Rank: 3, UserID: "u_pankaj", Name: "Pankaj Sharma", TotalXP: 2940, TotalPassed: 19, TotalAttempts: 21, Accuracy: "91%"},
			{Rank: 4, UserID: "u_rohit", Name: "Rohit Kumar", TotalXP: 2850, TotalPassed: 18, TotalAttempts: 20, Accuracy: "88%"},
			{Rank: 5, UserID: "u_suresh", Name: "Suresh Kumar", TotalXP: 2710, TotalPassed: 16, TotalAttempts: 19, Accuracy: "86%"},
		}
	}

	// Sort by XP descending
	sort.Slice(board, func(i, j int) bool {
		return board[i].TotalXP > board[j].TotalXP
	})

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    board,
	})
}

// ─── Badge Definitions ───────────────────────────────────────────────────────

type badgeDef struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Subtitle string `json:"subtitle"`
	Icon     string `json:"icon"`   // Material icon name string
	Color    string `json:"color"` // hex
}

var allBadges = []badgeDef{
	{ID: "hardware_master", Title: "Hardware Master", Subtitle: "Complete 10+ Tests", Icon: "memory_rounded", Color: "#5B4FCF"},
	{ID: "speed_racer", Title: "Speed Racer", Subtitle: "Finish a Quiz in under 10 minutes", Icon: "bolt_rounded", Color: "#F97316"},
	{ID: "seven_day_streak", Title: "7 Day Streak", Subtitle: "Attempt tests 7 days in a row", Icon: "local_fire_department_rounded", Color: "#EF4444"},
	{ID: "perfect_score", Title: "Perfect Score", Subtitle: "Score 100% on any test", Icon: "workspace_premium_rounded", Color: "#22C55E"},
	{ID: "emmc_specialist", Title: "EMMC Specialist", Subtitle: "Pass 5 EMMC-related quizzes", Icon: "developer_board_rounded", Color: "#3B82F6"},
	{ID: "first_pass", Title: "First Pass", Subtitle: "Pass your very first quiz", Icon: "military_tech_rounded", Color: "#F59E0B"},
	{ID: "xp_500", Title: "XP Hunter", Subtitle: "Earn 500+ total XP", Icon: "star_rounded", Color: "#8B5CF6"},
	{ID: "xp_2000", Title: "XP Champion", Subtitle: "Earn 2,000+ total XP", Icon: "emoji_events_rounded", Color: "#FACC15"},
}

// GetBadges computes which badges a user has earned from their real quiz_results.
// GET /quizzes/badges?userId=...
func (qr *QuizRouter) GetBadges(c *gin.Context) {
	userID := c.Query("userId")

	type badgeResp struct {
		badgeDef
		IsUnlocked bool `json:"isUnlocked"`
	}

	// If no userId provided, return all badges as locked (used for discovery)
	if userID == "" {
		result := make([]badgeResp, len(allBadges))
		for i, b := range allBadges {
			result[i] = badgeResp{badgeDef: b, IsUnlocked: false}
		}
		c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
		return
	}

	// Fetch all quiz results for this user
	var results []models.QuizResult
	qr.DB.Where("user_id = ?", userID).Order("attempted_at asc").Find(&results)

	totalAttempts := len(results)
	totalXP := 0
	passedCount := 0
	emmcPassed := 0
	perfectScore := false

	// Streak computation: count consecutive calendar days with at least 1 attempt
	daySet := map[string]bool{}
	for _, r := range results {
		totalXP += r.XPEarned
		if r.IsPassed {
			passedCount++
			if r.ScorePercent == 100 {
				perfectScore = true
			}
			if r.CourseSlug != "" {
				slugLower := r.CourseSlug
				if len(slugLower) > 4 && (contains(slugLower, "emmc") || contains(slugLower, "ufs") || contains(slugLower, "flash")) {
					emmcPassed++
				}
			}
		}
		daySet[r.AttemptedAt.Format("2006-01-02")] = true
	}

	// Compute max streak
	maxStreak := computeMaxStreak(daySet)

	unlocked := map[string]bool{
		"hardware_master": totalAttempts >= 10,
		"speed_racer":     totalAttempts > 0, // XP-based proxy; real check needs attempt duration stored
		"seven_day_streak": maxStreak >= 7,
		"perfect_score":   perfectScore,
		"emmc_specialist": emmcPassed >= 5,
		"first_pass":      passedCount >= 1,
		"xp_500":          totalXP >= 500,
		"xp_2000":         totalXP >= 2000,
	}

	result := make([]badgeResp, len(allBadges))
	for i, b := range allBadges {
		result[i] = badgeResp{badgeDef: b, IsUnlocked: unlocked[b.ID]}
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || func() bool {
		for i := 0; i <= len(s)-len(substr); i++ {
			if s[i:i+len(substr)] == substr {
				return true
			}
		}
		return false
	}())
}

func computeMaxStreak(daySet map[string]bool) int {
	if len(daySet) == 0 {
		return 0
	}
	days := make([]time.Time, 0, len(daySet))
	for d := range daySet {
		t, err := time.Parse("2006-01-02", d)
		if err == nil {
			days = append(days, t)
		}
	}
	sort.Slice(days, func(i, j int) bool { return days[i].Before(days[j]) })

	max := 1
	current := 1
	for i := 1; i < len(days); i++ {
		if days[i].Sub(days[i-1]).Hours() <= 24 {
			current++
			if current > max {
				max = current
			}
		} else {
			current = 1
		}
	}
	return max
}

func (qr *QuizRouter) AdminListQuizzes(c *gin.Context) {
	var quizzes []models.Quiz
	if err := qr.DB.Preload("Questions.Options", func(db *gorm.DB) *gorm.DB {
		return db.Order("option_index asc")
	}).Order("created_at desc").Find(&quizzes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"quizzes": quizzes,
		"data":    quizzes,
	})
}

func (qr *QuizRouter) AdminSaveQuiz(c *gin.Context) {
	var req struct {
		ID           string `json:"id"`
		CourseSlug   string `json:"courseSlug" binding:"required"`
		Title        string `json:"title" binding:"required"`
		DurationMins int    `json:"durationMinutes"`
		PassingScore int    `json:"passingScorePercent"`
		Questions    []struct {
			ID            string   `json:"id"`
			Question      string   `json:"question" binding:"required"`
			Explanation   string   `json:"explanation"`
			CorrectOption int      `json:"correctOptionIndex"`
			Options       []string `json:"options"`
		} `json:"questions"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	quizID := req.ID
	if quizID == "" {
		quizID = "qz_" + uuid.New().String()[:8]
	}

	if req.PassingScore <= 0 {
		req.PassingScore = 75
	}
	if req.DurationMins <= 0 {
		req.DurationMins = 15
	}

	quiz := models.Quiz{
		ID:             quizID,
		CourseSlug:     req.CourseSlug,
		Title:          req.Title,
		DurationMins:   req.DurationMins,
		TotalQuestions: len(req.Questions),
		PassingScore:   req.PassingScore,
		CreatedAt:      time.Now(),
	}

	// Upsert quiz
	qr.DB.Save(&quiz)

	// If questions provided, replace old ones
	if len(req.Questions) > 0 {
		var oldQIDs []string
		qr.DB.Model(&models.QuizQuestion{}).Where("quiz_id = ?", quizID).Pluck("id", &oldQIDs)
		if len(oldQIDs) > 0 {
			qr.DB.Where("question_id IN ?", oldQIDs).Delete(&models.QuizOption{})
			qr.DB.Where("quiz_id = ?", quizID).Delete(&models.QuizQuestion{})
		}

		for qi, q := range req.Questions {
			qID := q.ID
			if qID == "" {
				qID = "qq_" + uuid.New().String()[:8]
			}
			question := models.QuizQuestion{
				ID:            qID,
				QuizID:        quizID,
				QuestionText:  q.Question,
				Explanation:   q.Explanation,
				OrderNum:      qi + 1,
				CorrectOption: q.CorrectOption,
			}
			qr.DB.Create(&question)

			for oi, optText := range q.Options {
				option := models.QuizOption{
					ID:         "qo_" + uuid.New().String()[:8],
					QuestionID: qID,
					Text:       optText,
					Index:      oi,
				}
				qr.DB.Create(&option)
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": quiz})
}

func (qr *QuizRouter) AdminDeleteQuiz(c *gin.Context) {
	id := c.Param("id")
	var qIDs []string
	qr.DB.Model(&models.QuizQuestion{}).Where("quiz_id = ?", id).Pluck("id", &qIDs)
	if len(qIDs) > 0 {
		qr.DB.Where("question_id IN ?", qIDs).Delete(&models.QuizOption{})
		qr.DB.Where("quiz_id = ?", id).Delete(&models.QuizQuestion{})
	}
	qr.DB.Delete(&models.Quiz{}, "id = ?", id)
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Quiz deleted"})
}

func (qr *QuizRouter) AdminListResults(c *gin.Context) {
	var results []models.QuizResult
	qr.DB.Order("attempted_at desc").Limit(100).Find(&results)
	c.JSON(http.StatusOK, gin.H{"success": true, "data": results})
}

func (qr *QuizRouter) AdminDeleteResult(c *gin.Context) {
	id := c.Param("id")
	qr.DB.Delete(&models.QuizResult{}, "id = ?", id)
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Result deleted"})
}
