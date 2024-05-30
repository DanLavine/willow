package testhelpers

import (
	"context"

	"github.com/DanLavine/willow/internal/middleware"
	"go.uber.org/zap"
)

// NewContextWithLogger can be used to create a default background context with a Nop logger
func NewContextWithMiddlewareSetup() context.Context {
	ctx := context.Background()

	// setup logger
	return context.WithValue(ctx, middleware.LoggerCtxKey, zap.NewNop())
}

func NewCancelContextWithMiddlewareSetup() (context.Context, context.CancelFunc) {
	ctx := context.Background()

	// setup logger
	ctx = context.WithValue(ctx, middleware.LoggerCtxKey, zap.NewNop())

	return context.WithCancel(ctx)
}
