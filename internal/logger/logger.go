package logger

import (
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"

	"rainbow-backend/internal/config"
)

const (
	accessLogFile = "access.log"
	appLogFile    = "app.log"
)

type Runtime struct {
	AccessLogPath string
	AppLogPath    string
	closers       []io.Closer
}

func Setup(cfg config.LogConfig) (*Runtime, error) {
	if err := os.MkdirAll(cfg.RootDir, 0o755); err != nil {
		return nil, err
	}

	accessWriter, accessPath, err := openLogFile(cfg.RootDir, accessLogFile)
	if err != nil {
		return nil, err
	}

	appWriter, appPath, err := openLogFile(cfg.RootDir, appLogFile)
	if err != nil {
		_ = accessWriter.Close()
		return nil, err
	}

	gin.DefaultWriter = io.MultiWriter(os.Stdout, accessWriter)
	gin.DefaultErrorWriter = io.MultiWriter(os.Stderr, appWriter)
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
	log.SetOutput(io.MultiWriter(os.Stdout, appWriter))

	return &Runtime{
		AccessLogPath: accessPath,
		AppLogPath:    appPath,
		closers:       []io.Closer{accessWriter, appWriter},
	}, nil
}

func (r *Runtime) Close() error {
	var firstErr error
	for _, closer := range r.closers {
		if err := closer.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func openLogFile(rootDir, filename string) (*os.File, string, error) {
	path := filepath.Join(rootDir, filename)
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, "", err
	}

	return file, path, nil
}
