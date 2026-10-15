package model

import "time"

const (
	ChatStatusOK     = "ok"
	ChatStatusFailed = "failed"
)

type ChatPersona struct {
	SceneCode    string    `gorm:"column:scene_code;size:128;primaryKey" json:"scene_code"`
	DisplayName  string    `gorm:"column:display_name;size:64;not null;default:''" json:"-"`
	Hello        string    `gorm:"column:hello;size:2000;not null;default:''" json:"-"`
	BasePrompt   string    `gorm:"column:base_prompt;type:text;not null" json:"-"`
	ModePrompts  JSONMap   `gorm:"column:mode_prompts;type:json;not null" json:"-"`
	PresetInputs JSONMap   `gorm:"column:preset_inputs;type:json;not null" json:"-"`
	Temperature  float64   `gorm:"column:temperature;not null;default:0.8" json:"-"`
	MaxTokens    int       `gorm:"column:max_tokens;not null;default:400" json:"-"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

func (ChatPersona) TableName() string {
	return "chat_personas"
}

type ChatLog struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	UserID        uint      `gorm:"column:user_id;not null;index" json:"user_id"`
	SceneCode     string    `gorm:"column:scene_code;size:128;not null;index" json:"scene_code"`
	Mode          string    `gorm:"size:32;not null" json:"mode"`
	UserText      string    `gorm:"column:user_text;type:text;not null" json:"user_text"`
	AssistantText string    `gorm:"column:assistant_text;type:text" json:"assistant_text"`
	Status        string    `gorm:"size:16;not null;index" json:"status"`
	ErrorMessage  string    `gorm:"column:error_message;size:255" json:"error_message,omitempty"`
	CreatedAt     time.Time `gorm:"index" json:"createdAt"`
}

func (ChatLog) TableName() string {
	return "chat_logs"
}

type ChatTurn struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	SceneCode string     `json:"scene_code"`
	Mode      string     `json:"mode"`
	Messages  []ChatTurn `json:"messages"`
}

type ChatReplyResponse struct {
	Reply string `json:"reply"`
}

type ChatLogListRequest struct {
	SceneCode string `form:"scene_code"`
	UserID    uint   `form:"user_id"`
	Keyword   string `form:"keyword"`
	Page      int    `form:"page"`
	PageSize  int    `form:"pageSize"`
}

type ChatLogFilter struct {
	SceneCode string
	UserID    uint
	Keyword   string
	Page      int
	PageSize  int
}

type ChatLogListResponse struct {
	List     []*ChatLog `json:"list"`
	Total    int64      `json:"total"`
	Page     int        `json:"page"`
	PageSize int        `json:"pageSize"`
}

type ChatPersonaAdminResponse struct {
	SceneCode    string    `json:"scene_code"`
	DisplayName  string    `json:"display_name"`
	Hello        string    `json:"hello"`
	BasePrompt   string    `json:"base_prompt"`
	ModePrompts  JSONMap   `json:"mode_prompts"`
	PresetInputs JSONMap   `json:"preset_inputs"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type ChatPersonaWriteRequest struct {
	SceneCode    string  `json:"scene_code"`
	DisplayName  string  `json:"display_name"`
	Hello        string  `json:"hello"`
	BasePrompt   string  `json:"base_prompt"`
	ModePrompts  JSONMap `json:"mode_prompts"`
	PresetInputs JSONMap `json:"preset_inputs"`
}

func NewChatPersonaAdminResponse(persona *ChatPersona) *ChatPersonaAdminResponse {
	if persona == nil {
		return nil
	}
	return &ChatPersonaAdminResponse{
		SceneCode:    persona.SceneCode,
		DisplayName:  persona.DisplayName,
		Hello:        persona.Hello,
		BasePrompt:   persona.BasePrompt,
		ModePrompts:  copyJSONMap(persona.ModePrompts),
		PresetInputs: copyJSONMap(persona.PresetInputs),
		UpdatedAt:    persona.UpdatedAt,
	}
}

func copyJSONMap(source JSONMap) JSONMap {
	copied := JSONMap{}
	for key, value := range source {
		copied[key] = value
	}
	return copied
}
