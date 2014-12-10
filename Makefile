.PHONY: bindata fmt race test

# Name of project
PROJECT=wavepipe

# Flags passed to Go linker, used to inject commit hash
LDFLAGS=-ldflags "-X main.version `git rev-parse HEAD`"

# Build the binary for the current platform
make:
	go build ${LDFLAGS} -o bin/${PROJECT} ./cmd/${PROJECT}/

# Build binary assets
bindata:
	go-bindata -pkg bindata -o ./bindata/bindata.go res/...

# Format, vet, and lint all files
fmt:
	go fmt ./...
	go vet ./...
	golint .

# Build the binary with the race detector enabled
race:
	go build -race ${LDFLAGS} -o bin/${PROJECT} ./cmd/${PROJECT}/

# Run all tests
test:
	go test -v ./...
