all: local

fmt:
	go fmt ./...

lint:
	golangci-lint run

test:
	go test -race -cover ./...

local: lint test
	CGO_ENABLED=0 go build ${--local-args} -o . ./cmd/...
