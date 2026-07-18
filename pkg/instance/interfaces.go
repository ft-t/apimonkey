package instance

import (
	"context"

	"github.com/valyala/fastjson"

	"github.com/ft-t/apimonkey/pkg/executor"
)

//go:generate mockgen -destination interfaces_mocks_test.go -package instance_test -source=interfaces.go

type Executor interface {
	Execute(
		ctx context.Context,
		executeReq executor.ExecuteRequest,
	) (*executor.ExecuteResponse, error)
	ExecuteAction(
		ctx context.Context,
		executeReq executor.ExecuteActionRequest,
	) (*executor.ExecuteActionResponse, error)
}

type BrowserOpener interface {
	Open(url string) error
}

type BrowserOpenerFunc func(url string) error

func (f BrowserOpenerFunc) Open(url string) error {
	return f(url)
}

type ImageReader interface {
	ReadFile(filename string) ([]byte, error)
}

type ImageReaderFunc func(filename string) ([]byte, error)

func (f ImageReaderFunc) ReadFile(filename string) ([]byte, error) {
	return f(filename)
}

type SDK interface {
	ShowAlert(ctxID string)
	ShowOk(ctxID string)
	SetTitle(ctxID string, title string, target int)
	SetImage(ctxID string, imageData string, target int)
}

type Factory interface {
	Create(ctxID string) Instance
}

type Instance interface {
	SetConfig(payload *fastjson.Value) error
	StartAsync(ctx context.Context)
	Stop()
	KeyPressed(ctx context.Context) error
}
