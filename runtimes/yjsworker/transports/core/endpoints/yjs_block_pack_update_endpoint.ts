import {
  context as otelContext,
  propagation,
  SpanStatusCode,
} from "@opentelemetry/api";
import type { Hono } from "hono";
import { bodyLimit } from "hono/body-limit";
import type { UpdateBlockPackYjsDocumentRequest } from "../../../../../contracts/yjs-worker/v1/update_block_pack.js";
import { YjsMaintenanceMaximumPayloadBytes } from "../../../../../contracts/yjs-worker/v1/yjsworker_contract.js";
import type { Telemetry } from "../../../telemetry.js";
import type { RealtimeGateway } from "../../realtime/realtime_gateway.js";

export function configureYjsBlockPackUpdateEndpoint(
  app: Hono,
  realtimeGateway: RealtimeGateway,
  telemetry: Telemetry
): void {
  app.post(
    "/core/yjs-block-pack-update/v1",
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

      const parentContext = propagation.extract(
        otelContext.active(),
        context.req.raw.headers,
        {
          get: (headers, key) => headers.get(key) ?? undefined,
          keys: headers => [...headers.keys()],
        }
      );
      return otelContext.with(parentContext, async () => {
        const span = telemetry.startSpan("document.block_pack_update");
        try {
          const request =
            await context.req.json<UpdateBlockPackYjsDocumentRequest>();
          if (
            typeof request.blockPackId !== "string" ||
            !Array.isArray(request.blocks) ||
            request.blocks.length === 0 ||
            request.blocks.some(
              block =>
                typeof block.blockId !== "string" ||
                block.block === null ||
                typeof block.block !== "object"
            )
          ) {
            return context.body(null, 422);
          }

          const response = await realtimeGateway.updateBlockPack(request);
          telemetry.recordOperation({
            operation: "document.block_pack_update",
            outcome: "success",
            durationMilliseconds: performance.now() - startedAt,
            payloadBytes: contentLength,
          });

          return context.json(response);
        } catch (error) {
          span.recordException(error as Error);
          span.setStatus({ code: SpanStatusCode.ERROR });
          telemetry.recordOperation({
            operation: "document.block_pack_update",
            outcome: "error",
            durationMilliseconds: performance.now() - startedAt,
            payloadBytes: contentLength,
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
