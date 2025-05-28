# Makefile for datacenter-operator

# Variables
BINARY_NAME=datacenter-operator
GO=go
GOFMT=gofmt
DOCKER=docker
KUBECTL=kubectl
DOCKER_IMAGE=ghcr.io/vitistack/datacenter-operator
DOCKER_TAG=latest
MAIN_PATH=./cmd/api/main.go
CONTROLLER-GEN=controller-gen

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
run: check-go ## Run the application
	@echo "${GREEN}Running ${BINARY_NAME}...${RESET}"
	${GO} run ${MAIN_PATH}

##@ Docker
docker-build: check-docker ## Build Docker image
	@echo "${CYAN}Building Docker image...${RESET}"
	${DOCKER} build -t ${DOCKER_IMAGE}:${DOCKER_TAG} .

docker-push: check-docker ## Push Docker image
	@echo "${CYAN}Pushing Docker image...${RESET}"
	${DOCKER} push ${DOCKER_IMAGE}:${DOCKER_TAG}

##@ CRDs & Resources
generate-crds: check-go ## Generate CRDs and copy to Helm chart
	@echo "${GREEN}Generating CRDs...${RESET}"
	${CONTROLLER-GEN} object:headerFile="hacks/boilerplate.go.txt" paths="./pkg/crds/..."
	${CONTROLLER-GEN} crd paths=./pkg/crds/... output:crd:artifacts:config=hacks/crds
	@echo "${GREEN}Copying CRDs to Helm chart...${RESET}"
	@mkdir -p charts/datacenter-operator/crds
	@cp hacks/crds/*.yaml charts/datacenter-operator/crds/

install-configmap: check-kubectl ## Install configmap into cluster
	@echo "${GREEN}Installing configmap into cluster...${RESET}"
	${KUBECTL} apply -f hacks/test/manifests/configmap.yaml

uninstall-configmap: check-kubectl ## Uninstall configmap into cluster
	@echo "${GREEN}Installing configmap into cluster...${RESET}"
	${KUBECTL} delete -f hacks/test/manifests/configmap.yaml

install-crds: check-kubectl ## Install CRDs into a cluster
	@echo "${GREEN}Installing CRDs...${RESET}"
	${KUBECTL} apply -f hacks/crds/

uninstall-crds: check-kubectl ## Uninstall CRDs into a cluster
	@echo "${GREEN}Installing CRDs...${RESET}"
	${KUBECTL} delete -f hacks/crds/

install-crds-from-chart: check-kubectl ## Install CRDs from Helm chart
	@echo "${GREEN}Installing CRDs from Helm chart...${RESET}"
	${KUBECTL} apply -f charts/datacenter-operator/crds/

install-test-manifests: check-kubectl ## Install test resources (KubernetesProviders and MachineProviders)
	@echo "${GREEN}Installing test resources (KubernetesProviders and MachineProviders...)${RESET}"
	${KUBECTL} apply -f hacks/test/manifests/

##@ Installation
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
