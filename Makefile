.PHONY: dev-speed
dev-speed:
	go build -ldflags="-s -w" -o ./dev-speed ./main.go 