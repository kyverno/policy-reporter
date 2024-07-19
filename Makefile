GO ?= go
BUILD ?= build
REPO ?= ghcr.io/kyverno/policy-reporter
IMAGE_TAG ?= 2.20.1
LD_FLAGS=-s -w -linkmode external -extldflags "-static"
PLATFORMS ?= linux/arm64,linux/amd64,linux/s390x

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
	go test -v ./... -covermode=count -coverprofile=coverage.out.tmp -timeout=30s
	cat coverage.out.tmp | grep -v "github.com/kyverno/policy-reporter/cmd/" | grep -v "github.com/kyverno/policy-reporter/main.go" | grep -v "github.com/kyverno/policy-reporter/pkg/crd/" > coverage.out
	rm coverage.out.tmp

.PHONY: build
build: prepare
	CGO_ENABLED=1 $(GO) build -v -ldflags="-s -w" $(GOFLAGS) -o $(BUILD)/policyreporter .

.PHONY: docker-build
docker-build:
	@docker buildx build --progress plane --platform $(PLATFORMS)  --tag $(REPO):$(IMAGE_TAG) . --build-arg LD_FLAGS='$(LD_FLAGS) -X main.Version=$(IMAGE_TAG)'

.PHONY: docker-push
docker-push:
	@docker buildx build --progress plane --platform $(PLATFORMS)  --tag $(REPO):$(IMAGE_TAG) . --build-arg LD_FLAGS='$(LD_FLAGS) -X main.Version=$(IMAGE_TAG)' --push
	@docker buildx build --progress plane --platform $(PLATFORMS)  --tag $(REPO):latest . --build-arg LD_FLAGS='$(LD_FLAGS) -X main.Version=$(IMAGE_TAG)' --push

.PHONY: docker-push-dev
docker-push-dev:
	@docker buildx build --progress plane --platform $(PLATFORMS)  --tag $(REPO):dev . --build-arg LD_FLAGS='$(LD_FLAGS) -X main.Version=$(IMAGE_TAG)-dev' --push

.PHONY: fmt
fmt:
	$(call print-target)
	@echo "Running gci"
	@go run github.com/daixiang0/gci@v0.9.1 write -s standard -s default -s "prefix(github.com/kyverno/policy-reporter)" .
	@echo "Running gofumpt"
	@go run mvdan.cc/gofumpt@v0.4.0 -w .
