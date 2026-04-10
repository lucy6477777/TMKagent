.PHONY: build test test-integration lint clean

BINARY := bin/mini-tmk-agent

build:
	go build -o $(BINARY) ./cmd/mini-tmk-agent

test:
	go test ./tests/unit/... -v

test-integration:
	go test -tags integration ./tests/integration/... -v -timeout 60s

lint:
	golangci-lint run ./...

clean:
	rm -rf bin/
