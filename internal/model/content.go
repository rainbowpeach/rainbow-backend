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
