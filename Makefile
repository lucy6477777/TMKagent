.PHONY: build web-build web-dev test test-integration lint clean

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

test:
	go test ./... -v

test-integration:
	go test -tags integration ./tests/integration/... -v -timeout 60s

lint:
	golangci-lint run ./...

clean:
	rm -rf bin/
