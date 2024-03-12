package testhelpers

import (
	"context"

	"github.com/DanLavine/willow/internal/reporting"
	"go.uber.org/zap"
)

// NewContextWithLogger can be used to create a default background context with a Nop logger
func NewContextWithLogger() context.Context {
	return context.WithValue(context.Background(), reporting.CustomLogger, zap.NewNop())
}
