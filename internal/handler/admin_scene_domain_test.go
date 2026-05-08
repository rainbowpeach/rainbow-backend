package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"rainbow-backend/internal/model"
	"rainbow-backend/internal/service"
)

type stubSceneDomainRepo struct {
	createErr   error
	updateErr   error
	createdItem *model.SceneDomain
	updatedHost string
	updatedItem *model.SceneDomain
}

func (r *stubSceneDomainRepo) GetByHost(_ context.Context, _ string) (*model.SceneDomain, error) {
	return nil, nil
}

func (r *stubSceneDomainRepo) Create(_ context.Context, item *model.SceneDomain) error {
	if r.createErr != nil {
		return r.createErr
	}
	r.createdItem = item
	return nil
}

func (r *stubSceneDomainRepo) UpdateByHost(_ context.Context, currentHost string, item *model.SceneDomain) error {
	if r.updateErr != nil {
		return r.updateErr
	}
	r.updatedHost = currentHost
	r.updatedItem = item
	return nil
}

func (r *stubSceneDomainRepo) DeleteByHost(_ context.Context, _ string) error {
	return nil
}

func (r *stubSceneDomainRepo) List(_ context.Context, _ model.SceneDomainFilter) ([]model.SceneDomain, int64, error) {
	return nil, 0, nil
}

func TestAdminSceneDomainUpdateUsesPathHostWhenBodyHostMissing(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := &stubSceneDomainRepo{}
	handler := NewAdminSceneDomainHandler(service.NewSceneDomainService(repo))
	router := gin.New()
	router.PUT("/api/admin/scene-domains/:host", handler.Update)

	body := []byte(`{"scene_code":"sweet"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/admin/scene-domains/love.example.com", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d with body %s", recorder.Code, recorder.Body.String())
	}
	if repo.updatedHost != "love.example.com" {
		t.Fatalf("expected current host love.example.com, got %q", repo.updatedHost)
	}
	if repo.updatedItem == nil {
		t.Fatal("expected repo update to be called")
	}
	if repo.updatedItem.Host != "love.example.com" {
		t.Fatalf("expected updated item host love.example.com, got %q", repo.updatedItem.Host)
	}
	if repo.updatedItem.SceneCode != "sweet" {
		t.Fatalf("expected scene_code sweet, got %q", repo.updatedItem.SceneCode)
	}

	var resp model.Response
	if err := json.Unmarshal(recorder.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.Code != model.CodeOK {
		t.Fatalf("expected response code 0, got %d", resp.Code)
	}
}

func TestAdminSceneDomainCreateStillRequiresHost(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := &stubSceneDomainRepo{}
	handler := NewAdminSceneDomainHandler(service.NewSceneDomainService(repo))
	router := gin.New()
	router.POST("/api/admin/scene-domains", handler.Create)

	body := []byte(`{"scene_code":"sweet"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/admin/scene-domains", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d with body %s", recorder.Code, recorder.Body.String())
	}
	if repo.createdItem != nil {
		t.Fatal("expected repo create not to be called")
	}

	var resp model.Response
	if err := json.Unmarshal(recorder.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.Code != model.CodeInvalidParams {
		t.Fatalf("expected invalid params code, got %d", resp.Code)
	}
}
