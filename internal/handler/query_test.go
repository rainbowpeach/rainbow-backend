package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"rainbow-backend/internal/model"
	"rainbow-backend/internal/service"
)

func TestQueryScenePageConfigUsesExplicitSceneCodeWithoutHostResolution(t *testing.T) {
	gin.SetMode(gin.TestMode)

	scenePageConfigRepo := &stubHandlerScenePageConfigRepo{
		item: &model.ScenePageConfig{
			SceneCode:   "curry",
			TextDefault: "Today",
			TagsDefault: model.JSONStringArray{"daily"},
		},
	}
	handler := NewQueryHandler(
		service.NewContentService(&stubHandlerContentRepo{}),
		service.NewScenePageConfigService(scenePageConfigRepo),
	)

	router := gin.New()
	router.GET("/api/query/scene-page-config", handler.GetScenePageConfig)

	req := httptest.NewRequest(http.MethodGet, "/api/query/scene-page-config?scene_code=curry", nil)
	req.Host = "unknown.dapinsport.cn"
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
	if resp.Data.SceneCode != "curry" {
		t.Fatalf("expected scene_code curry, got %q", resp.Data.SceneCode)
	}
	if scenePageConfigRepo.getScene != "curry" {
		t.Fatalf("expected lookup by explicit scene_code curry, got %q", scenePageConfigRepo.getScene)
	}
}

func TestQueryContentUsesExplicitSceneCodeAndDateWithoutHostResolution(t *testing.T) {
	gin.SetMode(gin.TestMode)

	contentRepo := &stubHandlerContentRepo{
		item: &model.ContentItem{
			ID:        7,
			SceneCode: "curry",
			Date:      "2026-09-19",
			Text:      "Today",
			CreatedAt: time.Date(2026, 9, 19, 0, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2026, 9, 19, 0, 0, 0, 0, time.UTC),
		},
	}
	handler := NewQueryHandler(
		service.NewContentService(contentRepo),
		service.NewScenePageConfigService(&stubHandlerScenePageConfigRepo{}),
	)

	router := gin.New()
	router.GET("/api/query/content", handler.GetContent)

	req := httptest.NewRequest(http.MethodGet, "/api/query/content?scene_code=curry&date=2026-09-19", nil)
	req.Host = "admin.dapinsport.cn"
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", recorder.Code, recorder.Body.String())
	}

	var resp struct {
		Code int                   `json:"code"`
		Data model.ContentResponse `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.Code != model.CodeOK {
		t.Fatalf("expected response code 0, got %d", resp.Code)
	}
	if resp.Data.SceneCode != "curry" || resp.Data.Date != "2026-09-19" {
		t.Fatalf("unexpected content response: %+v", resp.Data)
	}
	if contentRepo.getScene != "curry" || contentRepo.getDate != "2026-09-19" {
		t.Fatalf("expected lookup by explicit scene/date, got scene=%q date=%q", contentRepo.getScene, contentRepo.getDate)
	}
}

func TestQueryRejectsMissingRequiredParameters(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewQueryHandler(
		service.NewContentService(&stubHandlerContentRepo{}),
		service.NewScenePageConfigService(&stubHandlerScenePageConfigRepo{}),
	)
	router := gin.New()
	router.GET("/api/query/scene-page-config", handler.GetScenePageConfig)
	router.GET("/api/query/content", handler.GetContent)

	tests := []struct {
		name string
		path string
	}{
		{name: "scene page config", path: "/api/query/scene-page-config"},
		{name: "content scene code", path: "/api/query/content?date=2026-09-19"},
		{name: "content date", path: "/api/query/content?scene_code=curry"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			recorder := httptest.NewRecorder()

			router.ServeHTTP(recorder, req)

			if recorder.Code != http.StatusBadRequest {
				t.Fatalf("expected status 400, got %d body=%s", recorder.Code, recorder.Body.String())
			}

			var resp model.Response
			if err := json.Unmarshal(recorder.Body.Bytes(), &resp); err != nil {
				t.Fatalf("unmarshal response: %v", err)
			}
			if resp.Code != model.CodeInvalidParams {
				t.Fatalf("expected code %d, got %d", model.CodeInvalidParams, resp.Code)
			}
		})
	}
}

func TestQueryMapsNotFoundAndInvalidDateErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)

	scenePageConfigHandler := NewQueryHandler(
		service.NewContentService(&stubHandlerContentRepo{}),
		service.NewScenePageConfigService(&stubHandlerScenePageConfigRepo{getErr: gorm.ErrRecordNotFound}),
	)
	scenePageConfigRouter := gin.New()
	scenePageConfigRouter.GET("/api/query/scene-page-config", scenePageConfigHandler.GetScenePageConfig)

	notFoundRecorder := httptest.NewRecorder()
	scenePageConfigRouter.ServeHTTP(
		notFoundRecorder,
		httptest.NewRequest(http.MethodGet, "/api/query/scene-page-config?scene_code=curry", nil),
	)
	if notFoundRecorder.Code != http.StatusNotFound {
		t.Fatalf("expected scene page config status 404, got %d", notFoundRecorder.Code)
	}

	contentHandler := NewQueryHandler(
		service.NewContentService(&stubHandlerContentRepo{}),
		service.NewScenePageConfigService(&stubHandlerScenePageConfigRepo{}),
	)
	contentRouter := gin.New()
	contentRouter.GET("/api/query/content", contentHandler.GetContent)

	invalidDateRecorder := httptest.NewRecorder()
	contentRouter.ServeHTTP(
		invalidDateRecorder,
		httptest.NewRequest(http.MethodGet, "/api/query/content?scene_code=curry&date=2026/09/19", nil),
	)
	if invalidDateRecorder.Code != http.StatusBadRequest {
		t.Fatalf("expected invalid date status 400, got %d", invalidDateRecorder.Code)
	}

	var resp model.Response
	if err := json.Unmarshal(invalidDateRecorder.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.Code != model.CodeInvalidDateFormat {
		t.Fatalf("expected code %d, got %d", model.CodeInvalidDateFormat, resp.Code)
	}
}
