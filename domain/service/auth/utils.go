package auth

import "context"

type UserIdKey struct{}

func CtxWithUser(ctx context.Context, userId int) context.Context {
	return context.WithValue(ctx, UserIdKey{}, userId)
}

func ExtractUser(ctx context.Context) int {
	return ctx.Value(UserIdKey{}).(int)
}
