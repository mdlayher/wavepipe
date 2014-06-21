.PHONY: bindata clean fmt test all

# Name of the output binary
WP=wavepipe

# Build the binary for the current platform
make:
	go build -o bin/${WP}

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
	golint .

# Run all tests
test:
	go test -v ./...
