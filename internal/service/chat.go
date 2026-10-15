package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"gorm.io/gorm"

	"rainbow-backend/internal/model"
	"rainbow-backend/internal/repo"
)

const (
	chatHistoryLimit     = 10
	chatUserTextLimit    = 500
	chatHistoryTextLimit = 2000
	chatRateLimit        = 20
	chatRateWindow       = time.Minute
	defaultTemperature   = 0.8
	defaultMaxTokens     = 400
)

var (
	ErrChatNotConfigured  = errors.New("chat not configured")
	ErrInvalidChat        = errors.New("invalid chat")
	ErrChatPersona        = errors.New("chat persona not found")
	ErrChatPersonaExists  = errors.New("chat persona exists")
	ErrChatRateLimited    = errors.New("chat rate limited")
)

var sceneAliases = map[string]string{
	"xiaozhizhu-live": "zhizhuxia-live",
	"xiaozhizhu":      "zhizhuxia-live",
	"spiderman1":      "zhizhuxia-live",
	"zhizhuxia":       "zhizhuxia-live",
	"kobe":            "kobe-live",
	"cluo":            "cluo-live",
	"cr7":             "cluo-live",
	"ronaldo":         "cluo-live",
}

var genderLabels = []string{"保密", "男", "女"}
var identityLabels = []string{"保密", "大学生", "初中生", "高中生", "社会牛马"}

type ChatService struct {
	personas repo.ChatPersonaRepository
	logs     repo.ChatLogRepository
	users    repo.UserProfileRepository
	client   ChatCompletionClient
	pages    scenePageConfigCreator
	timeout  time.Duration
	limiter  *userRateLimiter
	now      func() time.Time
}

type scenePageConfigCreator interface {
	Create(ctx context.Context, item *model.ScenePageConfig) error
}

func NewChatService(
	personas repo.ChatPersonaRepository,
	logs repo.ChatLogRepository,
	users repo.UserProfileRepository,
	client ChatCompletionClient,
	timeout time.Duration,
) *ChatService {
	if timeout <= 0 {
		timeout = 60 * time.Second
	}
	return &ChatService{
		personas: personas,
		logs:     logs,
		users:    users,
		client:   client,
		timeout:  timeout,
		limiter:  newUserRateLimiter(chatRateLimit, chatRateWindow),
		now:      time.Now,
	}
}

func (s *ChatService) SetScenePages(pages scenePageConfigCreator) {
	s.pages = pages
}

func (s *ChatService) Reply(ctx context.Context, userID uint, req *model.ChatRequest) (*model.ChatReplyResponse, error) {
	if s.client == nil {
		return nil, ErrChatNotConfigured
	}
	sceneCode, mode, history, userText, err := normalizeChatRequest(req)
	if err != nil {
		return nil, err
	}
	if !s.limiter.Allow(userID, s.now()) {
		return nil, ErrChatRateLimited
	}

	persona, err := s.personas.GetBySceneCode(ctx, sceneCode)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrChatPersona
		}
		return nil, fmt.Errorf("get chat persona: %w", err)
	}

	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get chat user: %w", err)
	}

	messages := []model.ChatTurn{{
		Role:    "system",
		Content: systemPrompt(persona, mode, user),
	}}
	messages = append(messages, history...)

	callCtx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	reply, callErr := s.client.Complete(callCtx, messages, personaTemperature(persona), personaMaxTokens(persona))
	logItem := &model.ChatLog{
		UserID:    userID,
		SceneCode: sceneCode,
		Mode:      mode,
		UserText:  userText,
		Status:    model.ChatStatusOK,
	}
	if callErr != nil {
		logItem.Status = model.ChatStatusFailed
		logItem.ErrorMessage = trimRunes(callErr.Error(), 255)
		_ = s.logs.Create(ctx, logItem)
		return nil, callErr
	}

	logItem.AssistantText = reply
	if err := s.logs.Create(ctx, logItem); err != nil {
		return nil, fmt.Errorf("save chat log: %w", err)
	}

	return &model.ChatReplyResponse{Reply: reply}, nil
}

func (s *ChatService) ListLogs(ctx context.Context, req model.ChatLogListRequest) (*model.ChatLogListResponse, error) {
	filter := model.ChatLogFilter{
		UserID:   req.UserID,
		Keyword:  strings.TrimSpace(req.Keyword),
		Page:     req.Page,
		PageSize: req.PageSize,
	}
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}
	if filter.PageSize > 100 {
		return nil, ErrInvalidChat
	}
	if req.SceneCode != "" {
		sceneCode, err := canonicalSceneCode(req.SceneCode)
		if err != nil {
			return nil, ErrInvalidChat
		}
		filter.SceneCode = sceneCode
	}

	items, total, err := s.logs.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("list chat logs: %w", err)
	}

	list := make([]*model.ChatLog, 0, len(items))
	for i := range items {
		list = append(list, &items[i])
	}

	return &model.ChatLogListResponse{
		List:     list,
		Total:    total,
		Page:     filter.Page,
		PageSize: filter.PageSize,
	}, nil
}

