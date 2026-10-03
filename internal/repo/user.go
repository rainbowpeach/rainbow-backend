package repo

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"rainbow-backend/internal/model"
)

type UserRepository interface {
	GetOrCreateByOpenID(ctx context.Context, openID string) (*model.User, error)
	UpdateLogin(ctx context.Context, userID uint, token string, tokenExpireAt, lastLoginAt time.Time) error
	GetByToken(ctx context.Context, token string) (*model.User, error)
}

type UserProfileRepository interface {
	GetByID(ctx context.Context, userID uint) (*model.User, error)
	UpdateProfile(ctx context.Context, userID uint, updates map[string]any) error
	ClearProfile(ctx context.Context, userID uint) error
}

type GormUserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *GormUserRepository {
	return &GormUserRepository{db: db}
}

func (r *GormUserRepository) GetOrCreateByOpenID(ctx context.Context, openID string) (*model.User, error) {
	if err := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "openid"}},
			DoNothing: true,
		}).
		Create(&model.User{OpenID: openID}).Error; err != nil {
		return nil, err
	}

	var user model.User
	if err := r.db.WithContext(ctx).Where("openid = ?", openID).First(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *GormUserRepository) UpdateLogin(ctx context.Context, userID uint, token string, tokenExpireAt, lastLoginAt time.Time) error {
	return r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", userID).Updates(map[string]any{
		"token":           token,
		"token_expire_at": tokenExpireAt,
		"last_login_at":   lastLoginAt,
	}).Error
}

func (r *GormUserRepository) GetByToken(ctx context.Context, token string) (*model.User, error) {
	var user model.User
	if err := r.db.WithContext(ctx).Where("token = ?", token).First(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *GormUserRepository) GetByID(ctx context.Context, userID uint) (*model.User, error) {
	var user model.User
	if err := r.db.WithContext(ctx).First(&user, userID).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *GormUserRepository) UpdateProfile(ctx context.Context, userID uint, updates map[string]any) error {
	result := r.db.WithContext(ctx).
		Model(&model.User{}).
		Where("id = ?", userID).
		Updates(updates)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (r *GormUserRepository) ClearProfile(ctx context.Context, userID uint) error {
	return r.UpdateProfile(ctx, userID, map[string]any{
		"scene_code": model.JSONStringArray{},
		"nickname":   nil,
		"avatar_url": nil,
		"birthday":   nil,
		"gender":     nil,
		"occupation": nil,
	})
}
