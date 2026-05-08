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

var ErrScenePageConfigNotFound = errors.New("scene page config not found")
var ErrDuplicateScenePageConfig = errors.New("duplicate scene page config")
var ErrInvalidScenePageConfigParams = errors.New("invalid scene page config params")

type ScenePageConfigService struct {
	repo repo.ScenePageConfigRepository
}

func NewScenePageConfigService(repo repo.ScenePageConfigRepository) *ScenePageConfigService {
	return &ScenePageConfigService{repo: repo}
}

func (s *ScenePageConfigService) List(ctx context.Context, req model.ScenePageConfigListRequest) (*model.ScenePageConfigListResponse, error) {
	filter, err := normalizeScenePageConfigFilter(req)
	if err != nil {
		return nil, err
	}

	items, total, err := s.repo.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("list scene page configs: %w", err)
	}

	list := make([]*model.ScenePageConfigResponse, 0, len(items))
	for i := range items {
		list = append(list, model.NewScenePageConfigResponse(&items[i]))
	}

	return &model.ScenePageConfigListResponse{
		List:     list,
		Total:    total,
		Page:     filter.Page,
		PageSize: filter.PageSize,
	}, nil
}

func (s *ScenePageConfigService) GetBySceneCode(ctx context.Context, sceneCode string) (*model.ScenePageConfigResponse, error) {
	normalized, err := model.ValidateSceneCode(sceneCode)
	if err != nil {
		return nil, ErrInvalidScenePageConfigParams
	}

	item, err := s.repo.GetBySceneCode(ctx, normalized)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrScenePageConfigNotFound
		}
		return nil, fmt.Errorf("get scene page config: %w", err)
	}

	return model.NewScenePageConfigResponse(item), nil
}

func (s *ScenePageConfigService) Create(ctx context.Context, req *model.ScenePageConfigUpsertRequest) (*model.ScenePageConfigResponse, error) {
	item, err := buildScenePageConfigModel(req)
	if err != nil {
		return nil, err
	}

	if err := s.repo.Create(ctx, item); err != nil {
		if isDuplicateScenePageConfigError(err) {
			return nil, ErrDuplicateScenePageConfig
		}
		return nil, fmt.Errorf("create scene page config: %w", err)
	}

	return model.NewScenePageConfigResponse(item), nil
}

func (s *ScenePageConfigService) Update(ctx context.Context, sceneCode string, req *model.ScenePageConfigUpsertRequest) (*model.ScenePageConfigResponse, error) {
	normalized, err := model.ValidateSceneCode(sceneCode)
	if err != nil {
		return nil, ErrInvalidScenePageConfigParams
	}

	item, err := buildScenePageConfigModel(req)
	if err != nil {
		return nil, err
	}

	if item.SceneCode != normalized {
		return nil, ErrInvalidScenePageConfigParams
	}

	if err := s.repo.UpdateBySceneCode(ctx, normalized, item); err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return nil, ErrScenePageConfigNotFound
		case isDuplicateScenePageConfigError(err):
			return nil, ErrDuplicateScenePageConfig
		default:
			return nil, fmt.Errorf("update scene page config: %w", err)
		}
	}

	return model.NewScenePageConfigResponse(item), nil
}

func (s *ScenePageConfigService) Delete(ctx context.Context, sceneCode string) (*model.ScenePageConfigResponse, error) {
	normalized, err := model.ValidateSceneCode(sceneCode)
	if err != nil {
		return nil, ErrInvalidScenePageConfigParams
	}

	item, err := s.repo.GetBySceneCode(ctx, normalized)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrScenePageConfigNotFound
		}
		return nil, fmt.Errorf("get scene page config before delete: %w", err)
	}

	if err := s.repo.DeleteBySceneCode(ctx, normalized); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrScenePageConfigNotFound
		}
		return nil, fmt.Errorf("delete scene page config: %w", err)
	}

	return model.NewScenePageConfigResponse(item), nil
}

