package instance_test

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/ft-t/apimonkey/pkg/instance"
)

func TestDefaultFactory(t *testing.T) {
	controller := gomock.NewController(t)
	sdk := NewMockSDK(controller)
	executor := NewMockExecutor(controller)
	browserOpener := NewMockBrowserOpener(controller)

	factory := instance.NewDefaultFactory(sdk, executor, browserOpener)

	instanceRef := factory.Create("1231231").(*instance.DefaultInstance)

	assert.Equal(t, sdk, instanceRef.SDK())
	assert.Equal(t, executor, instanceRef.Executor())
	assert.Equal(t, browserOpener, instanceRef.BrowserOpener())
	assert.Equal(t, "1231231", instanceRef.ContextID())
}
