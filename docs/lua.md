# Lua scripting

API Monkey embeds [GopherLua](https://github.com/yuin/gopher-lua), a Lua 5.1 implementation. Every execution uses a new Lua state with the standard libraries and a preloaded `json` module. State and globals are not shared between executions.

Two script types have different inputs and return contracts:

- **Response Lua** transforms an HTTP response before response mapping.
- **Action Lua** runs when a key configured with `Execute Lua` is pressed.

## Shared runtime

Load the JSON module when needed:

```lua
local json = require("json")
```

Decode JSON:

```lua
local value, err = json.decode('{"status":"ready"}')

if err ~= nil then
  error(err)
end

return value.status
```

Encode a Lua value:

```lua
local value, err = json.encode({status = "ready", count = 3})

if err ~= nil then
  error(err)
end

return value
```

JSON arrays use Lua's one-based numeric indexes. Objects use string keys. Invalid JSON, sparse arrays, mixed table key types, and recursive tables return `nil` plus an error string.

Scripts inherit the key or plugin lifecycle context. Long-running Lua code and HTTP calls are cancelled when that context is cancelled.

The `http.request(options)` function is available to both response and action scripts. See [HTTP requests](#http-requests) for its complete contract.

## Response Lua

Response Lua is configured in **Lua Script** and runs in the request pipeline:

```text
HTTP response -> JSON selector -> response Lua -> response mapping -> key
```

### Globals

| Global | Type | Meaning |
| --- | --- | --- |
| `ResponseBody` | string | Selected response value, or raw response body when no JSON selector is configured. |
| `ResponseStatusCode` | number | HTTP response status code. |

The names are normal Lua globals, so `ResponseBody` and `_G.ResponseBody` are equivalent.

### Return behavior

Return one display value. API Monkey converts the value at the top of the Lua stack to text and passes it to response mapping.

Use strings, numbers, or booleans for predictable results. Returning no value produces Lua's `nil` string representation, which is rarely useful.

### Example: count active alerts

```lua
local json = require("json")
local data, err = json.decode(ResponseBody)

if err ~= nil then
  error(err)
end

local active = 0

for _, alert in ipairs(data.alerts) do
  if alert.state == "firing" and alert.labels.alertname ~= "Watchdog" then
    active = active + 1
  end
end

return active
```

### Example: include HTTP status

```lua
if ResponseStatusCode >= 200 and ResponseStatusCode < 300 then
  return "healthy"
end

return "HTTP " .. tostring(ResponseStatusCode)
```

### Example: handle missing fields

```lua
local json = require("json")
local data, err = json.decode(ResponseBody)

if err ~= nil then
  error(err)
end

if data.queue == nil or data.queue.pending == nil then
  return "unknown"
end

return data.queue.pending
```

## Action Lua

Action Lua is configured in **Action Lua Script** and runs when **Button Action** is `Execute Lua`.

### Globals

| Global | Type | Meaning |
| --- | --- | --- |
| `ButtonContextID` | string | Stream Deck context ID for the pressed key. |
| `ButtonConfig` | table | Snapshot of the key configuration. `actionScript` is deliberately removed. |
| `http` | table | API Monkey HTTP module. |

`ButtonConfig` uses the saved JSON field names:

| Field | Type |
| --- | --- |
| `action` | string |
| `apiUrl` | string |
| `browserUrl` | string |
| `intervalSeconds` | number |
| `responseJSONSelector` | string |
| `responseMapper` | table |
| `headers` | table |
| `parameters` | table |
| `titlePrefix` | string |
| `bodyScript` | string |
| `showSuccessNotification` | boolean |
| `insecureSkipVerify` | boolean |
| `methodType` | string |
| `body` | string |

### Return behavior

| Return | Result |
| --- | --- |
| string, number, or boolean | Converted to text, then passed through title prefix and response mapping. |
| `nil` | Key remains unchanged. |
| table, function, thread, userdata, or channel | Script fails with an unsupported return type error. |

Only one action can run for a key at a time. Presses received while it is running are ignored.

## HTTP requests

Both response and action Lua can send a synchronous HTTP request with `http.request(options)`:

```lua
local response = http.request({
  method = "POST",
  url = "https://example.test/hooks/build",
  headers = {
    Authorization = "Bearer token",
    ["Content-Type"] = "application/json"
  },
  body = '{"ref":"main"}',
  timeout_seconds = 10
})
```

Request options:

| Field | Required | Type | Meaning |
| --- | --- | --- | --- |
| `method` | yes | non-empty string | HTTP method. |
| `url` | yes | non-empty string | Absolute request URL. |
| `headers` | no | table of string keys and string values | Request headers. |
| `body` | no | string | Request body; defaults to empty. |
| `timeout_seconds` | no | positive number | Per-request timeout. |

Response fields:

| Field | Type | Meaning |
| --- | --- | --- |
| `status_code` | number | HTTP status code. |
| `headers` | table of string arrays | Response headers. Each header can contain multiple values. |
| `body` | string | Complete response body. |

HTTP status codes such as `404` or `500` are returned normally. Invalid options, malformed URLs, transport failures, timeouts, cancellation, and response read failures raise Lua errors.

`http.request` uses the plugin's standard HTTP client. The key's **Ignore certificate errors** setting does not change Lua HTTP verification.

### Example: trigger a configured endpoint

```lua
local response = http.request({
  method = ButtonConfig.methodType,
  url = ButtonConfig.apiUrl,
  headers = ButtonConfig.headers,
  body = ButtonConfig.body,
  timeout_seconds = 10
})

return response.status_code
```

### Example: read a JSON response

```lua
local json = require("json")
local response = http.request({
  method = "GET",
  url = ButtonConfig.apiUrl,
  timeout_seconds = 5
})
local data, err = json.decode(response.body)

if err ~= nil then
  error(err)
end

if response.status_code ~= 200 then
  return "HTTP " .. tostring(response.status_code)
end

return data.status
```

### Example: preserve the current key display

```lua
local response = http.request({
  method = "POST",
  url = ButtonConfig.apiUrl,
  timeout_seconds = 10
})

if response.status_code == 202 then
  return nil
end

return response.status_code
```

## Errors and debugging

Any Lua error fails the request or action and triggers the Stream Deck alert indicator. Runtime and HTTP errors are written to `logs/log.log` inside the installed plugin directory.

Use Lua's `error` function to stop with a useful message:

```lua
local data, err = json.decode(ResponseBody)

if err ~= nil then
  error("invalid API response: " .. err)
end
```

Common failures:

| Error | Cause |
| --- | --- |
| `method is required` | `http.request` omitted a non-empty string method. |
| `url is required` | `http.request` omitted a non-empty string URL. |
| `headers must be a table` | `headers` was not a Lua table. |
| `headers must contain string keys and values` | A header name or value was not a string. |
| `timeout_seconds must be positive` | Timeout was zero, negative, or not numeric. |
| `unsupported Lua action return type: table` | Action returned a non-scalar value. |
| `context canceled` | Key or plugin lifecycle ended while the script was running. |
