package executor

import "context"

type ScriptExecutor interface {
	Execute(ctx context.Context, script string, rawBody string, statusCode int) (string, error)
}
