package expert

import (
	"context"
	"time"
)

const marketCleanupTimeout = 10 * time.Second

func marketCleanupContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.WithoutCancel(ctx), marketCleanupTimeout)
}
