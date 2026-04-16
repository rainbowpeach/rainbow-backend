package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"rainbow-backend/internal/config"
	"rainbow-backend/internal/model"
)

var ErrFileRequired = errors.New("file is required")
var ErrEmptyFile = errors.New("empty file")
var ErrUnsupportedFileType = errors.New("unsupported file type")
var ErrFileTooLarge = errors.New("file too large")

type UploadRequest struct {
	FileHeader *multipart.FileHeader
	BaseURL    string
}

type UploadService struct {
	cfg config.UploadConfig
}

type uploadCategory struct {
	subdir       string
	publicPath   string
	maxSize      int64
	allowedTypes map[string]map[string]struct{}
}

func NewUploadService(cfg config.UploadConfig) *UploadService {
	return &UploadService{cfg: cfg}
}

func (s *UploadService) UploadImage(ctx context.Context, req *UploadRequest) (*model.UploadResponse, error) {
	return s.upload(ctx, req, uploadCategory{
		subdir:     "images",
		publicPath: "/static/images/",
		maxSize:    s.cfg.ImageMaxSize,
		allowedTypes: map[string]map[string]struct{}{
			".jpg":  contentTypeSet("image/jpeg"),
			".jpeg": contentTypeSet("image/jpeg"),
			".png":  contentTypeSet("image/png"),
			".webp": contentTypeSet("image/webp"),
		},
	})
}

func (s *UploadService) UploadAudio(ctx context.Context, req *UploadRequest) (*model.UploadResponse, error) {
	return s.upload(ctx, req, uploadCategory{
		subdir:     "audio",
		publicPath: "/static/audio/",
		maxSize:    s.cfg.AudioMaxSize,
		allowedTypes: map[string]map[string]struct{}{
			".mp3": contentTypeSet("audio/mpeg", "audio/mp3"),
			".wav": contentTypeSet("audio/wav", "audio/wave", "audio/x-wav"),
			".ogg": contentTypeSet("audio/ogg", "application/ogg"),
			".m4a": contentTypeSet("audio/mp4", "video/mp4"),
		},
	})
}

func (s *UploadService) upload(_ context.Context, req *UploadRequest, category uploadCategory) (*model.UploadResponse, error) {
	if req == nil || req.FileHeader == nil {
		return nil, ErrFileRequired
	}
	if req.FileHeader.Size <= 0 {
		return nil, ErrEmptyFile
	}
	if req.FileHeader.Size > category.maxSize {
		return nil, ErrFileTooLarge
	}

	filename := sanitizeFilename(req.FileHeader.Filename)
	ext := strings.ToLower(filepath.Ext(filename))
	allowedContentTypes, ok := category.allowedTypes[ext]
	if !ok {
		return nil, ErrUnsupportedFileType
	}

	contentType, err := sniffContentType(req.FileHeader, allowedContentTypes)
	if err != nil {
		return nil, err
	}

	targetDir := filepath.Join(s.cfg.RootDir, category.subdir)
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return nil, fmt.Errorf("create upload directory: %w", err)
	}

	storedName, targetPath, err := createTargetFile(targetDir, ext)
	if err != nil {
		return nil, fmt.Errorf("create target file: %w", err)
	}

	if err := copyUploadedFile(req.FileHeader, targetPath); err != nil {
		return nil, fmt.Errorf("save uploaded file: %w", err)
	}

	return &model.UploadResponse{
		URL:         joinPublicURL(req.BaseURL, category.publicPath, storedName),
		Filename:    storedName,
		Size:        req.FileHeader.Size,
		ContentType: contentType,
	}, nil
}

func sniffContentType(fileHeader *multipart.FileHeader, allowed map[string]struct{}) (string, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return "", fmt.Errorf("open uploaded file: %w", err)
	}
	defer file.Close()

	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && !errors.Is(err, io.EOF) {
		return "", fmt.Errorf("read uploaded file header: %w", err)
	}

	detected := normalizeContentType(http.DetectContentType(buffer[:n]))
	declared := normalizeContentType(fileHeader.Header.Get("Content-Type"))

	if _, ok := allowed[detected]; ok {
		return detected, nil
	}
	if detected == "application/octet-stream" {
		if _, ok := allowed[declared]; ok {
			return declared, nil
		}
	}

	return "", ErrUnsupportedFileType
}

func createTargetFile(targetDir, ext string) (string, string, error) {
	for range 5 {
		filename, err := generateStoredFilename(ext)
		if err != nil {
			return "", "", err
		}

		targetPath := filepath.Join(targetDir, filename)
		file, err := os.OpenFile(targetPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
		if err == nil {
			if closeErr := file.Close(); closeErr != nil {
				return "", "", closeErr
			}
			return filename, targetPath, nil
		}
		if !errors.Is(err, os.ErrExist) {
			return "", "", err
		}
	}

	return "", "", errors.New("could not allocate unique filename")
}

func copyUploadedFile(fileHeader *multipart.FileHeader, targetPath string) error {
	src, err := fileHeader.Open()
	if err != nil {
		return fmt.Errorf("open uploaded file: %w", err)
	}
	defer src.Close()

	dst, err := os.OpenFile(targetPath, os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open destination file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		_ = os.Remove(targetPath)
		return fmt.Errorf("copy uploaded file: %w", err)
	}

	return nil
}

func generateStoredFilename(ext string) (string, error) {
	randomBytes := make([]byte, 4)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf("generate random suffix: %w", err)
	}

	timestamp := time.Now().UTC().Format("20060102150405")
	suffix := hex.EncodeToString(randomBytes)
	return fmt.Sprintf("%s-%s%s", timestamp, suffix, ext), nil
}

func sanitizeFilename(name string) string {
	name = strings.ReplaceAll(name, "\x00", "")
	name = filepath.Base(strings.TrimSpace(name))
	if name == "." || name == string(filepath.Separator) {
		return ""
	}
	return name
}

func joinPublicURL(baseURL, publicPath, filename string) string {
	return strings.TrimRight(baseURL, "/") + publicPath + filename
}

func normalizeContentType(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return ""
	}

	if idx := strings.Index(value, ";"); idx >= 0 {
		value = value[:idx]
	}

	return strings.TrimSpace(value)
}

func contentTypeSet(values ...string) map[string]struct{} {
	set := make(map[string]struct{}, len(values))
	for _, value := range values {
		set[normalizeContentType(value)] = struct{}{}
	}
	return set
}
