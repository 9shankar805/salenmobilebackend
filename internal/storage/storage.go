package storage

import (
"context"
"fmt"
"io"
"log"
"net/url"
"os"
"strings"
"time"

"github.com/minio/minio-go/v7"
"github.com/minio/minio-go/v7/pkg/credentials"
)

type StorageService struct {
R2Client    *minio.Client
R2Bucket    string
R2PublicURL string
R2Enabled   bool
}

var DefaultStorage *StorageService

func InitStorage() *StorageService {
endpoint := os.Getenv("R2_ENDPOINT")
if endpoint == "" {
endpoint = "c9c7fdbce6bcc769903cbe670932b2cb.r2.cloudflarestorage.com"
}
endpoint = strings.TrimPrefix(endpoint, "https://")
endpoint = strings.TrimPrefix(endpoint, "http://")
endpoint = strings.TrimSuffix(endpoint, "/")

accessKey := os.Getenv("R2_ACCESS_KEY_ID")
if accessKey == "" {
accessKey = "1e3b3a46cbd4a917cf2f23dd12ee88a4"
}

secretKey := os.Getenv("R2_SECRET_ACCESS_KEY")
if secretKey == "" {
secretKey = "97dbf6fed577279cf84b9a039867e474ad9cc253154399641cd49c314a5eb9c2"
}

bucket := os.Getenv("R2_BUCKET")
if bucket == "" {
bucket = "salenmobileguru"
}

publicURL := os.Getenv("R2_PUBLIC_URL")
if publicURL == "" {
publicURL = "https://pub-de6175c42cb241aab9f5c23b24292596.r2.dev"
}
publicURL = strings.TrimSuffix(publicURL, "/")

s := &StorageService{
R2Bucket:    bucket,
R2PublicURL: publicURL,
}

if accessKey != "" && secretKey != "" {
client, err := minio.New(endpoint, &minio.Options{
Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
Secure: true,
})
if err == nil {
s.R2Client = client
s.R2Enabled = true
log.Printf("☁️ Pure Cloudflare R2 Storage initialized (bucket: '%s', public CDN: '%s')", bucket, publicURL)
} else {
log.Printf("⚠️ Cloudflare R2 initialization warning: %v", err)
}
}

DefaultStorage = s
return s
}

// UploadDirect streams the file directly to Cloudflare R2 without touching local server disk
func (s *StorageService) UploadDirect(folder string, filename string, r io.Reader, size int64, contentType string) (publicURL string, r2Key string, err error) {
if !s.R2Enabled || s.R2Client == nil {
return "", "", fmt.Errorf("Cloudflare R2 is not configured")
}

if folder == "" {
folder = "uploads"
}
folder = strings.Trim(folder, "/")
r2Key = fmt.Sprintf("%s/%s", folder, filename)

if contentType == "" {
contentType = "application/octet-stream"
}

_, err = s.R2Client.PutObject(context.Background(), s.R2Bucket, r2Key, r, size, minio.PutObjectOptions{
ContentType: contentType,
})
if err != nil {
return "", "", fmt.Errorf("R2 upload error: %w", err)
}

publicURL = fmt.Sprintf("%s/%s", s.R2PublicURL, r2Key)
log.Printf("☁️ Directly stored to Cloudflare R2: %s", publicURL)
return publicURL, r2Key, nil
}

func (s *StorageService) GetObjectStream(objectKey string) (io.ReadCloser, minio.ObjectInfo, error) {
if !s.R2Enabled || s.R2Client == nil {
return nil, minio.ObjectInfo{}, fmt.Errorf("R2 is not configured")
}
obj, err := s.R2Client.GetObject(context.Background(), s.R2Bucket, objectKey, minio.GetObjectOptions{})
if err != nil {
return nil, minio.ObjectInfo{}, err
}
info, err := obj.Stat()
if err != nil {
_ = obj.Close()
return nil, minio.ObjectInfo{}, err
}
return obj, info, nil
}

func (s *StorageService) GetPresignedURL(objectKey string, expiry time.Duration) (string, error) {
if !s.R2Enabled || s.R2Client == nil {
return "", fmt.Errorf("R2 not configured")
}
reqParams := make(url.Values)
u, err := s.R2Client.PresignedGetObject(context.Background(), s.R2Bucket, objectKey, expiry, reqParams)
if err != nil {
return "", err
}
return u.String(), nil
}

func (s *StorageService) DeleteObject(objectKey string) error {
if !s.R2Enabled || s.R2Client == nil {
return fmt.Errorf("R2 not configured")
}
return s.R2Client.RemoveObject(context.Background(), s.R2Bucket, objectKey, minio.RemoveObjectOptions{})
}
