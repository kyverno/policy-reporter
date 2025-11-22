############
# DEFAULTS #
############

KIND_IMAGE           ?= kindest/node:v1.33.1
KIND_NAME            ?= kind
USE_CONFIG           ?= default
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
PACKAGE             ?= github.com/kyverno/policy-reporter

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
KIND_VERSION                  := v0.29.0
KO             				  := $(TOOLS_DIR)/ko
KO_VERSION     				  := v0.18.0
HELM                          := $(TOOLS_DIR)/helm
HELM_VERSION                  := v3.19.0
HELM_DOCS                     := $(TOOLS_DIR)/helm-docs
HELM_DOCS_VERSION             := v1.14.2
GCI                           := $(TOOLS_DIR)/gci
GCI_VERSION                   := v0.13.7
GOFUMPT                       := $(TOOLS_DIR)/gofumpt
GOFUMPT_VERSION               := v0.9.1
TOOLS                         := $(HELM) $(HELM_DOCS) $(GCI) $(GOFUMPT)


CONTROLLER_GEN                     := $(TOOLS_DIR)/controller-gen
CONTROLLER_GEN_VERSION             ?= v0.16.1
CLIENT_GEN                         ?= $(TOOLS_DIR)/client-gen
LISTER_GEN                         ?= $(TOOLS_DIR)/lister-gen
INFORMER_GEN                       ?= $(TOOLS_DIR)/informer-gen
OPENAPI_GEN                        ?= $(TOOLS_DIR)/openapi-gen
REGISTER_GEN                       ?= $(TOOLS_DIR)/register-gen
DEEPCOPY_GEN                       ?= $(TOOLS_DIR)/deepcopy-gen
APPLYCONFIGURATION_GEN             ?= $(TOOLS_DIR)/applyconfiguration-gen
CODE_GEN_VERSION                   ?= v0.28.0

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

$(CONTROLLER_GEN):
	@echo Install controller-gen... >&2
	@cd ./hack/controller-gen && GOBIN=$(TOOLS_DIR) go install

$(CLIENT_GEN):
	@echo Install client-gen... >&2
	@GOBIN=$(TOOLS_DIR) go install k8s.io/code-generator/cmd/client-gen@$(CODE_GEN_VERSION)

$(LISTER_GEN):
	@echo Install lister-gen... >&2
	@GOBIN=$(TOOLS_DIR) go install k8s.io/code-generator/cmd/lister-gen@$(CODE_GEN_VERSION)

$(INFORMER_GEN):
	@echo Install informer-gen... >&2
	@GOBIN=$(TOOLS_DIR) go install k8s.io/code-generator/cmd/informer-gen@$(CODE_GEN_VERSION)

$(OPENAPI_GEN):
	@echo Install openapi-gen... >&2
	@GOBIN=$(TOOLS_DIR) go install k8s.io/code-generator/cmd/openapi-gen@$(CODE_GEN_VERSION)

$(REGISTER_GEN):
	@echo Install register-gen... >&2
	@GOBIN=$(TOOLS_DIR) go install k8s.io/code-generator/cmd/register-gen@$(CODE_GEN_VERSION)

$(DEEPCOPY_GEN):
	@echo Install deepcopy-gen... >&2
	@GOBIN=$(TOOLS_DIR) go install k8s.io/code-generator/cmd/deepcopy-gen@$(CODE_GEN_VERSION)

$(GEN_CRD_API_REFERENCE_DOCS):
	@echo Install gen-crd-api-reference-docs... >&2
	@GOBIN=$(TOOLS_DIR) go install github.com/ahmetb/gen-crd-api-reference-docs@$(GEN_CRD_API_REFERENCE_DOCS_VERSION)

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
kind-load: $(KIND) docker-build ## Build playground image and load it in kind cluster
	@echo Load playground image... >&2
	@$(KIND) load docker-image --name $(KIND_NAME) ko.local/$(PACKAGE):$(GIT_SHA)


