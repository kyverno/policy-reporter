.DEFAULT_GOAL: build-all

############
# DEFAULTS #
############

GIT_SHA              := $(shell git rev-parse HEAD)
REGISTRY             ?= ghcr.io
REPO                 ?= kyverno
KIND_IMAGE           ?= kindest/node:v1.30.0
KIND_NAME            ?= kind
KIND_CONFIG          ?= default
GOOS                 ?= $(shell go env GOOS)
GOARCH               ?= $(shell go env GOARCH)
KOCACHE              ?= /tmp/ko-cache
BUILD_WITH           ?= ko
KYVERNOPRE_IMAGE     := kyvernopre
KYVERNO_IMAGE        := kyverno
CLI_IMAGE            := kyverno-cli
CLEANUP_IMAGE        := cleanup-controller
REPORTS_IMAGE        := reports-controller
BACKGROUND_IMAGE     := background-controller
REPO_KYVERNOPRE      := $(REGISTRY)/$(REPO)/$(KYVERNOPRE_IMAGE)
REPO_KYVERNO         := $(REGISTRY)/$(REPO)/$(KYVERNO_IMAGE)
REPO_CLI             := $(REGISTRY)/$(REPO)/$(CLI_IMAGE)
REPO_CLEANUP         := $(REGISTRY)/$(REPO)/$(CLEANUP_IMAGE)
REPO_REPORTS         := $(REGISTRY)/$(REPO)/$(REPORTS_IMAGE)
REPO_BACKGROUND      := $(REGISTRY)/$(REPO)/$(BACKGROUND_IMAGE)
USE_CONFIG           ?= standard
INSTALL_VERSION	     ?= 3.2.6

#########
# TOOLS #
#########

TOOLS_DIR                          ?= $(PWD)/.tools
KIND                               ?= $(TOOLS_DIR)/kind
KIND_VERSION                       ?= v0.23.0
CONTROLLER_GEN                     := $(TOOLS_DIR)/controller-gen
CONTROLLER_GEN_VERSION             ?= v0.16.1
CLIENT_GEN                         ?= $(TOOLS_DIR)/client-gen
LISTER_GEN                         ?= $(TOOLS_DIR)/lister-gen
INFORMER_GEN                       ?= $(TOOLS_DIR)/informer-gen
OPENAPI_GEN                        ?= $(TOOLS_DIR)/openapi-gen
REGISTER_GEN                       ?= $(TOOLS_DIR)/register-gen
DEEPCOPY_GEN                       ?= $(TOOLS_DIR)/deepcopy-gen
DEFAULTER_GEN                      ?= $(TOOLS_DIR)/defaulter-gen
APPLYCONFIGURATION_GEN             ?= $(TOOLS_DIR)/applyconfiguration-gen
CODE_GEN_VERSION                   ?= v0.28.0
GEN_CRD_API_REFERENCE_DOCS         ?= $(TOOLS_DIR)/gen-crd-api-reference-docs
GEN_CRD_API_REFERENCE_DOCS_VERSION ?= latest
GENREF                             ?= $(TOOLS_DIR)/genref
GENREF_VERSION                     ?= master
GO_ACC                             ?= $(TOOLS_DIR)/go-acc
GO_ACC_VERSION                     ?= latest
GOIMPORTS                          ?= $(TOOLS_DIR)/goimports
GOIMPORTS_VERSION                  ?= latest
HELM                               ?= $(TOOLS_DIR)/helm
HELM_VERSION                       ?= v3.12.3
HELM_DOCS                          ?= $(TOOLS_DIR)/helm-docs
HELM_DOCS_VERSION                  ?= v1.11.0
KO                                 ?= $(TOOLS_DIR)/ko
KO_VERSION                         ?= v0.17.1
KUBE_VERSION                       ?= v1.25.0
TOOLS                              := $(KIND) $(CONTROLLER_GEN) $(CLIENT_GEN) $(LISTER_GEN) $(INFORMER_GEN) $(OPENAPI_GEN) $(REGISTER_GEN) $(DEEPCOPY_GEN) $(DEFAULTER_GEN) $(APPLYCONFIGURATION_GEN) $(GEN_CRD_API_REFERENCE_DOCS) $(GENREF) $(GO_ACC) $(GOIMPORTS) $(HELM) $(HELM_DOCS) $(KO)
ifeq ($(GOOS), darwin)
SED                                := gsed
else
SED                                := sed
endif
COMMA                              := ,

