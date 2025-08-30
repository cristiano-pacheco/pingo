package event

const (
	IdentityUserAuthenticatedTopic = "identity.user.authenticated"
)

type UserAuthenticatedMessage struct {
	UserID string `json:"user_id"`
}
