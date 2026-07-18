package instance

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLifecycleRestartPollingPreservesLifecycleContext(t *testing.T) {
	lifecycle := newLifecycle(context.Background())
	lifecycleContext := lifecycle.Context()
	firstPollingContext := lifecycle.RestartPolling()
	secondPollingContext := lifecycle.RestartPolling()

	require.ErrorIs(t, firstPollingContext.Err(), context.Canceled)
	require.NoError(t, lifecycleContext.Err())
	require.NoError(t, secondPollingContext.Err())
	assert.Equal(t, lifecycleContext, lifecycle.Context())

	lifecycle.Stop()
	require.ErrorIs(t, lifecycleContext.Err(), context.Canceled)
	require.ErrorIs(t, secondPollingContext.Err(), context.Canceled)
}