$(KIND):
	@echo Install kind... >&2
	@GOBIN=$(TOOLS_DIR) go install sigs.k8s.io/kind@$(KIND_VERSION)

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

$(DEFAULTER_GEN):
	@echo Install defaulter-gen... >&2
	@GOBIN=$(TOOLS_DIR) go install k8s.io/code-generator/cmd/defaulter-gen@$(CODE_GEN_VERSION)

$(APPLYCONFIGURATION_GEN):
	@echo Install applyconfiguration-gen... >&2
	@GOBIN=$(TOOLS_DIR) go install k8s.io/code-generator/cmd/applyconfiguration-gen@$(CODE_GEN_VERSION)

$(GEN_CRD_API_REFERENCE_DOCS):
	@echo Install gen-crd-api-reference-docs... >&2
	@GOBIN=$(TOOLS_DIR) go install github.com/ahmetb/gen-crd-api-reference-docs@$(GEN_CRD_API_REFERENCE_DOCS_VERSION)

$(GENREF):
	@echo Install genref... >&2
	@GOBIN=$(TOOLS_DIR) go install github.com/kubernetes-sigs/reference-docs/genref@$(GENREF_VERSION)

$(GO_ACC):
	@echo Install go-acc... >&2
	@GOBIN=$(TOOLS_DIR) go install github.com/ory/go-acc@$(GO_ACC_VERSION)

$(GOIMPORTS):
	@echo Install goimports... >&2
	@GOBIN=$(TOOLS_DIR) go install golang.org/x/tools/cmd/goimports@$(GOIMPORTS_VERSION)

$(HELM):
	@echo Install helm... >&2
	@GOBIN=$(TOOLS_DIR) go install helm.sh/helm/v3/cmd/helm@$(HELM_VERSION)

$(HELM_DOCS):
	@echo Install helm-docs... >&2
	@GOBIN=$(TOOLS_DIR) go install github.com/norwoodj/helm-docs/cmd/helm-docs@$(HELM_DOCS_VERSION)

$(KO):
	@echo Install ko... >&2
	@GOBIN=$(TOOLS_DIR) go install github.com/google/ko@$(KO_VERSION)

.PHONY: install-tools
install-tools: $(TOOLS) ## Install tools

.PHONY: clean-tools
clean-tools: ## Remove installed tools
	@echo Clean tools... >&2
	@rm -rf $(TOOLS_DIR)

#################
# BUILD (LOCAL) #
#################

CMD_DIR        := cmd
KYVERNO_DIR    := $(CMD_DIR)/kyverno
KYVERNOPRE_DIR := $(CMD_DIR)/kyverno-init
CLI_DIR        := $(CMD_DIR)/cli/kubectl-kyverno
CLEANUP_DIR    := $(CMD_DIR)/cleanup-controller
REPORTS_DIR    := $(CMD_DIR)/reports-controller
BACKGROUND_DIR := $(CMD_DIR)/background-controller
KYVERNO_BIN    := $(KYVERNO_DIR)/kyverno
KYVERNOPRE_BIN := $(KYVERNOPRE_DIR)/kyvernopre
CLI_BIN        := $(CLI_DIR)/kubectl-kyverno
CLEANUP_BIN    := $(CLEANUP_DIR)/cleanup-controller
REPORTS_BIN    := $(REPORTS_DIR)/reports-controller
BACKGROUND_BIN := $(BACKGROUND_DIR)/background-controller
PACKAGE        ?= github.com/kyverno/policy-reporter
CGO_ENABLED    ?= 0
ifdef VERSION
LD_FLAGS       := "-s -w -X $(PACKAGE)/pkg/version.BuildVersion=$(VERSION)"
else
LD_FLAGS       := "-s -w"
endif

