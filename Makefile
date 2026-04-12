.PHONY: build web-build web-dev test test-cover test-integration lint clean

BINARY := bin/mini-tmk-agent

web-build:
	cd web && npm ci && npm run build
	rm -rf internal/web/static/*
	cp -r web/dist/* internal/web/static/
	touch internal/web/static/.gitkeep

web-dev:
	cd web && npm run dev

build: web-build
	go build -o $(BINARY) ./cmd/mini-tmk-agent

GO_PACKAGES := $(shell go list ./... | grep -v /web/node_modules)

test:
	go test $(GO_PACKAGES) -v

test-cover:
	go test ./config/... ./internal/... -cover

test-integration:
	go test -tags integration ./tests/integration/... -v -timeout 60s

lint:
	golangci-lint run ./...

clean:
	rm -rf bin/
