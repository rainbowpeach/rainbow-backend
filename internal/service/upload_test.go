package service

import (
	"bytes"
	"context"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"rainbow-backend/internal/config"
)

func TestUploadServiceUploadImageSuccess(t *testing.T) {
	rootDir := t.TempDir()
	service := NewUploadService(config.UploadConfig{
		RootDir:      rootDir,
		ImageMaxSize: 10 * 1024 * 1024,
		AudioMaxSize: 20 * 1024 * 1024,
	})

	fileHeader := newMultipartFileHeader(t, "file", "cover.PNG", "image/png", pngFixture())
	result, err := service.UploadImage(context.Background(), &UploadRequest{
		FileHeader: fileHeader,
		BaseURL:    "http://127.0.0.1:18080",
	})
	if err != nil {
		t.Fatalf("UploadImage() error = %v", err)
	}

	if !strings.HasPrefix(result.URL, "http://127.0.0.1:18080/static/images/") {
		t.Fatalf("expected image URL prefix, got %q", result.URL)
	}
	if filepath.Ext(result.Filename) != ".png" {
		t.Fatalf("expected stored extension .png, got %q", result.Filename)
	}
	if result.ContentType != "image/png" {
		t.Fatalf("expected image/png content type, got %q", result.ContentType)
	}

	savedPath := filepath.Join(rootDir, "images", result.Filename)
	info, err := os.Stat(savedPath)
	if err != nil {
		t.Fatalf("expected saved file, got stat error %v", err)
	}
	if info.Size() != int64(len(pngFixture())) {
		t.Fatalf("expected saved file size %d, got %d", len(pngFixture()), info.Size())
	}
}

func TestUploadServiceUploadAudioSuccess(t *testing.T) {
	rootDir := t.TempDir()
	service := NewUploadService(config.UploadConfig{
		RootDir:      rootDir,
		ImageMaxSize: 10 * 1024 * 1024,
		AudioMaxSize: 20 * 1024 * 1024,
	})

	fileHeader := newMultipartFileHeader(t, "file", "sound.wav", "audio/wav", wavFixture())
	result, err := service.UploadAudio(context.Background(), &UploadRequest{
		FileHeader: fileHeader,
		BaseURL:    "http://127.0.0.1:28080",
	})
	if err != nil {
		t.Fatalf("UploadAudio() error = %v", err)
	}

	if !strings.HasPrefix(result.URL, "http://127.0.0.1:28080/static/audio/") {
		t.Fatalf("expected audio URL prefix, got %q", result.URL)
	}
	if filepath.Ext(result.Filename) != ".wav" {
		t.Fatalf("expected stored extension .wav, got %q", result.Filename)
	}
	if result.ContentType == "" {
		t.Fatal("expected content type to be returned")
	}
}

func TestUploadServiceRejectsUnsupportedFileType(t *testing.T) {
	service := NewUploadService(config.UploadConfig{
		RootDir:      t.TempDir(),
		ImageMaxSize: 10 * 1024 * 1024,
		AudioMaxSize: 20 * 1024 * 1024,
	})

	fileHeader := newMultipartFileHeader(t, "file", "cover.png", "image/png", []byte("plain text"))
	_, err := service.UploadImage(context.Background(), &UploadRequest{
		FileHeader: fileHeader,
		BaseURL:    "http://127.0.0.1:18080",
	})
	if !errors.Is(err, ErrUnsupportedFileType) {
		t.Fatalf("expected ErrUnsupportedFileType, got %v", err)
	}
}

func TestUploadServiceRejectsOversizedFile(t *testing.T) {
	service := NewUploadService(config.UploadConfig{
		RootDir:      t.TempDir(),
		ImageMaxSize: 8,
		AudioMaxSize: 20 * 1024 * 1024,
	})

	fileHeader := newMultipartFileHeader(t, "file", "cover.png", "image/png", pngFixture())
	_, err := service.UploadImage(context.Background(), &UploadRequest{
		FileHeader: fileHeader,
		BaseURL:    "http://127.0.0.1:18080",
	})
	if !errors.Is(err, ErrFileTooLarge) {
		t.Fatalf("expected ErrFileTooLarge, got %v", err)
	}
}

func newMultipartFileHeader(t *testing.T, fieldName, filename, contentType string, data []byte) *multipart.FileHeader {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreatePart(textproto.MIMEHeader{
		"Content-Disposition": {`form-data; name="` + fieldName + `"; filename="` + filename + `"`},
		"Content-Type":        {contentType},
	})
	if err != nil {
		t.Fatalf("CreatePart() error = %v", err)
	}
	if _, err := part.Write(data); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/upload", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if err := req.ParseMultipartForm(int64(body.Len()) + 1024); err != nil {
		t.Fatalf("ParseMultipartForm() error = %v", err)
	}

	return req.MultipartForm.File[fieldName][0]
}

func pngFixture() []byte {
	return []byte{
		0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a,
		0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52,
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x06, 0x00, 0x00, 0x00, 0x1f, 0x15, 0xc4,
		0x89, 0x00, 0x00, 0x00, 0x0a, 0x49, 0x44, 0x41,
		0x54, 0x78, 0x9c, 0x63, 0x60, 0x00, 0x00, 0x00,
		0x02, 0x00, 0x01, 0xe5, 0x27, 0xd4, 0xa2, 0x00,
		0x00, 0x00, 0x00, 0x49, 0x45, 0x4e, 0x44, 0xae,
		0x42, 0x60, 0x82,
	}
}

func wavFixture() []byte {
	return []byte{
		0x52, 0x49, 0x46, 0x46, 0x24, 0x08, 0x00, 0x00,
		0x57, 0x41, 0x56, 0x45, 0x66, 0x6d, 0x74, 0x20,
		0x10, 0x00, 0x00, 0x00, 0x01, 0x00, 0x01, 0x00,
		0x44, 0xac, 0x00, 0x00, 0x88, 0x58, 0x01, 0x00,
		0x02, 0x00, 0x10, 0x00, 0x64, 0x61, 0x74, 0x61,
		0x00, 0x08, 0x00, 0x00,
	}
}
