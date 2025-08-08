package handler

import "net/http"

type AuthHandler struct {
}

func NewAuthHandler() *AuthHandler {
	return &AuthHandler{}
}

func (h *AuthHandler) GenerateJWTToken(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("GenerateJWTToken"))
}
