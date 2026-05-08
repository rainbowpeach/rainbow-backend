package service

import (
	"context"
	"errors"
	"testing"

	gmysql "github.com/go-sql-driver/mysql"
	"gorm.io/gorm"

	"rainbow-backend/internal/model"
)

type stubScenePageConfigRepo struct {
	item         *model.ScenePageConfig
	items        []model.ScenePageConfig
	total        int64
	getErr       error
	createErr    error
	updateErr    error
	deleteErr    error
	listErr      error
	createdItem  *model.ScenePageConfig
	updatedScene string
	updatedItem  *model.ScenePageConfig
	deletedScene string
}

func (r *stubScenePageConfigRepo) GetBySceneCode(_ context.Context, sceneCode string) (*model.ScenePageConfig, error) {
	if r.getErr != nil {
		return nil, r.getErr
	}
	r.updatedScene = sceneCode
	return r.item, nil
}

func (r *stubScenePageConfigRepo) Create(_ context.Context, item *model.ScenePageConfig) error {
	if r.createErr != nil {
		return r.createErr
	}
	r.createdItem = item
	return nil
}

func (r *stubScenePageConfigRepo) UpdateBySceneCode(_ context.Context, sceneCode string, item *model.ScenePageConfig) error {
	if r.updateErr != nil {
		return r.updateErr
	}
	r.updatedScene = sceneCode
	r.updatedItem = item
	return nil
}

func (r *stubScenePageConfigRepo) DeleteBySceneCode(_ context.Context, sceneCode string) error {
	if r.deleteErr != nil {
		return r.deleteErr
	}
	r.deletedScene = sceneCode
	return nil
}

func (r *stubScenePageConfigRepo) List(_ context.Context, _ model.ScenePageConfigFilter) ([]model.ScenePageConfig, int64, error) {
	if r.listErr != nil {
		return nil, 0, r.listErr
	}
	return r.items, r.total, nil
}

func TestScenePageConfigServiceCreateSuccess(t *testing.T) {
	repo := &stubScenePageConfigRepo{}
	service := NewScenePageConfigService(repo)

	result, err := service.Create(context.Background(), &model.ScenePageConfigUpsertRequest{
		SceneCode:        "love",
		Logo:             "/static/love/images/logo.png",
		Banner:           "/static/love/images/banner.png",
		BacImg:           "/static/love/images/bg.png",
		DefaultBgURL:     "/static/love/images/default_bg.png",
		DefaultMusic:     "/static/love/audio/default_music.mp3",
		TextDefault:      "hello",
		TagsDefault:      []string{"warm", "spring"},
		PlayButtonColor:  "#1A2B3C",
		TextDefaultColor: "#ffffff",
		TagsColor:        "#111111",
		TagsBacColor:     "#222222",
		DateColor:        "#333333",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if result.SceneCode != "love" {
		t.Fatalf("expected scene_code love, got %q", result.SceneCode)
	}
	if repo.createdItem == nil {
		t.Fatal("expected repo create to be called")
	}
	if repo.createdItem.PlayButtonColor != "#1a2b3c" {
		t.Fatalf("expected normalized color, got %q", repo.createdItem.PlayButtonColor)
	}
	if repo.createdItem.DefaultBgURL != "/static/love/images/default_bg.png" {
		t.Fatalf("expected default_bg_url to pass through, got %q", repo.createdItem.DefaultBgURL)
	}
	if repo.createdItem.DefaultMusic != "/static/love/audio/default_music.mp3" {
		t.Fatalf("expected default_music to pass through, got %q", repo.createdItem.DefaultMusic)
	}
}

func TestScenePageConfigServiceCreateRejectsInvalidColor(t *testing.T) {
	service := NewScenePageConfigService(&stubScenePageConfigRepo{})

	_, err := service.Create(context.Background(), &model.ScenePageConfigUpsertRequest{
		SceneCode:       "love",
		PlayButtonColor: "red",
	})
	if !errors.Is(err, ErrInvalidScenePageConfigParams) {
		t.Fatalf("expected ErrInvalidScenePageConfigParams, got %v", err)
	}
}

func TestScenePageConfigServiceCreateRejectsInvalidTags(t *testing.T) {
	service := NewScenePageConfigService(&stubScenePageConfigRepo{})

	_, err := service.Create(context.Background(), &model.ScenePageConfigUpsertRequest{
		SceneCode:   "love",
		TagsDefault: []string{"ok", " "},
	})
	if !errors.Is(err, ErrInvalidScenePageConfigParams) {
		t.Fatalf("expected ErrInvalidScenePageConfigParams, got %v", err)
	}
}

func TestScenePageConfigServiceGetNotFound(t *testing.T) {
	service := NewScenePageConfigService(&stubScenePageConfigRepo{getErr: gorm.ErrRecordNotFound})

	_, err := service.GetBySceneCode(context.Background(), "love")
	if !errors.Is(err, ErrScenePageConfigNotFound) {
		t.Fatalf("expected ErrScenePageConfigNotFound, got %v", err)
	}
}

func TestScenePageConfigServiceCreateDuplicate(t *testing.T) {
	service := NewScenePageConfigService(&stubScenePageConfigRepo{
		createErr: &gmysql.MySQLError{Number: 1062},
	})

	_, err := service.Create(context.Background(), &model.ScenePageConfigUpsertRequest{
		SceneCode: "love",
	})
	if !errors.Is(err, ErrDuplicateScenePageConfig) {
		t.Fatalf("expected ErrDuplicateScenePageConfig, got %v", err)
	}
}

func TestScenePageConfigServiceUpdateRequiresPathSceneMatch(t *testing.T) {
	service := NewScenePageConfigService(&stubScenePageConfigRepo{})

	_, err := service.Update(context.Background(), "love", &model.ScenePageConfigUpsertRequest{
		SceneCode: "sweet",
	})
	if !errors.Is(err, ErrInvalidScenePageConfigParams) {
		t.Fatalf("expected ErrInvalidScenePageConfigParams, got %v", err)
	}
}
