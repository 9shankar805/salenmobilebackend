package admin

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	RegisterRoutes(r, nil)
	return r
}

func TestCourseRoutesAndSubmodules(t *testing.T) {
	r := setupTestRouter()

	// 1. Test Courses Catalog
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/admin/courses", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200 for /admin/courses, got %d", w.Code)
	}
	if !strings.Contains(w.Header().Get("Content-Type"), "text/html") {
		t.Errorf("Expected text/html for /admin/courses, got %s", w.Header().Get("Content-Type"))
	}

	// 2. Test Course Workspace Shell
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/admin/courses/c_test_123", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200 for /admin/courses/c_test_123, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Course Workspace") {
		t.Errorf("Expected Course Workspace in body for /admin/courses/c_test_123")
	}

	// 3. Test All 12 Submodules: index.html, style.css, script.js
	modules := []string{
		"attendance",
		"marks",
		"lessons",
		"exercises",
		"submissions",
		"people",
		"certificates",
		"compliance",
		"landingpage",
		"ai-tutor",
		"settings",
		"analytics",
	}

	for _, mod := range modules {
		// Test index.html
		htmlURL := "/internal/routes/admin/courses/" + mod + "/index.html"
		w = httptest.NewRecorder()
		req, _ = http.NewRequest("GET", htmlURL, nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("Failed GET %s: status %d", htmlURL, w.Code)
		}
		if !strings.Contains(w.Header().Get("Content-Type"), "text/html") {
			t.Errorf("Expected text/html for %s, got %s", htmlURL, w.Header().Get("Content-Type"))
		}
		body := w.Body.String()
		if !strings.Contains(body, "style.css") {
			t.Errorf("Expected %s to reference style.css", htmlURL)
		}
		if !strings.Contains(body, "script.js") {
			t.Errorf("Expected %s to reference script.js", htmlURL)
		}

		// Test style.css
		cssURL := "/internal/routes/admin/courses/" + mod + "/style.css"
		w = httptest.NewRecorder()
		req, _ = http.NewRequest("GET", cssURL, nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("Failed GET %s: status %d", cssURL, w.Code)
		}
		if !strings.Contains(w.Header().Get("Content-Type"), "text/css") {
			t.Errorf("Expected text/css for %s, got %s", cssURL, w.Header().Get("Content-Type"))
		}
		if len(w.Body.Bytes()) == 0 {
			t.Errorf("Empty body returned for %s", cssURL)
		}

		// Test script.js
		jsURL := "/internal/routes/admin/courses/" + mod + "/script.js"
		w = httptest.NewRecorder()
		req, _ = http.NewRequest("GET", jsURL, nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("Failed GET %s: status %d", jsURL, w.Code)
		}
		if !strings.Contains(w.Header().Get("Content-Type"), "application/javascript") {
			t.Errorf("Expected application/javascript for %s, got %s", jsURL, w.Header().Get("Content-Type"))
		}
		if len(w.Body.Bytes()) == 0 {
			t.Errorf("Empty body returned for %s", jsURL)
		}
	}
}

