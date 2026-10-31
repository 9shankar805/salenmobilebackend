package media

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"smg-backend/internal/middleware"
	"smg-backend/internal/models"
	"smg-backend/internal/storage"
)

type MediaRouter struct {
	DB *gorm.DB
}

func RegisterRoutes(r *gin.Engine, db *gorm.DB) {
	mr := &MediaRouter{DB: db}

	r.GET("/media", mr.ListMedia)
	r.POST("/media", mr.CreateMedia)
	r.DELETE("/media/:id", mr.DeleteMedia)

	r.POST("/api/upload", mr.UploadFile)
	r.POST("/media/upload", mr.UploadFile)

	r.GET("/storage/*key", mr.StreamR2Object)
	r.HEAD("/storage/*key", mr.StreamR2Object)
	r.GET("/r2/*key", mr.StreamR2Object)
	r.HEAD("/r2/*key", mr.StreamR2Object)

	adm := r.Group("/api/admin/media", middleware.AuthMiddleware(), middleware.AdminMiddleware())
	{
		adm.GET("", mr.ListMedia)
		adm.POST("", mr.CreateMedia)
		adm.POST("/upload", mr.UploadFile)
		adm.DELETE("/:id", mr.DeleteMedia)
	}

	admRoot := r.Group("/api/admin", middleware.AuthMiddleware(), middleware.AdminMiddleware())
	{
		admRoot.POST("/upload", mr.UploadFile)
	}
}

func (mr *MediaRouter) ListMedia(c *gin.Context) {
var mediaList []models.MediaAsset
mr.DB.Order("created_at desc").Find(&mediaList)
c.JSON(http.StatusOK, gin.H{"success": true, "media": mediaList, "data": mediaList})
}

type CreateMediaRequest struct {
Name     string `json:"name" binding:"required"`
Kind     string `json:"kind"`
URL      string `json:"url" binding:"required"`
Size     string `json:"size"`
Category string `json:"category"`
}

func (mr *MediaRouter) CreateMedia(c *gin.Context) {
var req CreateMediaRequest
if err := c.ShouldBindJSON(&req); err != nil {
c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
return
}

asset := models.MediaAsset{
ID:            uuid.New().String(),
Name:          req.Name,
Kind:          req.Kind,
URL:           req.URL,
SizeFormatted: req.Size,
Category:      req.Category,
CreatedAt:     time.Now(),
}
if asset.Kind == "" {
asset.Kind = "document"
}
if asset.SizeFormatted == "" {
asset.SizeFormatted = "8.4 MB"
}

if err := mr.DB.Create(&asset).Error; err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
return
}
c.JSON(http.StatusCreated, gin.H{"success": true, "media": asset, "data": asset})
}

func (mr *MediaRouter) UploadFile(c *gin.Context) {
file, err := c.FormFile("file")
if err != nil {
c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "No file uploaded: " + err.Error()})
return
}

folder := c.DefaultPostForm("folder", "uploads")
folder = strings.Trim(folder, "/")

src, err := file.Open()
if err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to open upload stream: " + err.Error()})
return
}
defer src.Close()

ext := filepath.Ext(file.Filename)
base := strings.TrimSuffix(filepath.Base(file.Filename), ext)
base = strings.ReplaceAll(base, " ", "_")
uniqueFilename := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), base, ext)

cType := file.Header.Get("Content-Type")
if cType == "" {
cType = "application/octet-stream"
}

publicURL, r2Key, err := storage.DefaultStorage.UploadDirect(folder, uniqueFilename, src, file.Size, cType)
if err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Cloudflare R2 upload error: " + err.Error()})
return
}

kind := "document"
lowerExt := strings.ToLower(ext)
if lowerExt == ".jpg" || lowerExt == ".jpeg" || lowerExt == ".png" || lowerExt == ".webp" || lowerExt == ".gif" {
kind = "image"
} else if lowerExt == ".mp4" || lowerExt == ".mkv" || lowerExt == ".webm" {
kind = "video"
} else if lowerExt == ".pdf" {
kind = "schematic"
}