func (s *ChatService) ListPersonas(ctx context.Context) ([]*model.ChatPersonaAdminResponse, error) {
	items, err := s.personas.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list chat personas: %w", err)
	}
	list := make([]*model.ChatPersonaAdminResponse, 0, len(items))
	for i := range items {
		list = append(list, model.NewChatPersonaAdminResponse(&items[i]))
	}
	return list, nil
}

func (s *ChatService) CreatePersona(ctx context.Context, req *model.ChatPersonaWriteRequest) (*model.ChatPersonaAdminResponse, error) {
	if req == nil {
		return nil, ErrInvalidChat
	}
	code, err := canonicalSceneCode(req.SceneCode)
	if err != nil {
		return nil, ErrInvalidChat
	}
	displayName, hello, basePrompt, modes, presets, err := normalizePersonaWrite(req)
	if err != nil {
		return nil, err
	}

	existing, err := s.personas.GetBySceneCode(ctx, code)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("get chat persona: %w", err)
	}
	if err == nil && existing != nil {
		return nil, ErrChatPersonaExists
	}

	if err := s.ensureEmptyScenePage(ctx, code); err != nil {
		return nil, err
	}

	persona := &model.ChatPersona{
		SceneCode:    code,
		DisplayName:  displayName,
		Hello:        hello,
		BasePrompt:   basePrompt,
		ModePrompts:  modes,
		PresetInputs: presets,
		Temperature:  defaultTemperature,
		MaxTokens:    defaultMaxTokens,
	}
	if err := s.personas.Create(ctx, persona); err != nil {
		if isDuplicateScenePageConfigError(err) {
			return nil, ErrChatPersonaExists
		}
		return nil, fmt.Errorf("create chat persona: %w", err)
	}
	return model.NewChatPersonaAdminResponse(persona), nil
}

func (s *ChatService) UpdatePersona(ctx context.Context, sceneCode string, req *model.ChatPersonaWriteRequest) (*model.ChatPersonaAdminResponse, error) {
	code, err := canonicalSceneCode(sceneCode)
	if err != nil {
		return nil, ErrInvalidChat
	}
	displayName, hello, basePrompt, modeUpdates, presetUpdates, err := normalizePersonaWrite(req)
	if err != nil {
		return nil, err
	}

	persona, err := s.personas.GetBySceneCode(ctx, code)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrChatPersona
		}
		return nil, fmt.Errorf("get chat persona: %w", err)
	}

	persona.DisplayName = displayName
	persona.Hello = hello
	persona.BasePrompt = basePrompt
	persona.ModePrompts = mergeJSONMap(persona.ModePrompts, modeUpdates)
	persona.PresetInputs = mergeJSONMap(persona.PresetInputs, presetUpdates)
	if err := s.personas.UpdatePrompts(ctx, persona); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrChatPersona
		}
		return nil, fmt.Errorf("update chat persona: %w", err)
	}
	return model.NewChatPersonaAdminResponse(persona), nil
}

func (s *ChatService) ensureEmptyScenePage(ctx context.Context, sceneCode string) error {
	if s.pages == nil {
		return nil
	}
	err := s.pages.Create(ctx, &model.ScenePageConfig{
		SceneCode:   sceneCode,
		TagsDefault: model.JSONStringArray{},
		ImageURLs:   model.JSONStringArray{},
	})
	if err != nil && !isDuplicateScenePageConfigError(err) {
		return fmt.Errorf("create scene page config: %w", err)
	}
	return nil
}

func normalizePersonaWrite(req *model.ChatPersonaWriteRequest) (string, string, string, model.JSONMap, model.JSONMap, error) {
	if req == nil {
		return "", "", "", nil, nil, ErrInvalidChat
	}
	displayName := strings.TrimSpace(req.DisplayName)
	if displayName == "" || utf8.RuneCountInString(displayName) > 64 {
		return "", "", "", nil, nil, ErrInvalidChat
	}
	hello := strings.TrimSpace(req.Hello)
	if hello == "" || utf8.RuneCountInString(hello) > 500 {
		return "", "", "", nil, nil, ErrInvalidChat
	}
	basePrompt := strings.TrimSpace(req.BasePrompt)
	if basePrompt == "" || utf8.RuneCountInString(basePrompt) > 20000 {
		return "", "", "", nil, nil, ErrInvalidChat
	}
	modes, err := normalizePromptMap(req.ModePrompts)
	if err != nil {
		return "", "", "", nil, nil, err
	}
	presets, err := normalizePromptMap(req.PresetInputs)
	if err != nil {
		return "", "", "", nil, nil, err
	}
	return displayName, hello, basePrompt, modes, presets, nil
}

