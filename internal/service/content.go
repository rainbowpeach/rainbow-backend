package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	gmysql "github.com/go-sql-driver/mysql"
	"gorm.io/gorm"

	"rainbow-backend/internal/model"
	"rainbow-backend/internal/repo"
)

var ErrInvalidDateFormat = errors.New("invalid date format")
var ErrContentNotFound = errors.New("content not found")
var ErrDuplicateDate = errors.New("duplicate date")
var ErrInvalidContentParams = errors.New("invalid content params")

type ContentService struct {
	contentRepo repo.ContentRepository
}

func NewContentService(contentRepo repo.ContentRepository) *ContentService {
	return &ContentService{
		contentRepo: contentRepo,
	}
}

func (s *ContentService) GetBySceneAndDate(ctx context.Context, sceneCode, date string) (*model.ContentResponse, error) {
	normalizedSceneCode, err := model.ValidateSceneCode(sceneCode)
	if err != nil {
		return nil, ErrInvalidContentParams
	}
	date = strings.TrimSpace(date)
	if _, err := time.Parse("2006-01-02", date); err != nil {
		return nil, ErrInvalidDateFormat
	}

	item, err := s.contentRepo.GetBySceneAndDate(ctx, normalizedSceneCode, date)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrContentNotFound
		}
		return nil, fmt.Errorf("get content by date: %w", err)
	}

	return model.NewContentResponse(item), nil
}

func (s *ContentService) Create(ctx context.Context, req *model.ContentUpsertRequest) (*model.IDResponse, error) {
	if err := validateContentUpsert(req); err != nil {
		return nil, err
	}

	item := &model.ContentItem{
		SceneCode: model.NormalizeSceneCode(req.SceneCode),
		Date:      req.Date,
		Text:      req.Text,
		Tags:      model.JSONStringArray(req.Tags),
		BgURL:     req.BgURL,
		Music:     req.Music,
	}
	if err := s.contentRepo.Create(ctx, item); err != nil {
		if isDuplicateDateError(err) {
			return nil, ErrDuplicateDate
		}
		return nil, fmt.Errorf("create content: %w", err)
	}

	return &model.IDResponse{ID: item.ID}, nil
}

func (s *ContentService) Update(ctx context.Context, id uint, req *model.ContentUpsertRequest) (*model.IDResponse, error) {
	if err := validateContentUpsert(req); err != nil {
		return nil, err
	}

	item := &model.ContentItem{
		SceneCode: model.NormalizeSceneCode(req.SceneCode),
		Date:      req.Date,
		Text:      req.Text,
		Tags:      model.JSONStringArray(req.Tags),
		BgURL:     req.BgURL,
		Music:     req.Music,
	}
	if err := s.contentRepo.UpdateByID(ctx, id, item); err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return nil, ErrContentNotFound
		case isDuplicateDateError(err):
			return nil, ErrDuplicateDate
		default:
			return nil, fmt.Errorf("update content: %w", err)
		}
	}

	return &model.IDResponse{ID: id}, nil
}

func (s *ContentService) Delete(ctx context.Context, id uint) (*model.IDResponse, error) {
	if err := s.contentRepo.DeleteByID(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrContentNotFound
		}
		return nil, fmt.Errorf("delete content: %w", err)
	}

	return &model.IDResponse{ID: id}, nil
}

func (s *ContentService) List(ctx context.Context, filter model.ContentFilter, page, pageSize int) (*model.ContentListResponse, error) {
	if filter.Date != "" {
		filter.Date = strings.TrimSpace(filter.Date)
		if err := validateDate(filter.Date); err != nil {
			return nil, err
		}
	}
	if filter.SceneCode != "" {
		normalizedSceneCode, err := model.ValidateSceneCode(filter.SceneCode)
		if err != nil {
			return nil, ErrInvalidContentParams
		}
		filter.SceneCode = normalizedSceneCode
	}

	items, total, err := s.contentRepo.List(ctx, filter, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("list content: %w", err)
	}

	list := make([]*model.ContentResponse, 0, len(items))
	for i := range items {
		list = append(list, model.NewContentResponse(&items[i]))
	}

	return &model.ContentListResponse{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func validateDate(date string) error {
	date = strings.TrimSpace(date)
	if _, err := time.Parse("2006-01-02", date); err != nil {
		return ErrInvalidDateFormat
	}
	return nil
}

func validateContentUpsert(req *model.ContentUpsertRequest) error {
	if req == nil {
		return ErrInvalidContentParams
	}
	sceneCode, err := model.ValidateSceneCode(req.SceneCode)
	if err != nil {
		return ErrInvalidContentParams
	}
	req.SceneCode = sceneCode
	req.Date = strings.TrimSpace(req.Date)
	if err := validateDate(req.Date); err != nil {
		return err
	}
	for i, tag := range req.Tags {
		normalizedTag := strings.TrimSpace(tag)
		if normalizedTag == "" {
			return ErrInvalidContentParams
		}
		req.Tags[i] = normalizedTag
	}
	req.Text = strings.TrimSpace(req.Text)
	req.BgURL = strings.TrimSpace(req.BgURL)
	req.Music = strings.TrimSpace(req.Music)
	return nil
}

func isDuplicateDateError(err error) bool {
	var mysqlErr *gmysql.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062
}
