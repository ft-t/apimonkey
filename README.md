![build workflow](https://github.com/ft-t/apimonkey/actions/workflows/release.yaml/badge.svg?branch=master)
[![codecov](https://codecov.io/github/ft-t/apimonkey/graph/badge.svg?token=1DUN0Y78V4)](https://codecov.io/github/ft-t/apimonkey)
[![go-report](https://goreportcard.com/badge/github.com/ft-t/apimonkey?nocache=true)](https://goreportcard.com/report/github.com/ft-t/apimonkey)

# API Monkey

API Monkey is a Stream Deck plugin for sending HTTP requests and displaying the processed response on a key. It supports scheduled polling, on-demand actions, Go templates, JSON selection, Lua processing, and response-to-text or response-to-image mappings.

## Features

- Send `GET`, `POST`, `PUT`, `DELETE`, and `PATCH` requests.
- Configure headers, request bodies, and TLS certificate verification.
- Insert reusable parameters into URLs, bodies, headers, browser URLs, and title prefixes with Go templates.
- Extract values from JSON responses with GJSON selectors.
- Process responses with Lua 5.1 scripts and the bundled JSON module.
- Map processed values to key text or packaged PNG/SVG images.
- Poll in the background or refresh only when the key is pressed.
- Assign key presses to open a browser, refresh the request, or run a dedicated Lua action.
- Send HTTP requests from action Lua scripts.

## Supported platforms

- Windows 10 or later, x64
- macOS 10.15 or later, Apple silicon
- Stream Deck 6.9 or later

## Installation

Install [API Monkey from the Elgato Marketplace](https://marketplace.elgato.com/product/api-monkey-754d958d-1c8e-4734-bbc3-b01ab914e3bb), then add it to a Stream Deck profile.

For manual installation:

1. Download `com.ftt.apimonkey.sdPlugin.zip` from the [latest GitHub release](https://github.com/ft-t/apimonkey/releases/latest).
2. Extract the `com.ftt.apimonkey.sdPlugin` directory into the Stream Deck plugins directory:
   - Windows: `%APPDATA%\Elgato\StreamDeck\Plugins\`
   - macOS: `~/Library/Application Support/com.elgato.StreamDeck/Plugins/`
3. Restart Stream Deck.
4. Add **API Monkey** to a profile and open its Property Inspector.

## Quick start

Create an on-demand key that shows this repository's star count:

1. Set **Button Action** to `Refresh Request`.
2. Set **Request Type** to `GET`.
3. Set **API URL** to `https://api.github.com/repos/ft-t/apimonkey`.
4. Set **Polling** to `0`.
5. Set **JSON Selector** to `stargazers_count`.
6. Set **Title Prefix** to `Stars`.
7. Select **Save All**, then press the key.

## Documentation

- [Features, configuration, and examples](docs/usage.md)
- [Lua scripting reference](docs/lua.md)
- [Agent and contributor instructions](CLAUDE.md)

## License

[MIT](LICENSE)
