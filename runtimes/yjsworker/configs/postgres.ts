import { port, positiveInteger, required } from "../util/environment.js";
import { config as workerConfig } from "./config.js";

export type PostgresConfig = {
  host: string;
  port: number;
  user: string;
  password: string;
  database: string;
};

export const postgresConfig: PostgresConfig = {
  host: required("YJS_DB_HOST"),
  port: port("YJS_DB_PORT"),
  user: required("YJS_DB_USER"),
  password: required("YJS_DB_PASSWORD"),
  database: required("YJS_DB_NAME"),
};

export const yjsMaintenanceConfig = {
  maximumPendingHints: positiveInteger(
    "YJS_MAINTENANCE_MAXIMUM_PENDING_HINTS",
    1000
  ),
  maximumDispatchBatch: positiveInteger(
    "YJS_MAINTENANCE_MAXIMUM_DISPATCH_BATCH",
    32
  ),
  maximumDispatchWorkers: positiveInteger(
    "YJS_MAINTENANCE_MAXIMUM_DISPATCH_WORKERS",
    8
  ),
  maximumRequestAttempts: positiveInteger(
    "YJS_MAINTENANCE_MAXIMUM_REQUEST_ATTEMPTS",
    3
  ),
};

export const postgresPoolMaximumSize = positiveInteger(
  "YJS_DB_POOL_MAXIMUM_SIZE",
  10
);
export const postgresIdleTimeoutMilliseconds = positiveInteger(
  "YJS_DB_POOL_IDLE_TIMEOUT_MILLISECONDS",
  30000
);

export const postgresApplicationName =
  workerConfig.telemetry.serviceName ?? "notegic-yjs-worker";
