package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"rainbow-backend/internal/config"
	"rainbow-backend/internal/model"
	"rainbow-backend/internal/service"
)

type stubHandlerContentRepo struct {
	item     *model.ContentItem
	getErr   error
	getScene string
	getDate  string
}

func (r *stubHandlerContentRepo) GetBySceneAndDate(_ context.Context, sceneCode, date string) (*model.ContentItem, error) {
	if r.getErr != nil {
		return nil, r.getErr
	}
	r.getScene = sceneCode
	r.getDate = date
	return r.item, nil
}

func (r *stubHandlerContentRepo) GetByID(_ context.Context, _ uint) (*model.ContentItem, error) {
	if r.getErr != nil {
		return nil, r.getErr
	}
	return r.item, nil
}

func (r *stubHandlerContentRepo) Create(_ context.Context, _ *model.ContentItem) error {
	return nil
}

func (r *stubHandlerContentRepo) UpdateByID(_ context.Context, _ uint, _ *model.ContentItem) error {
	return nil
}

func (r *stubHandlerContentRepo) DeleteByID(_ context.Context, _ uint) error {
	return nil
}

func (r *stubHandlerContentRepo) List(_ context.Context, _ model.ContentFilter, _, _ int) ([]model.ContentItem, int64, error) {
	return nil, 0, nil
}

type stubPublicSceneDomainRepo struct {
	item    *model.SceneDomain
	getErr  error
	gotHost string
}

func (r *stubPublicSceneDomainRepo) GetByHost(_ context.Context, host string) (*model.SceneDomain, error) {
	if r.getErr != nil {
		return nil, r.getErr
	}
	r.gotHost = host
	return r.item, nil
}

func (r *stubPublicSceneDomainRepo) Create(_ context.Context, _ *model.SceneDomain) error {
	return nil
}

func (r *stubPublicSceneDomainRepo) UpdateByHost(_ context.Context, _ string, _ *model.SceneDomain) error {
	return nil
}

func (r *stubPublicSceneDomainRepo) DeleteByHost(_ context.Context, _ string) error {
	return nil
}

func (r *stubPublicSceneDomainRepo) List(_ context.Context, _ model.SceneDomainFilter) ([]model.SceneDomain, int64, error) {
	return nil, 0, nil
}

func TestAdminUploadAudioReturnsStableResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rootDir := t.TempDir()
	handler := NewAdminUploadHandler(service.NewUploadService(config.UploadConfig{
		RootDir:      rootDir,
		ImageMaxSize: 10 * 1024 * 1024,
		AudioMaxSize: 20 * 1024 * 1024,
	}))

	router := gin.New()
	router.POST("/api/admin/upload/audio", handler.UploadAudio)

	req := newMultipartUploadRequest(t, "http://admin.example.com/api/admin/upload/audio", "file", "demo.wav", "audio/wav", wavFixtureForHandlerTest(), "love")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d with body %s", recorder.Code, recorder.Body.String())
	}

	var resp struct {
		Code    int                  `json:"code"`
		Message string               `json:"message"`
		Data    model.UploadResponse `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if resp.Code != model.CodeOK {
		t.Fatalf("expected response code 0, got %d", resp.Code)
	}
	if resp.Message != "ok" {
		t.Fatalf("expected message ok, got %q", resp.Message)
	}
	if !strings.HasPrefix(resp.Data.URL, "http://admin.example.com/static/love/audio/") {
		t.Fatalf("expected audio URL prefix, got %q", resp.Data.URL)
	}
	if filepath.Ext(resp.Data.Filename) != ".wav" {
		t.Fatalf("expected stored extension .wav, got %q", resp.Data.Filename)
	}
	if resp.Data.ContentType != "audio/wav" {
		t.Fatalf("expected contentType audio/wav, got %q", resp.Data.ContentType)
	}
	if resp.Data.Size != int64(len(wavFixtureForHandlerTest())) {
		t.Fatalf("expected size %d, got %d", len(wavFixtureForHandlerTest()), resp.Data.Size)
	}

	savedPath := filepath.Join(rootDir, "love", "audio", resp.Data.Filename)
	if _, err := os.Stat(savedPath); err != nil {
		t.Fatalf("expected uploaded audio file to exist at %s: %v", savedPath, err)
	}
}

func TestAdminUploadVideoReturnsStableResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rootDir := t.TempDir()
	handler := NewAdminUploadHandler(service.NewUploadService(config.UploadConfig{
		RootDir:      rootDir,
		ImageMaxSize: 10 * 1024 * 1024,
		AudioMaxSize: 20 * 1024 * 1024,
		VideoMaxSize: 500 * 1024 * 1024,
	}))

	router := gin.New()
	router.POST("/api/admin/upload/video", handler.UploadVideo)

	req := newMultipartUploadRequest(t, "http://admin.example.com/api/admin/upload/video", "file", "intro.bin", "application/octet-stream", mp4VideoFixtureForHandlerTest(), "love")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d with body %s", recorder.Code, recorder.Body.String())
	}

	var resp struct {
		Code int                  `json:"code"`
		Data model.UploadResponse `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if resp.Code != model.CodeOK {
		t.Fatalf("expected response code 0, got %d", resp.Code)
	}
	if !strings.HasPrefix(resp.Data.URL, "http://admin.example.com/static/love/videos/") {
		t.Fatalf("expected video URL prefix, got %q", resp.Data.URL)
	}
	if filepath.Ext(resp.Data.Filename) != ".mp4" {
		t.Fatalf("expected stored extension .mp4, got %q", filepath.Ext(resp.Data.Filename))
	}
	if resp.Data.ContentType != "video/mp4" {
		t.Fatalf("expected video/mp4 content type, got %q", resp.Data.ContentType)
	}

	savedPath := filepath.Join(rootDir, "love", "videos", resp.Data.Filename)
	if _, err := os.Stat(savedPath); err != nil {
		t.Fatalf("expected uploaded video file to exist at %s: %v", savedPath, err)
	}
}

func TestAdminUploadAvatarReturnsStableResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rootDir := t.TempDir()
	handler := NewAdminUploadHandler(service.NewUploadService(config.UploadConfig{
		RootDir:      rootDir,
		ImageMaxSize: 10 * 1024 * 1024,
		AudioMaxSize: 20 * 1024 * 1024,
		VideoMaxSize: 500 * 1024 * 1024,
	}))

	router := gin.New()
	router.POST("/api/admin/upload/avatar", handler.UploadAvatar)

	req := newMultipartUploadRequest(t, "http://admin.example.com/api/admin/upload/avatar", "file", "avatar.png", "image/png", pngFixtureForHandlerTest(), "love")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d with body %s", recorder.Code, recorder.Body.String())
	}

	var resp struct {
		Code int                  `json:"code"`
		Data model.UploadResponse `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if resp.Code != model.CodeOK {
		t.Fatalf("expected response code 0, got %d", resp.Code)
	}
	if !strings.HasPrefix(resp.Data.URL, "http://admin.example.com/static/love/avatars/") {
		t.Fatalf("expected avatar URL prefix, got %q", resp.Data.URL)
	}
	if filepath.Ext(resp.Data.Filename) != ".png" {
		t.Fatalf("expected stored extension .png, got %q", filepath.Ext(resp.Data.Filename))
	}

	savedPath := filepath.Join(rootDir, "love", "avatars", resp.Data.Filename)
	if _, err := os.Stat(savedPath); err != nil {
		t.Fatalf("expected uploaded avatar file to exist at %s: %v", savedPath, err)
	}
}

func TestPublicContentReturnsMusicField(t *testing.T) {
	gin.SetMode(gin.TestMode)

	now := time.Date(2026, 4, 7, 9, 0, 0, 0, time.UTC)
	contentRepo := &stubHandlerContentRepo{
		item: &model.ContentItem{
			ID:        1,
			SceneCode: "love",
			Date:      "2026-04-07",
			Text:      "Today is a good day.",
			Tags:      model.JSONStringArray{"warm", "spring"},
			ImageURLs: model.JSONStringArray{"https://love.example.com/static/love/images/1.png"},
			BgURL:     "https://love.example.com/static/love/images/demo.png",
			Music:     "https://love.example.com/static/love/audio/demo.mp3",
			CreatedAt: now,
			UpdatedAt: now,
		},
	}
	sceneRepo := &stubPublicSceneDomainRepo{
		item: &model.SceneDomain{
			Host:      "love.example.com",
			SceneCode: "love",
		},
	}

	handler := NewPublicContentHandler(
		service.NewContentService(contentRepo),
		service.NewScenePageConfigService(&stubHandlerScenePageConfigRepo{}),
		service.NewSceneResolver(sceneRepo),
		config.SceneConfig{},
	)

	router := gin.New()
	router.GET("/api/public/content", handler.GetByDate)

	req := httptest.NewRequest(http.MethodGet, "/api/public/content?date=2026-04-07", nil)
	req.Host = "love.example.com"
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d with body %s", recorder.Code, recorder.Body.String())
	}

	var resp struct {
		Code    int                   `json:"code"`
		Message string                `json:"message"`
		Data    model.ContentResponse `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if resp.Code != model.CodeOK {
		t.Fatalf("expected response code 0, got %d", resp.Code)
	}
	if resp.Data.SceneCode != "love" {
		t.Fatalf("expected scene_code love, got %q", resp.Data.SceneCode)
	}
	if resp.Data.Music != "https://love.example.com/static/love/audio/demo.mp3" {
		t.Fatalf("expected music URL to be returned, got %q", resp.Data.Music)
	}
	if len(resp.Data.ImageURLs) != 1 || resp.Data.ImageURLs[0] != "https://love.example.com/static/love/images/1.png" {
		t.Fatalf("expected image_urls to be returned, got %#v", resp.Data.ImageURLs)
	}
	if contentRepo.getScene != "love" {
		t.Fatalf("expected content lookup scene love, got %q", contentRepo.getScene)
	}
	if contentRepo.getDate != "2026-04-07" {
		t.Fatalf("expected content lookup date 2026-04-07, got %q", contentRepo.getDate)
	}
	if sceneRepo.gotHost != "love.example.com" {
		t.Fatalf("expected resolved host love.example.com, got %q", sceneRepo.gotHost)
	}
}

