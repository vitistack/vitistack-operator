# datacenter-operator

API for centralizing data for datacenter infrastructure.

## Prerequsites

- Golang SDK

## Run

`go run cmd/api/main.go`

## Debug

- Open VS Code with the Golang extension installed
- Press `F5`
- Set some breakpoints

## Build a OCI container

- Build binary with `CGO_ENABLED=0 go build -o dist/dc-api -ldflags '-w -extldflags "-static"' cmd/api/main.go`
- Build the oci image: `docker build .`

## Modify CRDs

Prerequisites

- `go install sigs.k8s.io/controller-tools/cmd/controller-gen@latest`

1. edit files under `pkg/crds/*`
2. run `controller-gen object:headerFile="hacks/boilerplate.go.txt" paths="./pkg/crds/..."`
3. run `controller-gen crd paths=./pkg/crds/... output:crd:artifacts:config=hacks/crds`
4. install crds into cluster `kubectl apply -f hacks/crds`
