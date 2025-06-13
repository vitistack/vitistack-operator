# Makefile for datacenter-operator

# Variables
BINARY_NAME=datacenter-operator
GO=go
GOFMT=gofmt
DOCKER=docker
KUBECTL=kubectl
DOCKER_IMAGE=ghcr.io/vitistack/datacenter-operator
DOCKER_TAG=latest
MAIN_PATH=./cmd/datacenter-operator/main.go

NAMESPACE=default
POD ?= $(shell kubectl -n $(NAMESPACE) get pods -o jsonpath='{.items[0].metadata.name}')

# Get the currently used golang version
GO_VERSION=$(shell go version | sed -e 's/^[^0-9.]*\([0-9.]*\).*/\1/')

# COLORS
GREEN  := $(shell tput -Txterm setaf 2)
YELLOW := $(shell tput -Txterm setaf 3)
WHITE  := $(shell tput -Txterm setaf 7)
CYAN   := $(shell tput -Txterm setaf 6)
RED    := $(shell tput -Txterm setaf 1)
RESET  := $(shell tput -Txterm sgr0)

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
	${GO} build -o dist/${BINARY_NAME} ${MAIN_PATH}

build-static: check-go ## Build the application with static linking
	@echo "${GREEN}Building ${BINARY_NAME} with static linking...${RESET}"
	CGO_ENABLED=0 ${GO} build -ldflags '-extldflags "-static"' -o dist/${BINARY_NAME} ${MAIN_PATH}

clean: ## Clean build files and artifacts
	@echo "${YELLOW}Cleaning...${RESET}"
	${GO} clean
	rm -f ${BINARY_NAME}

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

lint: check-go ## Run golangci-lint
	@echo "${YELLOW}Running linter...${RESET}"
	golangci-lint run ./...

gosec: check-go ## Run gosec security scanner
	@echo "${YELLOW}Running gosec...${RESET}"
	gosec ./...
	@echo "${YELLOW}gosec completed.${RESET}"
	
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

##@ CRDs & Resources
.PHONY: install-configmap uninstall-configmap install-crds download-crds uninstall
install-configmap: check-kubectl ## Install configmap into cluster
	@echo "${GREEN}Installing configmap into cluster...${RESET}"
	${KUBECTL} apply -f hack/test/manifests/configmap.yaml

uninstall-configmap: check-kubectl ## Uninstall configmap into cluster
	@echo "${GREEN}Installing configmap into cluster...${RESET}"
	${KUBECTL} delete -f hack/test/manifests/configmap.yaml

