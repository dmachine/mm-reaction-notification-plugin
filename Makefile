PLUGIN_ID    ?= com.github.reaction-notification
PLUGIN_VERSION ?= $(shell python3 -c "import json; print(json.load(open('plugin.json'))['version'])" 2>/dev/null || echo "0.1.0")

# Local Mattermost instance for `make deploy`
MM_SERVICESETTINGS_SITEURL ?= http://localhost:8065
MM_ADMIN_USERNAME          ?= admin
MM_ADMIN_PASSWORD          ?= admin

.PHONY: all server webapp bundle dist deploy test check-style clean

## Default target: build everything and produce the deployable .tar.gz
all: dist

## Build the Go server binary for all supported platforms
server:
	mkdir -p server/dist
	cd server && \
	  GOOS=linux   GOARCH=amd64 go build -trimpath -o dist/plugin-linux-amd64   . && \
	  GOOS=linux   GOARCH=arm64 go build -trimpath -o dist/plugin-linux-arm64   . && \
	  GOOS=darwin  GOARCH=amd64 go build -trimpath -o dist/plugin-darwin-amd64  . && \
	  GOOS=darwin  GOARCH=arm64 go build -trimpath -o dist/plugin-darwin-arm64  . && \
	  GOOS=windows GOARCH=amd64 go build -trimpath -o dist/plugin-windows-amd64.exe .

## Build the React webapp bundle
webapp:
	cd webapp && npm install && npm run build

## Assemble the plugin directory and create the .tar.gz archive
bundle: server webapp
	rm -rf dist/$(PLUGIN_ID)
	mkdir -p dist/$(PLUGIN_ID)/server/dist
	mkdir -p dist/$(PLUGIN_ID)/webapp/dist
	cp plugin.json dist/$(PLUGIN_ID)/
	cp server/dist/* dist/$(PLUGIN_ID)/server/dist/
	cp webapp/dist/main.js dist/$(PLUGIN_ID)/webapp/dist/
	cd dist && tar -czf $(PLUGIN_ID)-$(PLUGIN_VERSION).tar.gz $(PLUGIN_ID)/
	rm -rf dist/$(PLUGIN_ID)
	@echo "Bundle created: dist/$(PLUGIN_ID)-$(PLUGIN_VERSION).tar.gz"

## Alias for bundle
dist: bundle

## Upload the plugin to a running local Mattermost instance via the REST API
deploy: dist
	@echo "Uploading $(PLUGIN_ID)-$(PLUGIN_VERSION).tar.gz to $(MM_SERVICESETTINGS_SITEURL)"
	@TOKEN=$$(curl -s -X POST "$(MM_SERVICESETTINGS_SITEURL)/api/v4/users/login" \
	    -H "Content-Type: application/json" \
	    -d "{\"login_id\":\"$(MM_ADMIN_USERNAME)\",\"password\":\"$(MM_ADMIN_PASSWORD)\"}" \
	    -D /tmp/mm-login-headers.txt > /dev/null && \
	    grep -i "^Token:" /tmp/mm-login-headers.txt | awk '{print $$2}' | tr -d '\r'); \
	curl -s -X POST "$(MM_SERVICESETTINGS_SITEURL)/api/v4/plugins" \
	    -H "Authorization: Bearer $$TOKEN" \
	    -F "plugin=@dist/$(PLUGIN_ID)-$(PLUGIN_VERSION).tar.gz" \
	    -F "force=true" | python3 -m json.tool

## Run Go unit tests
test:
	cd server && go test ./...

## Lint Go code
check-style-go:
	cd server && go vet ./...

## Lint JS code
check-style-js:
	cd webapp && npm run lint

check-style: check-style-go check-style-js

## Remove all build artifacts
clean:
	rm -rf server/dist webapp/dist dist
