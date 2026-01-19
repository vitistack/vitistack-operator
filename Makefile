# Makefile for vitistack-operator

# Variables
BINARY_NAME=vitistack-operator
GO=go
GOFMT=gofmt
DOCKER=docker
KUBECTL=kubectl
DOCKER_IMAGE=ghcr.io/vitistack/vitistack-operator
DOCKER_TAG=latest
MAIN_PATH=./cmd/vitistack-operator/main.go

NAMESPACE=default
POD ?= $(shell kubectl -n $(NAMESPACE) get pods -o jsonpath='{.items[0].metadata.name}')

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

# Get the currently used golang version
GO_VERSION=$(shell go version | sed -e 's/^[^0-9.]*\([0-9.]*\).*/\1/')
GOSEC ?= $(LOCALBIN)/gosec
GOSEC_VERSION ?= latest
GOVULNCHECK ?= $(LOCALBIN)/govulncheck
GOVULNCHECK_VERSION ?= latest

GOLANGCI_LINT = $(LOCALBIN)/golangci-lint
GOLANGCI_LINT_VERSION ?= latest

# Helper macro: installs a Go tool only if the target binary doesn't already exist.
define go-install-tool
@[ -f $(1) ] || { \
	echo "Installing $(2)@$(3) to $(LOCALBIN)"; \
	GOBIN=$(LOCALBIN) go install $(2)@$(3); \
}
endef

