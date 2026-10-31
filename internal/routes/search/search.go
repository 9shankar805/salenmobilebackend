package search

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"smg-backend/internal/models"
)

type SearchRouter struct {
	DB *gorm.DB
}

func RegisterRoutes(r *gin.Engine, db *gorm.DB) {
	sr := &SearchRouter{DB: db}

	// Global search — public, no auth required
	r.GET("/api/search", sr.GlobalSearch)
}

// GlobalSearch searches across courses, live sessions, schematics, and community posts.
// GET /api/search?q=...&type=...&limit=...
func (sr *SearchRouter) GlobalSearch(c *gin.Context) {
	q := strings.TrimSpace(c.Query("q"))
	typeFilter := strings.ToLower(strings.TrimSpace(c.Query("type"))) // courses|lessons|schematics|community
	limit := 20

	type SearchResult struct {
		ID       string `json:"id"`
		Title    string `json:"title"`
		Category string `json:"category"` // Courses, Recorded Lessons, Schematics & Guides, Community
		Subtitle string `json:"subtitle"`
		ImageURL string `json:"imageAsset"`
		Tag      string `json:"tag"`
		Slug     string `json:"slug,omitempty"`
		Price    float64 `json:"price,omitempty"`
	}

	results := []SearchResult{}

	// ─── 1. Courses ───────────────────────────────────────────────────────────
	if typeFilter == "" || typeFilter == "courses" {
		var courses []models.Course
		query := sr.DB.Where("is_published = ?", true).Limit(limit)
		if q != "" {
			like := "%" + q + "%"
			query = query.Where("title LIKE ? OR description LIKE ? OR tags LIKE ? OR category LIKE ?", like, like, like, like)
		}
		query.Find(&courses)

		for _, c := range courses {
			img := c.ImageAsset
			if img == "" {
				img = "assets/banner_bg_img/banner_1.webp"
			}
			tag := strings.ToUpper(c.Category)
			if tag == "" {
				tag = "COURSE"
			}
			results = append(results, SearchResult{
				ID:       c.ID,
				Title:    c.Title,
				Category: "Courses",
				Subtitle: c.Instructor + " • " + c.Level,
				ImageURL: img,
				Tag:      tag,
				Slug:     c.Slug,
				Price:    c.Price,
			})
		}
	}

	// ─── 2. Lessons ───────────────────────────────────────────────────────────
	if typeFilter == "" || typeFilter == "lessons" {
		var lessons []models.Lesson
		query := sr.DB.Limit(limit)
		if q != "" {
			like := "%" + q + "%"
			query = query.Where("title LIKE ? OR description LIKE ?", like, like)
		}
		query.Find(&lessons)

		for _, l := range lessons {
			results = append(results, SearchResult{
				ID:       l.ID,
				Title:    l.Title,
				Category: "Recorded Lessons",
				Subtitle: "Video Lesson • " + l.Duration,
				ImageURL: "assets/banner_bg_img/banner_2.webp",
				Tag:      "VIDEO",
				Slug:     l.Slug,
			})
		}
	}

	// ─── 3. Schematics ────────────────────────────────────────────────────────
	if typeFilter == "" || typeFilter == "schematics" {
		var schematics []models.Schematic
		query := sr.DB.Limit(limit)
		if q != "" {
			like := "%" + q + "%"
			query = query.Where("title LIKE ? OR category LIKE ?", like, like)
		}
		query.Find(&schematics)

		for _, s := range schematics {
			results = append(results, SearchResult{
				ID:       s.ID,
				Title:    s.Title,
				Category: "Schematics & Guides",
				Subtitle: "PDF Schematic • " + s.FileSize,
				ImageURL: "assets/banner_bg_img/banner_3.webp",
				Tag:      "SCHEMATIC",
			})
		}
	}

	// ─── 4. Community ─────────────────────────────────────────────────────────
	if typeFilter == "" || typeFilter == "community" {
		var posts []models.CommunityPost
		query := sr.DB.Limit(limit / 2)
		if q != "" {
			like := "%" + q + "%"
			query = query.Where("title LIKE ? OR body LIKE ?", like, like)
		}
		query.Find(&posts)

		for _, p := range posts {
			results = append(results, SearchResult{
				ID:       p.ID,
				Title:    p.Title,
				Category: "Community",
				Subtitle: p.AuthorName + " • " + p.AuthorRole,
				ImageURL: "assets/banner_bg_img/banner_4.webp",
				Tag:      "DISCUSSION",
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"query":   q,
		"total":   len(results),
		"data":    results,
	})
}
