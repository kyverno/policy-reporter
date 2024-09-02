############
# DEFAULTS #
############

KIND_IMAGE           ?= kindest/node:v1.30.2
KIND_NAME            ?= kyverno
USE_CONFIG           ?= standard,no-ingress,in-cluster,all-read-rbac
KUBECONFIG           ?= ""
PIP                  ?= "pip3"
GO 					 ?= go
BUILD 				 ?= build
IMAGE_TAG 			 ?= 3.0.0

#############
# VARIABLES #
#############

GIT_SHA             := $(shell git rev-parse HEAD)
TIMESTAMP           := $(shell date '+%Y-%m-%d_%I:%M:%S%p')
GOOS                ?= $(shell go env GOOS)
GOARCH              ?= $(shell go env GOARCH)
REGISTRY            ?= ghcr.io
OWNER               ?= kyverno
KO_REGISTRY         := ko.local
IMAGE               ?= policy-reporter
LD_FLAGS            := -s -w -linkmode external -extldflags "-static"
LOCAL_PLATFORM      := linux/$(GOARCH)
PLATFORMS           := linux/arm64,linux/amd64,linux/s390x
REPO                := $(REGISTRY)/$(OWNER)/$(IMAGE)
COMMA               := ,

ifndef VERSION
APP_VERSION         := $(GIT_SHA)
else
APP_VERSION         := $(VERSION)
endif

#########
# TOOLS #
#########

TOOLS_DIR                     := $(PWD)/.tools
KIND                 		  := $(TOOLS_DIR)/kind
KIND_VERSION         		  := v0.24.0
KO             				  := $(TOOLS_DIR)/ko
KO_VERSION     				  := v0.15.1
HELM                          := $(TOOLS_DIR)/helm
HELM_VERSION                  := v3.10.1
HELM_DOCS                     := $(TOOLS_DIR)/helm-docs
HELM_DOCS_VERSION             := v1.11.0
GCI                           := $(TOOLS_DIR)/gci
GCI_VERSION                   := v0.9.1
GOFUMPT                       := $(TOOLS_DIR)/gofumpt
GOFUMPT_VERSION               := v0.4.0
TOOLS                         := $(HELM) $(HELM_DOCS) $(GCI) $(GOFUMPT)

$(HELM):
	@echo Install helm... >&2
	@GOBIN=$(TOOLS_DIR) go install helm.sh/helm/v3/cmd/helm@$(HELM_VERSION)

$(HELM_DOCS):
	@echo Install helm-docs... >&2
	@GOBIN=$(TOOLS_DIR) go install github.com/norwoodj/helm-docs/cmd/helm-docs@$(HELM_DOCS_VERSION)

$(GCI):
	@echo Install gci... >&2
	@GOBIN=$(TOOLS_DIR) go install github.com/daixiang0/gci@$(GCI_VERSION)

$(GOFUMPT):
	@echo Install gofumpt... >&2
	@GOBIN=$(TOOLS_DIR) go install mvdan.cc/gofumpt@$(GOFUMPT_VERSION)

$(KIND):
	@echo Install kind... >&2
	@GOBIN=$(TOOLS_DIR) go install sigs.k8s.io/kind@$(KIND_VERSION)

$(KO):
	@echo Install ko... >&2
	@GOBIN=$(TOOLS_DIR) go install github.com/google/ko@$(KO_VERSION)

.PHONY: gci
gci: $(GCI)
	@echo "Running gci"
	@$(GCI) write -s standard -s default -s "prefix(github.com/kyverno/policy-reporter)" .

.PHONY: gofumpt
gofumpt: $(GOFUMPT)
	@echo "Running gofumpt"
	@$(GOFUMPT) -w .

.PHONY: fmt
fmt: gci gofumpt

.PHONY: install-tools
install-tools: $(TOOLS) ## Install tools

.PHONY: clean-tools
clean-tools: ## Remove installed tools
	@echo Clean tools... >&2
	@rm -rf $(TOOLS_DIR)

########
# KIND #
########

.PHONY: kind-create-cluster
kind-create-cluster: $(KIND) ## Create kind cluster
	@echo Create kind cluster... >&2
	@$(KIND) create cluster --name $(KIND_NAME) --image $(KIND_IMAGE) --config ./scripts/kind.yaml
	@kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml
	@sleep 15
	@kubectl wait --namespace ingress-nginx --for=condition=ready pod --selector=app.kubernetes.io/component=controller --timeout=90s

.PHONY: kind-delete-cluster
kind-delete-cluster: $(KIND) ## Delete kind cluster
	@echo Delete kind cluster... >&2
	@$(KIND) delete cluster --name $(KIND_NAME)

.PHONY: kind-load
kind-load: $(KIND) ko-build ## Build playground image and load it in kind cluster
	@echo Load playground image... >&2
	@$(KIND) load docker-image --name $(KIND_NAME) ko.local/github.com/kyverno/policy-reporter:$(GIT_SHA)

###########
# CODEGEN #
###########

.PHONY: codegen-helm-docs
codegen-helm-docs: ## Generate helm docs
	@echo Generate helm docs... >&2
	@docker run -v ${PWD}/charts:/work -w /work jnorwood/helm-docs:v1.11.0 -s file

.PHONY: verify-helm-docs
verify-helm-docs: codegen-helm-docs ## Check Helm charts are up to date
	@echo Checking helm charts are up to date... >&2
	@git --no-pager diff -- charts
	@echo 'If this test fails, it is because the git diff is non-empty after running "make codegen-helm-docs".' >&2
	@echo 'To correct this, locally run "make codegen-helm-docs", commit the changes, and re-run tests.' >&2
	@git diff --quiet --exit-code -- charts

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
