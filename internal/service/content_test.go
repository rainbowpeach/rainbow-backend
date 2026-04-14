package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"gorm.io/gorm"

	"rainbow-backend/internal/model"
)

type stubContentRepo struct {
	item *model.ContentItem
	err  error
}

func (r *stubContentRepo) GetByDate(context.Context, string) (*model.ContentItem, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.item, nil
}

func TestContentServiceGetByDateSuccess(t *testing.T) {
	now := time.Date(2026, 4, 14, 9, 0, 0, 0, time.UTC)
	service := NewContentService(&stubContentRepo{
		item: &model.ContentItem{
			ID:        1,
			Date:      "2026-04-14",
			Text:      "hello",
			Tags:      model.JSONStringArray{"心动", "温柔"},
			BgURL:     "https://example.com/bg.jpg",
			Music:     "https://example.com/music.mp3",
			CreatedAt: now,
			UpdatedAt: now,
		},
	})

	result, err := service.GetByDate(context.Background(), "2026-04-14")
	if err != nil {
		t.Fatalf("GetByDate() error = %v", err)
	}

	if result.ID != 1 {
		t.Fatalf("expected id 1, got %d", result.ID)
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

	_, err := service.GetByDate(context.Background(), "2026/04/14")
	if !errors.Is(err, ErrInvalidDateFormat) {
		t.Fatalf("expected ErrInvalidDateFormat, got %v", err)
	}
}

func TestContentServiceGetByDateNotFound(t *testing.T) {
	service := NewContentService(&stubContentRepo{err: gorm.ErrRecordNotFound})

	_, err := service.GetByDate(context.Background(), "2026-04-14")
	if !errors.Is(err, ErrContentNotFound) {
		t.Fatalf("expected ErrContentNotFound, got %v", err)
	}
}
