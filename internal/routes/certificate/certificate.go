package certificate

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"smg-backend/internal/middleware"
	"smg-backend/internal/models"
)

type CertificateRouter struct {
	DB *gorm.DB
}

func RegisterRoutes(r *gin.Engine, db *gorm.DB) {
	cr := &CertificateRouter{DB: db}
	cr.seedSampleCertificates()

	// Public and User APIs
	r.GET("/user/certificates", cr.GetUserCertificates)
	r.GET("/certificates/verify/:certNum", cr.VerifyCertificate)

	// Admin APIs
	adm := r.Group("/api/admin/certificates", middleware.AuthMiddleware(), middleware.AdminMiddleware())
	{
		adm.GET("", cr.GetAllCertificates)
		adm.POST("/issue", cr.IssueCertificate)
	}
}

func (cr *CertificateRouter) GetUserCertificates(c *gin.Context) {
	userID := c.Query("userId")
	if userID == "" {
		userID = "guest-tech"
	}

	var list []models.Certificate
	if err := cr.DB.Where("user_id = ?", userID).Order("issued_at desc").Find(&list).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	if len(list) == 0 {
		// Provide default sample verified certificate for smooth mobile app demo
		cr.DB.Where("user_id = ?", "demo-technician").Find(&list)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    list,
	})
}

func (cr *CertificateRouter) VerifyCertificate(c *gin.Context) {
	certNum := c.Param("certNum")

	var cert models.Certificate
	err := cr.DB.Where("certificate_number = ?", certNum).First(&cert).Error
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Certificate not found or invalid",
			"valid":   false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"valid":   true,
		"data": gin.H{
			"certificateNumber": cert.CertificateNumber,
			"studentName":       cert.StudentName,
			"courseTitle":       cert.CourseTitle,
			"courseSlug":        cert.CourseSlug,
			"grade":             cert.Grade,
			"issuedAt":          cert.IssuedAt.Format("2006-01-02"),
			"pdfUrl":            cert.PdfURL,
			"issuer":            "Salen Mobile Guru Academy",
			"signature":         "Salen Dangol (Master Hardware Engineer)",
		},
	})
}

type IssueCertRequest struct {
	UserID      string `json:"userId" binding:"required"`
	StudentName string `json:"studentName" binding:"required"`
	CourseID    string `json:"courseId"`
	CourseTitle string `json:"courseTitle" binding:"required"`
	CourseSlug  string `json:"courseSlug" binding:"required"`
	Grade       string `json:"grade"`
	Instructor  string `json:"instructor"`
}

func (cr *CertificateRouter) IssueCertificate(c *gin.Context) {
	var req IssueCertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	if req.Grade == "" {
		req.Grade = "A+ (Master Technician)"
	}
	if req.Instructor == "" {
		req.Instructor = "Salen Dangol"
	}

	now := time.Now()
	certNum := fmt.Sprintf("SMG-%d-%s", now.Year(), uuid.New().String()[:6])

	cert := models.Certificate{
		ID:                uuid.New().String(),
		UserID:            req.UserID,
		StudentName:       req.StudentName,
		CourseID:          req.CourseID,
		CourseTitle:       req.CourseTitle,
		CourseSlug:        req.CourseSlug,
		CertificateNumber: certNum,
		Instructor:        req.Instructor,
		IssuedAt:          now,
		IssueDate:         now, // populate both fields so MySQL never gets a zero date
		Grade:             req.Grade,
		PdfURL:            fmt.Sprintf("https://salenmobileguru.com/api/certificates/%s.pdf", certNum),
		VerificationURL:   fmt.Sprintf("https://salenmobileguru.com/certificates/verify/%s", certNum),
	}

	if err := cr.DB.Create(&cert).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    cert,
	})
}

func (cr *CertificateRouter) GetAllCertificates(c *gin.Context) {
	var list []models.Certificate
	if err := cr.DB.Order("issued_at desc").Find(&list).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": list})
}

func (cr *CertificateRouter) seedSampleCertificates() {
	var count int64
	cr.DB.Model(&models.Certificate{}).Count(&count)
	if count > 0 {
		return
	}

	samples := []models.Certificate{
		{
			ID:                uuid.New().String(),
			UserID:            "demo-technician",
			StudentName:       "Sunil Shrestha",
			CourseTitle:       "Android Hardware Repairing Level 1",
			CourseSlug:        "android-hardware-repairing-level-1",
			CertificateNumber: "SMG-2026-A10982",
			Instructor:        "Salen Dangol",
			IssuedAt:          time.Now().Add(-30 * 24 * time.Hour),
			IssueDate:         time.Now().Add(-30 * 24 * time.Hour),
			Grade:             "A+ (Mastery in PMIC & Micro-Soldering)",
			PdfURL:            "https://salenmobileguru.com/certificates/SMG-2026-A10982.pdf",
			VerificationURL:   "https://salenmobileguru.com/certificates/verify/SMG-2026-A10982",
		},
		{
			ID:                uuid.New().String(),
			UserID:            "guest-tech",
			StudentName:       "Technician Learner",
			CourseTitle:       "Borneo & ZXW Schematic Reading",
			CourseSlug:        "borneo-zxw-schematic-reading",
			CertificateNumber: "SMG-2026-B88201",
			Instructor:        "Rohit Sharma",
			IssuedAt:          time.Now().Add(-10 * 24 * time.Hour),
			IssueDate:         time.Now().Add(-10 * 24 * time.Hour),
			Grade:             "A (Circuit Tracing Specialist)",
			PdfURL:            "https://salenmobileguru.com/certificates/SMG-2026-B88201.pdf",
			VerificationURL:   "https://salenmobileguru.com/certificates/verify/SMG-2026-B88201",
		},
	}

	for _, s := range samples {
		cr.DB.Create(&s)
	}
}
