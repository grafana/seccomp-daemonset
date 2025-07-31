GO ?= go

BIN_DIR := bin

.PHONY: build
build: $(BIN_DIR)
	CGO_ENABLED=0 $(GO) build -ldflags="-extldflags=-static" -o $(BIN_DIR)/seccomp-operator .

.PHONY: test
test:
	$(GO) test ./...

.PHONY: clean
clean:
	rm -rf $(BIN_DIR)

$(BIN_DIR):
	mkdir -p $(BIN_DIR)
