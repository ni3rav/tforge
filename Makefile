.PHONY: fmt test build vet check

fmt:
	gofmt -w $(shell find . -name '*.go' -print)

test:
	go test -count=1 ./...

build:
	go build ./...

vet:
	go vet ./...

check: test build vet
