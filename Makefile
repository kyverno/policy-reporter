GO ?= go

BUILD ?= build

all: build

.PHONY: clean
clean:
	rm -rf $(BUILD)

.PHONY: prepare
prepare:
	mkdir -p $(BUILD)

.PHONY: test
test:
	go test -v ./...

.PHONY: coverage
coverage:
	go test -v ./... -covermode=count -coverprofile=coverage.out

.PHONY: build
build: prepare
	CGO_ENABLED=0 $(GO) build -v -ldflags="-s -w" $(GOFLAGS) -o $(BUILD)/policyreporter .
