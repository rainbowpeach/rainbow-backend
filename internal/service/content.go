package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"rainbow-backend/internal/model"
	"rainbow-backend/internal/repo"
)

var ErrInvalidDateFormat = errors.New("invalid date format")
var ErrContentNotFound = errors.New("content not found")

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