sizeMB := float64(file.Size) / (1024 * 1024)
sizeStr := fmt.Sprintf("%.1f MB", sizeMB)
if sizeMB < 1.0 {
sizeStr = fmt.Sprintf("%.1f KB", float64(file.Size)/1024)
}

category := c.DefaultPostForm("category", folder)
asset := models.MediaAsset{
ID:            uuid.New().String(),
Name:          file.Filename,
Kind:          kind,
URL:           publicURL,
SizeFormatted: sizeStr,
Category:      category,
CreatedAt:     time.Now(),
}
mr.DB.Create(&asset)

localProxyURL := fmt.Sprintf("/storage/%s", r2Key)

c.JSON(http.StatusOK, gin.H{
"success":       true,
"publicUrl":     publicURL,
"url":           publicURL,
"localProxyUrl": localProxyURL,
"r2Key":         r2Key,
"bucket":        storage.DefaultStorage.R2Bucket,
"folder":        folder,
"filename":      uniqueFilename,
"originalName":  file.Filename,
"size":          file.Size,
"media":         asset,
"data":          asset,
})
}

func (mr *MediaRouter) StreamR2Object(c *gin.Context) {
	key := c.Param("key")
	key = strings.TrimPrefix(key, "/")

	// 1. Check local storage folder first (e.g. PDFs, images, schematics, templates)
	candidates := []string{
		filepath.Join("storage", key),
		filepath.Join("/var/www/salenmobilebackend/storage", key),
	}
	if sp := os.Getenv("STORAGE_PATH"); sp != "" {
		candidates = append([]string{filepath.Join(sp, key)}, candidates...)
	}
	for _, localPath := range candidates {
		if fi, err := os.Stat(localPath); err == nil && !fi.IsDir() {
			ext := strings.ToLower(filepath.Ext(key))
			if ext == ".pdf" {
				c.Header("Content-Type", "application/pdf")
				c.Header("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", filepath.Base(key)))
			} else if ext == ".png" {
				c.Header("Content-Type", "image/png")
			} else if ext == ".jpg" || ext == ".jpeg" {
				c.Header("Content-Type", "image/jpeg")
			} else if ext == ".webp" {
				c.Header("Content-Type", "image/webp")
			} else if ext == ".svg" {
				c.Header("Content-Type", "image/svg+xml")
			} else if ext == ".json" {
				c.Header("Content-Type", "application/json")
			} else if ext == ".csv" {
				c.Header("Content-Type", "text/csv")
			}
			c.File(localPath)
			return
		}
	}

	// 2. Stream from Cloudflare R2
	if storage.DefaultStorage != nil && storage.DefaultStorage.R2Enabled {
		stream, info, err := storage.DefaultStorage.GetObjectStream(key)
		if err == nil {
			defer stream.Close()
			if info.ContentType != "" {
				c.Writer.Header().Set("Content-Type", info.ContentType)
			}
			c.Writer.Header().Set("Content-Length", fmt.Sprintf("%d", info.Size))
			c.Writer.Header().Set("ETag", info.ETag)
			c.Writer.Header().Set("Cache-Control", "public, max-age=2592000")
			c.Status(http.StatusOK)
			_, _ = io.Copy(c.Writer, stream)
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "Object not found: " + key})
}

func (mr *MediaRouter) DeleteMedia(c *gin.Context) {
id := c.Param("id")
var asset models.MediaAsset
if err := mr.DB.Where("id = ?", id).First(&asset).Error; err == nil {
if strings.Contains(asset.URL, storage.DefaultStorage.R2PublicURL) {
r2Key := strings.TrimPrefix(asset.URL, storage.DefaultStorage.R2PublicURL+"/")
_ = storage.DefaultStorage.DeleteObject(r2Key)
}
mr.DB.Delete(&asset)
}
c.JSON(http.StatusOK, gin.H{"success": true, "message": "Media asset deleted"})
}
