MAIN                 ?= main.go
GO                    = go
ALL_PKGS             := ./...
LOCALBIN             ?= $(shell pwd)/bin
BINDIR               ?= $(LOCALBIN)

MAIN_MODULE          := $(shell go list -m)
VERSION_PKG          ?= $(MAIN_MODULE)/pkg/version

GITHUB_ORGANIZATION  ?= $(shell echo $(MAIN_MODULE) | cut -d "/" -f2)
PROJECT_NAME         ?= $(shell echo $(MAIN_MODULE) | cut -d "/" -f3)


VERSION   ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "v0.0.0")
COMMIT     ?= $(strip $(shell git rev-parse --short HEAD 2>&1 || echo "dev"))
BRANCH     ?= $(strip $(shell git symbolic-ref --short HEAD 2>/dev/null || git rev-parse --short HEAD 2>/dev/null || echo "detached"))
BUILD_DATE = $(shell date +%FT%T%z)

# DOCKER_REPO_HOST defines the host of the docker registry to which images will be pushed.
DOCKER_REPO_HOST ?= ghcr.io

# DOCKER_REPO_BASE defines the base of the docker registry to which images will be pushed.
DOCKER_REPO_BASE ?= $(DOCKER_REPO_HOST)/$(GITHUB_ORGANIZATION)

# DOCKER_REPO_NAME defines the name of the docker repository to which images will be pushed.
# The default value is set to $(PROJECT_NAME), which will result in pushing to
# ghcr.io/$(GITHUB_ORGANIZATION)/$(PROJECT_NAME) by default.
# You can change this value to push to a different repository or registry.
DOCKER_REPO_NAME ?= $(PROJECT_NAME)

# DOCKER_TAG_BASE defines the base tag for the docker image,
# which is a combination of the repository host, base, and name.
# The default value is set to $(DOCKER_REPO_BASE)/$(DOCKER_REPO_NAME),
# which will result in a base tag of
# ghcr.io/$(GITHUB_ORGANIZATION)/$(PROJECT_NAME) by default.
# You can change this value to use a different base tag for your docker images.
DOCKER_TAG_BASE ?= $(DOCKER_REPO_BASE)/$(DOCKER_REPO_NAME)

# Image URL to use all building/pushing image targets
IMG ?= $(DOCKER_TAG_BASE):$(VERSION)$(GIT_TAG)

# DOCKER_BUILD_ARGS allows you to set custom build arguments for the docker build command,
# which can be used to pass additional information or configuration to the Dockerfile
# during the build process. You can set this variable when invoking make,
# like: make docker-build DOCKER_BUILD_ARGS="MY_ARG=value ANOTHER_ARG=value2"
# For example:
# DOCKER_BUILD_ARGS ?= GIT_USER=${GIT_USER} GIT_TOKEN=${GIT_TOKEN}
DOCKER_BUILD_ARGS ?= 

DEFAULT_BUILD_ARGS := VERSION=$(VERSION) \
					COMMIT=$(COMMIT) \
					BRANCH=$(BRANCH) \
					BUILD_DATE=$(BUILD_DATE)

# DOCKER_BUILD_ARG_FLAGS processes DOCKER_BUILD_ARGS and prepends --build-arg to each argument
DOCKER_BUILD_ARG_FLAGS := $(foreach arg,$(DEFAULT_BUILD_ARGS),--build-arg $(strip $(arg)))
DOCKER_BUILD_ARG_FLAGS += $(foreach arg,$(DOCKER_BUILD_ARGS),$(strip $(arg)))

# DOCKER_LOCAL_FILE allows you to specify a custom Dockerfile for local development,
# which can be useful to speed up build times by using a simpler Dockerfile that doesn't
# require multi-platform support or other features needed for production builds.
# You can set this variable when invoking make, like:
# DOCKER_LOCAL_FILE ?= local.Dockerfile
DOCKER_LOCAL_FILE ?=

# DOCKER_BUILD_COMMON is a variable that contains the common flags for the docker build command,
# which are used in both local and CI environments.
# It is defined here to avoid duplication of the flags in the different target/branches.
# The actual flags are set in the respective branches to allow for differences in how the
# build should be executed in local vs CI environments (e.g., using a local Dockerfile vs a default one,
# or including/excluding certain build arguments).
DOCKER_BUILD_COMMON := $(DOCKER_BUILD_ARG_FLAGS) \
					-t $(IMG)

