package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"rainbow-backend/internal/config"
	"rainbow-backend/internal/model"
	"rainbow-backend/internal/service"
)

type stubHandlerScenePageConfigRepo struct {
	item      *model.ScenePageConfig
	getErr    error
	createErr error
	created   *model.ScenePageConfig
	getScene  string
	updateErr error
	deleteErr error
}

func (r *stubHandlerScenePageConfigRepo) GetBySceneCode(_ context.Context, sceneCode string) (*model.ScenePageConfig, error) {
	r.getScene = sceneCode
	if r.getErr != nil {
		return nil, r.getErr
	}
	return r.item, nil
}

func (r *stubHandlerScenePageConfigRepo) Create(_ context.Context, item *model.ScenePageConfig) error {
	r.created = item
	return r.createErr
}

func (r *stubHandlerScenePageConfigRepo) UpdateBySceneCode(_ context.Context, _ string, _ *model.ScenePageConfig) error {
	return r.updateErr
}

func (r *stubHandlerScenePageConfigRepo) DeleteBySceneCode(_ context.Context, _ string) error {
	return r.deleteErr
}

func (r *stubHandlerScenePageConfigRepo) List(_ context.Context, _ model.ScenePageConfigFilter) ([]model.ScenePageConfig, int64, error) {
	return nil, 0, nil
}

func TestPublicScenePageConfigReturnsHostResolvedConfig(t *testing.T) {
	gin.SetMode(gin.TestMode)

	contentRepo := &stubHandlerContentRepo{}
	scenePageConfigRepo := &stubHandlerScenePageConfigRepo{
		item: &model.ScenePageConfig{
			SceneCode:        "love",
			Logo:             "/static/love/images/logo.png",
			Banner:           "/static/love/images/banner.png",
			BacImg:           "/static/love/images/bg.png",
			DefaultBgURL:     "/static/love/images/default_bg.png",
			DefaultMusic:     "/static/love/audio/default_music.mp3",
			TagsDefault:      model.JSONStringArray{"heart"},
			PlayButtonColor:  "#123456",
			TextDefaultColor: "#ffffff",
		},
	}
	sceneRepo := &stubPublicSceneDomainRepo{
		item: &model.SceneDomain{
			Host:      "love.dapinsport.cn",
			SceneCode: "love",
		},
	}

	handler := NewPublicContentHandler(
		service.NewContentService(contentRepo),
		service.NewScenePageConfigService(scenePageConfigRepo),
		service.NewSceneResolver(sceneRepo),
		config.SceneConfig{},
	)

	router := gin.New()
	router.GET("/api/public/scene-page-config", handler.GetScenePageConfig)

	req := httptest.NewRequest(http.MethodGet, "/api/public/scene-page-config", nil)
	req.Host = "love.dapinsport.cn"
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", recorder.Code, recorder.Body.String())
	}

	var resp struct {
		Code int                           `json:"code"`
		Data model.ScenePageConfigResponse `json:"data"`
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
	if resp.Data.DefaultBgURL != "/static/love/images/default_bg.png" {
		t.Fatalf("expected default_bg_url to be returned, got %q", resp.Data.DefaultBgURL)
	}
	if resp.Data.DefaultMusic != "/static/love/audio/default_music.mp3" {
		t.Fatalf("expected default_music to be returned, got %q", resp.Data.DefaultMusic)
	}
	if scenePageConfigRepo.getScene != "love" {
		t.Fatalf("expected config lookup by scene love, got %q", scenePageConfigRepo.getScene)
	}
}

func TestPublicScenePageConfigReturnsNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewPublicContentHandler(
		service.NewContentService(&stubHandlerContentRepo{}),
		service.NewScenePageConfigService(&stubHandlerScenePageConfigRepo{getErr: gorm.ErrRecordNotFound}),
		service.NewSceneResolver(&stubPublicSceneDomainRepo{
			item: &model.SceneDomain{
				Host:      "love.dapinsport.cn",
				SceneCode: "love",
			},
		}),
		config.SceneConfig{},
	)

	router := gin.New()
	router.GET("/api/public/scene-page-config", handler.GetScenePageConfig)

	req := httptest.NewRequest(http.MethodGet, "/api/public/scene-page-config", nil)
	req.Host = "love.dapinsport.cn"
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestAdminScenePageConfigCreateRejectsInvalidPayload(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewAdminScenePageConfigHandler(service.NewScenePageConfigService(&stubHandlerScenePageConfigRepo{}))
	router := gin.New()
	router.POST("/api/admin/scene-page-configs", handler.Create)

	body := []byte(`{"scene_code":"love","play_button_color":"red"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/admin/scene-page-configs", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d body=%s", recorder.Code, recorder.Body.String())
	}
}
