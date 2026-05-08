package model

import (
	"errors"
	"net/url"
	"regexp"
	"strings"
	"time"
)

var hexColorPattern = regexp.MustCompile(`^#(?:[0-9a-fA-F]{3}|[0-9a-fA-F]{6})$`)

type ScenePageConfig struct {
	SceneCode        string          `gorm:"column:scene_code;size:128;primaryKey" json:"scene_code"`
	Logo             string          `gorm:"size:1024" json:"logo"`
	Banner           string          `gorm:"size:1024" json:"banner"`
	BacImg           string          `gorm:"column:bac_img;size:1024" json:"bac_img"`
	DefaultBgURL     string          `gorm:"column:default_bg_url;size:1024" json:"default_bg_url"`
	DefaultMusic     string          `gorm:"column:default_music;size:1024" json:"default_music"`
	TextDefault      string          `gorm:"column:text_default;type:text" json:"text_default"`
	TagsDefault      JSONStringArray `gorm:"column:tags_default;type:json;not null" json:"tags_default"`
	PlayButtonColor  string          `gorm:"column:play_button_color;size:32" json:"play_button_color"`
	TextDefaultColor string          `gorm:"column:text_default_color;size:32" json:"text_default_color"`
	TagsColor        string          `gorm:"column:tags_color;size:32" json:"tags_color"`
	TagsBacColor     string          `gorm:"column:tags_bac_color;size:32" json:"tags_bac_color"`
	DateColor        string          `gorm:"column:date_color;size:32" json:"date_color"`
	CreatedAt        time.Time       `json:"createdAt"`
	UpdatedAt        time.Time       `json:"updatedAt"`
}

func (ScenePageConfig) TableName() string {
	return "scene_page_configs"
}

type ScenePageConfigUpsertRequest struct {
	SceneCode        string   `json:"scene_code"`
	Logo             string   `json:"logo"`
	Banner           string   `json:"banner"`
	BacImg           string   `json:"bac_img"`
	DefaultBgURL     string   `json:"default_bg_url"`
	DefaultMusic     string   `json:"default_music"`
	TextDefault      string   `json:"text_default"`
	TagsDefault      []string `json:"tags_default"`
	PlayButtonColor  string   `json:"play_button_color"`
	TextDefaultColor string   `json:"text_default_color"`
	TagsColor        string   `json:"tags_color"`
	TagsBacColor     string   `json:"tags_bac_color"`
	DateColor        string   `json:"date_color"`
}

type ScenePageConfigResponse struct {
	SceneCode        string   `json:"scene_code"`
	Logo             string   `json:"logo"`
	Banner           string   `json:"banner"`
	BacImg           string   `json:"bac_img"`
	DefaultBgURL     string   `json:"default_bg_url"`
	DefaultMusic     string   `json:"default_music"`
	TextDefault      string   `json:"text_default"`
	TagsDefault      []string `json:"tags_default"`
	PlayButtonColor  string   `json:"play_button_color"`
	TextDefaultColor string   `json:"text_default_color"`
	TagsColor        string   `json:"tags_color"`
	TagsBacColor     string   `json:"tags_bac_color"`
	DateColor        string   `json:"date_color"`
}

func NewScenePageConfigResponse(item *ScenePageConfig) *ScenePageConfigResponse {
	if item == nil {
		return nil
	}

	tags := []string(item.TagsDefault)
	if tags == nil {
		tags = []string{}
	}

	return &ScenePageConfigResponse{
		SceneCode:        item.SceneCode,
		Logo:             item.Logo,
		Banner:           item.Banner,
		BacImg:           item.BacImg,
		DefaultBgURL:     item.DefaultBgURL,
		DefaultMusic:     item.DefaultMusic,
		TextDefault:      item.TextDefault,
		TagsDefault:      tags,
		PlayButtonColor:  item.PlayButtonColor,
		TextDefaultColor: item.TextDefaultColor,
		TagsColor:        item.TagsColor,
		TagsBacColor:     item.TagsBacColor,
		DateColor:        item.DateColor,
	}
}

type ScenePageConfigListRequest struct {
	Page     int    `form:"page"`
	PageSize int    `form:"pageSize"`
	Scene    string `form:"scene"`
}

type ScenePageConfigFilter struct {
	Page     int
	PageSize int
	Scene    string
}

type ScenePageConfigListResponse struct {
	List     []*ScenePageConfigResponse `json:"list"`
	Total    int64                      `json:"total"`
	Page     int                        `json:"page"`
	PageSize int                        `json:"pageSize"`
}

func ValidateHexColor(value string) (string, error) {
	normalized := strings.TrimSpace(value)
	if normalized == "" {
		return "", nil
	}
	if !hexColorPattern.MatchString(normalized) {
		return "", errors.New("invalid color")
	}

	return strings.ToLower(normalized), nil
}

func ValidateOptionalAssetURL(value string) (string, error) {
	normalized := strings.TrimSpace(value)
	if normalized == "" {
		return "", nil
	}
	if strings.ContainsAny(normalized, "\r\n\t") {
		return "", errors.New("invalid asset url")
	}
	if strings.HasPrefix(normalized, "/") {
		return normalized, nil
	}

	parsed, err := url.Parse(normalized)
	if err != nil {
		return "", errors.New("invalid asset url")
	}
	if (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		return "", errors.New("invalid asset url")
	}

	return normalized, nil
}
