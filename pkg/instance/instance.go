package instance

import (
	"context"
	"encoding/json"
	"os/exec"
	"sync"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/valyala/fastjson"
	"meow.tf/streamdeck/sdk"

	"github.com/ft-t/apimonkey/pkg/common"
	"github.com/ft-t/apimonkey/pkg/utils"
)

type Instance struct {
	ctxID     string
	cfg       *common.Config
	mut       sync.Mutex
	ctx       context.Context
	ctxCancel context.CancelFunc
}

func NewInstance(
	ctxID string,
) *Instance {
	return &Instance{
		ctxID: ctxID,
		mut:   sync.Mutex{},
	}
}

func (i *Instance) SetConfig(payload *fastjson.Value) error {
	settingsBytes := payload.MarshalTo(nil)
	var tempConfig common.Config

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
	i.mut.Lock()
	defer i.mut.Unlock()

	i.stopWithoutLock()

	go i.run()
}

func (i *Instance) run() {
	ctx := i.ctx

	for ctx.Err() == nil {
		interval := 30
		if i.cfg.IntervalSeconds > 0 {
			interval = i.cfg.IntervalSeconds
		}

		time.Sleep(time.Duration(interval) * time.Second)
	}
}

func (i *Instance) Stop() {
	i.mut.Lock()
	defer i.mut.Unlock()

	i.stopWithoutLock()
}

func (i *Instance) stopWithoutLock() {
	if i.ctxCancel != nil {
		i.ctxCancel()
	}

	i.ctxCancel = nil
}

func (i *Instance) KeyPressed() error {
	targetUrl := i.cfg.BrowserUrl
	if targetUrl == "" {
		targetUrl = i.cfg.ApiUrl
	}

	targetUrl, err := utils.ExecuteTemplate(targetUrl, i.cfg)
	if err != nil {
		return errors.Wrap(err, "failed to execute template")
	}

	if err = exec.Command("rundll32",
		"url.dll,FileProtocolHandler", targetUrl).Start(); err != nil {
		return errors.Wrap(err, "failed to open url")
	}

	return nil
}
