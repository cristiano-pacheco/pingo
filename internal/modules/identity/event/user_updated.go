package event

const (
	IdentityUserUpdatedTopic = "identity.user.updated"
)

type UserUpdatedMessage struct {
	UserID string `json:"user_id"`
}