func normalizeScenePageConfigFilter(req model.ScenePageConfigListRequest) (model.ScenePageConfigFilter, error) {
	page := req.Page
	if page <= 0 {
		page = 1
	}

	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		return model.ScenePageConfigFilter{}, ErrInvalidScenePageConfigParams
	}

	filter := model.ScenePageConfigFilter{
		Page:     page,
		PageSize: pageSize,
	}
	if req.Scene != "" {
		sceneCode, err := model.ValidateSceneCode(req.Scene)
		if err != nil {
			return model.ScenePageConfigFilter{}, ErrInvalidScenePageConfigParams
		}
		filter.Scene = sceneCode
	}

	return filter, nil
}

func buildScenePageConfigModel(req *model.ScenePageConfigUpsertRequest) (*model.ScenePageConfig, error) {
	if req == nil {
		return nil, ErrInvalidScenePageConfigParams
	}

	sceneCode, err := model.ValidateSceneCode(req.SceneCode)
	if err != nil {
		return nil, ErrInvalidScenePageConfigParams
	}
	if len(req.TagsDefault) == 0 {
		req.TagsDefault = []string{}
	}

	normalizedLogo, err := model.ValidateOptionalAssetURL(req.Logo)
	if err != nil {
		return nil, ErrInvalidScenePageConfigParams
	}
	normalizedBanner, err := model.ValidateOptionalAssetURL(req.Banner)
	if err != nil {
		return nil, ErrInvalidScenePageConfigParams
	}
	normalizedBacImg, err := model.ValidateOptionalAssetURL(req.BacImg)
	if err != nil {
		return nil, ErrInvalidScenePageConfigParams
	}
	normalizedDefaultBgURL, err := model.ValidateOptionalAssetURL(req.DefaultBgURL)
	if err != nil {
		return nil, ErrInvalidScenePageConfigParams
	}
	normalizedDefaultMusic, err := model.ValidateOptionalAssetURL(req.DefaultMusic)
	if err != nil {
		return nil, ErrInvalidScenePageConfigParams
	}

	playButtonColor, err := model.ValidateHexColor(req.PlayButtonColor)
	if err != nil {
		return nil, ErrInvalidScenePageConfigParams
	}
	textDefaultColor, err := model.ValidateHexColor(req.TextDefaultColor)
	if err != nil {
		return nil, ErrInvalidScenePageConfigParams
	}
	tagsColor, err := model.ValidateHexColor(req.TagsColor)
	if err != nil {
		return nil, ErrInvalidScenePageConfigParams
	}
	tagsBacColor, err := model.ValidateHexColor(req.TagsBacColor)
	if err != nil {
		return nil, ErrInvalidScenePageConfigParams
	}
	dateColor, err := model.ValidateHexColor(req.DateColor)
	if err != nil {
		return nil, ErrInvalidScenePageConfigParams
	}

	for _, tag := range req.TagsDefault {
		if strings.TrimSpace(tag) == "" {
			return nil, ErrInvalidScenePageConfigParams
		}
	}

	req.SceneCode = sceneCode
	req.Logo = normalizedLogo
	req.Banner = normalizedBanner
	req.BacImg = normalizedBacImg
	req.DefaultBgURL = normalizedDefaultBgURL
	req.DefaultMusic = normalizedDefaultMusic
	req.PlayButtonColor = playButtonColor
	req.TextDefaultColor = textDefaultColor
	req.TagsColor = tagsColor
	req.TagsBacColor = tagsBacColor
	req.DateColor = dateColor

	return &model.ScenePageConfig{
		SceneCode:        sceneCode,
		Logo:             normalizedLogo,
		Banner:           normalizedBanner,
		BacImg:           normalizedBacImg,
		DefaultBgURL:     normalizedDefaultBgURL,
		DefaultMusic:     normalizedDefaultMusic,
		TextDefault:      strings.TrimSpace(req.TextDefault),
		TagsDefault:      model.JSONStringArray(req.TagsDefault),
		PlayButtonColor:  playButtonColor,
		TextDefaultColor: textDefaultColor,
		TagsColor:        tagsColor,
		TagsBacColor:     tagsBacColor,
		DateColor:        dateColor,
	}, nil
}

func isDuplicateScenePageConfigError(err error) bool {
	var mysqlErr *gmysql.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062
}
