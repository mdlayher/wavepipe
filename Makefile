.PHONY: bindata clean fmt install test

# Name of the output binary
WP=wavepipe

# Build the binary for the current platform
make:
	go build -ldflags "-X github.com/mdlayher/wavepipe/core.Revision `git rev-parse HEAD`" -o bin/${WP}

# Rebuild go-bindata files
bindata:
	go-bindata -ignore wavepipe.sql -o data/bindata.go res/...
	gofmt -r "main -> data" -w data/bindata.go

# Remove the bin folder
clean:
	rm -rf bin/

# Format and error-check all files
fmt:
	go fmt ./...
	go vet ./...
	golint .

# Copy binary into $GOPATH
install:
	cp bin/${WP} ${GOPATH}/bin/${WP}

# Run all tests
test:
	go test -v ./...