func normalizePromptMap(source model.JSONMap) (model.JSONMap, error) {
	updates := model.JSONMap{}
	for key, value := range source {
		name := strings.TrimSpace(key)
		if name == "" || utf8.RuneCountInString(name) > 32 {
			return nil, ErrInvalidChat
		}
		text := strings.TrimSpace(value)
		if utf8.RuneCountInString(text) > 8000 {
			return nil, ErrInvalidChat
		}
		updates[name] = text
	}
	return updates, nil
}

func mergeJSONMap(current, updates model.JSONMap) model.JSONMap {
	merged := model.JSONMap{}
	for key, value := range current {
		merged[key] = value
	}
	for key, value := range updates {
		merged[key] = value
	}
	return merged
}

func normalizeChatRequest(req *model.ChatRequest) (string, string, []model.ChatTurn, string, error) {
	if req == nil {
		return "", "", nil, "", ErrInvalidChat
	}

	sceneCode, err := canonicalSceneCode(req.SceneCode)
	if err != nil {
		return "", "", nil, "", ErrInvalidChat
	}

	mode := strings.TrimSpace(req.Mode)
	if mode == "" {
		mode = "chat"
	}
	if utf8.RuneCountInString(mode) > 32 {
		return "", "", nil, "", ErrInvalidChat
	}

	history := make([]model.ChatTurn, 0, len(req.Messages))
	for _, turn := range req.Messages {
		role := strings.TrimSpace(turn.Role)
		if role != "user" && role != "assistant" {
			continue
		}
		content := strings.TrimSpace(turn.Content)
		if content == "" {
			continue
		}
		history = append(history, model.ChatTurn{
			Role:    role,
			Content: trimRunes(content, chatHistoryTextLimit),
		})
	}
	if len(history) > chatHistoryLimit {
		history = history[len(history)-chatHistoryLimit:]
	}
	if len(history) == 0 || history[len(history)-1].Role != "user" {
		return "", "", nil, "", ErrInvalidChat
	}

	userText := trimRunes(history[len(history)-1].Content, chatUserTextLimit)
	history[len(history)-1].Content = userText
	return sceneCode, mode, history, userText, nil
}

func canonicalSceneCode(value string) (string, error) {
	normalized, err := model.ValidateSceneCode(value)
	if err != nil {
		return "", err
	}
	if alias, ok := sceneAliases[normalized]; ok {
		return alias, nil
	}
	return normalized, nil
}

func systemPrompt(persona *model.ChatPersona, mode string, user *model.User) string {
	extra := ""
	if persona.ModePrompts != nil {
		extra = persona.ModePrompts[mode]
		if extra == "" {
			extra = persona.ModePrompts["chat"]
		}
	}
	return persona.BasePrompt + extra + profileLine(user)
}

func profileLine(user *model.User) string {
	if user == nil {
		return "用户档案未知。"
	}

	parts := make([]string, 0, 2)
	if user.Gender != nil {
		if label := labelAt(genderLabels, *user.Gender); label != "" && label != "不透露" {
			parts = append(parts, "性别："+label)
		}
	}
	if user.Occupation != nil {
		if label := labelAt(identityLabels, *user.Occupation); label != "" && label != "未选择" {
			parts = append(parts, "身份："+label)
		}
	}
	if len(parts) == 0 {
		return "用户档案未知。"
	}
	return "用户档案：" + strings.Join(parts, "，") + "。"
}

func labelAt(labels []string, index int) string {
	if index < 0 || index >= len(labels) {
		return ""
	}
	return labels[index]
}

func personaTemperature(persona *model.ChatPersona) float64 {
	if persona.Temperature <= 0 {
		return defaultTemperature
	}
	return persona.Temperature
}

func personaMaxTokens(persona *model.ChatPersona) int {
	if persona.MaxTokens <= 0 {
		return defaultMaxTokens
	}
	return persona.MaxTokens
}

func trimRunes(value string, limit int) string {
	if limit <= 0 || utf8.RuneCountInString(value) <= limit {
		return value
	}
	runes := []rune(value)
	return string(runes[:limit])
}

type userRateLimiter struct {
	mu     sync.Mutex
	limit  int
	window time.Duration
	hits   map[uint][]time.Time
}

func newUserRateLimiter(limit int, window time.Duration) *userRateLimiter {
	return &userRateLimiter{
		limit:  limit,
		window: window,
		hits:   map[uint][]time.Time{},
	}
}

func (l *userRateLimiter) Allow(userID uint, now time.Time) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	cutoff := now.Add(-l.window)
	recent := l.hits[userID][:0]
	for _, hit := range l.hits[userID] {
		if hit.After(cutoff) {
			recent = append(recent, hit)
		}
	}
	if len(recent) >= l.limit {
		l.hits[userID] = recent
		return false
	}
	l.hits[userID] = append(recent, now)
	return true
}
