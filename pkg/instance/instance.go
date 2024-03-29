package instance

import (
	"encoding/json"

	"github.com/cockroachdb/errors"
	"github.com/valyala/fastjson"
	"meow.tf/streamdeck/sdk"
)

type Instance struct {
	ctxID string
	cfg   *Config
}

type Config struct {
}

func NewInstance(
	ctxID string,
) *Instance {
	return &Instance{
		ctxID: ctxID,
	}
}

func (i *Instance) SetConfig(payload *fastjson.Value) error {
	settingsBytes := payload.MarshalTo(nil)
	var tempConfig Config

	if err := json.Unmarshal(settingsBytes, &tempConfig); err != nil {
		return errors.Wrap(err, "failed to unmarshal settings")
	}

	i.cfg = &tempConfig

	return nil
}

func (i *Instance) ShowAlert() {
	sdk.ShowAlert(i.ctxID)
}

func (i *Instance) StartAsync() {

}
