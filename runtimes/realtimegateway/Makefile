.PHONY: test test-race test-e2e

test:
	go test ./...

test-race:
	go test -race ./...

test-e2e:
	@if [ -d test/e2e ]; then go test ./test/e2e/...; else echo "no RealtimeGateway E2E tests"; fi
