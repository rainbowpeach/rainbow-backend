package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"gorm.io/gorm"

	"rainbow-backend/internal/model"
	"rainbow-backend/internal/repo"
)

var (
	ErrInvalidUserProfile = errors.New("invalid user profile")
	ErrUserNotFound       = errors.New("user not found")
)

type UserProfileService struct {
	users repo.UserProfileRepository
}

func NewUserProfileService(users repo.UserProfileRepository) *UserProfileService {
	return &UserProfileService{users: users}
}

func (s *UserProfileService) Get(ctx context.Context, userID uint) (*model.UserProfileResponse, error) {
	user, err := s.getUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	return model.NewUserProfileResponse(user), nil
}

func (s *UserProfileService) Save(ctx context.Context, userID uint, req *model.UserProfileUpdateRequest) (*model.UserProfileResponse, error) {
	if _, err := s.getUser(ctx, userID); err != nil {
		return nil, err
	}

	updates, err := buildUserProfileUpdates(req)
	if err != nil {
		return nil, err
	}
	if len(updates) == 0 {
		return s.Get(ctx, userID)
	}

	if err := s.users.UpdateProfile(ctx, userID, updates); err != nil {
		return nil, fmt.Errorf("update user profile: %w", err)
	}

	return s.Get(ctx, userID)
}

func (s *UserProfileService) Clear(ctx context.Context, userID uint) (*model.UserProfileResponse, error) {
	if _, err := s.getUser(ctx, userID); err != nil {
		return nil, err
	}

	if err := s.users.ClearProfile(ctx, userID); err != nil {
		return nil, fmt.Errorf("clear user profile: %w", err)
	}

	return s.Get(ctx, userID)
}

func (s *UserProfileService) getUser(ctx context.Context, userID uint) (*model.User, error) {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get user: %w", err)
	}

	return user, nil
}

func buildUserProfileUpdates(req *model.UserProfileUpdateRequest) (map[string]any, error) {
	if req == nil {
		return nil, ErrInvalidUserProfile
	}

	updates := make(map[string]any)

	if req.SceneCode.Set {
		if req.SceneCode.Value == nil {
			updates["scene_code"] = model.JSONStringArray{}
		} else {
			sceneCodes := make(model.JSONStringArray, 0, len(*req.SceneCode.Value))
			for _, sceneCode := range *req.SceneCode.Value {
				normalized, err := model.ValidateSceneCode(sceneCode)
				if err != nil {
					return nil, ErrInvalidUserProfile
				}
				sceneCodes = append(sceneCodes, normalized)
			}
			updates["scene_code"] = sceneCodes
		}
	}

	if req.Nickname.Set {
		if req.Nickname.Value == nil {
			updates["nickname"] = nil
		} else {
			nickname := strings.TrimSpace(*req.Nickname.Value)
			if utf8.RuneCountInString(nickname) > 128 {
				return nil, ErrInvalidUserProfile
			}
			if nickname == "" {
				updates["nickname"] = nil
			} else {
				updates["nickname"] = nickname
			}
		}
	}

	if req.AvatarURL.Set {
		if req.AvatarURL.Value == nil {
			updates["avatar_url"] = nil
		} else {
			avatarURL := strings.TrimSpace(*req.AvatarURL.Value)
			if avatarURL == "" {
				updates["avatar_url"] = nil
			} else {
				normalized, err := model.ValidateOptionalAssetURL(avatarURL)
				if err != nil {
					return nil, ErrInvalidUserProfile
				}
				updates["avatar_url"] = normalized
			}
		}
	}

	if req.Birthday.Set {
		if req.Birthday.Value == nil {
			updates["birthday"] = nil
		} else {
			birthday := strings.TrimSpace(*req.Birthday.Value)
			if birthday == "" {
				updates["birthday"] = nil
			} else {
				parsed, err := time.Parse("2006-01-02", birthday)
				if err != nil {
					return nil, ErrInvalidUserProfile
				}
				updates["birthday"] = parsed
			}
		}
	}

	if req.Gender.Set {
		if req.Gender.Value == nil {
			updates["gender"] = nil
		} else {
			updates["gender"] = *req.Gender.Value
		}
	}

	if req.Occupation.Set {
		if req.Occupation.Value == nil {
			updates["occupation"] = nil
		} else {
			updates["occupation"] = *req.Occupation.Value
		}
	}

	return updates, nil
}
