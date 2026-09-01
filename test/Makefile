.PHONY: test test-architecture test-integration test-integration-run test-integration-kafka test-integration-kafka-run test-load-websocket test-soak-websocket test-load-kafka-lag

KAFKA_BROKERS ?= 127.0.0.1:19094

test:
	go test ./...

test-architecture:
	go test ./architecture

test-integration:
	$(MAKE) test-integration-run

test-integration-run:
	@if [ -d integration ]; then NOTEGIC_RUN_INTEGRATION=1 KAFKA_BROKERS=$(KAFKA_BROKERS) go test ./integration/...; else echo "no root integration tests"; fi

test-integration-kafka:
	$(MAKE) test-integration-kafka-run

test-integration-kafka-run:
	@NOTEGIC_RUN_INTEGRATION=1 KAFKA_BROKERS=$(KAFKA_BROKERS) go test -count=1 ./integration/kafka

test-load-websocket:
	@command -v k6 >/dev/null 2>&1 || { echo "k6 is required; see test/load/README.md"; exit 1; }
	REALTIME_GATEWAY_WS_URL="$(REALTIME_GATEWAY_WS_URL)" \
	REALTIME_CHANNEL_TICKET="$(REALTIME_CHANNEL_TICKET)" \
	K6_VUS="$(K6_VUS)" K6_DURATION="$(K6_DURATION)" \
	k6 run load/websocket.js

test-soak-websocket: test-load-websocket

test-load-kafka-lag:
	@command -v k6 >/dev/null 2>&1 || { echo "k6 is required; see test/load/README.md"; exit 1; }
	KAFKA_METRICS_URL="$(KAFKA_METRICS_URL)" \
	KAFKA_LAG_THRESHOLD="$(KAFKA_LAG_THRESHOLD)" \
	K6_VUS="$(K6_VUS)" K6_DURATION="$(K6_DURATION)" \
	k6 run load/kafka_consumer_lag.js
