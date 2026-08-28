import { port } from "../util/environment.js";

export const config = {
  host: process.env.YJS_WORKER_HOST ?? "0.0.0.0",
  port: port("YJS_WORKER_PORT", 8787),
  telemetry: {
    serviceName: process.env.OTEL_SERVICE_NAME ?? "notegic-yjs-worker",
    serviceVersion: process.env.OTEL_SERVICE_VERSION ?? "0.1.0",
    deploymentEnvironment:
      process.env.OTEL_DEPLOYMENT_ENVIRONMENT ??
      process.env.NODE_ENV ??
      "development",
    serviceInstanceId:
      process.env.OTEL_SERVICE_INSTANCE_ID ?? process.env.HOSTNAME ?? "unknown",
    otlpEndpoint: (
      process.env.OTEL_EXPORTER_OTLP_ENDPOINT ??
      "http://notegic-otel-collector:4318"
    ).replace(/\/$/, ""),
  },
};
