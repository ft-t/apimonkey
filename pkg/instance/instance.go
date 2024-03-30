package instance

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"sync"
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
	ctxID     string
	cfg       *common.Config
	mut       sync.Mutex
	ctx       context.Context
	ctxCancel context.CancelFunc
	executor  Executor
	sdk       SDK
}

func NewInstance(
	ctxID string,
	executor Executor,
	sdk SDK,
) *DefaultInstance {
	return &DefaultInstance{
		ctxID:    ctxID,
		mut:      sync.Mutex{},
		executor: executor,
		sdk:      sdk,
	}
}

func (i *DefaultInstance) SDK() SDK {
	return i.sdk
}

func (i *DefaultInstance) Executor() Executor {
	return i.executor
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

	i.cfg = &tempConfig

	return nil
}

func (i *DefaultInstance) ShowAlert() {
	i.sdk.ShowAlert(i.ctxID)
}

func (i *DefaultInstance) ShowOk() {
	i.sdk.ShowOk(i.ctxID)
}

func (i *DefaultInstance) StartAsync() {
	i.mut.Lock()
	defer i.mut.Unlock()

	i.stopWithoutLock()

	ctx, cancel := context.WithCancel(context.Background())
	i.ctx = ctx
	i.ctxCancel = cancel

	go i.run()
}

func (i *DefaultInstance) run() {
	ctx := i.ctx

	for ctx.Err() == nil {
		interval := 30
		if i.cfg.IntervalSeconds > 0 {
			interval = i.cfg.IntervalSeconds
		}

		newLogger := log.With().
			Str("id", uuid.NewString()).
			Str("ctxID", i.ctxID).
			Logger()

		innerCtx, innerCancel := context.WithCancel(ctx)
		innerCtx = newLogger.WithContext(innerCtx)

		i.ExecuteSingleRequest(innerCtx)
		innerCancel()

		time.Sleep(time.Duration(interval) * time.Second)
	}
}

func (i *DefaultInstance) ExecuteSingleRequest(
	ctx context.Context,
) {
	resp, err := i.executor.Execute(ctx, executor.ExecuteRequest{
		Config: *i.cfg,
	})
	if err != nil {
		zerolog.Ctx(ctx).Err(err).Msg("error executing request")
		i.ShowAlert()
		return
	}

	if handleErr := i.HandleResponse(ctx, resp); handleErr != nil {
		zerolog.Ctx(ctx).Err(handleErr).Msg("error handling response")
		i.ShowAlert()
		return
	}

	if i.cfg.ShowSuccessNotification {
		i.ShowOk()
	}
}

func (i *DefaultInstance) HandleResponse(
	ctx context.Context,
	response *executor.ExecuteResponse,
) error {
	var sb strings.Builder
	prefix, err := utils.ExecuteTemplate(i.cfg.TitlePrefix, i.cfg.TemplateParameters)
	if err != nil {
		return errors.Wrap(err, "failed to execute template on prefix")
	}

	if prefix != "" {
		sb.WriteString(strings.ReplaceAll(prefix, "\\n", "\n") + "\n")
	}

	if len(i.cfg.ResponseMapper) == 0 {
		sb.WriteString(response.Response)

		i.sdk.SetTitle(i.ctxID, sb.String(), 0)
		i.sdk.SetImage(i.ctxID, "", 0)

		return nil
	}

	def, defaultOk := i.cfg.ResponseMapper["*"]
	mapped, ok := i.cfg.ResponseMapper[response.Response]

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
}

func (i *DefaultInstance) stopWithoutLock() {
	if i.ctxCancel != nil {
		i.ctxCancel()
	}

	i.ctxCancel = nil
}

func (i *DefaultInstance) KeyPressed() error {
	targetUrl := i.cfg.BrowserUrl
	if targetUrl == "" {
		targetUrl = i.cfg.ApiUrl
	}

	targetUrl, err := utils.ExecuteTemplate(targetUrl, i.cfg.TemplateParameters)
	if err != nil {
		i.ShowAlert()
		return errors.Wrap(err, "failed to execute template")
	}

	if err = exec.Command("rundll32",
		"url.dll,FileProtocolHandler", targetUrl).Start(); err != nil {
		i.ShowAlert()
		return errors.Wrap(err, "failed to open url")
	}

	return nil
}