.PHONY: kind-install
kind-install: $(HELM) ## Install kyverno helm chart
	@echo Install policy-reporter chart... >&2
	@$(HELM) upgrade --install policy-reporter --namespace policy-reporter --create-namespace --wait ./charts/policy-reporter \
		--set image.registry=ko.local \
		--set image.repository=$(PACKAGE) \
		--set image.tag=$(GIT_SHA) \
		$(foreach CONFIG,$(subst $(COMMA), ,$(USE_CONFIG)),--values ./scripts/config/$(CONFIG)/values.yaml) \

###########
# CODEGEN #
###########

.PHONY: codegen-static-manifests
codegen-static-manifests: $(HELM) ## Generate helm docs
	@echo Generate static manifests... >&2
	@$(HELM) template policy-reporter ./charts/policy-reporter \
		--set static=true \
		--set metrics.enabled=true \
		--set rest.enabled=true \
		-n policy-reporter \
		--create-namespace > manifests/policy-reporter/install.yaml
	@$(HELM) template policy-reporter ./charts/policy-reporter \
		--set static=true \
		--set metrics.enabled=true \
		--set ui.enabled=true \
		-n policy-reporter \
		--create-namespace > manifests/policy-reporter-ui/install.yaml
	@$(HELM) template policy-reporter ./charts/policy-reporter --set static=true \
		--set metrics.enabled=true \
		--set ui.enabled=true \
		--set plugin.kyverno.enabled=true \
		-n policy-reporter \
		--create-namespace > manifests/policy-reporter-kyverno-ui/install.yaml
	@$(HELM) template policy-reporter ./charts/policy-reporter \
		--set static=true \
		--set metrics.enabled=true \
		--set ui.enabled=true \
		--set plugin.kyverno.enabled=true \
		--set replicaCount=2 \
		--set ui.replicaCount=2 \
		--set plugin.kyverno.replicaCount=2 \
		-n policy-reporter \
		--create-namespace > manifests/policy-reporter-kyverno-ui-ha/install.yaml

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
	go test -v ./... -timeout=30s

.PHONY: coverage
coverage:
	go test -v ./... -covermode=count -coverprofile=coverage.out.tmp -timeout=30s
	cat coverage.out.tmp | grep -v "github.com/kyverno/policy-reporter/cmd/" \
		| grep -v "github.com/kyverno/policy-reporter/main.go" \
		| grep -v "github.com/kyverno/policy-reporter/pkg/crd/" \
		| grep -v "github.com/kyverno/policy-reporter/hack/main.go" \
		| grep -v "github.com/kyverno/policy-reporter/hack/controller-gen/" \
		| grep -v "github.com/kyverno/policy-reporter/pkg/database/" \
		| grep -v "github.com/kyverno/policy-reporter/pkg/target/provider/" \
		| grep -v "github.com/kyverno/policy-reporter/pkg/kubernetes/pods" \
		| grep -v "github.com/kyverno/policy-reporter/pkg/kubernetes/jobs" > coverage.out
	rm coverage.out.tmp

.PHONY: build
build: prepare
	CGO_ENABLED=1 $(GO) build -v -ldflags="-s -w" $(GOFLAGS) -o $(BUILD)/policyreporter .

.PHONY: docker-build
docker-build:
	@docker buildx build --progress plain --platform $(LOCAL_PLATFORM)  --tag $(KO_REGISTRY)/$(PACKAGE):$(GIT_SHA) . --build-arg LD_FLAGS='$(LD_FLAGS) -X main.Version=$(IMAGE_TAG)'

.PHONY: docker-push
docker-push:
	@docker buildx build --progress plain --platform $(PLATFORMS)  --tag $(REPO):$(IMAGE_TAG) . --build-arg LD_FLAGS='$(LD_FLAGS) -X main.Version=$(IMAGE_TAG)' --push
	@docker buildx build --progress plain --platform $(PLATFORMS)  --tag $(REPO):latest . --build-arg LD_FLAGS='$(LD_FLAGS) -X main.Version=$(IMAGE_TAG)' --push