.PHONY: fmt
fmt: ## Run go fmt
	@echo Go fmt... >&2
	@go fmt ./...

.PHONY: vet
vet: ## Run go vet
	@echo Go vet... >&2
	@go vet ./...

.PHONY: imports
imports: $(GOIMPORTS)
	@echo Go imports... >&2
	@$(GOIMPORTS) -w .

.PHONY: fmt-check
fmt-check: fmt
	@echo Checking code format... >&2
	@git --no-pager diff .
	@echo 'If this test fails, it is because the git diff is non-empty after running "make fmt".' >&2
	@echo 'To correct this, locally run "make fmt" and commit the changes.' >&2
	@git diff --quiet --exit-code .

.PHONY: imports-check
imports-check: imports
	@echo Checking go imports... >&2
	@git --no-pager diff .
	@echo 'If this test fails, it is because the git diff is non-empty after running "make imports-check".' >&2
	@echo 'To correct this, locally run "make imports" and commit the changes.' >&2
	@git diff --quiet --exit-code .

.PHONY: unused-package-check
unused-package-check:
	@tidy=$$(go mod tidy); \
	if [ -n "$${tidy}" ]; then \
		echo "go mod tidy checking failed!"; echo "$${tidy}"; echo; \
	fi

$(KYVERNOPRE_BIN): fmt vet
	@echo Build kyvernopre binary... >&2
	@CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) \
		go build -o ./$(KYVERNOPRE_BIN) -ldflags=$(LD_FLAGS) ./$(KYVERNOPRE_DIR)

$(KYVERNO_BIN): fmt vet
	@echo Build kyverno binary... >&2
	@CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) \
		go build -o ./$(KYVERNO_BIN) -ldflags=$(LD_FLAGS) ./$(KYVERNO_DIR)

$(CLI_BIN): fmt vet
	@echo Build cli binary... >&2
	@CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) \
		go build -o ./$(CLI_BIN) -ldflags=$(LD_FLAGS) ./$(CLI_DIR)

$(CLEANUP_BIN): fmt vet
	@echo Build cleanup controller binary... >&2
	@CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) \
		go build -o ./$(CLEANUP_BIN) -ldflags=$(LD_FLAGS) ./$(CLEANUP_DIR)

$(REPORTS_BIN): fmt vet
	@echo Build reports controller binary... >&2
	@CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) \
		go build -o ./$(REPORTS_BIN) -ldflags=$(LD_FLAGS) ./$(REPORTS_DIR)

$(BACKGROUND_BIN): fmt vet
	@echo Build background controller binary... >&2
	@CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) \
		go build -o ./$(BACKGROUND_BIN) -ldflags=$(LD_FLAGS) ./$(BACKGROUND_DIR)

.PHONY: build-kyverno-init
build-kyverno-init: $(KYVERNOPRE_BIN) ## Build kyvernopre binary

.PHONY: build-kyverno
build-kyverno: $(KYVERNO_BIN) ## Build kyverno binary

.PHONY: build-cli
build-cli: $(CLI_BIN) ## Build cli binary

.PHONY: build-cleanup-controller
build-cleanup-controller: $(CLEANUP_BIN) ## Build cleanup controller binary

.PHONY: build-reports-controller
build-reports-controller: $(REPORTS_BIN) ## Build reports controller binary

.PHONY: build-background-controller
build-background-controller: $(BACKGROUND_BIN) ## Build background controller binary

build-all: build-kyverno-init build-kyverno build-cli build-cleanup-controller build-reports-controller build-background-controller ## Build all binaries

##############
# BUILD (KO) #
##############

LOCAL_PLATFORM      := linux/$(GOARCH)
KO_REGISTRY         := ko.local
ifndef VERSION
KO_TAGS             := $(GIT_SHA)
else ifeq ($(VERSION),main)
KO_TAGS             := $(GIT_SHA),latest
else
KO_TAGS             := $(GIT_SHA),$(subst /,-,$(VERSION))
endif

