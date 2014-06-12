# Name of the output binary
WP=wavepipe

# Build the binary for the current platform
make:
	go build -o bin/${WP}

# Rebuild go-bindata files
bindata:
	go-bindata -o data/bindata.go res/sqlite/
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

# Cross-compile the binary

all:
	make darwin_386
	make darwin_amd64
	make linux_386
	make linux_amd64
	make windows_386
	make windows_amd64

darwin_386:
	GOOS="darwin" GOARCH="386" go build -o bin/${WP}_darwin_386

darwin_amd64:
	GOOS="darwin" GOARCH="amd64" go build -o bin/${WP}_darwin_amd64

linux_386:
	GOOS="linux" GOARCH="386" go build -o bin/${WP}_linux_386

linux_amd64:
	GOOS="linux" GOARCH="amd64" go build -o bin/${WP}_linux_amd64

windows_386:
	GOOS="windows" GOARCH="386" go build -o bin/${WP}_windows_386.exe

windows_amd64:
	GOOS="windows" GOARCH="amd64" go build -o bin/${WP}_windows_amd64.exe
