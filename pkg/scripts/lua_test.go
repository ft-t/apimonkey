package scripts_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ft-t/apimonkey/pkg/scripts"
)

var jsonString = `{"alerts":[{"labels":{"alertname":"Watchdog"}},{"labels":{"alertname":"CriticalAlert"}},{"labels":{"alertname":"AnotherAlert"}}]}`

func TestLua(t *testing.T) {
	executor := scripts.NewLua()

	resp, err := executor.Execute(context.TODO(), `
   	    message = "Hello, world!" .. tostring(_G.ResponseStatusCode)
		return message .. _G.ResponseBody`, jsonString, 200)
	assert.NoError(t, err)

	assert.Equal(t, "Hello, world!200{\"alerts\":[{\"labels\":{\"alertname\":\"Watchdog\"}},{\"labels\":{\"alertname\":\"CriticalAlert\"}},{\"labels\":{\"alertname\":\"AnotherAlert\"}}]}", resp)
}