ifdef SSH_AUTH_SOCK
	DOCKER_BUILD_COMMON += --ssh default=$${SSH_AUTH_SOCK}
endif

ifneq ($(DOCKER_LOCAL_FILE),)
	DOCKER_BUILD_COMMON += -f $(DOCKER_LOCAL_FILE)
endif

# Setup the -ldflags option for go build here, interpolate the variable values
LDFLAGS = -ldflags "-X $(VERSION_PKG).VERSION=$(VERSION) -X $(VERSION_PKG).COMMIT=$(COMMIT) -X $(VERSION_PKG).BRANCH=$(BRANCH) -X $(VERSION_PKG).BUILDDATE=$(BUILD_DATE)"

GOOPTS ?= $(LDFLAGS)

RUN_ARGS ?=

BASE_DIR	:= $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))

NPROCS = $(shell nproc 2>/dev/null || sysctl -n hw.ncpu 2>/dev/null || echo 1)

# Automatically scale jobs to the number of CPU cores
.MAKEFLAGS += -j$(NPROCS)

# PLATFORMS defines the target platforms for the manager binary be built to provide support to multiple
# architectures. (i.e. make build-all). To use this option, you can set the PLATFORMS variable when invoking make, like:
# PLATFORMS="linux/amd64 linux/arm64" make build-all
PLATFORMS ?= linux/amd64 linux/arm linux/arm64 linux/ppc64le linux/s390x \
			windows/amd64 windows/arm64 \
			darwin/amd64 darwin/arm64

# DOCKER_PLATFORMS defines the target platforms for the manager image be built to provide support to multiple
# architectures. (i.e. make docker-buildx IMG=myregistry/mypoperator:0.0.1).
# To use this option you need to:
# - be able to use docker buildx. More info: https://docs.docker.com/build/buildx/
# - have enabled BuildKit. More info: https://docs.docker.com/develop/develop-images/build_enhancements/
# - be able to push the image to your registry (i.e. if you do not set a valid value via IMG=<myregistry/image:<tag>> then the export will fail)
# To adequately provide solutions that are compatible with multiple platforms, you should consider using this option.
#DOCKER_PLATFORMS ?= $(shell echo $(PLATFORMS) | tr ' ' ',')
DOCKER_PLATFORMS ?= linux/amd64 linux/arm linux/arm64 linux/ppc64le linux/s390x


SUPPRESS_OUTPUT ?= $(if $(CI),false,true)

OUTPUT_REDIRECT = 
ifeq ($(SUPPRESS_OUTPUT), true)
OUTPUT_REDIRECT = > /dev/null 2>&1
endif

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# CONTAINER_TOOL defines the container tool to be used for building images.
# Be aware that the target commands are only tested with Docker which is
# scaffolded by default. However, you might want to replace it to use other
# tools. (i.e. podman)
CONTAINER_TOOL ?= docker

# Setting SHELL to bash allows bash commands to be executed by recipes.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

ACTION ?= all

TEST_ARGS ?= -v -test.v
TEST_COVERAGE_FILE ?= cover.out
TEST_COVERAGE_CONFIG ?= ./.github/.testcoverage-local.yml

LINT_TIMEOUT ?= 10m

PKGS     := $(shell $(GO) list $(ALL_PKGS) | grep -Ev "e2e|examples|vendor")
#DEPENDENCIES := prometheus-operator certmanager

# COLORS

BOLD            := $(shell tput -Txterm bold)
UNDER           := $(shell tput -Txterm smul)
FG_BLACK        := $(shell tput -Txterm setaf 0)
FG_RED          := $(shell tput -Txterm setaf 1)
FG_GREEN        := $(shell tput -Txterm setaf 2)
FG_YELLOW       := $(shell tput -Txterm setaf 3)
FG_LIGHTPURPLE  := $(shell tput -Txterm setaf 4)
FG_PURPLE       := $(shell tput -Txterm setaf 5)
FG_BLUE         := $(shell tput -Txterm setaf 6)
FG_WHITE        := $(shell tput -Txterm setaf 7)
BG_BLACK        := $(shell tput -Txterm setab 0)
BG_RED          := $(shell tput -Txterm setab 1)
BG_GREEN        := $(shell tput -Txterm setab 2)
BG_YELLOW       := $(shell tput -Txterm setab 3)
BG_LIGHTPURPLE  := $(shell tput -Txterm setab 4)
BG_PURPLE       := $(shell tput -Txterm setab 5)
BG_BLUE         := $(shell tput -Txterm setab 6)
BG_WHITE        := $(shell tput -Txterm setab 7)
RESET           := $(shell tput -Txterm sgr0)