# Basic colors
BLACK=\033[0;30m
RED=\033[0;31m
GREEN=\033[0;32m
YELLOW=\033[0;33m
BLUE=\033[0;34m
PURPLE=\033[0;35m
CYAN=\033[0;36m
WHITE=\033[0;37m

# Text formatting
BOLD=\033[1m
UNDERLINE=\033[4m
RESET=\033[0m

# External CLI dependencies
CURL ?= curl
JQ ?= jq

## VitiStack CRDs (download/install)
# Override VITISTACK_CRDS_REF to pin a branch, tag, or commit (default: main)
VITISTACK_CRDS_REF ?= main
# GitHub API endpoint to list files under crds/ at a specific ref
VITISTACK_CRDS_API ?= https://api.github.com/repos/vitistack/common/contents/crds?ref=$(VITISTACK_CRDS_REF)
# Local directory where CRDs will be downloaded
CRDS_DOWNLOAD_DIR ?= hack/vitistack-crds

##@ Help
.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Prerequisites
.PHONY: check-go check-docker check-kubectl check-prereqs

check-go: ## Check if Go is installed
	@echo "${YELLOW}Checking if Go is installed...${RESET}"
	@which ${GO} > /dev/null || (echo "${RED}Error: Go is not installed or not in PATH${RESET}" && exit 1)
	@echo "${GREEN}Go is installed: $$(${GO} version)${RESET}"

check-docker: ## Check if Docker is installed
	@echo "${YELLOW}Checking if Docker is installed...${RESET}"
	@which ${DOCKER} > /dev/null || (echo "${RED}Error: Docker is not installed or not in PATH${RESET}" && exit 1)
	@echo "${GREEN}Docker is installed: $$(${DOCKER} --version)${RESET}"

check-kubectl: ## Check if kubectl is installed
	@echo "${YELLOW}Checking if kubectl is installed...${RESET}"
	@which ${KUBECTL} > /dev/null || (echo "${RED}Error: kubectl is not installed or not in PATH${RESET}" && exit 1)
	@echo "${GREEN}kubectl is installed: $$(${KUBECTL} version)${RESET}"

check-prereqs: check-go check-docker check-kubectl ## Check all prerequisites are installed
	@echo "${GREEN}All prerequisites are installed.${RESET}"

##@ Build
.PHONY: all build clean test lint fmt vet run docker-build docker-push help

all: check-prereqs clean test build ## Run all checks, clean, test and build

build: check-go ## Build the application binary
	@echo "${GREEN}Building ${BINARY_NAME}...${RESET}"
	${GO} vet ./...
	${GO} fmt ./...
	${GO} build -o dist/${BINARY_NAME} ${MAIN_PATH}

build-static: check-go ## Build the application with static linking
	@echo "${GREEN}Building ${BINARY_NAME} with static linking...${RESET}"
	${GO} vet ./...
	${GO} fmt ./...
	CGO_ENABLED=0 ${GO} build -ldflags '-extldflags "-static"' -o dist/${BINARY_NAME} ${MAIN_PATH}

clean: ## Clean build files and artifacts
	@echo "${YELLOW}Cleaning...${RESET}"
	${GO} clean
	rm -f ${BINARY_NAME}

.PHONY: lint
lint: golangci-lint ## Run golangci-lint linter
	$(GOLANGCI_LINT) run

.PHONY: lint-fix
lint-fix: golangci-lint ## Run golangci-lint linter and perform fixes
	$(GOLANGCI_LINT) run --fix

##@ SBOM (Software Bill of Materials)
SYFT ?= $(LOCALBIN)/syft
SYFT_VERSION ?= latest
SBOM_OUTPUT_DIR ?= sbom
SBOM_PROJECT_NAME ?= vitistack-operator

.PHONY: install-syft
install-syft: $(SYFT) ## Install syft SBOM generator locally
$(SYFT): $(LOCALBIN)
	@set -e; echo "Installing syft $(SYFT_VERSION)"; \
	curl -sSfL https://raw.githubusercontent.com/anchore/syft/main/install.sh | sh -s -- -b $(LOCALBIN)

.PHONY: sbom-source
sbom-source: install-syft ## Generate SBOMs for Go source code (CycloneDX + SPDX)
	@mkdir -p $(SBOM_OUTPUT_DIR)
	@echo "Generating source code SBOMs..."
	$(SYFT) dir:. --source-name=$(SBOM_PROJECT_NAME) -o cyclonedx-json=$(SBOM_OUTPUT_DIR)/sbom-source.cdx.json
	$(SYFT) dir:. --source-name=$(SBOM_PROJECT_NAME) -o spdx-json=$(SBOM_OUTPUT_DIR)/sbom-source.spdx.json
	@echo "SBOMs generated: $(SBOM_OUTPUT_DIR)/sbom-source.{cdx,spdx}.json"

.PHONY: sbom-container
sbom-container: install-syft ## Generate SBOMs for container image (CycloneDX + SPDX, requires IMG)
	@mkdir -p $(SBOM_OUTPUT_DIR)
	@echo "Generating container SBOMs for $(IMG)..."
	$(SYFT) $(IMG) -o cyclonedx-json=$(SBOM_OUTPUT_DIR)/sbom-container.cdx.json
	$(SYFT) $(IMG) -o spdx-json=$(SBOM_OUTPUT_DIR)/sbom-container.spdx.json
	@echo "SBOMs generated: $(SBOM_OUTPUT_DIR)/sbom-container.{cdx,spdx}.json"

.PHONY: sbom
sbom: sbom-source ## Alias for sbom-source

##@ Testing & Quality
test: check-go ## Run all tests
	@echo "${GREEN}Running tests...${RESET}"
	${GO} test -v ./...

test-coverage: check-go ## Run tests with coverage report
	@echo "${GREEN}Running tests with coverage...${RESET}"
	${GO} test -cover -coverprofile=coverage.out ./...
	${GO} tool cover -html=coverage.out -o coverage.html

fmt: check-go ## Format Go code
	@echo "${YELLOW}Formatting code...${RESET}"
	${GOFMT} -w -s .

vet: check-go ## Run go vet
	@echo "${YELLOW}Running go vet...${RESET}"
	${GO} vet ./...

	
##@ Development
.PHONY: run 
run: check-go ## Run the application
	@echo "${GREEN}Running ${BINARY_NAME}...${RESET}"
	${GO} run ${MAIN_PATH}

##@ Docker
.PHONY: docker-build docker-push
docker-build: check-docker ## Build Docker image
	@echo "${CYAN}Building Docker image...${RESET}"
	${DOCKER} build -t ${DOCKER_IMAGE}:${DOCKER_TAG} .

docker-push: check-docker ## Push Docker image
	@echo "${CYAN}Pushing Docker image...${RESET}"
	${DOCKER} push ${DOCKER_IMAGE}:${DOCKER_TAG}



##@ Installation
.PHONY: golangci-lint
golangci-lint: $(GOLANGCI_LINT) ## Download golangci-lint locally if necessary.
$(GOLANGCI_LINT): $(LOCALBIN)
	$(call go-install-tool,$(GOLANGCI_LINT),github.com/golangci/golangci-lint/v2/cmd/golangci-lint,$(GOLANGCI_LINT_VERSION))

.PHONY: install-lint install-gosec helm-install
install-lint: ## Install golangci-lint
	@echo "${YELLOW}Installing golangci-lint...${RESET}"
	${GO} install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
	@echo "${YELLOW}Then run 'make lint' to check for issues.${RESET}"

.PHONY: install-security-scanner
install-security-scanner: $(GOSEC) ## Install gosec security scanner locally (static analysis for security issues)
$(GOSEC): $(LOCALBIN)
	@set -e; echo "Attempting to install gosec $(GOSEC_VERSION)"; \
	if ! GOBIN=$(LOCALBIN) go install github.com/securego/gosec/v2/cmd/gosec@$(GOSEC_VERSION) 2>/dev/null; then \
		echo "Primary install failed, attempting install from @main (compatibility fallback)"; \
		if ! GOBIN=$(LOCALBIN) go install github.com/securego/gosec/v2/cmd/gosec@main; then \
			echo "gosec installation failed for versions $(GOSEC_VERSION) and @main"; \
			exit 1; \
		fi; \
	fi; \
	echo "gosec installed at $(GOSEC)"; \
	chmod +x $(GOSEC)

.PHONY: install-govulncheck
install-govulncheck: $(GOVULNCHECK) ## Install govulncheck locally (vulnerability scanner for Go)
$(GOVULNCHECK): $(LOCALBIN)
	@set -e; echo "Attempting to install govulncheck $(GOVULNCHECK_VERSION)"; \
	if ! GOBIN=$(LOCALBIN) go install golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION) 2>/dev/null; then \
		echo "Primary install failed, attempting install from @latest (compatibility fallback)"; \
		if ! GOBIN=$(LOCALBIN) go install golang.org/x/vuln/cmd/govulncheck@latest; then \
			echo "govulncheck installation failed for versions $(GOVULNCHECK_VERSION) and @latest"; \
			exit 1; \
		fi; \
	fi; \
	echo "govulncheck installed at $(GOVULNCHECK)"; \
	chmod +x $(GOVULNCHECK)

helm-install: check-kubectl ## Install Helm chart with CRDs
	@echo "${GREEN}Installing Helm chart with CRDs...${RESET}"
	helm upgrade --install vitistack-operator ./charts/vitistack-operator --namespace default

##@ Dependencies
.PHONY: deps update-deps
deps: ## Download and verify dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@go mod verify
	@go mod tidy
	@echo "Dependencies updated!"

update-deps: ## Update dependencies
	@echo "Updating dependencies..."
	@go get -u ./...
	@go mod tidy
	@echo "Dependencies updated!"

##@ Security
.PHONY: gosec
gosec: install-security-scanner ## Run gosec security scan (fails on findings)
	$(GOSEC) ./...

.PHONY: govulncheck
govulncheck: install-govulncheck ## Run govulncheck vulnerability scan (fails on findings)
	$(GOVULNCHECK) ./...

.PHONY: go-security-scan-docker
go-security-scan-docker: ## Run gosec scan using official container (alternative if local install fails)
	@echo "Running gosec via Docker container..."; \
	$(CONTAINER_TOOL) run --rm -v $(PWD):/workspace -w /workspace securego/gosec/gosec:latest ./...

##@ Kubernetes
.PHONY: k8s-token-kubernetes k8s-token-kubernetes-cleanup

k8s-token-kubernetes: check-kubectl ## Read the service account token from a pod in the specified namespace, specified by NAMESPACE variable (ex: make k8s-token-kubernetes NAMESPACE=default)
	@echo "${YELLOW}>>> Checking for pods in namespace $(NAMESPACE)...${RESET}"
	@POD_COUNT=$$(kubectl get pods -n $(NAMESPACE) --no-headers 2>/dev/null | wc -l | tr -d ' '); \
	if [ "$$POD_COUNT" -eq 0 ]; then \
		echo "${YELLOW}No pods found in namespace $(NAMESPACE). Creating namespace and temporary pod...${RESET}"; \
		kubectl create namespace $(NAMESPACE) --dry-run=client -o yaml | kubectl apply -f - 2>/dev/null || true; \
		echo "${GREEN}Creating temporary pod 'token-extractor' in namespace $(NAMESPACE)...${RESET}"; \
		kubectl run token-extractor -n $(NAMESPACE) --image=busybox --restart=Never --command -- sleep infinity; \
		echo "${YELLOW}Waiting for pod to be ready...${RESET}"; \
		kubectl wait --for=condition=Ready pod/token-extractor -n $(NAMESPACE) --timeout=60s; \
		echo "${GREEN}>>> Reading token from temporary pod token-extractor in namespace $(NAMESPACE)...\n${RESET}"; \
		kubectl -n $(NAMESPACE) exec token-extractor -- cat /var/run/secrets/kubernetes.io/serviceaccount/token; \
	else \
		FIRST_POD=$$(kubectl get pods -n $(NAMESPACE) --no-headers -o custom-columns=":metadata.name" | head -n 1); \
		echo "${GREEN}>>> Reading token from existing pod $$FIRST_POD in namespace $(NAMESPACE)...${RESET}"; \
		kubectl -n $(NAMESPACE) exec $$FIRST_POD -- cat /var/run/secrets/kubernetes.io/serviceaccount/token; \
	fi

k8s-token-kubernetes-cleanup: check-kubectl ## Cleanup temporary pod created by token-kubernetes
	@echo "${YELLOW}>>> Cleaning up temporary pod token-extractor in namespace $(NAMESPACE)...${RESET}"
	@kubectl delete pod token-extractor -n $(NAMESPACE) --ignore-not-found
	@echo "${GREEN}Temporary pod cleaned up.${RESET}"

##@ CRDs & Resources
.PHONY: k8s-install-configmap k8s-uninstall-configmap k8s-install-vitistack-crds k8s-download-vitistack-crds k8s-uninstall-crds
k8s-install-configmap: check-kubectl ## Install configmap into cluster
	@echo "${GREEN}Installing configmap into cluster...${RESET}"
	${KUBECTL} apply -f hack/test/manifests/configmap.yaml

k8s-uninstall-configmap: check-kubectl ## Uninstall configmap from cluster
	@echo "${RED}Uninstalling configmap from cluster...${RESET}"
	${KUBECTL} delete -f hack/test/manifests/configmap.yaml

.PHONY: k8s-download-vitistack-crds
k8s-download-vitistack-crds: require-curl-jq ## Download all VitiStack CRDs from vitistack/crds@$(VITISTACK_CRDS_REF) into $(CRDS_DOWNLOAD_DIR)
	@echo -e "$(CYAN)Fetching CRD list from$(RESET) $(VITISTACK_CRDS_API)"
	@echo -e "$(YELLOW)Clearing existing contents of$(RESET) $(CRDS_DOWNLOAD_DIR)"
	@rm -rf "$(CRDS_DOWNLOAD_DIR)"
	@mkdir -p $(CRDS_DOWNLOAD_DIR)
	@$(CURL) -fsSL "$(VITISTACK_CRDS_API)" | $(JQ) -r '.[] | select(.type=="file") | select(.name | test("\\.(ya?ml)$$")) | .download_url' > $(CRDS_DOWNLOAD_DIR)/.crd_urls
	@if [ ! -s $(CRDS_DOWNLOAD_DIR)/.crd_urls ]; then echo "No CRD files found at ref $(VITISTACK_CRDS_REF)."; exit 1; fi
	@echo -e "$(CYAN)Downloading CRDs into$(RESET) $(CRDS_DOWNLOAD_DIR)"
	@while read -r url; do \
		fname=$$(basename $$url); \
		echo "- $$fname"; \
		$(CURL) -fsSL "$$url" -o "$(CRDS_DOWNLOAD_DIR)/$$fname"; \
	done < $(CRDS_DOWNLOAD_DIR)/.crd_urls
	@echo -e "$(GREEN)CRDs downloaded to$(RESET) $(CRDS_DOWNLOAD_DIR)"

.PHONY: k8s-install-vitistack-crds
k8s-install-vitistack-crds: require-kubectl k8s-download-vitistack-crds ## Apply downloaded VitiStack CRDs to the current kube-context
	@echo -e "$(CYAN)Applying CRDs from$(RESET) $(CRDS_DOWNLOAD_DIR)"
	@$(KUBECTL) apply -f $(CRDS_DOWNLOAD_DIR)
	@echo -e "$(GREEN)VitiStack CRDs installed successfully.$(RESET)"

k8s-uninstall-crds: check-kubectl ## Uninstall CRDs from the cluster
	@echo "${RED}Uninstalling CRDs...${RESET}"
	${KUBECTL} delete -f $(CRDS_DOWNLOAD_DIR)


# Dependency checks used by the CRD targets
.PHONY: require-kubectl
require-kubectl: ## Verify kubectl is installed
	@command -v $(KUBECTL) >/dev/null 2>&1 || { echo "Error: $(KUBECTL) is required (see https://kubernetes.io/docs/tasks/tools/)."; exit 1; }

## Shared dependency check
.PHONY: require-curl-jq
require-curl-jq: ## Verify curl and jq are installed
	@command -v $(CURL) >/dev/null 2>&1 || { echo "Error: $(CURL) is required (e.g., brew install curl)."; exit 1; }
	@command -v $(JQ) >/dev/null 2>&1 || { echo "Error: $(JQ) is required (e.g., brew install jq)."; exit 1; }