KO_CLI_REPO         := $(PACKAGE)/$(CLI_DIR)
KO_KYVERNOPRE_REPO  := $(PACKAGE)/$(KYVERNOPRE_DIR)
KO_KYVERNO_REPO     := $(PACKAGE)/$(KYVERNO_DIR)
KO_CLEANUP_REPO     := $(PACKAGE)/$(CLEANUP_DIR)
KO_REPORTS_REPO     := $(PACKAGE)/$(REPORTS_DIR)
KO_BACKGROUND_REPO  := $(PACKAGE)/$(BACKGROUND_DIR)

.PHONY: ko-build-kyverno-init
ko-build-kyverno-init: $(KO) ## Build kyvernopre local image (with ko)
	@echo Build kyvernopre local image with ko... >&2
	@LD_FLAGS=$(LD_FLAGS) KOCACHE=$(KOCACHE) KO_DOCKER_REPO=$(KO_REGISTRY) \
		$(KO) build ./$(KYVERNOPRE_DIR) --preserve-import-paths --tags=$(KO_TAGS) --platform=$(LOCAL_PLATFORM)

.PHONY: ko-build-kyverno
ko-build-kyverno: $(KO) ## Build kyverno local image (with ko)
	@echo Build kyverno local image with ko... >&2
	@LD_FLAGS=$(LD_FLAGS) KOCACHE=$(KOCACHE) KO_DOCKER_REPO=$(KO_REGISTRY) \
		$(KO) build ./$(KYVERNO_DIR) --preserve-import-paths --tags=$(KO_TAGS) --platform=$(LOCAL_PLATFORM)

.PHONY: ko-build-cli
ko-build-cli: $(KO) ## Build cli local image (with ko)
	@echo Build cli local image with ko... >&2
	@LD_FLAGS=$(LD_FLAGS) KOCACHE=$(KOCACHE) KO_DOCKER_REPO=$(KO_REGISTRY) \
		$(KO) build ./$(CLI_DIR) --preserve-import-paths --tags=$(KO_TAGS) --platform=$(LOCAL_PLATFORM)

.PHONY: ko-build-cleanup-controller
ko-build-cleanup-controller: $(KO) ## Build cleanup controller local image (with ko)
	@echo Build cleanup controller local image with ko... >&2
	@LD_FLAGS=$(LD_FLAGS) KOCACHE=$(KOCACHE) KO_DOCKER_REPO=$(KO_REGISTRY) \
		$(KO) build ./$(CLEANUP_DIR) --preserve-import-paths --tags=$(KO_TAGS) --platform=$(LOCAL_PLATFORM)

.PHONY: ko-build-reports-controller
ko-build-reports-controller: $(KO) ## Build reports controller local image (with ko)
	@echo Build reports controller local image with ko... >&2
	@LD_FLAGS=$(LD_FLAGS) KOCACHE=$(KOCACHE) KO_DOCKER_REPO=$(KO_REGISTRY) \
		$(KO) build ./$(REPORTS_DIR) --preserve-import-paths --tags=$(KO_TAGS) --platform=$(LOCAL_PLATFORM)

.PHONY: ko-build-background-controller
ko-build-background-controller: $(KO) ## Build background controller local image (with ko)
	@echo Build background controller local image with ko... >&2
	@LD_FLAGS=$(LD_FLAGS) KOCACHE=$(KOCACHE) KO_DOCKER_REPO=$(KO_REGISTRY) \
		$(KO) build ./$(BACKGROUND_DIR) --preserve-import-paths --tags=$(KO_TAGS) --platform=$(LOCAL_PLATFORM)

.PHONY: ko-build-all
ko-build-all: ko-build-kyverno-init ko-build-kyverno ko-build-cli ko-build-cleanup-controller ko-build-reports-controller ko-build-background-controller ## Build all local images (with ko)

################
# PUBLISH (KO) #
################

REGISTRY_USERNAME   ?= dummy
PLATFORMS           := all

