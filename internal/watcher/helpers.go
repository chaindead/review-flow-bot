package watcher

import "context"

func contextFromChan(doneCh <-chan struct{}) context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		select {
		case <-doneCh:
			cancel()
		case <-ctx.Done():
			// context already canceled
		}
	}()
	return ctx
}
