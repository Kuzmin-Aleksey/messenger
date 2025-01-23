package service

import "context"

func getUserId(ctx context.Context) int {
	return ctx.Value("UserId").(int)
}
