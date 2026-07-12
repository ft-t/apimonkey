package executor

import (
	"context"

	"github.com/ft-t/apimonkey/pkg/common"
)

//go:generate mockgen -destination interfaces_mocks_test.go -package executor_test -source=interfaces.go

type ScriptExecutor interface {
	Execute(ctx context.Context, script string, rawBody string, statusCode int) (string, error)
	ExecuteAction(
		ctx context.Context,
		script string,
		buttonContextID string,
		config common.Config,
	) (*string, error)
}
