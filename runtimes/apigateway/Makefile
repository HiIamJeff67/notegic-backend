.PHONY: test test-race test-e2e build-api-gateway

build-api-gateway:
	go build -o ./tmp/apigateway ./commands

test:
	go test ./...

test-race:
	go test -race ./...

test-e2e:
	go test ./test/e2e/...
