.PHONY: build test render-example

build:
	go build -o bin/deployctl ./cmd/deployctl

test:
	go test ./...

render-example:
	go run ./cmd/deployctl render --config deploykit.example.yaml
