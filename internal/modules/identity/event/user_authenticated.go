package event

const (
	IdentityUserAuthenticatedTopic = "identity.user.authenticated"
)

type UserAuthenticatedMessage struct {
	UserID uint64 `json:"user_id"`
}
