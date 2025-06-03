# datacenter-operator

API for centralizing data for datacenter infrastructure.

## Prerequsites

- Golang SDK

## Run

`go run cmd/datacenter-operator/main.go`

## Debug

- Open VS Code with the Golang extension installed
- Press `F5`
- Set some breakpoints

## Build a OCI container

- Build binary with `CGO_ENABLED=0 go build -o dist/datacenter-operator -ldflags '-w -extldflags "-static"' cmd/datacenter-operator/main.go`
- Build the oci image: `docker build .`
