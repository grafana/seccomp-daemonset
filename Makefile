GO ?= go

BIN_DIR := bin

.PHONY: build
build: $(BIN_DIR)
	CGO_ENABLED=0 $(GO) build -ldflags="-extldflags=-static" -o $(BIN_DIR)/seccomp-daemonset .

.PHONY: test
test:
	$(GO) test ./...

.PHONY: fmt
fmt:
	$(GO) run golang.org/x/tools/cmd/goimports@v0.35.0 -w .

.PHONY: lint
lint:
	$(GO) run golang.org/x/tools/cmd/goimports@v0.35.0 -d .
	golangci-lint run

.PHONY: clean
clean:
	rm -rf $(BIN_DIR)

$(BIN_DIR):
	mkdir -p $(BIN_DIR)