.PHONY: docker-push-dev
docker-push-dev:
	@docker buildx build --progress plane --platform $(PLATFORMS)  --tag $(REPO):dev . --build-arg LD_FLAGS='$(LD_FLAGS) -X main.Version=$(IMAGE_TAG)-dev' --push

###########
# CODEGEN #
###########

OUT_PACKAGE                 := $(PACKAGE)/pkg/crd/client/targetconfig
INPUT_DIRS                  := $(PACKAGE)/pkg/crd/api/targetconfig/v1alpha1
CLIENT_INPUT_DIRS           := $(PACKAGE)/pkg/crd/api/targetconfig/v1alpha1
CLIENTSET_PACKAGE           := $(OUT_PACKAGE)/clientset
LISTERS_PACKAGE             := $(OUT_PACKAGE)/listers
INFORMERS_PACKAGE           := $(OUT_PACKAGE)/informers
CRDS_PATH                   := ${PWD}/config/crds

.PHONY: codegen-client-clientset
codegen-client-clientset: $(CLIENT_GEN) ## Generate clientset
	@echo Generate clientset... >&2
	@rm -rf $(CLIENTSET_PACKAGE) && mkdir -p $(CLIENTSET_PACKAGE)
	@$(CLIENT_GEN) \
		--go-header-file ./scripts/boilerplate.go.txt \
		--clientset-name versioned \
		--output-package $(CLIENTSET_PACKAGE) \
		--input-base "" \
		--input $(CLIENT_INPUT_DIRS)

.PHONY: codegen-client-listers
codegen-client-listers: $(LISTER_GEN) ## Generate listers
	@echo Generate listers... >&2
	@rm -rf $(LISTERS_PACKAGE) && mkdir -p $(LISTERS_PACKAGE)
	@$(LISTER_GEN) \
		--go-header-file ./scripts/boilerplate.go.txt \
		--output-package $(LISTERS_PACKAGE) \
		--input-dirs $(CLIENT_INPUT_DIRS)

.PHONY: codegen-client-informers
codegen-client-informers: $(INFORMER_GEN) ## Generate informers
	@$(INFORMER_GEN) \
		--go-header-file ./scripts/boilerplate.go.txt \
		--output-package $(INFORMERS_PACKAGE) \
		--input-dirs $(CLIENT_INPUT_DIRS) \
		--versioned-clientset-package $(CLIENTSET_PACKAGE)/versioned \
		--listers-package $(LISTERS_PACKAGE)

.PHONY: codegen-client-wrappers
codegen-client-wrappers: codegen-client-clientset $(GOIMPORTS) ## Generate client wrappers
	@echo Generate client wrappers... >&2
	@go run ./hack/main.go
	@$(GOIMPORTS) -w ./pkg/clients
	@go fmt ./pkg/clients/...

.PHONY: codegen-register
codegen-register: $(REGISTER_GEN) ## Generate types registrations
	@echo Generate registration... >&2
	@$(REGISTER_GEN) \
		--go-header-file=./scripts/boilerplate.go.txt \
		--input-dirs=$(INPUT_DIRS)

.PHONY: codegen-deepcopy
codegen-deepcopy: $(DEEPCOPY_GEN) ## Generate deep copy functions
	@echo Generate deep copy functions... >&2
	@$(DEEPCOPY_GEN) \
		--go-header-file=./scripts/boilerplate.go.txt \
		--input-dirs=$(INPUT_DIRS) \
		--output-file-base=zz_generated.deepcopy

.PHONY: codegen-crds
codegen-crds: $(CONTROLLER_GEN)
	@echo Generate policy reporter crds... >&2
	@rm -rf $(CRDS_PATH) && mkdir -p $(CRDS_PATH)
	@$(CONTROLLER_GEN) paths=./pkg/crd/api/targetconfig/... crd:crdVersions=v1,ignoreUnexportedFields=true,generateEmbeddedObjectMeta=false output:dir=$(CRDS_PATH)
