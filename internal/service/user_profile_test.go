package service

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"gorm.io/gorm"

	"rainbow-backend/internal/model"
)

type stubUserProfileRepo struct {
	user         *model.User
	updates      map[string]any
	clearCalled  bool
	getByIDError error
	updateError  error
	clearError   error
}

func (r *stubUserProfileRepo) GetByID(context.Context, uint) (*model.User, error) {
	if r.getByIDError != nil {
		return nil, r.getByIDError
	}
	if r.user == nil {
		return nil, gorm.ErrRecordNotFound
	}

	return r.user, nil
}

func (r *stubUserProfileRepo) UpdateProfile(_ context.Context, _ uint, updates map[string]any) error {
	if r.updateError != nil {
		return r.updateError
	}

	r.updates = updates
	for key, value := range updates {
		switch key {
		case "scene_code":
			r.user.SceneCodes = value.(model.JSONStringArray)
		case "nickname":
			if value == nil {
				r.user.Nickname = nil
			} else {
				normalized := value.(string)
				r.user.Nickname = &normalized
			}
		case "avatar_url":
			if value == nil {
				r.user.AvatarURL = nil
			} else {
				normalized := value.(string)
				r.user.AvatarURL = &normalized
			}
		case "birthday":
			if value == nil {
				r.user.Birthday = nil
			} else {
				parsed := value.(time.Time)
				r.user.Birthday = &parsed
			}
		case "gender":
			if value == nil {
				r.user.Gender = nil
			} else {
				normalized := value.(int)
				r.user.Gender = &normalized
			}
		case "occupation":
			if value == nil {
				r.user.Occupation = nil
			} else {
				normalized := value.(int)
				r.user.Occupation = &normalized
			}
		}
	}

	return nil
}

func (r *stubUserProfileRepo) ClearProfile(context.Context, uint) error {
	if r.clearError != nil {
		return r.clearError
	}

	r.clearCalled = true
	r.user.SceneCodes = model.JSONStringArray{}
	r.user.Nickname = nil
	r.user.AvatarURL = nil
	r.user.Birthday = nil
	r.user.Gender = nil
	r.user.Occupation = nil
	return nil
}

func TestUserProfileServiceSave(t *testing.T) {
	repository := &stubUserProfileRepo{user: &model.User{ID: 7}}
	service := NewUserProfileService(repository)

	sceneCodes := []string{" Curry ", "love"}
	nickname := "  小明  "
	avatarURL := "/static/user/avatars/avatar.png"
	birthday := "1990-01-02"
	gender := 1
	occupation := 2

	result, err := service.Save(context.Background(), 7, &model.UserProfileUpdateRequest{
		SceneCode:  model.OptionalField[[]string]{Set: true, Value: &sceneCodes},
		Nickname:   model.OptionalField[string]{Set: true, Value: &nickname},
		AvatarURL:  model.OptionalField[string]{Set: true, Value: &avatarURL},
		Birthday:   model.OptionalField[string]{Set: true, Value: &birthday},
		Gender:     model.OptionalField[int]{Set: true, Value: &gender},
		Occupation: model.OptionalField[int]{Set: true, Value: &occupation},
	})
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	if len(result.SceneCode) != 2 || result.SceneCode[0] != "curry" || result.SceneCode[1] != "love" {
		t.Fatalf("unexpected scene_code: %#v", result.SceneCode)
	}
	if result.Nickname != "小明" {
		t.Fatalf("nickname = %q", result.Nickname)
	}
	if result.AvatarURL != avatarURL {
		t.Fatalf("avatar_url = %q", result.AvatarURL)
	}
	if result.Birthday == nil || *result.Birthday != birthday {
		t.Fatalf("birthday = %#v", result.Birthday)
	}
	if result.Gender == nil || *result.Gender != gender {
		t.Fatalf("gender = %#v", result.Gender)
	}
	if result.Occupation == nil || *result.Occupation != occupation {
		t.Fatalf("occupation = %#v", result.Occupation)
	}
}

func TestUserProfileServiceSaveSupportsClearingFields(t *testing.T) {
	repository := &stubUserProfileRepo{
		user: &model.User{ID: 7},
	}
	service := NewUserProfileService(repository)

	var req model.UserProfileUpdateRequest
	if err := json.Unmarshal([]byte(`{
		"scene_code": [],
		"nickname": "",
		"avatar_url": null,
		"birthday": "",
		"gender": null,
		"occupation": null
	}`), &req); err != nil {
		t.Fatalf("unmarshal request: %v", err)
	}

	if _, err := service.Save(context.Background(), 7, &req); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	if len(repository.user.SceneCodes) != 0 ||
		repository.user.Nickname != nil ||
		repository.user.AvatarURL != nil ||
		repository.user.Birthday != nil ||
		repository.user.Gender != nil ||
		repository.user.Occupation != nil {
		t.Fatalf("expected profile fields to be cleared: %#v", repository.user)
	}
}

func TestUserProfileServiceRejectsInvalidValues(t *testing.T) {
	tests := []struct {
		name string
		req  *model.UserProfileUpdateRequest
	}{
		{
			name: "invalid scene code",
			req: &model.UserProfileUpdateRequest{
				SceneCode: model.OptionalField[[]string]{
					Set: true,
					Value: func() *[]string {
						value := []string{"invalid scene"}
						return &value
					}(),
				},
			},
		},
		{
			name: "invalid birthday",
			req: &model.UserProfileUpdateRequest{
				Birthday: model.OptionalField[string]{
					Set: true,
					Value: func() *string {
						value := "1990-02-30"
						return &value
					}(),
				},
			},
		},
		{
			name: "invalid avatar url",
			req: &model.UserProfileUpdateRequest{
				AvatarURL: model.OptionalField[string]{
					Set: true,
					Value: func() *string {
						value := "javascript:alert(1)"
						return &value
					}(),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewUserProfileService(&stubUserProfileRepo{user: &model.User{ID: 7}})
			if _, err := service.Save(context.Background(), 7, tt.req); !errors.Is(err, ErrInvalidUserProfile) {
				t.Fatalf("Save() error = %v, want ErrInvalidUserProfile", err)
			}
		})
	}
}

func TestUserProfileServiceClear(t *testing.T) {
	repository := &stubUserProfileRepo{user: &model.User{ID: 7}}
	service := NewUserProfileService(repository)

	if _, err := service.Clear(context.Background(), 7); err != nil {
		t.Fatalf("Clear() error = %v", err)
	}
	if !repository.clearCalled {
		t.Fatal("expected ClearProfile() to be called")
	}
}
