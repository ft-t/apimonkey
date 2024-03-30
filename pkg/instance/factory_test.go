package instance_test

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/ft-t/apimonkey/pkg/instance"
)

func TestDefaultFactory(t *testing.T) {
	sdk := NewMockSDK(gomock.NewController(t))
	executor := NewMockExecutor(gomock.NewController(t))

	factory := instance.NewDefaultFactory(sdk, executor)

	instanceRef := factory.Create("1231231")

	assert.Equal(t, sdk, instanceRef.SDK())
	assert.Equal(t, executor, instanceRef.Executor())
	assert.Equal(t, "1231231", instanceRef.ContextID())
}
