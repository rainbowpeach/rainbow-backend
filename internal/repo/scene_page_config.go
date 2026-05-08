package repo

import (
	"context"

	"gorm.io/gorm"

	"rainbow-backend/internal/model"
)

type ScenePageConfigRepository interface {
	GetBySceneCode(ctx context.Context, sceneCode string) (*model.ScenePageConfig, error)
	Create(ctx context.Context, item *model.ScenePageConfig) error
	UpdateBySceneCode(ctx context.Context, sceneCode string, item *model.ScenePageConfig) error
	DeleteBySceneCode(ctx context.Context, sceneCode string) error
	List(ctx context.Context, filter model.ScenePageConfigFilter) ([]model.ScenePageConfig, int64, error)
}

type GormScenePageConfigRepository struct {
	db *gorm.DB
}

func NewScenePageConfigRepository(db *gorm.DB) *GormScenePageConfigRepository {
	return &GormScenePageConfigRepository{db: db}
}

func (r *GormScenePageConfigRepository) GetBySceneCode(ctx context.Context, sceneCode string) (*model.ScenePageConfig, error) {
	var item model.ScenePageConfig
	if err := r.db.WithContext(ctx).Where("scene_code = ?", sceneCode).First(&item).Error; err != nil {
		return nil, err
	}

	return &item, nil
}

func (r *GormScenePageConfigRepository) Create(ctx context.Context, item *model.ScenePageConfig) error {
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *GormScenePageConfigRepository) UpdateBySceneCode(ctx context.Context, sceneCode string, item *model.ScenePageConfig) error {
	result := r.db.WithContext(ctx).
		Model(&model.ScenePageConfig{}).
		Where("scene_code = ?", sceneCode).
		Updates(map[string]any{
			"logo":               item.Logo,
			"banner":             item.Banner,
			"bac_img":            item.BacImg,
			"default_bg_url":     item.DefaultBgURL,
			"default_music":      item.DefaultMusic,
			"text_default":       item.TextDefault,
			"tags_default":       item.TagsDefault,
			"play_button_color":  item.PlayButtonColor,
			"text_default_color": item.TextDefaultColor,
			"tags_color":         item.TagsColor,
			"tags_bac_color":     item.TagsBacColor,
			"date_color":         item.DateColor,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *GormScenePageConfigRepository) DeleteBySceneCode(ctx context.Context, sceneCode string) error {
	result := r.db.WithContext(ctx).Where("scene_code = ?", sceneCode).Delete(&model.ScenePageConfig{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *GormScenePageConfigRepository) List(ctx context.Context, filter model.ScenePageConfigFilter) ([]model.ScenePageConfig, int64, error) {
	query := r.db.WithContext(ctx).Model(&model.ScenePageConfig{})
	if filter.Scene != "" {
		query = query.Where("scene_code = ?", filter.Scene)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []model.ScenePageConfig
	offset := (filter.Page - 1) * filter.PageSize
	if err := query.
		Order("scene_code ASC").
		Offset(offset).
		Limit(filter.PageSize).
		Find(&items).Error; err != nil {
		return nil, 0, err
	}

	return items, total, nil
}