# END COLORS

# This rule is used to forward a target like "build" to "common-build".  This
# allows a new "build" target to be defined in a Makefile which includes this
# one and override "common-build" without override warnings.
%: common-% ;

.PHONY: common-all
common-all: precheck check_license deps lint build

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk command is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

MAX_LINE_LENGTH ?= 78
# 20 for target name
TARGET_WIDTH ?= 22

# This is used to calculate the maximum length of the description
# part of the help text, based on the total max line length and the
# width allocated for the target name.
# The '-7' accounts for the spacing and formatting characters between
# the target and description.
#DESC_MAX_LEN := $(shell echo $(MAX_LINE_LENGTH)-$(TARGET_WIDTH)-7 | bc)
DESC_MAX_LEN := $(shell echo $$(($(MAX_LINE_LENGTH)-$(TARGET_WIDTH)-2)))

# This is used to split long descriptions into multiple lines with proper 
# indentation, without breaking words in half.
# It looks for the last space before the max line length to break the line.
# It's also very verbose written, since "printf" take colorcodes into
# account for string length, while "awk" does not, so we need to manually
# calculate the spacing for alignment.
.PHONY: common-help
common-help: ## Display this help.
	@echo ''
	@echo '$(BOLD)Usage:$(RESET)'
	@echo '  $(FG_YELLOW)make$(RESET) $(FG_GREEN)<target>$(RESET)'
	@echo ''
	@echo '$(BOLD)Targets:$(RESET)'
	@awk 'BEGIN {FS = ":.*##";} /^[a-zA-Z_0-9-]+:.*?##/ { \
		target = $$1; \
		sub(/^common-/, "", target); \
		desc = $$2; \
		printf "  "; \
		printf "$(FG_YELLOW)%s$(RESET)", target; \
		printf "%*s", $(TARGET_WIDTH) - length(target) - 5, ""; \
		\
		if (length(desc) > $(DESC_MAX_LEN)) { \
			while (length(desc) > $(DESC_MAX_LEN)) { \
				pos = $(DESC_MAX_LEN); \
				while (pos > 0 && substr(desc, pos, 1) != " ") pos--; \
				if (pos == 0) pos = $(DESC_MAX_LEN); \
				\
				printf "$(FG_GREEN)%s$(RESET)\n  ", substr(desc, 1, pos); \
				printf "%*s", $(TARGET_WIDTH) - 4, ""; \
				\
				desc = substr(desc, pos + 1); \
			} \
		} \
		printf "$(FG_GREEN)%s$(RESET)\n", desc \
	} \
	/^##@/ { \
		printf "\n$(BOLD)%s:$(RESET)\n", substr($$0, 5) \
	} ' $(MAKEFILE_LIST)
	@echo ''

##@ Development
.PHONY: precheck
precheck::

define PRECHECK_COMMAND_template =
precheck:: $(1)_precheck

PRECHECK_COMMAND_$(1) ?= $(1) $$(strip $$(PRECHECK_OPTIONS_$(1)))
.PHONY: $(1)_precheck
$(1)_precheck:
	@if ! $$(PRECHECK_COMMAND_$(1)) 1>/dev/null 2>&1; then \
		echo "Execution of '$$(PRECHECK_COMMAND_$(1))' command failed. Is $(1) installed?"; \
		exit 1; \
	fi
endef

.PHONY: common-fmt
common-fmt: ## Run go fmt against code.
	@echo "$(FG_BLUE)>> $(FG_GREEN)running go fmt$(RESET)"
	$(GO) fmt $(ALL_PKGS)

.PHONY: common-vet
common-vet: ## Run go vet against code.
	@echo "$(FG_BLUE)>> $(FG_GREEN)running go vet$(RESET)"
	$(GO) vet $(ALL_PKGS)

