# CLAUDE.md

## Project

API Monkey is a Go 1.26 Stream Deck plugin. It sends configured HTTP requests, processes responses, and updates Stream Deck keys. Releases package Windows x64 and macOS arm64 binaries with `resources/` into `com.ftt.apimonkey.sdPlugin.zip`.

Read [README.md](README.md), [docs/usage.md](docs/usage.md), and [docs/lua.md](docs/lua.md) before changing user-visible behavior.

## Repository map

- `main.go`: application wiring and Stream Deck event handlers.
- `pkg/common`: saved configuration and validation.
- `pkg/executor`: request templates, HTTP execution, JSON selection, and response Lua dispatch.
- `pkg/instance`: per-key lifecycle, polling, actions, response mapping, and Stream Deck updates.
- `pkg/scripts`: response Lua, action Lua, JSON module setup, and action HTTP API.
- `pkg/sdk`: Stream Deck SDK adapter.
- `pkg/utils`: browser, file, and Go template helpers.
- `resources/manifest.json`: plugin metadata and supported platforms.
- `resources/pi/pi.html`: Property Inspector fields and saved settings.
- `docs/usage.md`: canonical user-facing feature and configuration guide.
- `docs/lua.md`: canonical Lua scripting contract.

## Required workflow

- Run `gopls go_workspace` before working in this Go workspace.
- Read a package's `interfaces.go` before modifying that package.
- After first reading a Go file, inspect its `gopls go_file_context`.
- Use `gopls` for Go symbols, references, renames, diagnostics, and package APIs.
- Use Codebase Memory for architecture, call chains, change impact, and structural discovery.
- Prefer repository Makefile targets over direct tool commands when a matching target exists.
- Keep changes inside requested scope. Ask before changing public APIs, schemas, dependencies, or unrelated packages.
- Do not delete or rename existing assets, configuration, or fixtures to fix a failure.

## Code rules

- Keep code explicit, small, and boring. Avoid unrelated refactors.
- Public blocking or I/O methods accept `context.Context` first.
- Define minimal interfaces where consumed.
- Never pass dependencies as `nil` unless a constructor explicitly documents a nil sentinel.
- Use `github.com/golang/mock/gomock`. Generate mocks; never hand-edit generated mock files.
- Use the logger carried by `context.Context` through `zerolog.Ctx(ctx)`.
- Preserve error identity with wrapping. Use `errors.Is` and `errors.As`; never compare error strings.
- Handle every returned error. Log only documented best-effort cleanup failures.
- Tests must not perform real network, filesystem, clock, or random I/O. Database tests are the only real-I/O exception.
- Keep success and failure test tables separate. Do not branch on expected outcomes inside tests.

## Commands

```bash
make generate
make lint
go test -p 1 -timeout 60s ./...
go build ./...
```

Run `make generate` after interface changes. Before declaring work complete, `make lint`, tests for modified packages with `-p 1 -timeout 60s`, and `go build ./...` must pass.

`make build-apimonkey-windows` builds a Windows development package in `dist/`. `Dockerfile` is the cross-platform release packaging source of truth.

## Lua changes

Response and action Lua are separate contracts:

- Response Lua: `Lua.Execute`, globals `ResponseBody` and `ResponseStatusCode`.
- Action Lua: `Lua.ExecuteAction`, globals `ButtonContextID`, `ButtonConfig`, and `http`.

When changing Lua behavior:

1. Read `pkg/scripts/interfaces.go` and `pkg/scripts/lua.go` with gopls context.
2. Update `pkg/scripts/lua_test.go` without real HTTP calls.
3. Update [docs/lua.md](docs/lua.md) in the same change.
4. Update [docs/usage.md](docs/usage.md) when configuration or visible behavior changes.
5. Update `README.md` only when the high-level feature list, install steps, platform support, or quick start changes.

## Documentation rules

- Treat current code, tests, Property Inspector, manifest, and release workflow as source of truth.
- Keep README concise; link to detailed guides instead of duplicating them.
- Put configuration and end-user examples in `docs/usage.md`.
- Put all Lua globals, APIs, return rules, errors, and Lua examples in `docs/lua.md`.
- Keep examples executable against the documented contract. Never document planned behavior as available.
- Preserve existing images under `docs/` unless replacement is explicitly requested.

## Git

- Do not commit or push to `master`, `main`, `qa`, `uat`, or `Release/*` without explicit approval.
- Sign every commit with `git commit -S`. Never bypass signing.
- Do not add an AI co-author.
