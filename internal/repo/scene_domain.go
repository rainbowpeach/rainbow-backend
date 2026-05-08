package repo

import (
	"context"

	"gorm.io/gorm"

	"rainbow-backend/internal/model"
)

type SceneDomainRepository interface {
	GetByHost(ctx context.Context, host string) (*model.SceneDomain, error)
	Create(ctx context.Context, item *model.SceneDomain) error
	UpdateByHost(ctx context.Context, currentHost string, item *model.SceneDomain) error
	DeleteByHost(ctx context.Context, host string) error
	List(ctx context.Context, filter model.SceneDomainFilter) ([]model.SceneDomain, int64, error)
}

type GormSceneDomainRepository struct {
	db *gorm.DB
}

func NewSceneDomainRepository(db *gorm.DB) *GormSceneDomainRepository {
	return &GormSceneDomainRepository{db: db}
}

func (r *GormSceneDomainRepository) GetByHost(ctx context.Context, host string) (*model.SceneDomain, error) {
	var item model.SceneDomain
	if err := r.db.WithContext(ctx).Where("host = ?", host).First(&item).Error; err != nil {
		return nil, err
	}

	return &item, nil
}

func (r *GormSceneDomainRepository) Create(ctx context.Context, item *model.SceneDomain) error {
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *GormSceneDomainRepository) UpdateByHost(ctx context.Context, currentHost string, item *model.SceneDomain) error {
	result := r.db.WithContext(ctx).
		Model(&model.SceneDomain{}).
		Where("host = ?", currentHost).
		Updates(map[string]any{
			"host":       item.Host,
			"scene_code": item.SceneCode,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *GormSceneDomainRepository) DeleteByHost(ctx context.Context, host string) error {
	result := r.db.WithContext(ctx).Where("host = ?", host).Delete(&model.SceneDomain{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *GormSceneDomainRepository) List(ctx context.Context, filter model.SceneDomainFilter) ([]model.SceneDomain, int64, error) {
	query := r.db.WithContext(ctx).Model(&model.SceneDomain{})
	if filter.Host != "" {
		query = query.Where("host LIKE ?", "%"+filter.Host+"%")
	}
	if filter.Scene != "" {
		query = query.Where("scene_code = ?", filter.Scene)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []model.SceneDomain
	offset := (filter.Page - 1) * filter.PageSize
	if err := query.
		Order("host ASC").
		Offset(offset).
		Limit(filter.PageSize).
		Find(&items).Error; err != nil {
		return nil, 0, err
	}

	return items, total, nil
}
