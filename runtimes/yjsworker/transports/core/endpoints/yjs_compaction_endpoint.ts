import {
  context as otelContext,
  propagation,
  SpanStatusCode,
} from "@opentelemetry/api";
import type { Hono } from "hono";
import { bodyLimit } from "hono/body-limit";
import { YjsMaintenanceMaximumPayloadBytes } from "../../../../../contracts/yjs-worker/v1/yjsworker_contract.js";
import type { YjsCompactionService } from "../../../services/yjs_compaction_service.js";
import type { Telemetry } from "../../../telemetry.js";

export function configureYjsCompactionEndpoint(
  app: Hono,
  yjsCompactionService: YjsCompactionService,
  telemetry: Telemetry
): void {
  app.post(
    "/core/yjs-compaction/v1",
    bodyLimit({
      maxSize: YjsMaintenanceMaximumPayloadBytes,
      onError: context => context.body(null, 413),
    }),
    async context => {
      const startedAt = performance.now();
      const contentLength = Number(context.req.header("content-length") ?? 0);
      if (
        !Number.isSafeInteger(contentLength) ||
        contentLength <= 0 ||
        contentLength > YjsMaintenanceMaximumPayloadBytes
      ) {
        return context.body(null, 413);
      }

      const payload = Buffer.from(await context.req.arrayBuffer());
      const parentContext = propagation.extract(
        otelContext.active(),
        context.req.raw.headers,
        {
          get: (headers, key) => headers.get(key) ?? undefined,
          keys: headers => [...headers.keys()],
        }
      );
      return otelContext.with(parentContext, () => {
        const span = telemetry.startSpan("maintenance.compaction_batch");
        span.setAttribute("yjs.payload_bytes", payload.length);

        try {
          const result = yjsCompactionService.compactBatch(payload);
          telemetry.recordOperation({
            operation: "maintenance.compaction_batch",
            outcome: "success",
            durationMilliseconds: performance.now() - startedAt,
            payloadBytes: payload.length,
          });

          return context.body(Uint8Array.from(result), 200, {
            "content-type": "application/octet-stream",
          });
        } catch (error) {
          span.recordException(error as Error);
          span.setStatus({ code: SpanStatusCode.ERROR });
          telemetry.recordOperation({
            operation: "maintenance.compaction_batch",
            outcome: "error",
            durationMilliseconds: performance.now() - startedAt,
            payloadBytes: payload.length,
            error,
          });
          return context.body(null, 422);
        } finally {
          span.end();
        }
      });
    }
  );
}
