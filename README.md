# vitistack-operator

This operator provides an API for centralizing data for viti stack infrastructure.

## Prerequsites

- Golang SDK

## Run

`go run cmd/vitistack-operator/main.go`

## Debug

- Open VS Code with the Golang extension installed
- Press `F5`
- Set some breakpoints

## Build a OCI container

- Build binary with `CGO_ENABLED=0 go build -o dist/vitistack-operator -ldflags '-w -extldflags "-static"' cmd/vitistack-operator/main.go`
- Build the oci image: `docker build .`
