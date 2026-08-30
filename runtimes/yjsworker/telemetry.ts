import { monitorEventLoopDelay, PerformanceObserver } from "node:perf_hooks";

import {
  type Context,
  context,
  isSpanContextValid,
  metrics,
  type Span,
  trace,
} from "@opentelemetry/api";
import { logs, SeverityNumber } from "@opentelemetry/api-logs";
import { OTLPLogExporter } from "@opentelemetry/exporter-logs-otlp-proto";
import { OTLPMetricExporter } from "@opentelemetry/exporter-metrics-otlp-proto";
import { OTLPTraceExporter } from "@opentelemetry/exporter-trace-otlp-proto";
import { resourceFromAttributes } from "@opentelemetry/resources";
import { BatchLogRecordProcessor } from "@opentelemetry/sdk-logs";
import { PeriodicExportingMetricReader } from "@opentelemetry/sdk-metrics";
import { NodeSDK } from "@opentelemetry/sdk-node";
import {
  ATTR_SERVICE_INSTANCE_ID,
  ATTR_SERVICE_NAME,
  ATTR_SERVICE_VERSION,
} from "@opentelemetry/semantic-conventions";
import { config } from "./configs/config.js";
import { Logger } from "./util/logger.js";

type RoomState = {
  activeRooms: number;
  activeSubscribers: number;
  internalSockets: number;
};

export class Telemetry {
  private readonly eventLoopDelay = monitorEventLoopDelay({ resolution: 20 });
  private readonly gcObserver: PerformanceObserver;
  private readonly gcDuration;
  private readonly internalSocketCount;
  private readonly logger;
  private readonly operationCount;
  private readonly operationDuration;
  private readonly payloadBytes;
  private readonly sdk: NodeSDK;
  private readonly tracer;
  private readonly consoleLogger: Logger;
  private roomStateProvider: () => RoomState = () => ({
    activeRooms: 0,
    activeSubscribers: 0,
    internalSockets: 0,
  });

  private constructor(sdk: NodeSDK, consoleLogger: Logger) {
    this.sdk = sdk;
    this.consoleLogger = consoleLogger;

    const meter = metrics.getMeter("notegic.yjs-worker");
    this.logger = logs.getLogger("notegic.yjs-worker");
    this.tracer = trace.getTracer("notegic.yjs-worker");
    this.operationCount = meter.createCounter(
      "notegic.yjs.worker.operation.count"
    );
    this.operationDuration = meter.createHistogram(
      "notegic.yjs.worker.operation.duration"
    );
    this.payloadBytes = meter.createCounter("notegic.yjs.worker.payload.bytes");
    this.internalSocketCount = meter.createUpDownCounter(
      "notegic.yjs.worker.internal_socket.count"
    );
    this.gcDuration = meter.createHistogram("notegic.yjs.worker.gc.duration");
    this.gcObserver = new PerformanceObserver(entries => {
      for (const entry of entries.getEntries()) {
        this.gcDuration.record(entry.duration);
      }
    });

    const activeRoomGauge = meter.createObservableGauge(
      "notegic.yjs.worker.active_room.count"
    );
    const activeSubscriberGauge = meter.createObservableGauge(
      "notegic.yjs.worker.active_subscriber.count"
    );
    const heapGauge = meter.createObservableGauge(
      "notegic.yjs.worker.process.heap.bytes"
    );
    const rssGauge = meter.createObservableGauge(
      "notegic.yjs.worker.process.rss.bytes"
    );
    const externalGauge = meter.createObservableGauge(
      "notegic.yjs.worker.process.external.bytes"
    );
    const arrayBuffersGauge = meter.createObservableGauge(
      "notegic.yjs.worker.process.array_buffers.bytes"
    );
    const eventLoopDelayGauge = meter.createObservableGauge(
      "notegic.yjs.worker.event_loop_delay.milliseconds"
    );
    const uptimeGauge = meter.createObservableGauge(
      "notegic.yjs.worker.process.uptime.seconds"
    );
    meter.addBatchObservableCallback(
      result => {
        const roomState = this.roomStateProvider();
        const memoryUsage = process.memoryUsage();
        result.observe(activeRoomGauge, roomState.activeRooms);
        result.observe(activeSubscriberGauge, roomState.activeSubscribers);
        result.observe(heapGauge, memoryUsage.heapUsed);
        result.observe(rssGauge, memoryUsage.rss);
        result.observe(externalGauge, memoryUsage.external);
        result.observe(arrayBuffersGauge, memoryUsage.arrayBuffers);

        const mean = this.eventLoopDelay.mean / 1_000_000;
        if (!Number.isNaN(mean)) {
          result.observe(eventLoopDelayGauge, mean);
        }
        result.observe(uptimeGauge, process.uptime());
      },
      [
        activeRoomGauge,
        activeSubscriberGauge,
        heapGauge,
        rssGauge,
        externalGauge,
        arrayBuffersGauge,
        eventLoopDelayGauge,
        uptimeGauge,
      ]
    );

    this.eventLoopDelay.enable();
    this.gcObserver.observe({ entryTypes: ["gc"] });
  }