.PHONY: ko-login
ko-login: $(KO)
	@$(KO) login $(REGISTRY) --username $(REGISTRY_USERNAME) --password $(REGISTRY_PASSWORD)

.PHONY: ko-publish-kyverno-init
ko-publish-kyverno-init: ko-login ## Build and publish kyvernopre image (with ko)
	@LD_FLAGS=$(LD_FLAGS) KOCACHE=$(KOCACHE) KO_DOCKER_REPO=$(REPO_KYVERNOPRE) \
		$(KO) build ./$(KYVERNOPRE_DIR) --bare --tags=$(KO_TAGS) --platform=$(PLATFORMS) \
		--image-annotation 'org.opencontainers.image.authors'='The Kyverno team','org.opencontainers.image.source'='github.com/kyverno/kyverno/commit/${GIT_SHA}','org.opencontainers.image.vendor'='Kyverno','org.opencontainers.image.url'='ghcr.io/kyverno/kyvernopre'

.PHONY: ko-publish-kyverno
ko-publish-kyverno: ko-login ## Build and publish kyverno image (with ko)
	@LD_FLAGS=$(LD_FLAGS) KOCACHE=$(KOCACHE) KO_DOCKER_REPO=$(REPO_KYVERNO) \
		$(KO) build ./$(KYVERNO_DIR) --bare --tags=$(KO_TAGS) --platform=$(PLATFORMS) \
		--image-annotation 'org.opencontainers.image.authors'='The Kyverno team','org.opencontainers.image.source'='github.com/kyverno/kyverno/commit/${GIT_SHA}','org.opencontainers.image.vendor'='Kyverno','org.opencontainers.image.url'='ghcr.io/kyverno/kyverno'

.PHONY: ko-publish-cli
ko-publish-cli: ko-login ## Build and publish cli image (with ko)
	@LD_FLAGS=$(LD_FLAGS) KOCACHE=$(KOCACHE) KO_DOCKER_REPO=$(REPO_CLI) \
		$(KO) build ./$(CLI_DIR) --bare --tags=$(KO_TAGS) --platform=$(PLATFORMS) \
		--image-annotation 'org.opencontainers.image.authors'='The Kyverno Team','org.opencontainers.image.source'='github.com/kyverno/kyverno/commit/${GIT_SHA}','org.opencontainers.image.vendor'='Kyverno','org.opencontainers.image.url'='ghcr.io/kyverno/kyverno-cli'

.PHONY: ko-publish-cleanup-controller
ko-publish-cleanup-controller: ko-login ## Build and publish cleanup controller image (with ko)
	@LD_FLAGS=$(LD_FLAGS) KOCACHE=$(KOCACHE) KO_DOCKER_REPO=$(REPO_CLEANUP) \
		$(KO) build ./$(CLEANUP_DIR) --bare --tags=$(KO_TAGS) --platform=$(PLATFORMS) \
		--image-annotation 'org.opencontainers.image.authors'='The Kyverno Team','org.opencontainers.image.source'='github.com/kyverno/kyverno/commit/${GIT_SHA}','org.opencontainers.image.vendor'='Kyverno','org.opencontainers.image.url'='ghcr.io/kyverno/cleanup-controller'

.PHONY: ko-publish-reports-controller
ko-publish-reports-controller: ko-login ## Build and publish reports controller image (with ko)
	@LD_FLAGS=$(LD_FLAGS) KOCACHE=$(KOCACHE) KO_DOCKER_REPO=$(REPO_REPORTS) \
		$(KO) build ./$(REPORTS_DIR) --bare --tags=$(KO_TAGS) --platform=$(PLATFORMS) \
		--image-annotation 'org.opencontainers.image.authors'='The Kyverno team','org.opencontainers.image.source'='github.com/kyverno/kyverno/commit/${GIT_SHA}','org.opencontainers.image.vendor'='Kyverno','org.opencontainers.image.url'='ghcr.io/kyverno/reports-controller'

