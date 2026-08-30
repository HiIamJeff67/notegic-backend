.PHONY: test test-race test-e2e build-client-gateway

build-client-gateway:
	go build -o ./tmp/clientgateway ./commands

test:
	go test ./...

test-race:
	go test -race ./...

test-e2e:
	go test ./test/e2e/...
