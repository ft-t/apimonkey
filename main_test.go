package main

import (
	"context"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStartKeyPressDoesNotBlockCaller(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})
	completed := make(chan struct{})
	var finished atomic.Bool
	keyPress := func(context.Context, string) error {
		close(started)
		<-release
		finished.Store(true)
		close(completed)

		return nil
	}

	startKeyPress(context.Background(), "button-1", keyPress)
	<-started
	close(release)
	<-completed

	assert.True(t, finished.Load())
}
