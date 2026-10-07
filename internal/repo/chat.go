package repo

import (
	"context"
	"strings"

	"gorm.io/gorm"

	"rainbow-backend/internal/model"
)

type ChatPersonaRepository interface {
	GetBySceneCode(ctx context.Context, sceneCode string) (*model.ChatPersona, error)
}

type ChatLogRepository interface {
	Create(ctx context.Context, logItem *model.ChatLog) error
	List(ctx context.Context, filter model.ChatLogFilter) ([]model.ChatLog, int64, error)
}

type GormChatPersonaRepository struct {
	db *gorm.DB
}

func NewChatPersonaRepository(db *gorm.DB) *GormChatPersonaRepository {
	return &GormChatPersonaRepository{db: db}
}

func (r *GormChatPersonaRepository) GetBySceneCode(ctx context.Context, sceneCode string) (*model.ChatPersona, error) {
	var persona model.ChatPersona
	if err := r.db.WithContext(ctx).Where("scene_code = ?", sceneCode).First(&persona).Error; err != nil {
		return nil, err
	}

	return &persona, nil
}

type GormChatLogRepository struct {
	db *gorm.DB
}

func NewChatLogRepository(db *gorm.DB) *GormChatLogRepository {
	return &GormChatLogRepository{db: db}
}

func (r *GormChatLogRepository) Create(ctx context.Context, logItem *model.ChatLog) error {
	return r.db.WithContext(ctx).Create(logItem).Error
}

func (r *GormChatLogRepository) List(ctx context.Context, filter model.ChatLogFilter) ([]model.ChatLog, int64, error) {
	query := r.db.WithContext(ctx).Model(&model.ChatLog{})
	if filter.SceneCode != "" {
		query = query.Where("scene_code = ?", filter.SceneCode)
	}
	if filter.UserID > 0 {
		query = query.Where("user_id = ?", filter.UserID)
	}
	if keyword := strings.TrimSpace(filter.Keyword); keyword != "" {
		escaped := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(keyword)
		query = query.Where("user_text LIKE ?", "%"+escaped+"%")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []model.ChatLog
	offset := (filter.Page - 1) * filter.PageSize
	if err := query.Order("id DESC").Offset(offset).Limit(filter.PageSize).Find(&items).Error; err != nil {
		return nil, 0, err
	}

	return items, total, nil
}
