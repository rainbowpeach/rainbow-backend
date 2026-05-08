package service

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"rainbow-backend/internal/model"
	"rainbow-backend/internal/repo"
)

var ErrSceneNotConfigured = errors.New("scene not configured")

type ResolvedScene struct {
	Host      string
	SceneCode string
}

type SceneResolver struct {
	sceneDomainRepo repo.SceneDomainRepository
}

func NewSceneResolver(sceneDomainRepo repo.SceneDomainRepository) *SceneResolver {
	return &SceneResolver{sceneDomainRepo: sceneDomainRepo}
}

func (r *SceneResolver) ResolveHost(ctx context.Context, host string) (*ResolvedScene, error) {
	normalizedHost := model.NormalizeHost(host)
	if normalizedHost == "" {
		return nil, ErrSceneNotConfigured
	}

	item, err := r.sceneDomainRepo.GetByHost(ctx, normalizedHost)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSceneNotConfigured
		}
		return nil, fmt.Errorf("resolve host %s: %w", normalizedHost, err)
	}

	return &ResolvedScene{
		Host:      item.Host,
		SceneCode: item.SceneCode,
	}, nil
}
