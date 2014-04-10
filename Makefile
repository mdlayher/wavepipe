make:
	go build -o bin/wavepipe

fmt:
	go fmt
	golint .
	errcheck github.com/mdlayher/wavepipe

darwin_386:
	GOOS="darwin" GOARCH="386" go build -o bin/wavepipe_darwin_386

darwin_amd64:
	GOOS="darwin" GOARCH="amd64" go build -o bin/wavepipe_darwin_amd64

linux_386:
	GOOS="linux" GOARCH="386" go build -o bin/wavepipe_linux_386

linux_amd64:
	GOOS="linux" GOARCH="amd64" go build -o bin/wavepipe_linux_amd64

windows_386:
	GOOS="windows" GOARCH="386" go build -o bin/wavepipe_windows_386

windows_amd64:
	GOOS="windows" GOARCH="amd64" go build -o bin/wavepipe_windows_amd64
