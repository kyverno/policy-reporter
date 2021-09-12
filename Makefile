GO ?= go
BUILD ?= build
REPO ?= kyverno/policy-reporter
IMAGE_TAG ?= 1.8.5
LD_FLAGS="-s -w"

all: build

.PHONY: clean
clean:
	rm -rf $(BUILD)

.PHONY: prepare
prepare:
	mkdir -p $(BUILD)

.PHONY: test
test:
	go test -v ./... -timeout=120s

.PHONY: coverage
coverage:
	go test -v ./... -covermode=count -coverprofile=coverage.out -timeout=120s

.PHONY: build
build: prepare
	CGO_ENABLED=0 $(GO) build -v -ldflags="-s -w" $(GOFLAGS) -o $(BUILD)/policyreporter .

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
