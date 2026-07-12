package common_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ft-t/apimonkey/pkg/common"
)

func TestConfigValidateSuccess(t *testing.T) {
	tests := []struct {
		name           string
		action         common.Action
		expectedAction common.Action
	}{
		{
			name:           "missing action defaults to browser",
			expectedAction: common.ActionOpenBrowser,
		},
		{
			name:           "browser action",
			action:         common.ActionOpenBrowser,
			expectedAction: common.ActionOpenBrowser,
		},
		{
			name:           "refresh action",
			action:         common.ActionRefreshRequest,
			expectedAction: common.ActionRefreshRequest,
		},
		{
			name:           "lua action",
			action:         common.ActionExecuteLua,
			expectedAction: common.ActionExecuteLua,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			config := common.Config{Action: test.action}

			err := config.Validate()

			require.NoError(t, err)
			assert.Equal(t, test.expectedAction, config.Action)
		})
	}
}

func TestConfigValidateFailure(t *testing.T) {
	tests := []struct {
		name          string
		config        common.Config
		expectedError string
	}{
		{
			name:          "unknown action",
			config:        common.Config{Action: common.Action("unknown")},
			expectedError: "unknown button action: unknown",
		},
		{
			name:          "negative interval",
			config:        common.Config{IntervalSeconds: -1},
			expectedError: "interval seconds cannot be negative",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			config := test.config

			err := config.Validate()

			require.EqualError(t, err, test.expectedError)
		})
	}
}

func TestConfigClone(t *testing.T) {
	config := common.Config{
		ResponseMapper:     map[string]string{"response": "mapped"},
		Headers:            map[string]string{"header": "value"},
		TemplateParameters: map[string]string{"parameter": "value"},
	}

	cloned := config.Clone()
	cloned.ResponseMapper["response"] = "changed"
	cloned.Headers["header"] = "changed"
	cloned.TemplateParameters["parameter"] = "changed"

	assert.Equal(t, "mapped", config.ResponseMapper["response"])
	assert.Equal(t, "value", config.Headers["header"])
	assert.Equal(t, "value", config.TemplateParameters["parameter"])
}
