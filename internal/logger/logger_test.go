package logger

import (
	"log"
	"os"
	"testing"

	"github.com/gin-gonic/gin"

	"rainbow-backend/internal/config"
)

func TestSetupCreatesLogFiles(t *testing.T) {
	originalWriter := gin.DefaultWriter
	originalErrorWriter := gin.DefaultErrorWriter
	originalLogWriter := log.Writer()
	t.Cleanup(func() {
		gin.DefaultWriter = originalWriter
		gin.DefaultErrorWriter = originalErrorWriter
		log.SetOutput(originalLogWriter)
	})

	rootDir := t.TempDir()
	runtime, err := Setup(config.LogConfig{RootDir: rootDir})
	if err != nil {
		t.Fatalf("Setup() error = %v", err)
	}
	t.Cleanup(func() {
		if closeErr := runtime.Close(); closeErr != nil {
			t.Fatalf("Close() error = %v", closeErr)
		}
	})

	for _, path := range []string{runtime.AccessLogPath, runtime.AppLogPath} {
		info, statErr := os.Stat(path)
		if statErr != nil {
			t.Fatalf("expected log file %q to exist: %v", path, statErr)
		}
		if info.IsDir() {
			t.Fatalf("expected log path %q to be a file", path)
		}
	}
}
