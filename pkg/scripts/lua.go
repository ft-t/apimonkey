package scripts

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	lua "github.com/yuin/gopher-lua"
	luajson "layeh.com/gopher-json"

	"github.com/ft-t/apimonkey/pkg/common"
)

type Lua struct {
	httpClient HTTPClient
}

func NewLua(httpClient HTTPClient) *Lua {
	return &Lua{
		httpClient: httpClient,
	}
}

func (e *Lua) Execute(
	ctx context.Context,
	script string,
	rawBody string,
	statusCode int,
) (string, error) {
	l := e.newState(ctx)
	defer l.Close()

	l.SetGlobal("ResponseBody", lua.LString(rawBody))
	l.SetGlobal("ResponseStatusCode", lua.LNumber(statusCode))

	if err := l.DoString(script); err != nil {
		return "", scriptError(ctx, err)
	}

	vv := l.Get(-1)

	return vv.String(), nil
}

func (e *Lua) ExecuteAction(
	ctx context.Context,
	script string,
	buttonContextID string,
	config common.Config,
) (*string, error) {
	l := e.newState(ctx)
	defer l.Close()

	l.SetGlobal("ButtonContextID", lua.LString(buttonContextID))
	buttonConfig, err := luaButtonConfig(l, config)
	if err != nil {
		return nil, err
	}

	l.SetGlobal("ButtonConfig", buttonConfig)

	if err = l.DoString(script); err != nil {
		return nil, scriptError(ctx, err)
	}

	result := l.Get(-1)
	switch result.Type() {
	case lua.LTNil:
		return nil, nil
	case lua.LTString, lua.LTNumber, lua.LTBool:
		value := result.String()

		return &value, nil
	default:
		return nil, errors.Newf("unsupported Lua action return type: %s", result.Type())
	}
}

func (e *Lua) newState(ctx context.Context) *lua.LState {
	l := lua.NewState()
	l.SetContext(ctx)
	luajson.Preload(l)

	httpModule := l.NewTable()
	l.SetField(httpModule, "request", l.NewFunction(e.executeHTTPRequest))
	l.SetGlobal("http", httpModule)

	return l
}

func (e *Lua) executeHTTPRequest(l *lua.LState) int {
	options := l.CheckTable(1)
	method := requiredString(l, options, "method")
	url := requiredString(l, options, "url")
	body := optionalString(l, options, "body")

	ctx, cancel := requestContext(l, options)
	defer cancel()

	request, err := http.NewRequestWithContext(ctx, method, url, strings.NewReader(body))
	if err != nil {
		l.RaiseError("create HTTP request: %v", err)
	}

	setRequestHeaders(l, request, options.RawGetString("headers"))

	response, err := e.httpClient.Do(request)
	if err != nil {
		l.RaiseError("send HTTP request: %v", err)
	}

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		closeErr := response.Body.Close()
		l.RaiseError("read HTTP response body: %v", errors.CombineErrors(err, closeErr))
	}

	if err = response.Body.Close(); err != nil {
		l.RaiseError("close HTTP response body: %v", err)
	}

	result := l.NewTable()
	result.RawSetString("status_code", lua.LNumber(response.StatusCode))
	result.RawSetString("headers", responseHeaders(l, response.Header))
	result.RawSetString("body", lua.LString(responseBody))
	l.Push(result)

	return 1
}

func luaButtonConfig(l *lua.LState, config common.Config) (lua.LValue, error) {
	config = config.Clone()
	config.ActionScript = ""

	data, err := json.Marshal(config)
	if err != nil {
		return nil, errors.Wrap(err, "marshal button config")
	}

	value, err := luajson.Decode(l, data)
	if err != nil {
		return nil, errors.Wrap(err, "decode button config for Lua")
	}

	return value, nil
}

func requiredString(l *lua.LState, options *lua.LTable, name string) string {
	value := options.RawGetString(name)
	stringValue, ok := value.(lua.LString)
	if !ok || strings.TrimSpace(string(stringValue)) == "" {
		l.RaiseError("%s is required", name)
	}

	return string(stringValue)
}

func optionalString(l *lua.LState, options *lua.LTable, name string) string {
	value := options.RawGetString(name)
	if value == lua.LNil {
		return ""
	}

	stringValue, ok := value.(lua.LString)
	if !ok {
		l.RaiseError("%s must be a string", name)
	}

	return string(stringValue)
}

func requestContext(l *lua.LState, options *lua.LTable) (context.Context, context.CancelFunc) {
	value := options.RawGetString("timeout_seconds")
	if value == lua.LNil {
		return l.Context(), func() {}
	}

	timeout, ok := value.(lua.LNumber)
	if !ok || timeout <= 0 {
		l.RaiseError("timeout_seconds must be positive")
	}

	return context.WithTimeout(l.Context(), time.Duration(timeout*lua.LNumber(time.Second)))
}

func setRequestHeaders(l *lua.LState, request *http.Request, value lua.LValue) {
	if value == lua.LNil {
		return
	}

	headers, ok := value.(*lua.LTable)
	if !ok {
		l.RaiseError("headers must be a table")
	}

	headers.ForEach(func(key lua.LValue, value lua.LValue) {
		headerName, nameOK := key.(lua.LString)
		headerValue, valueOK := value.(lua.LString)
		if !nameOK || !valueOK {
			l.RaiseError("headers must contain string keys and values")
		}

		request.Header.Set(string(headerName), string(headerValue))
	})
}

func responseHeaders(l *lua.LState, headers http.Header) *lua.LTable {
	result := l.NewTable()
	for name, values := range headers {
		luaValues := l.NewTable()
		for _, value := range values {
			luaValues.Append(lua.LString(value))
		}

		result.RawSetString(name, luaValues)
	}

	return result
}

func scriptError(ctx context.Context, err error) error {
	if ctx.Err() != nil {
		return errors.WithStack(ctx.Err())
	}

	return errors.WithStack(err)
}
