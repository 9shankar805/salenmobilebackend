package community

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"smg-backend/internal/models"
)

type CommunityRouter struct {
	DB *gorm.DB
}

func RegisterRoutes(r *gin.Engine, db *gorm.DB) {
	cr := &CommunityRouter{DB: db}

	// Reading community posts is open — discovery & peer knowledgebase
	r.GET("/community", cr.GetQuestions)
	r.GET("/org-site/community/:orgId", cr.GetQuestions)

	// Creating posts, upvoting and replying in community forum & Q&A
	r.POST("/community", cr.CreatePost)
	r.POST("/community/:id/upvote", cr.UpvotePost)
	r.POST("/community/:id/reply", cr.AddReply)
}

func (cr *CommunityRouter) GetQuestions(c *gin.Context) {
	orgId := c.Param("orgId")
	if orgId == "" {
		orgId = c.DefaultQuery("orgId", "smg")
	}

	var posts []models.CommunityPost
	err := cr.DB.Preload("Replies").Where("organization_id = ?", orgId).Order("created_at desc").Find(&posts).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to fetch community posts",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    posts,
	})
}

type CreatePostRequest struct {
	Title          string `json:"title" binding:"required"`
	Body           string `json:"body" binding:"required"`
	OrganizationID string `json:"organizationId"`
	CourseID       string `json:"courseId"`
	AuthorName     string `json:"authorName"`
	AuthorRole     string `json:"authorRole"`
}

func (cr *CommunityRouter) CreatePost(c *gin.Context) {
	var req CreatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	if req.OrganizationID == "" {
		req.OrganizationID = "smg"
	}
	if req.AuthorName == "" {
		if headerName := c.GetHeader("X-User-Name"); headerName != "" {
			req.AuthorName = headerName
		} else {
			req.AuthorName = "Student Technician"
		}
	}
	if req.AuthorRole == "" {
		req.AuthorRole = "Mobile Technician"
	}

	post := models.CommunityPost{
		ID:             uuid.New().String(),
		OrganizationID: req.OrganizationID,
		CourseID:       req.CourseID,
		Title:          req.Title,
		Body:           req.Body,
		AuthorName:     req.AuthorName,
		AuthorRole:     req.AuthorRole,
		AuthorAvatar:   "assets/banner_bg_img/banner_1.webp",
		Upvotes:        1,
		RepliesCount:   0,
		IsSolved:       false,
		CreatedAt:      time.Now(),
	}

	if err := cr.DB.Create(&post).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to create post"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    post,
	})
}

func (cr *CommunityRouter) UpvotePost(c *gin.Context) {
	id := c.Param("id")
	cr.DB.Model(&models.CommunityPost{}).Where("id = ?", id).UpdateColumn("upvotes", gorm.Expr("upvotes + 1"))
	c.JSON(http.StatusOK, gin.H{"success": true})
}

type AddReplyRequest struct {
	AuthorName string `json:"authorName"`
	AuthorRole string `json:"authorRole"`
	Body       string `json:"body" binding:"required"`
}

func (cr *CommunityRouter) AddReply(c *gin.Context) {
	postID := c.Param("id")
	var req AddReplyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	if req.AuthorName == "" {
		if headerName := c.GetHeader("X-User-Name"); headerName != "" {
			req.AuthorName = headerName
		} else {
			req.AuthorName = "Master Technician"
		}
	}
	if req.AuthorRole == "" {
		req.AuthorRole = "Certified Hardware Instructor"
	}

	reply := models.CommunityReply{
		ID:           uuid.New().String(),
		PostID:       postID,
		AuthorName:   req.AuthorName,
		AuthorRole:   req.AuthorRole,
		AuthorAvatar: "assets/banner_bg_img/banner_3.webp",
		Body:         req.Body,
		Upvotes:      0,
		CreatedAt:    time.Now(),
	}

	if err := cr.DB.Create(&reply).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to add reply"})
		return
	}

	cr.DB.Model(&models.CommunityPost{}).Where("id = ?", postID).UpdateColumn("replies_count", gorm.Expr("replies_count + 1"))
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": reply})
}
