GO ?= go

BIN_DIR := bin
BIN_NAME ?= seccomp-daemonset

.PHONY: build
build: $(BIN_DIR)
	CGO_ENABLED=0 $(GO) build -ldflags="-extldflags=-static" -o $(BIN_DIR)/$(BIN_NAME) .

.PHONY: test
test:
	$(GO) test ./...

.PHONY: fmt
fmt:
	$(GO) run golang.org/x/tools/cmd/goimports@v0.35.0 -w .
	go mod tidy

.PHONY: lint
lint:
	$(GO) run golang.org/x/tools/cmd/goimports@v0.35.0 -d .
	go mod tidy -diff
# Only run golangci-lint if it "CI" is not set. On CI, we will use a separate action for it.
	if [ -z "${CI}" ]; then golangci-lint run; fi

.PHONY: clean
clean:
	rm -rf $(BIN_DIR)

$(BIN_DIR):
	mkdir -p $(BIN_DIR)
