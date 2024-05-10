.PHONY: build
build:
	go build -ldflags="-s -w" -o ./dev-speed ./main.go 

.PHONY: linux
linux:
	CGO_ENABLED=0  GOOS=linux  GOARCH=amd64 go build -ldflags="-s -w" -o ./dev-speed-linux ./main.go 