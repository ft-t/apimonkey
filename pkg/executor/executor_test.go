package executor_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ft-t/apimonkey/pkg/common"
	"github.com/ft-t/apimonkey/pkg/executor"
)

func TestExecutorExecuteAction(t *testing.T) {
	ctx := context.Background()
	config := common.Config{
		ActionScript:       "return 42",
		Headers:            map[string]string{"Authorization": "token"},
		TemplateParameters: map[string]string{"Project": "api-monkey"},
	}
	expectedConfig := config.Clone()
	value := "42"
	scriptExecutor := NewMockScriptExecutor(gomock.NewController(t))
	scriptExecutor.EXPECT().ExecuteAction(ctx, config.ActionScript, "button-1", expectedConfig).Return(&value, nil)
	subject := executor.NewExecutor(scriptExecutor)

	response, err := subject.ExecuteAction(ctx, executor.ExecuteActionRequest{
		ButtonContextID: "button-1",
		Config:          config,
	})

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, &value, response.Value)
}

func TestExecutorExecuteActionFailure(t *testing.T) {
	ctx := context.Background()
	config := common.Config{ActionScript: "invalid"}
	scriptExecutor := NewMockScriptExecutor(gomock.NewController(t))
	scriptExecutor.EXPECT().ExecuteAction(ctx, config.ActionScript, "button-1", config).Return(nil, errors.New("script failed"))
	subject := executor.NewExecutor(scriptExecutor)

	response, err := subject.ExecuteAction(ctx, executor.ExecuteActionRequest{
		ButtonContextID: "button-1",
		Config:          config,
	})

	assert.Nil(t, response)
	require.EqualError(t, err, "execute action script: script failed")
}
