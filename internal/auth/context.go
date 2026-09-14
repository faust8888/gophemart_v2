package auth

import "context"

type contextKey struct{}

// ContextWithUserID сохраняет идентификатор аутентифицированного пользователя в контексте.
func ContextWithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, contextKey{}, userID)
}

// UserIDFromContext возвращает идентификатор пользователя из контекста запроса.
func UserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(contextKey{}).(string)
	return userID, ok && userID != ""
}
