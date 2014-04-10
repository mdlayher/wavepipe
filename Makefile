# Name of the output binary
WP=wavepipe
# Full go import path of the project
WPPATH=github.com/mdlayher/${WP}

# Build the binary
make:
	go build -o bin/${WP}

# Format and error-check all files
fmt:
	go fmt ${WPPATH}
	go fmt ${WPPATH}/core
	golint .
	golint ./core
	errcheck ${WPPATH}
	errcheck ${WPPATH}/core

# Cross-compile the binary

darwin_386:
	GOOS="darwin" GOARCH="386" go build -o bin/${WP}_darwin_386

darwin_amd64:
	GOOS="darwin" GOARCH="amd64" go build -o bin/${WP}_darwin_amd64

linux_386:
	GOOS="linux" GOARCH="386" go build -o bin/${WP}_linux_386

linux_amd64:
	GOOS="linux" GOARCH="amd64" go build -o bin/${WP}_linux_amd64

windows_386:
	GOOS="windows" GOARCH="386" go build -o bin/${WP}_windows_386

windows_amd64:
	GOOS="windows" GOARCH="amd64" go build -o bin/${WP}_windows_amd64
