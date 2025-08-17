package dto

type AuthLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthGenerateJWTRequest struct {
	UserID uint64 `json:"user_id"`
	Code   string `json:"code"`
}

type AuthGenerateJWTResponse struct {
	Token string `json:"token"`
}
