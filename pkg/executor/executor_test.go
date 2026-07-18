package executor_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/imroc/req/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ft-t/apimonkey/pkg/common"
	"github.com/ft-t/apimonkey/pkg/executor"
)

func TestExecutorExecuteAction(t *testing.T) {
	controller := gomock.NewController(t)
	ctx := context.Background()
	config := common.Config{
		ActionScript:       "return 42",
		Headers:            map[string]string{"Authorization": "token"},
		TemplateParameters: map[string]string{"Project": "api-monkey"},
	}
	expectedConfig := config.Clone()
	value := "42"
	scriptExecutor := NewMockScriptExecutor(controller)
	scriptExecutor.EXPECT().ExecuteAction(ctx, config.ActionScript, "button-1", expectedConfig).Return(&value, nil)
	subject := executor.NewExecutor(scriptExecutor, NewMockHTTPClient(controller), NewMockHTTPClient(controller))

	response, err := subject.ExecuteAction(ctx, executor.ExecuteActionRequest{
		ButtonContextID: "button-1",
		Config:          config,
	})

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, &value, response.Value)
}

func TestExecutorExecuteActionFailure(t *testing.T) {
	controller := gomock.NewController(t)
	ctx := context.Background()
	config := common.Config{ActionScript: "invalid"}
	scriptExecutor := NewMockScriptExecutor(controller)
	scriptExecutor.EXPECT().ExecuteAction(ctx, config.ActionScript, "button-1", config).Return(nil, errors.New("script failed"))
	subject := executor.NewExecutor(scriptExecutor, NewMockHTTPClient(controller), NewMockHTTPClient(controller))

	response, err := subject.ExecuteAction(ctx, executor.ExecuteActionRequest{
		ButtonContextID: "button-1",
		Config:          config,
	})

	assert.Nil(t, response)
	require.EqualError(t, err, "execute action script: script failed")
}

func TestExecutorExecuteUsesSecureClient(t *testing.T) {
	controller := gomock.NewController(t)
	secureClient := NewMockHTTPClient(controller)
	secureClient.EXPECT().NewRequest().Return(successfulRequest("ready"))
	subject := executor.NewExecutor(NewMockScriptExecutor(controller), secureClient, NewMockHTTPClient(controller))

	response, err := subject.Execute(context.Background(), executor.ExecuteRequest{Config: common.Config{
		ApiUrl:     "https://example.test",
		MethodType: http.MethodGet,
	}})

	require.NoError(t, err)
	assert.Equal(t, "ready", response.Response)
}

func TestExecutorExecuteUsesInsecureClient(t *testing.T) {
	controller := gomock.NewController(t)
	insecureClient := NewMockHTTPClient(controller)
	insecureClient.EXPECT().NewRequest().Return(successfulRequest("ready"))
	subject := executor.NewExecutor(NewMockScriptExecutor(controller), NewMockHTTPClient(controller), insecureClient)

	response, err := subject.Execute(context.Background(), executor.ExecuteRequest{Config: common.Config{
		ApiUrl:             "https://example.test",
		MethodType:         http.MethodGet,
		InsecureSkipVerify: true,
	}})

	require.NoError(t, err)
	assert.Equal(t, "ready", response.Response)
}

func successfulRequest(body string) *req.Request {
	client := req.C()
	client.GetTransport().WrapRoundTripFunc(func(http.RoundTripper) req.HttpRoundTripFunc {
		return func(request *http.Request) (*http.Response, error) {
			return &http.Response{
				Status:     "200 OK",
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader(body)),
				Request:    request,
				ProtoMajor: 1,
				ProtoMinor: 1,
			}, nil
		}
	})

	return client.NewRequest()
}