.PHONY: common-deps
common-deps: ## Run go mod tidy and go mod download.
	@echo "$(FG_BLUE)>> $(FG_GREEN)running go mod tidy$(RESET)"
	$(GO) mod tidy -v
	@echo "$(FG_BLUE)>> $(FG_GREEN)running go mod download$(RESET)"
	$(GO) mod download

.PHONY: common-test
common-test: fmt vet ## Run unit tests with the Go testing framework.
	@echo "$(FG_BLUE)>> $(FG_GREEN)running unit tests$(RESET)"
	$(GO) test $(GOOPTS) $(PKGS) $(TEST_ARGS)

.PHONY: common-test-coverage
common-test-coverage: ## Run unit tests with coverage collection.
	@{ \
		echo "$(FG_BLUE)>> $(FG_GREEN)running unit tests with coverage$(RESET)"; \
		TEST_ARGS="$(TEST_ARGS) -coverprofile $(TEST_COVERAGE_FILE) -covermode=atomic -coverpkg=$(ALL_PKGS)" \
		$(MAKE) test ; \
	}

.PHONY: common-test-e2e
common-test-e2e: ## Run e2e tests with the Go testing framework.
	@echo "$(FG_BLUE)>> $(FG_GREEN)running E2E tests$(RESET)"
	$(GO) test $(GOOPTS) ./test/e2e/ $(TEST_ARGS)

.PHONY: common-test-e2e-coverage
common-test-e2e-coverage: ## Run e2e tests with coverage collection.
	@{ \
		echo "$(FG_BLUE)>> $(FG_GREEN)running E2E tests with coverage$(RESET)"; \
		TEST_ARGS="$(TEST_ARGS) -coverprofile $(TEST_COVERAGE_FILE) -covermode=atomic -coverpkg=$(ALL_PKGS)" \
		$(MAKE) test-e2e ; \
	}

.PHONY: common-test-all-coverage
common-test-all-coverage: gocovmerge gotestcoverage ## Run all tests (unit + e2e) and merge coverage.
	@{ \
		COVER_UNIT="cover-unit.out"; \
		COVER_E2E="cover-e2e.out"; \
		TEST_COVERAGE_FILE="$$COVER_UNIT" \
		TEST_ARGS="" \
		$(MAKE) common-test-coverage; \
		cover_unit_exit_code=$$?; \
		if [ $$cover_unit_exit_code -ne 0 ]; then \
			echo "$(FG_BLUE)>> $(FG_RED)Unit tests failed.$(RESET)"; \
			exit $$cover_unit_exit_code; \
		fi; \
		\
		TEST_COVERAGE_FILE="$$COVER_E2E" \
		TEST_ARGS="" \
		$(MAKE) common-test-e2e-coverage; \
		cover_e2e_exit_code=$$?; \
		if [ $$cover_e2e_exit_code -ne 0 ]; then \
			echo "$(FG_BLUE)>> $(FG_RED)E2E tests failed.$(RESET)"; \
			exit $$cover_e2e_exit_code; \
		fi; \
		\
		echo "$(FG_BLUE)>> $(FG_GREEN)merging coverage profiles$(RESET)"; \
		$(GOCOVMERGE) $$COVER_UNIT $$COVER_E2E > $(TEST_COVERAGE_FILE); \
		\
		echo "$(FG_BLUE)>> $(FG_GREEN)summarizing merged coverage$(RESET)"; \
		$(MAKE) coverage; \
	}

.PHONY: common-lint
common-lint: golangci-lint ## Run golangci-lint linter
	@echo "$(FG_BLUE)>> $(FG_GREEN)running golangci-lint$(RESET)"
	$(GOLANGCI_LINT) run --timeout=$(LINT_TIMEOUT)

.PHONY: common-lint-fix
common-lint-fix: golangci-lint ## Run golangci-lint linter and perform fixes
	@echo "$(FG_BLUE)>> $(FG_GREEN)running golangci-lint --fix$(RESET)"
	$(GOLANGCI_LINT) run --fix

.PHONY: common-coverage
common-coverage: gotestcoverage ## Check code coverage against the defined threshold in .testcoverage-local.yml
	@echo "$(FG_BLUE)>> $(FG_GREEN)summarizing coverage$(RESET)"
	$(GOTESTCOVERAGE) --config=$(TEST_COVERAGE_CONFIG)

