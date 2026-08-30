import type { Hono } from "hono";
import type { YjsCompactionService } from "../../../services/yjs_compaction_service.js";
import type { YjsDocumentInitializationService } from "../../../services/yjs_document_initialization_service.js";
import type { YjsProjectionService } from "../../../services/yjs_projection_service.js";
import type { Telemetry } from "../../../telemetry.js";
import { configureYjsCompactionEndpoint } from "../endpoints/yjs_compaction_endpoint.js";
import { configureYjsDocumentInitializationEndpoint } from "../endpoints/yjs_document_initialization_endpoint.js";
import { configureYjsProjectionEndpoint } from "../endpoints/yjs_projection_endpoint.js";

export function configureCoreRouter(
  app: Hono,
  yjsCompactionService: YjsCompactionService,
  yjsDocumentInitializationService: YjsDocumentInitializationService,
  yjsProjectionService: YjsProjectionService,
  telemetry: Telemetry
): void {
  configureYjsCompactionEndpoint(app, yjsCompactionService, telemetry);
  configureYjsDocumentInitializationEndpoint(
    app,
    yjsDocumentInitializationService,
    telemetry
  );
  configureYjsProjectionEndpoint(app, yjsProjectionService, telemetry);
}
