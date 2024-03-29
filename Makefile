.PHONY: build-apimonkey-windows
build-apimonkey-windows:
	@cd cmd/client/apimonkey && rm -rf dist && GOOS=windows go build -o dist/com.ftt.apimonkey.exe
	@cd cmd/client/apimonkey/resources && cp -a . ../dist/

.PHONY: dev-apimonkey
dev-apimonkey: build-apimonkey-windows
	@cd /mnt/c/Users/iqpir/AppData/Roaming/Elgato/StreamDeck/Plugins/com.ftt.apimonkey.sdPlugin/ && rm -rf *
	@cd cmd/client/apimonkey/dist/ && cp -a . /mnt/c/Users/iqpir/AppData/Roaming/Elgato/StreamDeck/Plugins/com.ftt.apimonkey.sdPlugin/ -f

.PHONY: tidy-modules
tidy-modules:
	@find . -type d \( -name build -prune \) -o -name go.mod -print | while read -r gomod_path; do \
		dir_path=$$(dirname "$$gomod_path"); \
		echo "Executing 'go mod tidy' in directory: $$dir_path"; \
		(cd "$$dir_path" && GOPROXY=$(GOPROXY) go mod tidy) || exit 1; \
	done

.PHONY: lint
lint:
	golangci-lint run