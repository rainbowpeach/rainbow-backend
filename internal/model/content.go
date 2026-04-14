package model

import "time"

type ContentItem struct {
	ID        uint            `gorm:"primaryKey" json:"id"`
	Date      string          `gorm:"size:10;uniqueIndex;not null" json:"date"`
	Text      string          `gorm:"type:text;not null" json:"text"`
	Tags      JSONStringArray `gorm:"type:json;not null" json:"tags"`
	BgURL     string          `gorm:"column:bg_url;type:text;not null" json:"bg_url"`
	Music     string          `gorm:"type:text;not null" json:"music"`
	CreatedAt time.Time       `json:"createdAt"`
	UpdatedAt time.Time       `json:"updatedAt"`
}

func (ContentItem) TableName() string {
	return "content_items"
}

type ContentUpsertRequest struct {
	Date  string   `json:"date" binding:"required"`
	Text  string   `json:"text" binding:"required"`
	Tags  []string `json:"tags" binding:"required,min=1,dive,required"`
	BgURL string   `json:"bg_url" binding:"required"`
	Music string   `json:"music" binding:"required"`
}

type ContentResponse struct {
	ID        uint     `json:"id"`
	Date      string   `json:"date"`
	Text      string   `json:"text"`
	Tags      []string `json:"tags"`
	BgURL     string   `json:"bg_url"`
	Music     string   `json:"music"`
	CreatedAt string   `json:"createdAt"`
	UpdatedAt string   `json:"updatedAt"`
}

func NewContentResponse(item *ContentItem) *ContentResponse {
	if item == nil {
		return nil
	}

	return &ContentResponse{
		ID:        item.ID,
		Date:      item.Date,
		Text:      item.Text,
		Tags:      []string(item.Tags),
		BgURL:     item.BgURL,
		Music:     item.Music,
		CreatedAt: item.CreatedAt.Format("2006-01-02"),
		UpdatedAt: item.UpdatedAt.Format("2006-01-02"),
	}
}

type ContentListRequest struct {
	Page     int `form:"page" binding:"required,min=1"`
	PageSize int `form:"pageSize" binding:"required,min=1,max=100"`
}

type ContentListResponse struct {
	List     []*ContentResponse `json:"list"`
	Total    int64              `json:"total"`
	Page     int                `json:"page"`
	PageSize int                `json:"pageSize"`
}

type IDResponse struct {
	ID uint `json:"id"`
}
