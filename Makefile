GO ?= go
BUILD ?= build
REPO ?= ghcr.io/kyverno/policy-reporter
IMAGE_TAG ?= 1.11.0
LD_FLAGS='-s -w -linkmode external -extldflags "-static"'

all: build

.PHONY: clean
clean:
	rm -rf $(BUILD)

.PHONY: prepare
prepare:
	mkdir -p $(BUILD)

.PHONY: test
test:
	go test -v ./... -timeout=10s

.PHONY: coverage
coverage:
	go test -v ./... -covermode=count -coverprofile=coverage.out -timeout=30s

.PHONY: build
build: prepare
	CGO_ENABLED=1 $(GO) build -v -ldflags="-s -w" $(GOFLAGS) -o $(BUILD)/policyreporter .

.PHONY: docker-build
docker-build:
	@docker buildx build --progress plane --platform linux/arm64,linux/amd64 --tag $(REPO):$(IMAGE_TAG) . --build-arg LD_FLAGS=$(LD_FLAGS)

.PHONY: docker-push
docker-push:
	@docker buildx build --progress plane --platform linux/arm64,linux/amd64 --tag $(REPO):$(IMAGE_TAG) . --build-arg LD_FLAGS=$(LD_FLAGS) --push
	@docker buildx build --progress plane --platform linux/arm64,linux/amd64 --tag $(REPO):latest . --build-arg LD_FLAGS=$(LD_FLAGS) --push

.PHONY: docker-push-dev
docker-push-dev:
	@docker buildx build --progress plane --platform linux/arm64,linux/amd64 --tag $(REPO):dev . --build-arg LD_FLAGS=$(LD_FLAGS) --push
