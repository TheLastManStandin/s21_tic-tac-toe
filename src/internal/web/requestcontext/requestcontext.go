package requestcontext

import "context"

type userUUIDkey struct {
}

func WithUserUUID(ctx context.Context, userUUID string) context.Context {
	return context.WithValue(ctx, userUUIDkey{}, userUUID)
}

func UserUUID(ctx context.Context) (string, bool) {
	uuid, ok := ctx.Value(userUUIDkey{}).(string)
	return uuid, ok
}
