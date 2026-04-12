.PHONY: build web-build web-dev test test-cover test-integration lint clean

BINARY := bin/mini-tmk-agent
GO_TEST_PKGS = $$(GOCACHE=/tmp/go-build-cache go list ./... | grep -v '/web/node_modules/')

web-build:
	cd web && npm ci && npm run build
	rm -rf internal/web/static/*
	cp -r web/dist/* internal/web/static/
	touch internal/web/static/.gitkeep

web-dev:
	cd web && npm run dev

build: web-build
	go build -o $(BINARY) ./cmd/mini-tmk-agent

test:
	PKGS="$(GO_TEST_PKGS)"; \
	GOCACHE=/tmp/go-build-cache go test $$PKGS -v

test-cover:
	PKGS="$(GO_TEST_PKGS)"; \
	GOCACHE=/tmp/go-build-cache go test $$PKGS -covermode=atomic -coverpkg=$$(echo $$PKGS | tr ' ' ',') -coverprofile=coverage.out
	GOCACHE=/tmp/go-build-cache go tool cover -func=coverage.out

test-integration:
	PKGS="$(GO_TEST_PKGS)"; \
	GOCACHE=/tmp/go-build-cache go test -tags integration $$PKGS ./tests/integration/... -v -timeout 60s

lint:
	golangci-lint run ./...

clean:
	rm -rf bin/
