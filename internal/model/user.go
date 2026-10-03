package model

import (
	"bytes"
	"encoding/json"
	"time"
)

type User struct {
	ID            uint            `gorm:"primaryKey" json:"id"`
	OpenID        string          `gorm:"column:openid;size:191;not null;uniqueIndex" json:"-"`
	Token         *string         `gorm:"size:128;uniqueIndex" json:"-"`
	TokenExpireAt *time.Time      `gorm:"column:token_expire_at" json:"-"`
	LastLoginAt   *time.Time      `gorm:"column:last_login_at" json:"-"`
	SceneCodes    JSONStringArray `gorm:"column:scene_code;type:json" json:"-"`
	Nickname      *string         `gorm:"size:128" json:"-"`
	AvatarURL     *string         `gorm:"column:avatar_url;size:1024" json:"-"`
	Birthday      *time.Time      `gorm:"column:birthday;type:date" json:"-"`
	Gender        *int            `gorm:"column:gender" json:"-"`
	Occupation    *int            `gorm:"column:occupation" json:"-"`
	CreatedAt     time.Time       `json:"createdAt"`
	UpdatedAt     time.Time       `json:"updatedAt"`
}

func (User) TableName() string {
	return "users"
}

type OptionalField[T any] struct {
	Set   bool
	Value *T
}

func (field *OptionalField[T]) UnmarshalJSON(data []byte) error {
	field.Set = true
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		field.Value = nil
		return nil
	}

	var value T
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}

	field.Value = &value
	return nil
}

type UserProfileUpdateRequest struct {
	SceneCode  OptionalField[[]string] `json:"scene_code"`
	Nickname   OptionalField[string]   `json:"nickname"`
	AvatarURL  OptionalField[string]   `json:"avatar_url"`
	Birthday   OptionalField[string]   `json:"birthday"`
	Gender     OptionalField[int]      `json:"gender"`
	Occupation OptionalField[int]      `json:"occupation"`
}

type UserProfileResponse struct {
	UserID     uint     `json:"user_id"`
	SceneCode  []string `json:"scene_code"`
	Nickname   string   `json:"nickname"`
	AvatarURL  string   `json:"avatar_url"`
	Birthday   *string  `json:"birthday"`
	Gender     *int     `json:"gender"`
	Occupation *int     `json:"occupation"`
}

func NewUserProfileResponse(user *User) *UserProfileResponse {
	if user == nil {
		return nil
	}

	sceneCodes := []string(user.SceneCodes)
	if sceneCodes == nil {
		sceneCodes = []string{}
	}

	nickname := ""
	if user.Nickname != nil {
		nickname = *user.Nickname
	}

	avatarURL := ""
	if user.AvatarURL != nil {
		avatarURL = *user.AvatarURL
	}

	var birthday *string
	if user.Birthday != nil {
		value := user.Birthday.Format("2006-01-02")
		birthday = &value
	}

	return &UserProfileResponse{
		UserID:     user.ID,
		SceneCode:  sceneCodes,
		Nickname:   nickname,
		AvatarURL:  avatarURL,
		Birthday:   birthday,
		Gender:     user.Gender,
		Occupation: user.Occupation,
	}
}
