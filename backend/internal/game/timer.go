package game

import (
	"context"
	"time"
)

// StartTimer fires onExpire after d unless the returned cancel is called first.
func StartTimer(ctx context.Context, d time.Duration, onExpire func()) (cancel context.CancelFunc) {
	ctx, cancel = context.WithCancel(ctx)
	go func() {
		select {
		case <-time.After(d):
			onExpire()
		case <-ctx.Done():
		}
	}()
	return cancel
}