install-viti-crds: ## Install CRDs into a cluster
	@echo "${GREEN}Installing CRDs...${RESET}"
	@if [ ! -d "hack/crds" ]; then \
		echo "${RED}Error: hack/crds directory does not exist${RESET}"; \
		echo "${YELLOW}Run 'make download-viti-crds' first to download the CRDs (requires GITHUB_TOKEN)${RESET}"; \
		exit 1; \
	fi
	@if [ -z "$$(find hack/crds -name '*.yaml' -type f 2>/dev/null)" ]; then \
		echo "${RED}Error: No YAML files found in hack/crds directory${RESET}"; \
		echo "${YELLOW}Run 'make download-viti-crds' first to download the CRDs (requires GITHUB_TOKEN)${RESET}"; \
		exit 1; \
	fi
	@echo "Found CRD files:"
	@ls -1 hack/crds/*.yaml | sed 's/^/  - /'
	${KUBECTL} apply -f hack/crds/
	@echo "${GREEN}CRDs installed successfully${RESET}"

download-viti-crds: ## Download CRDs from private repository (requires GITHUB_TOKEN)
	@echo "${GREEN}Downloading CRDs from private repository...${RESET}"
	@if [ -z "$$GITHUB_TOKEN" ]; then \
		echo "${RED}Error: GITHUB_TOKEN environment variable is required for private repository access${RESET}"; \
		exit 1; \
	fi
	@mkdir -p hack/crds
	@echo "Downloading vitistack.io_datacenters.yaml..."
	@if ! curl --fail -H "Authorization: token $$GITHUB_TOKEN" \
		-H "Accept: application/vnd.github.v3.raw" \
		-o hack/crds/vitistack.io_datacenters.yaml \
		https://api.github.com/repos/vitistack/crds/contents/crds/vitistack.io_datacenters.yaml; then \
		echo "${RED}Error: Failed to download vitistack.io_datacenters.yaml${RESET}"; \
		exit 1; \
	fi
	@echo "Downloading vitistack.io_kubernetesproviders.yaml..."
	@if ! curl --fail -H "Authorization: token $$GITHUB_TOKEN" \
		-H "Accept: application/vnd.github.v3.raw" \
		-o hack/crds/vitistack.io_kubernetesproviders.yaml \
		https://api.github.com/repos/vitistack/crds/contents/crds/vitistack.io_kubernetesproviders.yaml; then \
		echo "${RED}Error: Failed to download vitistack.io_kubernetesproviders.yaml${RESET}"; \
		exit 1; \
	fi
	@echo "Downloading vitistack.io_machineproviders.yaml..."
	@if ! curl --fail -H "Authorization: token $$GITHUB_TOKEN" \
		-H "Accept: application/vnd.github.v3.raw" \
		-o hack/crds/vitistack.io_machineproviders.yaml \
		https://api.github.com/repos/vitistack/crds/contents/crds/vitistack.io_machineproviders.yaml; then \
		echo "${RED}Error: Failed to download vitistack.io_machineproviders.yaml${RESET}"; \
		exit 1; \
	fi
	@echo "Downloading vitistack.io_machines.yaml..."
	@if ! curl --fail -H "Authorization: token $$GITHUB_TOKEN" \
		-H "Accept: application/vnd.github.v3.raw" \
		-o hack/crds/vitistack.io_machines.yaml \
		https://api.github.com/repos/vitistack/crds/contents/crds/vitistack.io_machines.yaml; then \
		echo "${RED}Error: Failed to download vitistack.io_machines.yaml${RESET}"; \
		exit 1; \
	fi
	@echo "Downloading vitistack.io_kubernetesclusters.yaml..."
	@if ! curl --fail -H "Authorization: token $$GITHUB_TOKEN" \
		-H "Accept: application/vnd.github.v3.raw" \
		-o hack/crds/vitistack.io_kubernetesclusters.yaml \
		https://api.github.com/repos/vitistack/crds/contents/crds/vitistack.io_kubernetesclusters.yaml; then \
		echo "${RED}Error: Failed to download vitistack.io_kubernetesclusters.yaml${RESET}"; \
		exit 1; \
	fi
	@echo "${GREEN}CRDs downloaded successfully${RESET}"

uninstall-crds: check-kubectl ## Uninstall CRDs into a cluster
	@echo "${RED}Uninstalling CRDs...${RESET}"
	${KUBECTL} delete -f hack/crds/

install-test-manifests: check-kubectl ## Install test resources (KubernetesProviders and MachineProviders)
	@echo "${GREEN}Installing test resources (KubernetesProviders and MachineProviders...)${RESET}"
	${KUBECTL} apply -f hack/test/manifests/

##@ Installation
.PHONY: install-lint install-gosec helm-install
install-lint: ## Install golangci-lint
	@echo "${YELLOW}Installing golangci-lint...${RESET}"
	${GO} install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
	@echo "${YELLOW}Then run 'make lint' to check for issues.${RESET}"

install-gosec: ## Install gosec security scanner
	@echo "${YELLOW}Installing gosec...${RESET}"
	${GO} install github.com/securego/gosec/v2/cmd/gosec@latest
	@echo "${YELLOW}Then run 'make gosec' to check for security issues.${RESET}"

helm-install: check-kubectl ## Install Helm chart with CRDs
	@echo "${GREEN}Installing Helm chart with CRDs...${RESET}"
	helm upgrade --install datacenter-operator ./charts/datacenter-operator --namespace default

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

##@ Kubernetes
.PHONY: token-kubernetes token-kubernetes-cleanup

token-kubernetes: check-kubectl ## Read the service account token from a pod in the specified namespace, specified by NAMESPACE variable (ex: make token-kubernetes NAMESPACE=default)
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

token-kubernetes-cleanup: check-kubectl ## Cleanup temporary pod created by token-kubernetes
	@echo "${YELLOW}>>> Cleaning up temporary pod token-extractor in namespace $(NAMESPACE)...${RESET}"
	@kubectl delete pod token-extractor -n $(NAMESPACE) --ignore-not-found
	@echo "${GREEN}Temporary pod cleaned up.${RESET}"