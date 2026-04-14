package repo

import (
	"context"

	"gorm.io/gorm"

	"rainbow-backend/internal/model"
)

type AdminRepository interface {
	GetByUsername(ctx context.Context, username string) (*model.Admin, error)
}

type GormAdminRepository struct {
	db *gorm.DB
}

func NewAdminRepository(db *gorm.DB) *GormAdminRepository {
	return &GormAdminRepository{db: db}
}

func (r *GormAdminRepository) GetByUsername(ctx context.Context, username string) (*model.Admin, error) {
	var admin model.Admin
	if err := r.db.WithContext(ctx).Where("username = ?", username).First(&admin).Error; err != nil {
		return nil, err
	}

	return &admin, nil
}
