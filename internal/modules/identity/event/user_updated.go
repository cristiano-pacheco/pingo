package event

const (
	IdentityUserUpdatedTopic = "identity.user.updated"
)

type UserUpdatedMessage struct {
	UserID uint64 `json:"user_id"`
}
