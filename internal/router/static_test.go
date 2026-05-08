package router

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRegisterStaticRoutesServesSceneScopedUploads(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rootDir := t.TempDir()
	filePath := filepath.Join(rootDir, "cluo", "images", "demo.jpg")
	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	content := []byte("demo-image")
	if err := os.WriteFile(filePath, content, 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	engine := gin.New()
	registerStaticRoutes(engine, rootDir)

	req := httptest.NewRequest(http.MethodGet, "/static/cluo/images/demo.jpg", nil)
	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}
	if recorder.Body.String() != string(content) {
		t.Fatalf("expected body %q, got %q", string(content), recorder.Body.String())
	}
}
