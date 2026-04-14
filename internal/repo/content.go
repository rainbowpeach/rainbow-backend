package repo

import (
	"context"

	"gorm.io/gorm"

	"rainbow-backend/internal/model"
)

type ContentRepository interface {
	GetByDate(ctx context.Context, date string) (*model.ContentItem, error)
}

type GormContentRepository struct {
	db *gorm.DB
}

func NewContentRepository(db *gorm.DB) *GormContentRepository {
	return &GormContentRepository{db: db}
}

func (r *GormContentRepository) GetByDate(ctx context.Context, date string) (*model.ContentItem, error) {
	var item model.ContentItem
	if err := r.db.WithContext(ctx).Where("date = ?", date).First(&item).Error; err != nil {
		return nil, err
	}

	return &item, nil
}