##@ Build

.PHONY: common-check_license
common-check_license:
	@echo "$(FG_BLUE)>> $(FG_GREEN)checking license header$(RESET)"
	@licRes=$$(for file in $$(find . -type f -iname '*.go' ! -path './vendor/*') ; do \
		awk 'NR<=3' $$file | sed -e 's/\r//g' | grep -Eq "(Copyright|generated|GENERATED)" || echo $$file; \
	done); \
	if [ -n "$${licRes}" ]; then \
		echo "license header checking failed:"; \
		for file in $${licRes}; do \
			echo "  - $$file"; \
			awk 'NR<=3' $$file; \
		done; \
		exit 1; \
	fi

.PHONY: common-build
common-build: fmt vet ## Build binary.
	@echo "$(FG_BLUE)>> $(FG_GREEN)building $(PROJECT_NAME) binary$(RESET)"
	$(GO) build $(GOOPTS) -o $(BINDIR)/$(PROJECT_NAME) cmd/$(MAIN)

BUILD_TARGETS = $(foreach p,$(PLATFORMS),$(BINDIR)/$(PROJECT_NAME)-$(subst /,-,$(p)))

.PHONY: common-build-all
common-build-all: fmt vet $(BUILD_TARGETS) ## Build binaries for all platforms.
	cd $(BINDIR) && \
	sha256sum $(PROJECT_NAME)-* > checksums.txt
	@echo "$(FG_BLUE)>> $(FG_GREEN)Checksums saved to $(BINDIR)/checksums.txt$(RESET)";

$(BINDIR)/$(PROJECT_NAME)-%: ## Build binary with default name.
	$(eval PLATFORM := $(subst -, ,$*))
	$(eval OS       := $(word 1,$(PLATFORM)))
	$(eval ARCH     := $(word 2,$(PLATFORM)))
	$(eval EXT      := $(if $(filter windows,$(OS)),.exe,))
	@echo "$(FG_BLUE)>> $(FG_GREEN)building $(PROJECT_NAME) binary for $(OS)/$(ARCH)$(RESET)";

	CGO_ENABLED=0 GOOS=$(OS) GOARCH=$(ARCH) \
	$(GO) build $(GOOPTS) -o $(BINDIR)/$(PROJECT_NAME)-$(OS)-$(ARCH)$(EXT) ./cmd/$(MAIN)

.PHONY: common-run
common-run: fmt vet ## Run a binary from your host.
	@echo "$(FG_BLUE)>> $(FG_GREEN)running $(PROJECT_NAME)$(RESET)"
	$(GO) run $(GOOPTS) ./cmd/$(MAIN) $(RUN_ARGS)

# If you wish to build the manager image targeting other platforms you can use the --platform flag.
# (i.e. docker build --platform linux/arm64). However, you must enable docker buildKit for it.
# More info: https://docs.docker.com/develop/develop-images/build_enhancements/
.PHONY: common-docker-build
common-docker-build: ## Build docker image with the manager.
	@echo "$(FG_BLUE)>> $(FG_GREEN)building docker image $(IMG)$(RESET)"
	$(CONTAINER_TOOL) build $(DOCKER_BUILD_COMMON) .

.PHONY: common-docker-push
common-docker-push: ## Push docker image with the manager.
	@echo "$(FG_BLUE)>> $(FG_GREEN)pushing docker image $(IMG)$(RESET)"
	$(CONTAINER_TOOL) push $(IMG)

.PHONY: common-docker-buildx
common-docker-buildx: ## Build and push docker image for the manager for cross-platform support
	@echo "$(FG_BLUE)>> $(FG_GREEN)building and pushing docker image $(IMG) for platforms: $(DOCKER_PLATFORMS)$(RESET)"
	@set -euo pipefail; \
	# make a temp Dockerfile that injects --platform=$${BUILDPLATFORM} in the FIRST FROM
	tmp="$$(mktemp -q $(BASE_DIR)/Dockerfile.cross.XXXXXX)"; \
	awk 'BEGIN{done=0} !done && $$1=="FROM"{ sub(/^FROM/,"FROM --platform=$${BUILDPLATFORM}"); done=1 } {print}' Dockerfile > "$$tmp"; \
	# ensure the builder exists (idempotent)
	if ! $(CONTAINER_TOOL) buildx inspect $(PROJECT_NAME)-builder >/dev/null 2>&1; then \
	  $(CONTAINER_TOOL) buildx create \
	  	--name $(PROJECT_NAME)-builder \
	    --driver docker-container >/dev/null; \
	fi; \
	# build using that builder (don't mutate global builder state)
	$(CONTAINER_TOOL) buildx build \
	  --push \
	  --builder $(PROJECT_NAME)-builder \
	  --platform="$(shell echo $(DOCKER_PLATFORMS) | tr ' ' ',')" \
	  $(DOCKER_BUILD_COMMON) \
	  -f "$$tmp" . && \
	rm -f "$$tmp"


