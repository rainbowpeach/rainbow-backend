package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	gmysql "github.com/go-sql-driver/mysql"
	"gorm.io/gorm"

	"rainbow-backend/internal/model"
	"rainbow-backend/internal/repo"
)

var ErrInvalidDateFormat = errors.New("invalid date format")
var ErrContentNotFound = errors.New("content not found")
var ErrDuplicateDate = errors.New("duplicate date")

type ContentService struct {
	contentRepo repo.ContentRepository
}

func NewContentService(contentRepo repo.ContentRepository) *ContentService {
	return &ContentService{
		contentRepo: contentRepo,
	}
}

func (s *ContentService) GetByDate(ctx context.Context, date string) (*model.ContentResponse, error) {
	if _, err := time.Parse("2006-01-02", date); err != nil {
		return nil, ErrInvalidDateFormat
	}

	item, err := s.contentRepo.GetByDate(ctx, date)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrContentNotFound
		}
		return nil, fmt.Errorf("get content by date: %w", err)
	}

	return model.NewContentResponse(item), nil
}

func (s *ContentService) Create(ctx context.Context, req *model.ContentUpsertRequest) (*model.IDResponse, error) {
	if err := validateDate(req.Date); err != nil {
		return nil, err
	}

	item := &model.ContentItem{
		Date:  req.Date,
		Text:  req.Text,
		Tags:  model.JSONStringArray(req.Tags),
		BgURL: req.BgURL,
		Music: req.Music,
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
	if err := validateDate(req.Date); err != nil {
		return nil, err
	}

	item := &model.ContentItem{
		Date:  req.Date,
		Text:  req.Text,
		Tags:  model.JSONStringArray(req.Tags),
		BgURL: req.BgURL,
		Music: req.Music,
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

func (s *ContentService) List(ctx context.Context, page, pageSize int) (*model.ContentListResponse, error) {
	items, total, err := s.contentRepo.List(ctx, page, pageSize)
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
	if _, err := time.Parse("2006-01-02", date); err != nil {
		return ErrInvalidDateFormat
	}
	return nil
}

func isDuplicateDateError(err error) bool {
	var mysqlErr *gmysql.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062
}
