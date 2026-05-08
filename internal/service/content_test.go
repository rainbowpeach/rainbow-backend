package service

import (
	"context"
	"errors"
	"testing"
	"time"

	gmysql "github.com/go-sql-driver/mysql"
	"gorm.io/gorm"

	"rainbow-backend/internal/model"
)

type stubContentRepo struct {
	item       *model.ContentItem
	items      []model.ContentItem
	total      int64
	getErr     error
	createErr  error
	updateErr  error
	deleteErr  error
	listErr    error
	created    *model.ContentItem
	updatedID  uint
	updated    *model.ContentItem
	deletedID  uint
	getScene   string
	getDate    string
	listPage   int
	listSize   int
	listFilter model.ContentFilter
}

func (r *stubContentRepo) GetBySceneAndDate(_ context.Context, sceneCode, date string) (*model.ContentItem, error) {
	if r.getErr != nil {
		return nil, r.getErr
	}
	r.getScene = sceneCode
	r.getDate = date
	return r.item, nil
}

func (r *stubContentRepo) Create(_ context.Context, item *model.ContentItem) error {
	if r.createErr != nil {
		return r.createErr
	}
	r.created = item
	item.ID = 99
	return nil
}

func (r *stubContentRepo) UpdateByID(_ context.Context, id uint, item *model.ContentItem) error {
	if r.updateErr != nil {
		return r.updateErr
	}
	r.updatedID = id
	r.updated = item
	return nil
}

func (r *stubContentRepo) DeleteByID(_ context.Context, id uint) error {
	if r.deleteErr != nil {
		return r.deleteErr
	}
	r.deletedID = id
	return nil
}

func (r *stubContentRepo) List(_ context.Context, filter model.ContentFilter, page, pageSize int) ([]model.ContentItem, int64, error) {
	if r.listErr != nil {
		return nil, 0, r.listErr
	}
	r.listFilter = filter
	r.listPage = page
	r.listSize = pageSize
	return r.items, r.total, nil
}

func TestContentServiceGetByDateSuccess(t *testing.T) {
	now := time.Date(2026, 4, 14, 9, 0, 0, 0, time.UTC)
	service := NewContentService(&stubContentRepo{
		item: &model.ContentItem{
			ID:        1,
			SceneCode: "love",
			Date:      "2026-04-14",
			Text:      "hello",
			Tags:      model.JSONStringArray{"心动", "温柔"},
			BgURL:     "https://example.com/bg.jpg",
			Music:     "https://example.com/music.mp3",
			CreatedAt: now,
			UpdatedAt: now,
		},
	})

	result, err := service.GetBySceneAndDate(context.Background(), "love", "2026-04-14")
	if err != nil {
		t.Fatalf("GetBySceneAndDate() error = %v", err)
	}

	if result.ID != 1 {
		t.Fatalf("expected id 1, got %d", result.ID)
	}
	if result.SceneCode != "love" {
		t.Fatalf("expected scene_code love, got %q", result.SceneCode)
	}
	if result.BgURL != "https://example.com/bg.jpg" {
		t.Fatalf("expected bg_url to match, got %q", result.BgURL)
	}
	if result.CreatedAt != "2026-04-14" {
		t.Fatalf("expected createdAt 2026-04-14, got %q", result.CreatedAt)
	}
}

func TestContentServiceGetByDateInvalidDate(t *testing.T) {
	service := NewContentService(&stubContentRepo{})

	_, err := service.GetBySceneAndDate(context.Background(), "default", "2026/04/14")
	if !errors.Is(err, ErrInvalidDateFormat) {
		t.Fatalf("expected ErrInvalidDateFormat, got %v", err)
	}
}

func TestContentServiceGetByDateNotFound(t *testing.T) {
	service := NewContentService(&stubContentRepo{getErr: gorm.ErrRecordNotFound})

	_, err := service.GetBySceneAndDate(context.Background(), "default", "2026-04-14")
	if !errors.Is(err, ErrContentNotFound) {
		t.Fatalf("expected ErrContentNotFound, got %v", err)
	}
}

