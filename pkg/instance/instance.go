package instance

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/valyala/fastjson"

	"github.com/ft-t/apimonkey/pkg/common"
	"github.com/ft-t/apimonkey/pkg/executor"
	"github.com/ft-t/apimonkey/pkg/utils"
)

type DefaultInstance struct {
	ctxID          string
	cfg            common.Config
	mut            sync.Mutex
	parentCtx      context.Context
	ctx            context.Context
	ctxCancel      context.CancelFunc
	started        bool
	actionInFlight atomic.Bool
	executor       Executor
	sdk            SDK
	browser        BrowserOpener
}

func NewInstance(
	ctxID string,
	executor Executor,
	sdk SDK,
	browserOpener BrowserOpener,
) *DefaultInstance {
	return &DefaultInstance{
		ctxID:    ctxID,
		mut:      sync.Mutex{},
		executor: executor,
		sdk:      sdk,
		browser:  browserOpener,
	}
}

func (i *DefaultInstance) SDK() SDK {
	return i.sdk
}

func (i *DefaultInstance) Executor() Executor {
	return i.executor
}

func (i *DefaultInstance) BrowserOpener() BrowserOpener {
	return i.browser
}

func (i *DefaultInstance) ContextID() string {
	return i.ctxID
}

func (i *DefaultInstance) SetConfig(payload *fastjson.Value) error {
	settingsBytes := payload.MarshalTo(nil)
	var tempConfig common.Config

	if err := json.Unmarshal(settingsBytes, &tempConfig); err != nil {
		i.ShowAlert()
		return errors.Wrap(err, "failed to unmarshal settings")
	}

	if err := tempConfig.Validate(); err != nil {
		i.ShowAlert()
		return err
	}

	i.mut.Lock()
	i.cfg = tempConfig.Clone()
	if i.started {
		i.restartPollingWithoutLock()
	}
	i.mut.Unlock()

	return nil
}

func (i *DefaultInstance) ShowAlert() {
	i.sdk.ShowAlert(i.ctxID)
}

func (i *DefaultInstance) ShowOk() {
	i.sdk.ShowOk(i.ctxID)
}

func (i *DefaultInstance) StartAsync(parent context.Context) {
	i.mut.Lock()
	defer i.mut.Unlock()

	i.parentCtx = parent
	i.started = true
	i.restartPollingWithoutLock()
}

func (i *DefaultInstance) restartPollingWithoutLock() {
	i.stopWithoutLock()

	i.ctx, i.ctxCancel = context.WithCancel(i.parentCtx)
	if i.cfg.IntervalSeconds > 0 {
		go i.run(i.ctx)
	}
}

func (i *DefaultInstance) run(ctx context.Context) {
	for {
		config := i.configSnapshot()

		newLogger := log.With().
			Str("id", uuid.NewString()).
			Str("ctxID", i.ctxID).
			Logger()

		requestCtx := newLogger.WithContext(ctx)
		if err := i.executeSingleRequest(requestCtx, config); err != nil {
			zerolog.Ctx(requestCtx).Err(err).Msg("error refreshing request")
			i.ShowAlert()
		} else if config.ShowSuccessNotification {
			i.ShowOk()
		}

		timer := time.NewTimer(time.Duration(config.IntervalSeconds) * time.Second)
		select {
		case <-ctx.Done():
			if !timer.Stop() {
				<-timer.C
			}

			return
		case <-timer.C:
		}
	}
}

func (i *DefaultInstance) ExecuteSingleRequest(
	ctx context.Context,
) error {
	return i.executeSingleRequest(ctx, i.configSnapshot())
}

func (i *DefaultInstance) executeSingleRequest(
	ctx context.Context,
	config common.Config,
) error {
	resp, err := i.executor.Execute(ctx, executor.ExecuteRequest{
		Config: config,
	})
	if err != nil {
		return errors.WithStack(err)
	}

	return i.handleResponse(ctx, resp, config)
}

func (i *DefaultInstance) HandleResponse(
	ctx context.Context,
	response *executor.ExecuteResponse,
) error {
	return i.handleResponse(ctx, response, i.configSnapshot())
}

