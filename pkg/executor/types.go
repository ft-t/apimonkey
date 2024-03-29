package executor

import "github.com/ft-t/apimonkey/pkg/common"

type ExecuteRequest struct {
	Config common.Config
}

type ExecuteResponse struct {
	Response string
	Code     int
}
