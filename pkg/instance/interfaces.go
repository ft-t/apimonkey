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
	StartAsync()
	Stop()
	KeyPressed() error
}
