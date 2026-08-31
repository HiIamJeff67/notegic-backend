import { serve } from "@hono/node-server";
import { SeverityNumber } from "@opentelemetry/api-logs";
import { Hono } from "hono";
import { WebSocketServer } from "ws";

import { config } from "./configs/config.js";
import { YjsPostgresRepository } from "./data/postgres/repository.js";
import { BlockPackProjector } from "./services/block_pack_projector.js";
import { YjsBlockPackUpdateService } from "./services/yjs_block_pack_update_service.js";
import { YjsCompactionService } from "./services/yjs_compaction_service.js";
import { YjsDocumentInitializationService } from "./services/yjs_document_initialization_service.js";
import { YjsProjectionService } from "./services/yjs_projection_service.js";
import type { Telemetry } from "./telemetry.js";
import { YjsMaintenanceConsumer } from "./transports/core/consumers/yjs_maintenance_consumer.js";
import { CoreCommandDispatcher } from "./transports/core/dispatchers/core_command_dispatcher.js";
import { configureCoreRouter } from "./transports/core/routers/core_router.js";
import { RealtimeGateway } from "./transports/realtime/realtime_gateway.js";
import { configureRealtimeRouter } from "./transports/realtime/realtime_router.js";
import { RoomRegistry } from "./transports/realtime/room_registry.js";
import { configureHealthRouter } from "./transports/status/health_router.js";
import { configureStartedRouter } from "./transports/status/started_router.js";
import { Logger } from "./util/logger.js";

export class YjsWorkerServer {
  private readonly server: ReturnType<typeof serve>;
  private readonly webSocketServer: WebSocketServer;
  private readonly realtimeGateway: RealtimeGateway;
  private readonly coreCommandDispatcher: CoreCommandDispatcher;
  private readonly yjsMaintenanceConsumer: YjsMaintenanceConsumer;
  private readonly yjsPostgresRepository: YjsPostgresRepository;
  private healthy = false;
  private ready = false;

  constructor(telemetry: Telemetry, logger = new Logger()) {
    const app = new Hono();
    const blockPackProjector = new BlockPackProjector();
    const yjsCompactionService = new YjsCompactionService(telemetry);
    const yjsDocumentInitializationService =
      new YjsDocumentInitializationService();
    const yjsProjectionService = new YjsProjectionService(
      blockPackProjector,
      telemetry
    );
    const roomRegistry = new RoomRegistry(telemetry);
    this.yjsPostgresRepository = new YjsPostgresRepository();
    this.coreCommandDispatcher = new CoreCommandDispatcher(logger);
    const yjsBlockPackUpdateService = new YjsBlockPackUpdateService(
      this.coreCommandDispatcher,
      blockPackProjector
    );
    this.yjsMaintenanceConsumer = new YjsMaintenanceConsumer(
      this.yjsPostgresRepository,
      yjsCompactionService,
      yjsProjectionService,
      logger
    );
    this.webSocketServer = new WebSocketServer({ noServer: true });
    this.realtimeGateway = new RealtimeGateway(
      roomRegistry,
      yjsCompactionService,
      this.coreCommandDispatcher,
      telemetry,
      logger,
      yjsBlockPackUpdateService
    );
    void this.yjsMaintenanceConsumer
      .start()
      .then(() => {
        this.ready = true;
      })
      .catch(error => {
        telemetry.log(
          SeverityNumber.ERROR,
          "yjs_maintenance_consumer.start_failed",
          {
            error: error instanceof Error ? error.message : String(error),
          }
        );
      });

    configureStartedRouter(app, () => this.isHealthy());
    configureHealthRouter(app, () => this.isReady());
    configureRealtimeRouter(
      app,
      this.realtimeGateway.handleConnection.bind(this.realtimeGateway)
    );
    configureCoreRouter(
      app,
      yjsCompactionService,
      yjsDocumentInitializationService,
      yjsProjectionService,
      this.realtimeGateway,
      telemetry
    );

    this.server = serve(
      {
        fetch: app.fetch,
        hostname: config.host,
        port: config.port,
        websocket: { server: this.webSocketServer },
      },
      () => {
        this.healthy = true;
        telemetry.log(SeverityNumber.INFO, "yjs_worker.started", {
          host: config.host,
          port: config.port,
        });
      }
    );
  }

  isHealthy(): boolean {
    return this.healthy;
  }

  isReady(): boolean {
    return this.ready;
  }

  async shutdown(): Promise<void> {
    this.ready = false;
    this.healthy = false;
    const closeServer = new Promise<void>(resolve => {
      this.server.close(() => resolve());
    });

    await this.realtimeGateway.shutdown();
    await this.yjsMaintenanceConsumer.shutdown();
    await this.coreCommandDispatcher.shutdown();
    await this.yjsPostgresRepository.close();
    await closeServer;
  }
}