##@ Helm

HELM_CHART_VERSION ?= $(shell yq '.version' $(HELM_CHART_DIR)/Chart.yaml)
HELM_CHART_DIR ?= chart/$(PROJECT_NAME)
HELM_REGISTRY ?= $(DOCKER_REPO_BASE)/charts
HELM := $(shell command -v helm 2>/dev/null || echo $(LOCALBIN)/helm)

.PHONY: common-helm-reqs
common-helm-reqs: ## Check that all required variables for helm packaging and release are set
	@missing=0; \
	if [ ! -d "$(HELM_CHART_DIR)" ]; then echo "$(FG_RED)ERROR: HELM_CHART_DIR ($(HELM_CHART_DIR)) does not exist$(RESET)"; missing=1; fi; \
	if [ ! -f "$(HELM_CHART_DIR)/Chart.yaml" ]; then echo "$(FG_RED)ERROR: Chart.yaml not found in $(HELM_CHART_DIR)$(RESET)"; missing=1; fi; \
	if [ -z "$(GIT_USER)" ]; then echo "$(FG_RED)ERROR: GIT_USER is not set$(RESET)"; missing=1; fi; \
	if [ -z "$(GIT_TOKEN)" ]; then echo "$(FG_RED)ERROR: GIT_TOKEN is not set$(RESET)"; missing=1; fi; \
	if [ -z "$(IMG)" ]; then echo "$(FG_RED)ERROR: IMG is not set$(RESET)"; missing=1; fi; \
	if [ "$${missing}" -ne 0 ]; then echo "$(FG_BLUE)>> $(FG_RED)One or more required variables are missing. Aborting.$(RESET)"; exit 1; fi

.PHONY: common-helm-install
common-helm-install: ## Install helm locally if not already installed
	@echo "$(FG_BLUE)>> $(FG_GREEN)Checking for helm...$(RESET)"
	@if command -v helm >/dev/null 2>&1; then \
		echo "$(FG_BLUE)>> $(FG_GREEN)helm found at $$(command -v helm) - skipping install$(RESET)"; \
	elif [ -x "$(LOCALBIN)/helm" ]; then \
		echo "$(FG_BLUE)>> $(FG_GREEN)helm already installed at $(LOCALBIN)/helm - skipping install$(RESET)"; \
	else \
		echo "$(FG_BLUE)>> $(FG_GREEN)Installing helm to $(LOCALBIN)...$(RESET)"; \
		mkdir -p $(LOCALBIN); \
		curl -sSL https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | \
		HELM_INSTALL_DIR=$(LOCALBIN) bash; \
	fi

.PHONY: common-helm-lint
common-helm-lint: common-helm-install ## Lint the Helm chart
	@echo "$(FG_BLUE)>> $(FG_GREEN)Linting Helm chart...$(RESET)"
	@command -v $(HELM) >/dev/null 2>&1 || { \
		echo "$(FG_RED)ERROR: $(HELM) not found. Install it: https://helm.sh/docs/intro/install/$(RESET)"; \
		exit 1; \
	}
	$(HELM) lint $(HELM_CHART_DIR) --strict
	@echo "$(FG_BLUE)>> $(FG_GREEN)✓ Helm chart linting passed!$(RESET)"

.PHONY: common-helm-package
common-helm-package: common-helm-lint common-helm-reqs ## Package the Helm chart
	@echo "$(FG_BLUE)>> $(FG_GREEN)Packaging Helm chart...$(RESET)"
	@mkdir -p dist
	$(HELM) package $(HELM_CHART_DIR) -d dist --version $(HELM_CHART_VERSION)
	@echo "$(FG_BLUE)>> $(FG_GREEN)✓ Chart packaged: dist/$(PROJECT_NAME)-$(HELM_CHART_VERSION).tgz$(RESET)"

