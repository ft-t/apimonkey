package instance_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fastjson"

	"github.com/ft-t/apimonkey/pkg/common"
	"github.com/ft-t/apimonkey/pkg/executor"
	"github.com/ft-t/apimonkey/pkg/instance"
)

func TestInstanceKeyPressedOpenBrowser(t *testing.T) {
	tests := []struct {
		name        string
		config      common.Config
		expectedURL string
	}{
		{
			name: "browser URL",
			config: common.Config{
				Action:             common.ActionOpenBrowser,
				BrowserUrl:         "https://example.test/{{.Project}}",
				TemplateParameters: map[string]string{"Project": "api-monkey"},
			},
			expectedURL: "https://example.test/api-monkey",
		},
		{
			name: "API URL fallback",
			config: common.Config{
				Action: common.ActionOpenBrowser,
				ApiUrl: "https://example.test/api",
			},
			expectedURL: "https://example.test/api",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			controller := gomock.NewController(t)
			browser := NewMockBrowserOpener(controller)
			browser.EXPECT().Open(test.expectedURL).Return(nil)
			subject := instance.NewInstance("button-1", NewMockExecutor(controller), NewMockSDK(controller), browser)
			require.NoError(t, subject.SetConfig(configValue(t, test.config)))

			err := subject.KeyPressed(context.Background())

			require.NoError(t, err)
		})
	}
}

func TestInstanceKeyPressedRefreshRequest(t *testing.T) {
	controller := gomock.NewController(t)
	requestExecutor := NewMockExecutor(controller)
	sdk := NewMockSDK(controller)
	config := common.Config{
		Action:                  common.ActionRefreshRequest,
		ApiUrl:                  "https://example.test/api",
		TitlePrefix:             "Status",
		ShowSuccessNotification: true,
	}
	requestExecutor.EXPECT().Execute(gomock.Any(), executor.ExecuteRequest{Config: config}).Return(&executor.ExecuteResponse{
		Response: "ready",
		Code:     200,
	}, nil)
	sdk.EXPECT().SetTitle("button-1", "Status\nready", 0)
	sdk.EXPECT().SetImage("button-1", "", 0)
	sdk.EXPECT().ShowOk("button-1")
	subject := instance.NewInstance("button-1", requestExecutor, sdk, NewMockBrowserOpener(controller))
	require.NoError(t, subject.SetConfig(configValue(t, config)))

	err := subject.KeyPressed(context.Background())

	require.NoError(t, err)
}

func TestInstanceKeyPressedExecuteLua(t *testing.T) {
	controller := gomock.NewController(t)
	actionExecutor := NewMockExecutor(controller)
	sdk := NewMockSDK(controller)
	config := common.Config{
		Action:       common.ActionExecuteLua,
		ActionScript: "return 'done'",
	}
	value := "done"
	actionExecutor.EXPECT().ExecuteAction(gomock.Any(), executor.ExecuteActionRequest{
		ButtonContextID: "button-1",
		Config:          config,
	}).Return(&executor.ExecuteActionResponse{Value: &value}, nil)
	sdk.EXPECT().SetTitle("button-1", value, 0)
	sdk.EXPECT().SetImage("button-1", "", 0)
	subject := instance.NewInstance("button-1", actionExecutor, sdk, NewMockBrowserOpener(controller))
	require.NoError(t, subject.SetConfig(configValue(t, config)))

	err := subject.KeyPressed(context.Background())

	require.NoError(t, err)
}

func TestInstanceKeyPressedExecuteLuaNil(t *testing.T) {
	controller := gomock.NewController(t)
	actionExecutor := NewMockExecutor(controller)
	config := common.Config{
		Action:       common.ActionExecuteLua,
		ActionScript: "return nil",
	}
	actionExecutor.EXPECT().ExecuteAction(gomock.Any(), executor.ExecuteActionRequest{
		ButtonContextID: "button-1",
		Config:          config,
	}).Return(&executor.ExecuteActionResponse{}, nil)
	subject := instance.NewInstance("button-1", actionExecutor, NewMockSDK(controller), NewMockBrowserOpener(controller))
	require.NoError(t, subject.SetConfig(configValue(t, config)))

	err := subject.KeyPressed(context.Background())

	require.NoError(t, err)
}

func TestInstanceSetConfigFailure(t *testing.T) {
	controller := gomock.NewController(t)
	sdk := NewMockSDK(controller)
	sdk.EXPECT().ShowAlert("button-1")
	subject := instance.NewInstance("button-1", NewMockExecutor(controller), sdk, NewMockBrowserOpener(controller))

	err := subject.SetConfig(configValue(t, common.Config{Action: common.Action("unknown")}))

	require.EqualError(t, err, "unknown button action: unknown")
}

func TestInstanceKeyPressedFailure(t *testing.T) {
	tests := []struct {
		name          string
		config        common.Config
		expectFailure func(*MockExecutor, *MockBrowserOpener)
		expectedError string
	}{
		{
			name:   "browser",
			config: common.Config{Action: common.ActionOpenBrowser, BrowserUrl: "https://example.test"},
			expectFailure: func(_ *MockExecutor, browser *MockBrowserOpener) {
				browser.EXPECT().Open("https://example.test").Return(errors.New("browser failed"))
			},
			expectedError: "browser failed",
		},
		{
			name:   "refresh",
			config: common.Config{Action: common.ActionRefreshRequest},
			expectFailure: func(requestExecutor *MockExecutor, _ *MockBrowserOpener) {
				requestExecutor.EXPECT().Execute(gomock.Any(), gomock.Any()).Return(nil, errors.New("request failed"))
			},
			expectedError: "request failed",
		},
		{
			name:   "lua",
			config: common.Config{Action: common.ActionExecuteLua, ActionScript: "invalid"},
			expectFailure: func(actionExecutor *MockExecutor, _ *MockBrowserOpener) {
				actionExecutor.EXPECT().ExecuteAction(gomock.Any(), gomock.Any()).Return(nil, errors.New("lua failed"))
			},
			expectedError: "lua failed",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			controller := gomock.NewController(t)
			dependencyExecutor := NewMockExecutor(controller)
			browser := NewMockBrowserOpener(controller)
			sdk := NewMockSDK(controller)
			test.expectFailure(dependencyExecutor, browser)
			sdk.EXPECT().ShowAlert("button-1")
			subject := instance.NewInstance("button-1", dependencyExecutor, sdk, browser)
			require.NoError(t, subject.SetConfig(configValue(t, test.config)))

			err := subject.KeyPressed(context.Background())

			require.ErrorContains(t, err, test.expectedError)
		})
	}
}

