package event

const (
	IdentityUserCreatedTopic = "identity.user.created"
)

type UserCreatedMessage struct {
	UserID string `json:"user_id"`
}
