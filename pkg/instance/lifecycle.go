package instance

import "context"

type lifecycle struct {
	ctx           context.Context
	cancel        context.CancelFunc
	pollingCancel context.CancelFunc
}

func newLifecycle(parent context.Context) *lifecycle {
	ctx, cancel := context.WithCancel(parent)

	return &lifecycle{
		ctx:    ctx,
		cancel: cancel,
	}
}

func (l *lifecycle) Context() context.Context {
	return l.ctx
}

func (l *lifecycle) RestartPolling() context.Context {
	l.StopPolling()
	pollingCtx, cancel := context.WithCancel(l.ctx)
	l.pollingCancel = cancel

	return pollingCtx
}

func (l *lifecycle) StopPolling() {
	if l.pollingCancel != nil {
		l.pollingCancel()
	}

	l.pollingCancel = nil
}

func (l *lifecycle) Stop() {
	l.StopPolling()
	l.cancel()
}
