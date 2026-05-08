package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	gmysql "github.com/go-sql-driver/mysql"
	"gorm.io/gorm"

	"rainbow-backend/internal/model"
	"rainbow-backend/internal/repo"
)

var ErrSceneDomainNotFound = errors.New("scene domain not found")
var ErrDuplicateHost = errors.New("duplicate host")

type SceneDomainService struct {
	sceneDomainRepo repo.SceneDomainRepository
}

func NewSceneDomainService(sceneDomainRepo repo.SceneDomainRepository) *SceneDomainService {
	return &SceneDomainService{
		sceneDomainRepo: sceneDomainRepo,
	}
}

func (s *SceneDomainService) List(ctx context.Context, req model.SceneDomainListRequest) (*model.SceneDomainListResponse, error) {
	filter, err := normalizeSceneDomainFilter(req)
	if err != nil {
		return nil, err
	}

	items, total, err := s.sceneDomainRepo.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("list scene domains: %w", err)
	}

	list := make([]*model.SceneDomainResponse, 0, len(items))
	for i := range items {
		list = append(list, model.NewSceneDomainResponse(&items[i]))
	}

	return &model.SceneDomainListResponse{
		List:     list,
		Total:    total,
		Page:     filter.Page,
		PageSize: filter.PageSize,
	}, nil
}

func (s *SceneDomainService) GetByHost(ctx context.Context, host string) (*model.SceneDomainResponse, error) {
	normalizedHost, err := model.ValidateHost(host)
	if err != nil {
		return nil, ErrInvalidContentParams
	}

	item, err := s.sceneDomainRepo.GetByHost(ctx, normalizedHost)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSceneDomainNotFound
		}
		return nil, fmt.Errorf("get scene domain: %w", err)
	}

	return model.NewSceneDomainResponse(item), nil
}

func (s *SceneDomainService) Create(ctx context.Context, req *model.SceneDomainUpsertRequest) (*model.SceneDomainResponse, error) {
	item, err := buildSceneDomainModel(req)
	if err != nil {
		return nil, err
	}

	if err := s.sceneDomainRepo.Create(ctx, item); err != nil {
		if isDuplicateSceneDomainError(err) {
			return nil, ErrDuplicateHost
		}
		return nil, fmt.Errorf("create scene domain: %w", err)
	}

	return model.NewSceneDomainResponse(item), nil
}

func (s *SceneDomainService) Update(ctx context.Context, host string, req *model.SceneDomainUpsertRequest) (*model.SceneDomainResponse, error) {
	normalizedHost, err := model.ValidateHost(host)
	if err != nil {
		return nil, ErrInvalidContentParams
	}

	item, err := buildSceneDomainModel(req)
	if err != nil {
		return nil, err
	}

	if err := s.sceneDomainRepo.UpdateByHost(ctx, normalizedHost, item); err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return nil, ErrSceneDomainNotFound
		case isDuplicateSceneDomainError(err):
			return nil, ErrDuplicateHost
		default:
			return nil, fmt.Errorf("update scene domain: %w", err)
		}
	}

	return model.NewSceneDomainResponse(item), nil
}

func (s *SceneDomainService) Delete(ctx context.Context, host string) (*model.SceneDomainResponse, error) {
	normalizedHost, err := model.ValidateHost(host)
	if err != nil {
		return nil, ErrInvalidContentParams
	}

	item, err := s.sceneDomainRepo.GetByHost(ctx, normalizedHost)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSceneDomainNotFound
		}
		return nil, fmt.Errorf("get scene domain before delete: %w", err)
	}

	if err := s.sceneDomainRepo.DeleteByHost(ctx, normalizedHost); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSceneDomainNotFound
		}
		return nil, fmt.Errorf("delete scene domain: %w", err)
	}

	return model.NewSceneDomainResponse(item), nil
}

func buildSceneDomainModel(req *model.SceneDomainUpsertRequest) (*model.SceneDomain, error) {
	if req == nil {
		return nil, ErrInvalidContentParams
	}

	host, err := model.ValidateHost(req.Host)
	if err != nil {
		return nil, ErrInvalidContentParams
	}
	sceneCode, err := model.ValidateSceneCode(req.SceneCode)
	if err != nil {
		return nil, ErrInvalidContentParams
	}

	req.Host = host
	req.SceneCode = sceneCode

	return &model.SceneDomain{
		Host:      host,
		SceneCode: sceneCode,
	}, nil
}

func normalizeSceneDomainFilter(req model.SceneDomainListRequest) (model.SceneDomainFilter, error) {
	page := req.Page
	if page <= 0 {
		page = 1
	}

	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		return model.SceneDomainFilter{}, ErrInvalidContentParams
	}

	filter := model.SceneDomainFilter{
		Page:     page,
		PageSize: pageSize,
		Host:     strings.ToLower(strings.TrimSpace(req.Host)),
	}
	if req.Scene != "" {
		sceneCode, err := model.ValidateSceneCode(req.Scene)
		if err != nil {
			return model.SceneDomainFilter{}, ErrInvalidContentParams
		}
		filter.Scene = sceneCode
	}

	return filter, nil
}

func isDuplicateSceneDomainError(err error) bool {
	var mysqlErr *gmysql.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062
}
