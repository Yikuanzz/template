package meta

// ContextKey 上下文键
type ContextKey string

const (
	ContextKeyAccessToken ContextKey = "access_token"
	ContextKeyUserID      ContextKey = "user_id"
)
