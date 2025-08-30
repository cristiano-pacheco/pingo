package event

const (
	IdentityUserCreatedTopic = "identity.user.created"
)

type UserCreatedMessage struct {
	UserID uint64 `json:"user_id"`
}