.PHONY: ko-publish-background-controller
ko-publish-background-controller: ko-login ## Build and publish background controller image (with ko)
	@LD_FLAGS=$(LD_FLAGS) KOCACHE=$(KOCACHE) KO_DOCKER_REPO=$(REPO_BACKGROUND) \
		$(KO) build ./$(BACKGROUND_DIR) --bare --tags=$(KO_TAGS) --platform=$(PLATFORMS) \
		--image-annotation 'org.opencontainers.image.authors'='The Kyverno team','org.opencontainers.image.source'='github.com/kyverno/kyverno/commit/${GIT_SHA}','org.opencontainers.image.vendor'='Kyverno','org.opencontainers.image.url'='ghcr.io/kyverno/background-controller'

.PHONY: ko-publish-all
ko-publish-all: ko-publish-kyverno-init ko-publish-kyverno ko-publish-cli ko-publish-cleanup-controller ko-publish-reports-controller ko-publish-background-controller ## Build and publish all images (with ko)

#################
# BUILD (IMAGE) #
#################

LOCAL_REGISTRY         := $($(shell echo $(BUILD_WITH) | tr '[:lower:]' '[:upper:]')_REGISTRY)
LOCAL_CLI_REPO         := $($(shell echo $(BUILD_WITH) | tr '[:lower:]' '[:upper:]')_CLI_REPO)
LOCAL_KYVERNOPRE_REPO  := $($(shell echo $(BUILD_WITH) | tr '[:lower:]' '[:upper:]')_KYVERNOPRE_REPO)
LOCAL_KYVERNO_REPO     := $($(shell echo $(BUILD_WITH) | tr '[:lower:]' '[:upper:]')_KYVERNO_REPO)
LOCAL_CLEANUP_REPO     := $($(shell echo $(BUILD_WITH) | tr '[:lower:]' '[:upper:]')_CLEANUP_REPO)
LOCAL_REPORTS_REPO     := $($(shell echo $(BUILD_WITH) | tr '[:lower:]' '[:upper:]')_REPORTS_REPO)
LOCAL_BACKGROUND_REPO  := $($(shell echo $(BUILD_WITH) | tr '[:lower:]' '[:upper:]')_BACKGROUND_REPO)

.PHONY: image-build-kyverno-init
image-build-kyverno-init: $(BUILD_WITH)-build-kyverno-init

.PHONY: image-build-kyverno
image-build-kyverno: $(BUILD_WITH)-build-kyverno

.PHONY: image-build-cli
image-build-cli: $(BUILD_WITH)-build-cli

.PHONY: image-build-cleanup-controller
image-build-cleanup-controller: $(BUILD_WITH)-build-cleanup-controller

.PHONY: image-build-reports-controller
image-build-reports-controller: $(BUILD_WITH)-build-reports-controller

.PHONY: image-build-background-controller
image-build-background-controller: $(BUILD_WITH)-build-background-controller

.PHONY: image-build-all
image-build-all: $(BUILD_WITH)-build-all

###########
# CODEGEN #
###########

GOPATH_SHIM                 := ${PWD}/.gopath
PACKAGE_SHIM                := $(GOPATH_SHIM)/src/$(PACKAGE)
OUT_PACKAGE                 := $(PACKAGE)/pkg/crd/client/targetconfig
INPUT_DIRS                  := $(PACKAGE)/pkg/crd/api/targetconfig/v1alpha1
CLIENT_INPUT_DIRS           := $(PACKAGE)/pkg/crd/api/targetconfig/v1alpha1
CLIENTSET_PACKAGE           := $(OUT_PACKAGE)/clientset
LISTERS_PACKAGE             := $(OUT_PACKAGE)/listers
INFORMERS_PACKAGE           := $(OUT_PACKAGE)/informers
APPLYCONFIGURATIONS_PACKAGE := $(OUT_PACKAGE)/applyconfigurations
CRDS_PATH                   := ${PWD}/config/crds
INSTALL_MANIFEST_PATH       := ${PWD}/config/install-latest-testing.yaml
KYVERNO_CHART_VERSION       ?= v0.0.0
POLICIES_CHART_VERSION      ?= v0.0.0
APP_CHART_VERSION           ?= latest
KUBE_CHART_VERSION          ?= ">=1.25.0-0"

