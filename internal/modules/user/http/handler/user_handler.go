package handler

import "net/http"

type UserHandler struct {
}

func NewUserHandler() *UserHandler {
	return &UserHandler{}
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("CreateUser"))
}

func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("UpdateUser"))
}

func (h *UserHandler) ActivateUser(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ActivateUser"))
}
