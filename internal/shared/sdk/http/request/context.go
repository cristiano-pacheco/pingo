package request

import (
	"net/http"
)

type contextKey string

const UserIDKey contextKey = "user_id"

func GetUserID(r *http.Request) uint64 {
	userID, ok := r.Context().Value(UserIDKey).(uint64)
	if !ok {
		return 0
	}
	return userID
}
