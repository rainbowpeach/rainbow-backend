package model

type MiniProgramLoginRequest struct {
	Code string `json:"code" binding:"required"`
}

type MiniProgramLoginResponse struct {
	Token     string `json:"token"`
	ExpiresIn int64  `json:"expiresIn"`
}
