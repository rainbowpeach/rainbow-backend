package service

import (
	"context"
	"errors"
	"testing"

	"gorm.io/gorm"

	"rainbow-backend/internal/model"
)

type stubSceneDomainResolverRepo struct {
	byHost    *model.SceneDomain
	byHostErr error
}

func (r *stubSceneDomainResolverRepo) GetByHost(context.Context, string) (*model.SceneDomain, error) {
	if r.byHostErr != nil {
		return nil, r.byHostErr
	}
	if r.byHost == nil {
		return nil, gorm.ErrRecordNotFound
	}
	return r.byHost, nil
}

func (r *stubSceneDomainResolverRepo) Create(context.Context, *model.SceneDomain) error {
	return nil
}

func (r *stubSceneDomainResolverRepo) UpdateByHost(context.Context, string, *model.SceneDomain) error {
	return nil
}

func (r *stubSceneDomainResolverRepo) DeleteByHost(context.Context, string) error {
	return nil
}

func (r *stubSceneDomainResolverRepo) List(context.Context, model.SceneDomainFilter) ([]model.SceneDomain, int64, error) {
	return nil, 0, nil
}

func TestSceneResolverExactHostMatch(t *testing.T) {
	resolver := NewSceneResolver(&stubSceneDomainResolverRepo{
		byHost: &model.SceneDomain{
			Host:      "love.example.com",
			SceneCode: "love",
		},
	})

	result, err := resolver.ResolveHost(context.Background(), "love.example.com:80")
	if err != nil {
		t.Fatalf("ResolveHost() error = %v", err)
	}

	if result.SceneCode != "love" {
		t.Fatalf("expected scene love, got %q", result.SceneCode)
	}
	if result.Host != "love.example.com" {
		t.Fatalf("expected normalized host love.example.com, got %q", result.Host)
	}
}

func TestSceneResolverRequiresConfiguredHost(t *testing.T) {
	resolver := NewSceneResolver(&stubSceneDomainResolverRepo{})

	_, err := resolver.ResolveHost(context.Background(), "unknown.example.com")
	if !errors.Is(err, ErrSceneNotConfigured) {
		t.Fatalf("expected ErrSceneNotConfigured, got %v", err)
	}
}

func TestSceneResolverReturnsRepositoryError(t *testing.T) {
	resolver := NewSceneResolver(&stubSceneDomainResolverRepo{
		byHostErr: errors.New("db down"),
	})

	_, err := resolver.ResolveHost(context.Background(), "love.example.com")
	if err == nil {
		t.Fatal("expected error")
	}
}
