package etlworld

import "context"

type WorldInstanceKey string

const WorldKey WorldInstanceKey = "world"

func WithWorld(ctx context.Context, geth *ETLWorld) context.Context {
	return context.WithValue(ctx, WorldKey, geth)
}

func GetWorld(ctx context.Context) *ETLWorld {
	return ctx.Value(WorldKey).(*ETLWorld)
}
