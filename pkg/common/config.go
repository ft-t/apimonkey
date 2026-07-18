package common

import (
	"maps"

	"github.com/cockroachdb/errors"
)

type Action string

const (
	ActionOpenBrowser    Action = "openBrowser"
	ActionRefreshRequest Action = "refreshRequest"
	ActionExecuteLua     Action = "executeLua"
)

type Config struct {
	Action                  Action            `json:"action"`
	ActionScript            string            `json:"actionScript,omitempty"`
	ApiUrl                  string            `json:"apiUrl"`
	BrowserUrl              string            `json:"browserUrl"`
	IntervalSeconds         int               `json:"intervalSeconds"`
	ResponseJSONSelector    string            `json:"responseJSONSelector"`
	ResponseMapper          map[string]string `json:"responseMapper"`
	Headers                 map[string]string `json:"headers"`
	TemplateParameters      map[string]string `json:"parameters"`
	TitlePrefix             string            `json:"titlePrefix"`
	BodyScript              string            `json:"bodyScript"`
	ShowSuccessNotification bool              `json:"showSuccessNotification"`
	InsecureSkipVerify      bool              `json:"insecureSkipVerify"`
	MethodType              string            `json:"methodType"`
	Body                    string            `json:"body"`
}

func (c *Config) Validate() error {
	if c.Action == "" {
		c.Action = ActionRefreshRequest
		if c.BrowserUrl != "" {
			c.Action = ActionOpenBrowser
		}
	}

	switch c.Action {
	case ActionOpenBrowser, ActionRefreshRequest, ActionExecuteLua:
	default:
		return errors.Newf("unknown button action: %s", c.Action)
	}

	if c.IntervalSeconds < 0 {
		return errors.New("interval seconds cannot be negative")
	}

	return nil
}

func (c *Config) Clone() Config {
	cloned := *c
	cloned.ResponseMapper = maps.Clone(c.ResponseMapper)
	cloned.Headers = maps.Clone(c.Headers)
	cloned.TemplateParameters = maps.Clone(c.TemplateParameters)

	return cloned
}
