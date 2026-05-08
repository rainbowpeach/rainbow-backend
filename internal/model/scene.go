package model

import (
	"errors"
	"net"
	"regexp"
	"strings"
)

var (
	sceneCodePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{0,63}$`)
	hostPattern      = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?(?:\.[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?)*$`)
)

type SceneDomain struct {
	Host      string `gorm:"size:191;not null;uniqueIndex" json:"host"`
	SceneCode string `gorm:"column:scene_code;size:64;not null;index" json:"scene_code"`
}

func (SceneDomain) TableName() string {
	return "scene_domains"
}

type SceneDomainUpsertRequest struct {
	Host      string `json:"host" binding:"required"`
	SceneCode string `json:"scene_code" binding:"required"`
}

type SceneDomainUpdateRequest struct {
	Host      string `json:"host"`
	SceneCode string `json:"scene_code" binding:"required"`
}

type SceneDomainResponse struct {
	Host      string `json:"host"`
	SceneCode string `json:"scene_code"`
}

func NewSceneDomainResponse(item *SceneDomain) *SceneDomainResponse {
	if item == nil {
		return nil
	}

	return &SceneDomainResponse{
		Host:      item.Host,
		SceneCode: item.SceneCode,
	}
}

type SceneDomainListRequest struct {
	Page     int    `form:"page"`
	PageSize int    `form:"pageSize"`
	Host     string `form:"host"`
	Scene    string `form:"scene"`
}

type SceneDomainListResponse struct {
	List     []*SceneDomainResponse `json:"list"`
	Total    int64                  `json:"total"`
	Page     int                    `json:"page"`
	PageSize int                    `json:"pageSize"`
}

type PublicSceneDomainMappingResponse struct {
	Host      string `json:"host"`
	SceneCode string `json:"scene_code"`
}

type ContentFilter struct {
	SceneCode string
	Date      string
}

type SceneDomainFilter struct {
	Page     int
	PageSize int
	Host     string
	Scene    string
}

func NormalizeSceneCode(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func ValidateSceneCode(value string) (string, error) {
	normalized := NormalizeSceneCode(value)
	if !sceneCodePattern.MatchString(normalized) {
		return "", errors.New("invalid scene code")
	}

	return normalized, nil
}

func NormalizeHost(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	value = strings.TrimSuffix(value, ".")

	if host, _, err := net.SplitHostPort(value); err == nil {
		value = host
	}

	return strings.TrimSuffix(value, ".")
}

func ValidateHost(value string) (string, error) {
	host := NormalizeHost(value)
	if host == "" {
		return "", errors.New("invalid host")
	}
	if net.ParseIP(host) != nil || host == "localhost" {
		return host, nil
	}
	if !hostPattern.MatchString(host) {
		return "", errors.New("invalid host")
	}

	return host, nil
}
