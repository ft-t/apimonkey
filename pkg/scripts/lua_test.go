package scripts_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ft-t/apimonkey/pkg/common"
	"github.com/ft-t/apimonkey/pkg/scripts"
)

const jsonString = `{"alerts":[{"labels":{"alertname":"Watchdog"}},{"labels":{"alertname":"CriticalAlert"}},{"labels":{"alertname":"AnotherAlert"}}]}`

func TestLuaExecute(t *testing.T) {
	executor := scripts.NewLua(NewMockHTTPClient(gomock.NewController(t)))

	result, err := executor.Execute(context.Background(), `
		message = "Hello, world!" .. tostring(_G.ResponseStatusCode)
		return message .. _G.ResponseBody`, jsonString, http.StatusOK)

	require.NoError(t, err)
	assert.Equal(t, "Hello, world!200"+jsonString, result)
}

func TestLuaExecuteActionScalar(t *testing.T) {
	tests := []struct {
		name           string
		script         string
		expectedResult string
	}{
		{
			name:           "string",
			script:         `return ButtonContextID .. "|" .. ButtonConfig.apiUrl .. "|" .. ButtonConfig.headers.Authorization .. "|" .. tostring(ButtonConfig.actionScript)`,
			expectedResult: "button-1|https://example.test|token|nil",
		},
		{
			name:           "number",
			script:         `return 42`,
			expectedResult: "42",
		},
		{
			name:           "boolean",
			script:         `return true`,
			expectedResult: "true",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			executor := scripts.NewLua(NewMockHTTPClient(gomock.NewController(t)))
			config := common.Config{
				ActionScript: "secret script",
				ApiUrl:       "https://example.test",
				Headers:      map[string]string{"Authorization": "token"},
			}

			result, err := executor.ExecuteAction(context.Background(), test.script, "button-1", config)

			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Equal(t, test.expectedResult, *result)
		})
	}
}

func TestLuaExecuteActionNil(t *testing.T) {
	executor := scripts.NewLua(NewMockHTTPClient(gomock.NewController(t)))

	result, err := executor.ExecuteAction(context.Background(), `return nil`, "button-1", common.Config{})

	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestLuaExecuteActionHTTP(t *testing.T) {
	httpClient := NewMockHTTPClient(gomock.NewController(t))
	httpClient.EXPECT().Do(gomock.Any()).DoAndReturn(func(request *http.Request) (*http.Response, error) {
		assert.Equal(t, http.MethodPost, request.Method)
		assert.Equal(t, "https://example.test/hook", request.URL.String())
		assert.Equal(t, "token", request.Header.Get("Authorization"))
		body, err := io.ReadAll(request.Body)
		require.NoError(t, err)
		assert.Equal(t, "payload", string(body))
		deadline, ok := request.Context().Deadline()
		require.True(t, ok)
		assert.WithinDuration(t, time.Now().Add(10*time.Second), deadline, time.Second)

		return &http.Response{
			StatusCode: http.StatusAccepted,
			Header: http.Header{
				"X-Result": []string{"first", "second"},
			},
			Body: io.NopCloser(strings.NewReader("accepted")),
		}, nil
	})
	executor := scripts.NewLua(httpClient)

	result, err := executor.ExecuteAction(context.Background(), `
		local response = http.request({
			method = "POST",
			url = "https://example.test/hook",
			headers = { Authorization = "token" },
			body = "payload",
			timeout_seconds = 10
		})
		return response.status_code .. "|" .. response.headers["X-Result"][2] .. "|" .. response.body
	`, "button-1", common.Config{})

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "202|second|accepted", *result)
}

func TestLuaExecuteActionFailure(t *testing.T) {
	tests := []struct {
		name          string
		script        string
		expectedError string
	}{
		{
			name:          "missing method",
			script:        `return http.request({url = "https://example.test"})`,
			expectedError: "method is required",
		},
		{
			name:          "missing url",
			script:        `return http.request({method = "GET"})`,
			expectedError: "url is required",
		},
		{
			name:          "invalid timeout",
			script:        `return http.request({method = "GET", url = "https://example.test", timeout_seconds = 0})`,
			expectedError: "timeout_seconds must be positive",
		},
		{
			name:          "unsupported return",
			script:        `return {value = "unsupported"}`,
			expectedError: "unsupported Lua action return type: table",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			executor := scripts.NewLua(NewMockHTTPClient(gomock.NewController(t)))

			result, err := executor.ExecuteAction(context.Background(), test.script, "button-1", common.Config{})

			assert.Nil(t, result)
			require.ErrorContains(t, err, test.expectedError)
		})
	}
}

func TestLuaExecuteActionTransportFailure(t *testing.T) {
	httpClient := NewMockHTTPClient(gomock.NewController(t))
	httpClient.EXPECT().Do(gomock.Any()).Return(nil, errors.New("connection failed"))
	executor := scripts.NewLua(httpClient)

	result, err := executor.ExecuteAction(context.Background(), `return http.request({method = "GET", url = "https://example.test"})`, "button-1", common.Config{})

	assert.Nil(t, result)
	require.ErrorContains(t, err, "connection failed")
}

func TestLuaExecuteCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	executor := scripts.NewLua(NewMockHTTPClient(gomock.NewController(t)))

	result, err := executor.Execute(ctx, `while true do end`, "", http.StatusOK)

	assert.Empty(t, result)
	require.ErrorIs(t, err, context.Canceled)
}
