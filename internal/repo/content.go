package repo

import (
	"context"

	"gorm.io/gorm"

	"rainbow-backend/internal/model"
)

type ContentRepository interface {
	GetBySceneAndDate(ctx context.Context, sceneCode, date string) (*model.ContentItem, error)
	Create(ctx context.Context, item *model.ContentItem) error
	UpdateByID(ctx context.Context, id uint, item *model.ContentItem) error
	DeleteByID(ctx context.Context, id uint) error
	List(ctx context.Context, filter model.ContentFilter, page, pageSize int) ([]model.ContentItem, int64, error)
}

type GormContentRepository struct {
	db *gorm.DB
}

func NewContentRepository(db *gorm.DB) *GormContentRepository {
	return &GormContentRepository{db: db}
}

func (r *GormContentRepository) GetBySceneAndDate(ctx context.Context, sceneCode, date string) (*model.ContentItem, error) {
	var item model.ContentItem
	if err := r.db.WithContext(ctx).
		Where("scene_code = ? AND date = ?", sceneCode, date).
		First(&item).Error; err != nil {
		return nil, err
	}

	return &item, nil
}

func (r *GormContentRepository) Create(ctx context.Context, item *model.ContentItem) error {
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *GormContentRepository) UpdateByID(ctx context.Context, id uint, item *model.ContentItem) error {
	result := r.db.WithContext(ctx).
		Model(&model.ContentItem{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"date":       item.Date,
			"scene_code": item.SceneCode,
			"text":       item.Text,
			"tags":       item.Tags,
			"bg_url":     item.BgURL,
			"music":      item.Music,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *GormContentRepository) DeleteByID(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.ContentItem{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *GormContentRepository) List(ctx context.Context, filter model.ContentFilter, page, pageSize int) ([]model.ContentItem, int64, error) {
	query := r.db.WithContext(ctx).Model(&model.ContentItem{})
	if filter.SceneCode != "" {
		query = query.Where("scene_code = ?", filter.SceneCode)
	}
	if filter.Date != "" {
		query = query.Where("date = ?", filter.Date)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []model.ContentItem
	offset := (page - 1) * pageSize
	if err := query.
		Order("scene_code ASC").
		Order("date DESC").
		Order("id DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&items).Error; err != nil {
		return nil, 0, err
	}

	return items, total, nil
}
