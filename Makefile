.PHONY: build-linux build-linux-aarch64 build-darwin build-darwin-aarch64 build-win build-win-aarch64 

all: build-linux build-linux-aarch64 build-darwin build-darwin-aarch64 build-win build-win-aarch64

build-linux:
	GOOS=linux GOARCH=amd64 go build -o rndua-linux-amd64 rndua.go

build-linux-aarch64:
	GOOS=linux GOARCH=arm64 go build -o rndua-linux-arm64 rndua.go

build-darwin:
	GOOS=darwin GOARCH=amd64 go build -o rndua-darwin-amd64 rndua.go

build-darwin-aarch64:
	GOOS=darwin GOARCH=arm64 go build -o rndua-darwin-arm64 rndua.go

build-win:
	GOOS=windows GOARCH=amd64 go build -o rndua-windows-amd64.exe rndua.go

build-win-aarch64:
	GOOS=windows GOARCH=arm64 go build -o rndua-windows-arm64.exe rndua.go