func TestInstanceKeyPressedSuppressesOverlap(t *testing.T) {
	controller := gomock.NewController(t)
	requestExecutor := NewMockExecutor(controller)
	started := make(chan struct{})
	release := make(chan struct{})
	requestExecutor.EXPECT().Execute(gomock.Any(), gomock.Any()).DoAndReturn(func(context.Context, executor.ExecuteRequest) (*executor.ExecuteResponse, error) {
		close(started)
		<-release

		return &executor.ExecuteResponse{Response: "ready", Code: 200}, nil
	})
	sdk := NewMockSDK(controller)
	sdk.EXPECT().SetTitle("button-1", "ready", 0)
	sdk.EXPECT().SetImage("button-1", "", 0)
	subject := instance.NewInstance("button-1", requestExecutor, sdk, NewMockBrowserOpener(controller))
	require.NoError(t, subject.SetConfig(configValue(t, common.Config{Action: common.ActionRefreshRequest})))
	firstResult := make(chan error, 1)

	go func() {
		firstResult <- subject.KeyPressed(context.Background())
	}()
	<-started
	secondErr := subject.KeyPressed(context.Background())
	close(release)
	firstErr := <-firstResult

	require.NoError(t, secondErr)
	require.NoError(t, firstErr)
}

func TestInstanceKeyPressedReleasesFlagAfterFailure(t *testing.T) {
	controller := gomock.NewController(t)
	requestExecutor := NewMockExecutor(controller)
	requestExecutor.EXPECT().Execute(gomock.Any(), gomock.Any()).Return(nil, errors.New("request failed")).Times(2)
	sdk := NewMockSDK(controller)
	sdk.EXPECT().ShowAlert("button-1").Times(2)
	subject := instance.NewInstance("button-1", requestExecutor, sdk, NewMockBrowserOpener(controller))
	require.NoError(t, subject.SetConfig(configValue(t, common.Config{Action: common.ActionRefreshRequest})))

	firstErr := subject.KeyPressed(context.Background())
	secondErr := subject.KeyPressed(context.Background())

	require.ErrorContains(t, firstErr, "request failed")
	require.ErrorContains(t, secondErr, "request failed")
}

func TestInstancePollingDisabled(t *testing.T) {
	controller := gomock.NewController(t)
	subject := instance.NewInstance("button-1", NewMockExecutor(controller), NewMockSDK(controller), NewMockBrowserOpener(controller))
	require.NoError(t, subject.SetConfig(configValue(t, common.Config{IntervalSeconds: 0})))

	subject.StartAsync(context.Background())
	subject.Stop()
}

func TestInstancePollingEnabled(t *testing.T) {
	controller := gomock.NewController(t)
	requestExecutor := NewMockExecutor(controller)
	called := make(chan struct{})
	requestExecutor.EXPECT().Execute(gomock.Any(), gomock.Any()).DoAndReturn(func(context.Context, executor.ExecuteRequest) (*executor.ExecuteResponse, error) {
		close(called)

		return &executor.ExecuteResponse{Response: "ready", Code: 200}, nil
	})
	sdk := NewMockSDK(controller)
	sdk.EXPECT().SetTitle("button-1", "ready", 0)
	sdk.EXPECT().SetImage("button-1", "", 0)
	subject := instance.NewInstance("button-1", requestExecutor, sdk, NewMockBrowserOpener(controller))
	require.NoError(t, subject.SetConfig(configValue(t, common.Config{IntervalSeconds: 60})))

	subject.StartAsync(context.Background())
	<-called
	subject.Stop()
}

func TestInstanceStopCancelsAction(t *testing.T) {
	controller := gomock.NewController(t)
	actionExecutor := NewMockExecutor(controller)
	started := make(chan struct{})
	actionExecutor.EXPECT().ExecuteAction(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, _ executor.ExecuteActionRequest) (*executor.ExecuteActionResponse, error) {
		close(started)
		<-ctx.Done()

		return nil, ctx.Err()
	})
	sdk := NewMockSDK(controller)
	sdk.EXPECT().ShowAlert("button-1")
	subject := instance.NewInstance("button-1", actionExecutor, sdk, NewMockBrowserOpener(controller))
	require.NoError(t, subject.SetConfig(configValue(t, common.Config{Action: common.ActionExecuteLua})))
	subject.StartAsync(context.Background())
	result := make(chan error, 1)

	go func() {
		result <- subject.KeyPressed(context.Background())
	}()
	<-started
	subject.Stop()
	err := <-result

	require.ErrorIs(t, err, context.Canceled)
}

func configValue(t *testing.T, config common.Config) *fastjson.Value {
	data, err := json.Marshal(config)
	require.NoError(t, err)
	value, err := fastjson.ParseBytes(data)
	require.NoError(t, err)

	return value
}