$(GOPATH_SHIM):
	@echo Create gopath shim... >&2
	@mkdir -p $(GOPATH_SHIM)

.INTERMEDIATE: $(PACKAGE_SHIM)
$(PACKAGE_SHIM): $(GOPATH_SHIM)
	@echo Create package shim... >&2
	@mkdir -p $(GOPATH_SHIM)/src/github.com/kyverno && ln -s -f ${PWD} $(PACKAGE_SHIM)

.PHONY: codegen-client-clientset
codegen-client-clientset: $(PACKAGE_SHIM) $(CLIENT_GEN) ## Generate clientset
	@echo Generate clientset... >&2
	@rm -rf $(CLIENTSET_PACKAGE) && mkdir -p $(CLIENTSET_PACKAGE)
	GOPATH=$(GOPATH_SHIM) $(CLIENT_GEN) \
		--go-header-file ./scripts/boilerplate.go.txt \
		--clientset-name versioned \
		--output-package $(CLIENTSET_PACKAGE) \
		--input-base "" \
		--input $(CLIENT_INPUT_DIRS)

.PHONY: codegen-client-listers
codegen-client-listers: $(PACKAGE_SHIM) $(LISTER_GEN) ## Generate listers
	@echo Generate listers... >&2
	@rm -rf $(LISTERS_PACKAGE) && mkdir -p $(LISTERS_PACKAGE)
	GOPATH=$(GOPATH_SHIM) $(LISTER_GEN) \
		--go-header-file ./scripts/boilerplate.go.txt \
		--output-package $(LISTERS_PACKAGE) \
		--input-dirs $(CLIENT_INPUT_DIRS)

.PHONY: codegen-client-informers
codegen-client-informers: $(PACKAGE_SHIM) $(INFORMER_GEN) ## Generate informers
	GOPATH=$(GOPATH_SHIM) $(INFORMER_GEN) \
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
codegen-register: $(PACKAGE_SHIM) $(REGISTER_GEN) ## Generate types registrations
	@echo Generate registration... >&2
	@GOPATH=$(GOPATH_SHIM) $(REGISTER_GEN) \
		--go-header-file=./scripts/boilerplate.go.txt \
		--input-dirs=$(INPUT_DIRS)

.PHONY: codegen-deepcopy
codegen-deepcopy: $(PACKAGE_SHIM) $(DEEPCOPY_GEN) ## Generate deep copy functions
	echo Generate deep copy functions... >&2
	GOPATH=$(GOPATH_SHIM) $(DEEPCOPY_GEN) \
		--go-header-file=./scripts/boilerplate.go.txt \
		--input-dirs=$(INPUT_DIRS) \
		--output-file-base=zz_generated.deepcopy

.PHONY: codegen-defaulters
codegen-defaulters: $(PACKAGE_SHIM) $(DEFAULTER_GEN) ## Generate defaulters
	@echo Generate defaulters... >&2
	@GOPATH=$(GOPATH_SHIM) $(DEFAULTER_GEN) --go-header-file=./scripts/boilerplate.go.txt --input-dirs=$(INPUT_DIRS)

.PHONY: codegen-applyconfigurations
codegen-applyconfigurations: $(PACKAGE_SHIM) $(APPLYCONFIGURATION_GEN) ## Generate apply configurations
	@echo Generate applyconfigurations... >&2
	@rm -rf $(APPLYCONFIGURATIONS_PACKAGE) && mkdir -p $(APPLYCONFIGURATIONS_PACKAGE)
	@GOPATH=$(GOPATH_SHIM) $(APPLYCONFIGURATION_GEN) \
		--go-header-file=./scripts/boilerplate.go.txt \
		--input-dirs=$(INPUT_DIRS) \
		--output-package $(APPLYCONFIGURATIONS_PACKAGE)