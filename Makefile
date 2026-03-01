.PHONY: build build-gui test test-gui

build:
	go build -o vb .

build-gui:
	CGO_ENABLED=1 go build -tags gui -o vb .

test:
	go test ./...

test-gui:
	CGO_ENABLED=1 go test -tags gui ./...
