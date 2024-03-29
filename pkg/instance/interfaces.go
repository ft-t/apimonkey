package instance

import (
	"context"

	"github.com/ft-t/apimonkey/pkg/executor"
)

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
	Create(ctxID string) *Instance
}