.PHONY: common-helm-release
common-helm-release: common-helm-package common-helm-reqs ## Release Helm chart to OCI registry
	@echo "$(FG_BLUE)>> $(FG_GREEN)Releasing Helm chart to OCI registry...$(RESET)"
	@command -v $(HELM) >/dev/null 2>&1 || { \
		echo "$(FG_RED)ERROR: $(HELM) not found$(RESET)"; \
		exit 1; \
	}
	@if [ -z "$(GIT_TOKEN)" ]; then \
		echo "$(FG_RED)ERROR: GIT_TOKEN not set$(RESET)"; \
		echo "For local dev, export GIT_TOKEN with your GitHub PAT:"; \
		echo "  export GIT_TOKEN=ghp_xxxxxxxxxxxx"; \
		echo ""; \
		echo "Attempting to use existing Docker credentials..."; \
	else \
		echo "$(FG_BLUE)>> $(FG_GREEN)Logging in to GitHub Container Registry...$(RESET)"; \
		echo "$(GIT_TOKEN)" | $(HELM) registry login ghcr.io -u $(GIT_USER) --password-stdin; \
	fi
	@echo "$(FG_BLUE)>> $(FG_GREEN)Pushing chart to oci://$(HELM_REGISTRY)/$(PROJECT_NAME):$(HELM_CHART_VERSION)$(RESET)"
	$(HELM) push dist/$(PROJECT_NAME)-$(HELM_CHART_VERSION).tgz oci://$(HELM_REGISTRY)
	@echo "$(FG_BLUE)>> $(FG_GREEN)✓ Chart released to: oci://$(HELM_REGISTRY)/$(PROJECT_NAME):$(HELM_CHART_VERSION)$(RESET)"

##@ Dependencies

## Location to install dependencies to
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

## Tool Binaries
GOVULNCHECK ?= $(LOCALBIN)/govulncheck
GOLANGCI_LINT = $(LOCALBIN)/golangci-lint
GOTESTCOVERAGE ?= $(LOCALBIN)/go-test-coverage
GOCOVMERGE ?= $(LOCALBIN)/gocovmerge
KUSTOMIZE ?= $(LOCALBIN)/kustomize

## Tool Versions
GOVULNCHECK_VERSION ?= latest
GOCOVMERGE_VERSION ?= latest
GOLANGCI_LINT_VERSION ?= v2.9.0
GOTESTCOVERAGE_VERSION ?= latest
KUSTOMIZE_VERSION ?= v5.8.2

.PHONY: common-gotestcoverage
common-gotestcoverage: $(GOTESTCOVERAGE) ## Download go-test-coverage locally if necessary.
$(GOTESTCOVERAGE): $(LOCALBIN)
	@echo "$(FG_BLUE)>> $(FG_GREEN)Installing/ensuring go-test-coverage $(GOTESTCOVERAGE_VERSION) to $(GOTESTCOVERAGE)$(RESET)"
	$(call go-install-tool,$(GOTESTCOVERAGE),github.com/vladopajic/go-test-coverage/v2,$(GOTESTCOVERAGE_VERSION))

.PHONY: common-gocovmerge
common-gocovmerge: $(GOCOVMERGE) ## Download gocovmerge locally if necessary.
$(GOCOVMERGE): $(LOCALBIN)
	@echo "$(FG_BLUE)>> $(FG_GREEN)Installing/ensuring gocovmerge $(GOCOVMERGE_VERSION) to $(GOCOVMERGE)$(RESET)"
	$(call go-install-tool,$(GOCOVMERGE),github.com/wadey/gocovmerge,$(GOCOVMERGE_VERSION))

.PHONY: common-govulncheck
common-govulncheck: $(GOVULNCHECK) ## Download govulncheck locally if necessary.
$(GOVULNCHECK): $(LOCALBIN)
	@echo "$(FG_BLUE)>> $(FG_GREEN)Installing/ensuring govulncheck $(GOVULNCHECK_VERSION) to $(GOVULNCHECK)$(RESET)"
	$(call go-install-tool,$(GOVULNCHECK),golang.org/x/vuln/cmd/govulncheck,$(GOVULNCHECK_VERSION))

