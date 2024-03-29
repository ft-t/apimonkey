package executor

import (
	"context"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/imroc/req/v3"
	"github.com/rs/zerolog"
	"github.com/tidwall/gjson"

	"github.com/ft-t/apimonkey/pkg/utils"
)

type Executor struct {
	executor ScriptExecutor
}

func NewExecutor(
	executor ScriptExecutor,
) *Executor {
	return &Executor{
		executor: executor,
	}
}

func (e *Executor) Execute(
	ctx context.Context,
	executeReq ExecuteRequest,
) (*ExecuteResponse, error) {
	httpReq := req.C().NewRequest()
	httpReq = httpReq.SetContext(ctx)
	httpReq.Method = executeReq.Config.MethodType

	apiUrl, err := utils.ExecuteTemplate(executeReq.Config.ApiUrl, executeReq.Config)
	if err != nil {
		return nil, err
	}
	httpReq = httpReq.SetURL(apiUrl)

	body, err := utils.ExecuteTemplate(executeReq.Config.Body, executeReq.Config)

	if !strings.EqualFold(httpReq.Method, "GET") && body != "" {
		httpReq = httpReq.SetBodyString(body)
	}

	for k, v := range executeReq.Config.Headers {
		val, renderErr := utils.ExecuteTemplate(v, executeReq.Config)
		if renderErr != nil {
			return nil, errors.Wrapf(renderErr, "error rendering header. input was %v", v)
		}

		httpReq = httpReq.SetHeader(k, val)
	}

	zerolog.Ctx(ctx).Trace().Str("url", apiUrl).Msg("sending request")
	resp := httpReq.Do(ctx)
	if resp == nil {
		return nil, errors.New("response is nil")
	}

	if resp.Err != nil {
		return nil, errors.Wrap(resp.Err, "error sending request")
	}

	value := resp.String()
	zerolog.Ctx(ctx).Debug().Str("response", value).Msg("got raw response")

	if strings.TrimSpace(executeReq.Config.ResponseJSONSelector) != "" {
		selectorVal := gjson.Get(value, executeReq.Config.ResponseJSONSelector)

		if selectorVal.Type == gjson.Null {
			return nil, errors.New("no data found by ResponseJSONSelector")
		}

		value = selectorVal.String()

		if value == "" {
			return nil, errors.New("empty value got from ResponseJSONSelector")
		}
	}

	if strings.TrimSpace(executeReq.Config.BodyScript) != "" {
		zerolog.Ctx(ctx).Trace().Str("script", executeReq.Config.BodyScript).Msg("executing script")

		scriptResult, scriptErr := e.executor.Execute(ctx, executeReq.Config.BodyScript, value, resp.StatusCode)
		if scriptErr != nil {
			return nil, errors.Wrap(scriptErr, "error executing script")
		}

		value = scriptResult

		zerolog.Ctx(ctx).Trace().Str("result", value).Msg("script executed")
	}

	zerolog.Ctx(ctx).Debug().
		Str("response", value).
		Str("selector", executeReq.Config.ResponseJSONSelector).
		Msg("post script processing")

	zerolog.Ctx(ctx).Debug().Str("final_result", value).Msgf("final")

	return &ExecuteResponse{
		Response: value,
		Code:     resp.StatusCode,
	}, nil
}
