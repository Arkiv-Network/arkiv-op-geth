package testutil

import "context"

type WorldInstanceKey string

const WorldKey WorldInstanceKey = "world"

func WithWorld(ctx context.Context, geth *World) context.Context {
	return context.WithValue(ctx, WorldKey, geth)
}

func GetWorld(ctx context.Context) *World {
	return ctx.Value(WorldKey).(*World)
}