.PHONY: common-golangci-lint
common-golangci-lint: $(GOLANGCI_LINT) ## Download golangci-lint locally if necessary.
$(GOLANGCI_LINT): $(LOCALBIN)
	@echo "$(FG_BLUE)>> $(FG_GREEN)Installing/ensuring golangci-lint $(GOLANGCI_LINT_VERSION) to $(GOLANGCI_LINT)$(RESET)"
	$(call go-install-tool,$(GOLANGCI_LINT),github.com/golangci/golangci-lint/v2/cmd/golangci-lint,$(GOLANGCI_LINT_VERSION))
.PHONY: common-kustomize
common-kustomize: $(KUSTOMIZE) ## Download kustomize locally if necessary.
$(KUSTOMIZE): $(LOCALBIN)
	@echo "$(FG_BLUE)>> $(FG_GREEN)Installing/ensuring kustomize $(KUSTOMIZE_VERSION) to $(KUSTOMIZE)$(RESET)"
	$(call go-install-tool,$(KUSTOMIZE),sigs.k8s.io/kustomize/kustomize/v5,$(KUSTOMIZE_VERSION))

# go-install-tool will 'go install' any package with custom target and name of binary, if it doesn't exist
# It first checks if the binary exists locally, then checks if it exists in PATH
# If found in PATH, it creates a symlink to the system version
# If not found anywhere, it installs and creates a versioned binary
# $1 - target path with name of binary
# $2 - package url which can be installed
# $3 - specific version of package
define go-install-tool
@binary_name=$$(basename "$(1)"); \
if [ -f "$(1)" ]; then \
	echo "$(FG_BLUE)>> $(FG_GREEN)$${binary_name} ensured.$(RESET)" ;\
elif command -v $${binary_name} >/dev/null 2>&1; then \
	echo "$(FG_BLUE)>> $(FG_GREEN)$${binary_name} found in PATH, creating symlink$(RESET)"; \
	ln -sf $$(which $${binary_name}) $(1); \
elif [ "$(if $(CI),false,true)" = "true" ]; then \
	{ \
		set -e; \
		package=$(2)@$(3) ;\
		echo "$(FG_BLUE)>> $(FG_YELLOW)$${binary_name} not found, installing $${package}$(RESET)" ;\
		echo "$(FG_BLUE)>> $(FG_YELLOW)location: default$(RESET)" ;\
		$(GO) install $${package} ;\
	} ;\
else \
	[ -f "$(1)-$(3)" ] || { \
		set -e; \
		package=$(2)@$(3) ;\
		echo "$(FG_BLUE)>> $(FG_YELLOW)$${binary_name} not found, installing $${package}$(RESET)" ;\
		echo "$(FG_BLUE)>> $(FG_YELLOW)location: $(LOCALBIN)$(RESET)" ;\
		rm -f $(1) || true ;\
		GOBIN=$(LOCALBIN) GOTOOLCHAIN=$(GO_TOOLCHAIN) \
		$(GO) install $${package} ;\
		mv $(1) $(1)-$(3) ;\
	} ;\
	ln -sf $(1)-$(3) $(1) ;\
fi
endef

# fetch-script will fetch anything and make it executable, if it doesn't exist
# $1 - url which can be called
# $2 - local path to save the file
define fetch-script
{ \
set -e; \
if [ ! -f "$(2)" ]; then \
	echo "$(FG_BLUE)>> $(FG_GREEN)Fetching: $(2)$(RESET)" ; \
	curl -sSLo $(2) "$(1)" ;\
	chmod +x $(2) ;\
fi \
}
endef

.PHONY: common-yq
YQ_VERSION := v4.13.0
YQ_BIN := $(LOCALBIN)/yq
common-yq:
ifeq (,$(wildcard $(YQ_BIN)))
ifeq (,$(shell which yq 2>/dev/null))
	mkdir -p $(dir $(YQ_BIN)) ;\
	OS=$(shell go env GOOS) && ARCH=$(shell go env GOARCH) && \
	$(call fetch-script,https://github.com/mikefarah/yq/releases/download/$(YQ_VERSION)/yq_$${OS}_$${ARCH},$(YQ_BIN))
else
YQ_BIN = $(shell which yq)
endif
endif