  static initialize(logger = new Logger()): Telemetry {
    const resource = resourceFromAttributes({
      [ATTR_SERVICE_NAME]: config.telemetry.serviceName,
      [ATTR_SERVICE_VERSION]: config.telemetry.serviceVersion,
      [ATTR_SERVICE_INSTANCE_ID]: config.telemetry.serviceInstanceId,
      "deployment.environment": config.telemetry.deploymentEnvironment,
    });
    const endpoint = config.telemetry.otlpEndpoint;
    const sdk = new NodeSDK({
      autoDetectResources: false,
      resource,
      traceExporter: new OTLPTraceExporter({ url: `${endpoint}/v1/traces` }),
      metricReaders: [
        new PeriodicExportingMetricReader({
          exporter: new OTLPMetricExporter({ url: `${endpoint}/v1/metrics` }),
          exportIntervalMillis: 15_000,
        }),
      ],
      logRecordProcessors: [
        new BatchLogRecordProcessor({
          exporter: new OTLPLogExporter({ url: `${endpoint}/v1/logs` }),
        }),
      ],
    });

    sdk.start();

    return new Telemetry(sdk, logger);
  }

  log(
    severityNumber: SeverityNumber,
    eventName: string,
    attributes: Record<string, boolean | number | string> = {},
    error?: unknown
  ): void {
    const level =
      severityNumber >= SeverityNumber.ERROR
        ? "error"
        : severityNumber >= SeverityNumber.WARN
          ? "warn"
          : "info";
    const activeSpanContext = trace.getActiveSpan()?.spanContext();
    const record = {
      level,
      event: eventName,
      ...attributes,
      ...(error instanceof Error ? { errorType: error.name } : {}),
      ...(activeSpanContext !== undefined &&
      isSpanContextValid(activeSpanContext)
        ? { traceId: activeSpanContext.traceId }
        : {}),
    };

    this.consoleLogger.log(level, JSON.stringify(record));
    this.logger.emit({
      context: context.active(),
      severityNumber,
      severityText: SeverityNumber[severityNumber],
      eventName,
      body: eventName,
      attributes: {
        ...attributes,
        ...(error instanceof Error ? { "error.type": error.name } : {}),
      },
    });
  }

  recordInternalSocket(delta: number): void {
    this.internalSocketCount.add(delta);
  }

  recordOperation({
    operation,
    outcome,
    durationMilliseconds,
    payloadBytes = 0,
    error,
  }: {
    operation: string;
    outcome: "success" | "error";
    durationMilliseconds: number;
    payloadBytes?: number;
    error?: unknown;
  }): void {
    const attributes = { operation, outcome };
    this.operationCount.add(1, attributes);
    this.operationDuration.record(durationMilliseconds, attributes);

    if (payloadBytes > 0) {
      this.payloadBytes.add(payloadBytes, { operation });
    }
    if (outcome === "error") {
      this.log(
        SeverityNumber.ERROR,
        "yjs.operation.failed",
        { operation },
        error
      );
    }
  }

  setRoomStateProvider(provider: () => RoomState): void {
    this.roomStateProvider = provider;
  }

  startSpan(
    operation: string,
    parentContext: Context = context.active()
  ): Span {
    return this.tracer.startSpan(`yjs.${operation}`, undefined, parentContext);
  }

  async shutdown(): Promise<void> {
    this.eventLoopDelay.disable();
    this.gcObserver.disconnect();

    try {
      await this.sdk.shutdown();
    } catch (error) {
      this.consoleLogger.warn(
        JSON.stringify({
          level: "warn",
          event: "telemetry.shutdown_failed",
          ...(error instanceof Error ? { errorType: error.name } : {}),
        })
      );
    }
  }
}
