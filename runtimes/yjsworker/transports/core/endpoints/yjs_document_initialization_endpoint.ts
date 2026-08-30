import {
  context as otelContext,
  propagation,
  SpanStatusCode,
} from "@opentelemetry/api";
import type { Hono } from "hono";
import { bodyLimit } from "hono/body-limit";
import { YjsMaintenanceMaximumPayloadBytes } from "../../../../../contracts/yjs-worker/v1/yjsworker_contract.js";
import type { YjsDocumentInitializationService } from "../../../services/yjs_document_initialization_service.js";
import type { Telemetry } from "../../../telemetry.js";
import type { YjsDocumentInitializationRequest } from "../../../types/yjs_document_initialization_type.js";

export function configureYjsDocumentInitializationEndpoint(
  app: Hono,
  yjsDocumentInitializationService: YjsDocumentInitializationService,
  telemetry: Telemetry
): void {
  app.post(
    "/core/yjs-document-initialization/v1",
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
        const span = telemetry.startSpan("document.initialization_batch");
        try {
          const request =
            await context.req.json<YjsDocumentInitializationRequest>();
          if (
            !Array.isArray(request.documents) ||
            request.documents.length === 0 ||
            request.documents.some(document => !Array.isArray(document.blocks))
          ) {
            return context.body(null, 422);
          }

          const documents = request.documents.map(document => {
            const result = yjsDocumentInitializationService.initialize(
              document.blocks
            );
            return {
              snapshot: result.snapshot.toString("base64"),
              stateVector: result.stateVector.toString("base64"),
            };
          });
          telemetry.recordOperation({
            operation: "document.initialization_batch",
            outcome: "success",
            durationMilliseconds: performance.now() - startedAt,
            payloadBytes: contentLength,
          });

          return context.json({ documents });
        } catch (error) {
          span.recordException(error as Error);
          span.setStatus({ code: SpanStatusCode.ERROR });
          telemetry.recordOperation({
            operation: "document.initialization_batch",
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