func TestPublicContentReturnsRawEmptyFields(t *testing.T) {
	gin.SetMode(gin.TestMode)

	now := time.Date(2026, 4, 7, 9, 0, 0, 0, time.UTC)
	contentRepo := &stubHandlerContentRepo{
		item: &model.ContentItem{
			ID:        1,
			SceneCode: "love",
			Date:      "2026-04-07",
			Text:      "",
			Tags:      model.JSONStringArray{},
			BgURL:     "",
			Music:     "",
			CreatedAt: now,
			UpdatedAt: now,
		},
	}
	sceneRepo := &stubPublicSceneDomainRepo{
		item: &model.SceneDomain{
			Host:      "love.example.com",
			SceneCode: "love",
		},
	}

	handler := NewPublicContentHandler(
		service.NewContentService(contentRepo),
		service.NewScenePageConfigService(&stubHandlerScenePageConfigRepo{}),
		service.NewSceneResolver(sceneRepo),
		config.SceneConfig{},
	)

	router := gin.New()
	router.GET("/api/public/content", handler.GetByDate)

	req := httptest.NewRequest(http.MethodGet, "/api/public/content?date=2026-04-07", nil)
	req.Host = "love.example.com"
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d with body %s", recorder.Code, recorder.Body.String())
	}

	var resp struct {
		Code int                   `json:"code"`
		Data model.ContentResponse `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if resp.Data.Text != "" {
		t.Fatalf("expected raw empty text, got %q", resp.Data.Text)
	}
	if len(resp.Data.Tags) != 0 {
		t.Fatalf("expected raw empty tags, got %#v", resp.Data.Tags)
	}
	if resp.Data.BgURL != "" {
		t.Fatalf("expected raw empty bg_url, got %q", resp.Data.BgURL)
	}
	if resp.Data.Music != "" {
		t.Fatalf("expected raw empty music, got %q", resp.Data.Music)
	}
}

func newMultipartUploadRequest(t *testing.T, targetURL, fieldName, filename, contentType string, data []byte, sceneCode string) *http.Request {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	if err := writer.WriteField("scene_code", sceneCode); err != nil {
		t.Fatalf("WriteField() error = %v", err)
	}

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

	req := httptest.NewRequest(http.MethodPost, targetURL, &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}

func wavFixtureForHandlerTest() []byte {
	return []byte{
		0x52, 0x49, 0x46, 0x46, 0x24, 0x08, 0x00, 0x00,
		0x57, 0x41, 0x56, 0x45, 0x66, 0x6d, 0x74, 0x20,
		0x10, 0x00, 0x00, 0x00, 0x01, 0x00, 0x01, 0x00,
		0x44, 0xac, 0x00, 0x00, 0x88, 0x58, 0x01, 0x00,
		0x02, 0x00, 0x10, 0x00, 0x64, 0x61, 0x74, 0x61,
		0x00, 0x08, 0x00, 0x00,
	}
}

func pngFixtureForHandlerTest() []byte {
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

func mp4VideoFixtureForHandlerTest() []byte {
	return []byte{
		0x00, 0x00, 0x00, 0x18,
		'f', 't', 'y', 'p',
		'i', 's', 'o', 'm',
		0x00, 0x00, 0x02, 0x00,
		'i', 's', 'o', 'm',
		'i', 's', 'o', '2',
	}
}