func (i *DefaultInstance) handleResponse(
	ctx context.Context,
	response *executor.ExecuteResponse,
	config common.Config,
) error {
	var sb strings.Builder
	prefix, err := utils.ExecuteTemplate(config.TitlePrefix, config.TemplateParameters)
	if err != nil {
		return errors.Wrap(err, "failed to execute template on prefix")
	}

	if prefix != "" {
		sb.WriteString(strings.ReplaceAll(prefix, "\\n", "\n") + "\n")
	}

	if len(config.ResponseMapper) == 0 {
		sb.WriteString(response.Response)

		i.sdk.SetTitle(i.ctxID, sb.String(), 0)
		i.sdk.SetImage(i.ctxID, "", 0)

		return nil
	}

	def, defaultOk := config.ResponseMapper["*"]
	mapped, ok := config.ResponseMapper[response.Response]

	if !ok && defaultOk {
		mapped = def
	}

	if mapped == "" {
		return errors.New("no mapping found")
	}

	if strings.HasSuffix(mapped, ".png") || strings.HasSuffix(mapped, ".svg") {
		if sb.Len() > 0 {
			i.sdk.SetTitle(i.ctxID, sb.String(), 0)
		}

		return i.handleImageMapping(ctx, mapped)
	} else {
		sb.WriteString(mapped)
		i.sdk.SetTitle(i.ctxID, sb.String(), 0)
		i.sdk.SetImage(i.ctxID, "", 0)
	}

	return nil
}

func (i *DefaultInstance) handleImageMapping(_ context.Context, mapped string) error {
	fileData, err := utils.ReadFile(mapped)

	if err != nil {
		return errors.Join(err, errors.New("image file not found"))
	}

	imageData := ""
	if strings.HasSuffix(mapped, ".png") {
		imageData = fmt.Sprintf("data:image/png;base64, %v", base64.StdEncoding.EncodeToString(fileData))
	} else if strings.HasSuffix(mapped, ".svg") {
		imageData = fmt.Sprintf("data:image/svg+xml;charset=utf8,%v", string(fileData))
	}

	i.sdk.SetImage(i.ctxID, imageData, 0)

	return nil
}

func (i *DefaultInstance) Stop() {
	i.mut.Lock()
	defer i.mut.Unlock()

	i.stopWithoutLock()
	i.parentCtx = nil
	i.ctx = nil
	i.started = false
}

func (i *DefaultInstance) stopWithoutLock() {
	if i.ctxCancel != nil {
		i.ctxCancel()
	}

	i.ctxCancel = nil
}

func (i *DefaultInstance) KeyPressed(ctx context.Context) error {
	if !i.actionInFlight.CompareAndSwap(false, true) {
		return nil
	}
	defer i.actionInFlight.Store(false)

	config, lifecycleCtx := i.actionSnapshot()
	actionCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	stopLifecycleCancellation := func() bool { return true }
	if lifecycleCtx != nil {
		stopLifecycleCancellation = context.AfterFunc(lifecycleCtx, cancel)
	}
	defer stopLifecycleCancellation()

	newLogger := log.With().
		Str("id", uuid.NewString()).
		Str("ctxID", i.ctxID).
		Logger()
	actionCtx = newLogger.WithContext(actionCtx)

	var err error
	switch config.Action {
	case common.ActionOpenBrowser:
		err = i.openBrowser(config)
	case common.ActionRefreshRequest:
		err = i.executeSingleRequest(actionCtx, config)
	case common.ActionExecuteLua:
		err = i.executeLuaAction(actionCtx, config)
	default:
		err = errors.Newf("unknown button action: %s", config.Action)
	}
	if err != nil {
		i.ShowAlert()
		return errors.WithStack(err)
	}

	if config.ShowSuccessNotification {
		i.ShowOk()
	}

	return nil
}

func (i *DefaultInstance) openBrowser(config common.Config) error {
	targetUrl := config.BrowserUrl
	if targetUrl == "" {
		targetUrl = config.ApiUrl
	}

	targetUrl, err := utils.ExecuteTemplate(targetUrl, config.TemplateParameters)
	if err != nil {
		return errors.Wrap(err, "failed to execute template")
	}

	if err = i.browser.Open(targetUrl); err != nil {
		return errors.Wrap(err, "open browser")
	}

	return nil
}

func (i *DefaultInstance) executeLuaAction(ctx context.Context, config common.Config) error {
	response, err := i.executor.ExecuteAction(ctx, executor.ExecuteActionRequest{
		ButtonContextID: i.ctxID,
		Config:          config,
	})
	if err != nil {
		return errors.WithStack(err)
	}

	if response.Value == nil {
		return nil
	}

	return i.handleResponse(ctx, &executor.ExecuteResponse{Response: *response.Value}, config)
}

func (i *DefaultInstance) configSnapshot() common.Config {
	i.mut.Lock()
	defer i.mut.Unlock()

	return i.cfg.Clone()
}

func (i *DefaultInstance) actionSnapshot() (common.Config, context.Context) {
	i.mut.Lock()
	defer i.mut.Unlock()

	return i.cfg.Clone(), i.ctx
}
