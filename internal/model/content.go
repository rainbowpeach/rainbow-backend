package model

import "time"

type ContentItem struct {
	ID        uint            `gorm:"primaryKey" json:"id"`
	SceneCode string          `gorm:"column:scene_code;size:64;not null;index;uniqueIndex:idx_content_scene_date" json:"scene_code"`
	Date      string          `gorm:"size:10;not null;uniqueIndex:idx_content_scene_date" json:"date"`
	Text      string          `gorm:"type:text" json:"text"`
	Tags      JSONStringArray `gorm:"type:json" json:"tags"`
	BgURL     string          `gorm:"column:bg_url;type:text" json:"bg_url"`
	Music     string          `gorm:"type:text" json:"music"`
	CreatedAt time.Time       `json:"createdAt"`
	UpdatedAt time.Time       `json:"updatedAt"`
}

func (ContentItem) TableName() string {
	return "content_items"
}

type ContentUpsertRequest struct {
	SceneCode string   `json:"scene_code" binding:"required"`
	Date      string   `json:"date" binding:"required"`
	Text      string   `json:"text"`
	Tags      []string `json:"tags"`
	BgURL     string   `json:"bg_url"`
	Music     string   `json:"music"`
}

type ContentResponse struct {
	ID        uint     `json:"id"`
	SceneCode string   `json:"scene_code"`
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

	tags := []string(item.Tags)
	if tags == nil {
		tags = []string{}
	}

	return &ContentResponse{
		ID:        item.ID,
		SceneCode: item.SceneCode,
		Date:      item.Date,
		Text:      item.Text,
		Tags:      tags,
		BgURL:     item.BgURL,
		Music:     item.Music,
		CreatedAt: item.CreatedAt.Format("2006-01-02"),
		UpdatedAt: item.UpdatedAt.Format("2006-01-02"),
	}
}

type ContentListRequest struct {
	Page     int    `form:"page" binding:"required,min=1"`
	PageSize int    `form:"pageSize" binding:"required,min=1,max=100"`
	Scene    string `form:"scene"`
	Date     string `form:"date"`
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