func TestContentServiceCreateSuccess(t *testing.T) {
	repo := &stubContentRepo{}
	service := NewContentService(repo)

	result, err := service.Create(context.Background(), &model.ContentUpsertRequest{
		SceneCode: "love",
		Date:      "2026-04-14",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if result.ID != 99 {
		t.Fatalf("expected id 99, got %d", result.ID)
	}
	if repo.created == nil || len(repo.created.Tags) != 0 {
		t.Fatal("expected content item to be passed to repo")
	}
	if repo.created.SceneCode != "love" {
		t.Fatalf("expected scene_code love, got %q", repo.created.SceneCode)
	}
	if repo.created.Text != "" || repo.created.BgURL != "" || repo.created.Music != "" {
		t.Fatalf("expected optional fields to default to empty strings, got %+v", repo.created)
	}
}

func TestContentServiceCreateAcceptsPartialPayload(t *testing.T) {
	repo := &stubContentRepo{}
	service := NewContentService(repo)

	_, err := service.Create(context.Background(), &model.ContentUpsertRequest{
		SceneCode: "love",
		Date:      "2026-04-14",
		Text:      " Today is a good day. ",
		Tags:      []string{" warm ", "spring"},
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if repo.created.Text != "Today is a good day." {
		t.Fatalf("expected normalized text, got %q", repo.created.Text)
	}
	if len(repo.created.Tags) != 2 || repo.created.Tags[0] != "warm" || repo.created.Tags[1] != "spring" {
		t.Fatalf("expected normalized tags, got %#v", repo.created.Tags)
	}
}

func TestContentServiceCreateDuplicateDate(t *testing.T) {
	repo := &stubContentRepo{
		createErr: &gmysql.MySQLError{Number: 1062},
	}
	service := NewContentService(repo)

	_, err := service.Create(context.Background(), &model.ContentUpsertRequest{
		SceneCode: "love",
		Date:      "2026-04-14",
	})
	if !errors.Is(err, ErrDuplicateDate) {
		t.Fatalf("expected ErrDuplicateDate, got %v", err)
	}
}

func TestContentServiceCreateAcceptsOptionalFields(t *testing.T) {
	repo := &stubContentRepo{}
	service := NewContentService(repo)

	_, err := service.Create(context.Background(), &model.ContentUpsertRequest{
		SceneCode: "love",
		Date:      "2026-04-14",
		Text:      "   ",
		Tags:      []string{},
		BgURL:     "   ",
		Music:     "   ",
	})
	if err != nil {
		t.Fatalf("expected optional fields to be accepted, got %v", err)
	}
	if repo.created == nil {
		t.Fatal("expected repo create to be called")
	}
	if repo.created.Text != "" || repo.created.BgURL != "" || repo.created.Music != "" {
		t.Fatalf("expected blank optional strings to be normalized, got %+v", repo.created)
	}
}

func TestContentServiceCreateRejectsBlankTags(t *testing.T) {
	repo := &stubContentRepo{}
	service := NewContentService(repo)

	_, err := service.Create(context.Background(), &model.ContentUpsertRequest{
		SceneCode: "love",
		Date:      "2026-04-14",
		Tags:      []string{" ", "温柔"},
	})
	if !errors.Is(err, ErrInvalidContentParams) {
		t.Fatalf("expected ErrInvalidContentParams, got %v", err)
	}
}

func TestContentServiceUpdateNotFound(t *testing.T) {
	service := NewContentService(&stubContentRepo{updateErr: gorm.ErrRecordNotFound})

	_, err := service.Update(context.Background(), 1, &model.ContentUpsertRequest{
		SceneCode: "love",
		Date:      "2026-04-14",
	})
	if !errors.Is(err, ErrContentNotFound) {
		t.Fatalf("expected ErrContentNotFound, got %v", err)
	}
}

func TestContentServiceDeleteSuccess(t *testing.T) {
	repo := &stubContentRepo{}
	service := NewContentService(repo)

	result, err := service.Delete(context.Background(), 7)
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if result.ID != 7 {
		t.Fatalf("expected id 7, got %d", result.ID)
	}
	if repo.deletedID != 7 {
		t.Fatalf("expected repo delete id 7, got %d", repo.deletedID)
	}
}

func TestContentServiceListSuccess(t *testing.T) {
	now := time.Date(2026, 4, 14, 9, 0, 0, 0, time.UTC)
	service := NewContentService(&stubContentRepo{
		items: []model.ContentItem{
			{
				ID:        1,
				SceneCode: "love",
				Date:      "2026-04-14",
				Text:      "",
				Tags:      model.JSONStringArray{},
				BgURL:     "",
				Music:     "",
				CreatedAt: now,
				UpdatedAt: now,
			},
		},
		total: 1,
	})

	result, err := service.List(context.Background(), model.ContentFilter{
		SceneCode: "love",
		Date:      "2026-04-14",
	}, 1, 10)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if result.Total != 1 || result.Page != 1 || result.PageSize != 10 {
		t.Fatalf("unexpected pagination result: %+v", result)
	}
	if len(result.List) != 1 {
		t.Fatalf("expected one result, got %d", len(result.List))
	}
	if repo := service.contentRepo.(*stubContentRepo); repo.listFilter.SceneCode != "love" || repo.listFilter.Date != "2026-04-14" {
		t.Fatalf("expected list filter to be passed through, got %+v", repo.listFilter)
	}
}
