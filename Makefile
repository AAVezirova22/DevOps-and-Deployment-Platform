.PHONY: build test render-example smoke-test check-prereqs

export GOCACHE ?= $(CURDIR)/.cache/go-build

build:
	go build -o bin/deployctl ./cmd/deployctl

test:
	go test ./...

render-example:
	go run ./cmd/deployctl render --config deploykit.example.yaml

smoke-test:
	scripts/kind-smoke-test.sh

check-prereqs:
	scripts/check-prereqs.sh
