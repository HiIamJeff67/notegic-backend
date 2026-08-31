import { randomUUID } from "node:crypto";
import { WebSocket } from "ws";
import * as Y from "yjs";
import type {
  UpdateBlockPackYjsDocumentRequest,
  UpdateBlockPackYjsDocumentResponse,
} from "../../../../contracts/yjs-worker/v1/update_block_pack.js";
import {
  BlockPackDocumentQuotaPolicyVersion,
  type BlockPackQuotaPolicy,
  YjsCompactionUpdateThreshold,
} from "../../../../contracts/yjs-worker/v1/yjsworker_contract.js";
import { YjsPersistenceBatchShutdownTimeoutMilliseconds } from "../../configs/persistence_config.js";
import {
  YjsPendingDocumentMaximumPayloadBytes,
  YjsPendingDocumentMaximumUpdateCount,
} from "../../configs/realtime_config.js";
import { BlockPackProjector } from "../../services/block_pack_projector.js";
import type { YjsBlockPackUpdateService } from "../../services/yjs_block_pack_update_service.js";
import type { YjsCompactionService } from "../../services/yjs_compaction_service.js";
import type { Telemetry } from "../../telemetry.js";
import {
  createInternalFrame,
  type InternalFrame,
  parseInternalFrame,
} from "../../types/internal_frame.js";
import { InternalFrameType } from "../../types/internal_frame_type.js";
import type { Room } from "../../types/room.js";
import {
  createYjsCompactionResult,
  parseYjsCompactionInput,
} from "../../types/yjs_compaction.js";
import {
  parseYjsDocumentState,
  parseYjsUpdateSequence,
} from "../../types/yjs_document_state.js";
import { Logger } from "../../util/logger.js";
import {
  CoreCommandDispatcher,
  CoreCommandError,
} from "../core/dispatchers/core_command_dispatcher.js";
import type { RoomRegistry } from "./room_registry.js";
import { YjsDebouncer } from "./yjs_debouncer.js";

export class RealtimeGateway {
  private readonly roomRegistry: RoomRegistry;
  private readonly blockPackProjector: BlockPackProjector;
  private readonly yjsCompactionService: YjsCompactionService;
  private readonly coreCommandDispatcher: CoreCommandDispatcher;
  private readonly webSockets = new Set<WebSocket>();
  private readonly pendingPersistenceBatches = new Map<string, string>();
  private readonly yjsDebouncer: YjsDebouncer;
  private readonly telemetry: Telemetry;
  private readonly logger: Logger;
  private readonly yjsBlockPackUpdateService: YjsBlockPackUpdateService | null;

  constructor(
    roomRegistry: RoomRegistry,
    yjsCompactionService: YjsCompactionService,
    coreCommandDispatcher: CoreCommandDispatcher,
    telemetry: Telemetry,
    logger = new Logger(),
    yjsBlockPackUpdateService?: YjsBlockPackUpdateService
  ) {
    this.roomRegistry = roomRegistry;
    this.blockPackProjector = new BlockPackProjector();
    this.yjsCompactionService = yjsCompactionService;
    this.coreCommandDispatcher = coreCommandDispatcher;
    this.yjsBlockPackUpdateService = yjsBlockPackUpdateService ?? null;
    this.logger = logger;
    this.yjsDebouncer = new YjsDebouncer(
      telemetry,
      this.resyncRoom.bind(this),
      async (blockPackId, persistenceBatchId, originConnectionId, payload) => {
        const startedAt = performance.now();
        let commandId: string | null = null;
        try {
          const dispatched = await this.coreCommandDispatcher.dispatchAsync<
            {
              persistenceBatchId: string;
              originConnectionId: string | null;
              payload: string;
            },
            { updateSequence: number }
          >("AppendYjsUpdate", blockPackId, {
            persistenceBatchId,
            originConnectionId,
            payload: payload.toString("base64"),
          });
          commandId = dispatched.commandId;
          this.pendingPersistenceBatches.set(persistenceBatchId, blockPackId);
          const response = await dispatched.reply;
          this.pendingPersistenceBatches.delete(persistenceBatchId);

          return response.updateSequence;
        } catch (error) {
          this.pendingPersistenceBatches.delete(persistenceBatchId);
          this.logCoreCommandFailure(
            "AppendYjsUpdate",
            blockPackId,
            commandId,
            startedAt,
            error,
            persistenceBatchId
          );
          throw error;
        }
      },
      this.handleYjsUpdatePersisted.bind(this)
    );
    this.telemetry = telemetry;
    this.telemetry.setRoomStateProvider(() => ({
      activeRooms: this.roomRegistry.size,
      activeSubscribers: this.roomRegistry.subscriberCount,
      internalSockets: this.webSockets.size,
    }));
  }

  /* ============================== Internal frame delivery ============================== */

  private static broadcastInternalFrame(
    room: Room,
    type: InternalFrameType,
    blockPackId: string,
    payload: Buffer
  ): void {
    for (const subscriber of room.subscribers.values()) {
      if (
        !subscriber.isReady &&
        (type === InternalFrameType.InternalFrameType_YjsDocument ||
          type === InternalFrameType.InternalFrameType_Awareness)
      ) {
        continue;
      }

      if (subscriber.webSocket.readyState === WebSocket.OPEN) {
        subscriber.webSocket.send(
          createInternalFrame(
            type,
            subscriber.connectionId,
            subscriber.connectorChannelId,
            blockPackId,
            payload
          )
        );
      }
    }
  }

  private sendRoomInitialState(room: Room, blockPackId: string): void {
    if (room.document === null) {
      return;
    }

    const payload = Buffer.from(Y.encodeStateAsUpdate(room.document));
    const awarenessPayload = this.roomRegistry.getAwarenessSnapshot(room);
    for (const subscriber of room.subscribers.values()) {
      if (subscriber.isReady) {
        continue;
      }

      if (subscriber.webSocket.readyState !== WebSocket.OPEN) {
        continue;
      }

      subscriber.isReady = true;
      subscriber.webSocket.send(
        createInternalFrame(
          InternalFrameType.InternalFrameType_Attached,
          subscriber.connectionId,
          subscriber.connectorChannelId,
          blockPackId
        )
      );
      subscriber.webSocket.send(
        createInternalFrame(
          InternalFrameType.InternalFrameType_YjsDocument,
          subscriber.connectionId,
          subscriber.connectorChannelId,
          blockPackId,
          payload
        )
      );
      if (awarenessPayload !== null) {
        subscriber.webSocket.send(
          createInternalFrame(
            InternalFrameType.InternalFrameType_Awareness,
            subscriber.connectionId,
            subscriber.connectorChannelId,
            blockPackId,
            awarenessPayload
          )
        );
      }
    }
  }

  private handleYjsDocumentUpdate(
    room: Room,
    webSocket: WebSocket | null,
    frame: InternalFrame
  ): void {
    const subscriber = room.subscribers.get(
      `${frame.connectionId}:${frame.connectorChannelId}`
    );
    if (subscriber === undefined) {
      if (webSocket !== null && webSocket.readyState === WebSocket.OPEN) {
        webSocket.send(
          createInternalFrame(
            InternalFrameType.InternalFrameType_ResyncRequired,
            frame.connectionId,
            frame.connectorChannelId,
            frame.blockPackId
          )
        );
      }

      return;
    }
    if (subscriber.quotaRecoveryRequired) {
      if (webSocket !== null && webSocket.readyState === WebSocket.OPEN) {
        webSocket.send(
          createInternalFrame(
            InternalFrameType.InternalFrameType_BlockPackQuotaExceeded,
            frame.connectionId,
            frame.connectorChannelId,
            frame.blockPackId
          )
        );
      }

      return;
    }
    if (room.document === null) {
      if (
        room.pendingYjsUpdates.length >= YjsPendingDocumentMaximumUpdateCount ||
        room.pendingYjsPayloadBytes + frame.payload.length >
          YjsPendingDocumentMaximumPayloadBytes
      ) {
        if (webSocket !== null && webSocket.readyState === WebSocket.OPEN) {
          webSocket.send(
            createInternalFrame(
              InternalFrameType.InternalFrameType_ResyncRequired,
              frame.connectionId,
              frame.connectorChannelId,
              frame.blockPackId
            )
          );
        }

        return;
      }

      room.pendingYjsUpdates.push({ webSocket, frame });
      room.pendingYjsPayloadBytes += frame.payload.length;

      return;
    }
    if (room.validationDocument === null) {
      if (webSocket !== null && webSocket.readyState === WebSocket.OPEN) {
        webSocket.send(
          createInternalFrame(
            InternalFrameType.InternalFrameType_ResyncRequired,
            frame.connectionId,
            frame.connectorChannelId,
            frame.blockPackId
          )
        );
      }

      return;
    }

    let blockCount: number;
    try {
      Y.applyUpdate(room.validationDocument, frame.payload);
      blockCount = this.blockPackProjector.countYjsDocumentBlocks(
        room.validationDocument
      );
    } catch {
      room.validationDocument.destroy();
      room.validationDocument = new Y.Doc();
      Y.applyUpdate(
        room.validationDocument,
        Y.encodeStateAsUpdate(room.document)
      );
      if (webSocket !== null && webSocket.readyState === WebSocket.OPEN) {
        webSocket.send(
          createInternalFrame(
            InternalFrameType.InternalFrameType_ResyncRequired,
            frame.connectionId,
            frame.connectorChannelId,
            frame.blockPackId
          )
        );
      }

      return;
    }

    if (blockCount > room.maximumBlockCount) {
      room.validationDocument.destroy();
      room.validationDocument = new Y.Doc();
      Y.applyUpdate(
        room.validationDocument,
        Y.encodeStateAsUpdate(room.document)
      );
      subscriber.quotaRecoveryRequired = true;
      this.telemetry.recordOperation({
        operation: "document.update",
        outcome: "error",
        durationMilliseconds: 0,
        payloadBytes: frame.payload.length,
      });
      if (webSocket !== null && webSocket.readyState === WebSocket.OPEN) {
        webSocket.send(
          createInternalFrame(
            InternalFrameType.InternalFrameType_BlockPackQuotaExceeded,
            frame.connectionId,
            frame.connectorChannelId,
            frame.blockPackId
          )
        );
      }

      return;
    }

    this.yjsDebouncer.queueUpdate(room, { webSocket, frame });
  }

  private resyncRoom(room: Room, blockPackId: string): void {
    const resyncError = new Error("Yjs room resync requested");
    this.logger.error("[YjsWorker] room resync requested", {
      blockPackId,
      lastUpdateSequence: room.lastUpdateSequence,
      compactedUntilSequence: room.compactedUntilSequence,
      projectedUntilSequence: room.projectedUntilSequence,
      pendingYjsUpdates: room.pendingYjsUpdates.length,
      pendingPersistenceUpdates: room.pendingPersistenceUpdates.length,
      pendingPersistenceBatches: [
        ...this.pendingPersistenceBatches.values(),
      ].filter(pendingBlockPackId => pendingBlockPackId === blockPackId).length,
      inFlightPersistenceBatchId:
        room.inFlightPersistenceBatch?.persistenceBatchId ?? null,
      subscribers: room.subscribers.size,
      isLoading: room.isLoading,
      stack: resyncError.stack,
    });
    for (const [persistenceBatchId, pendingBlockPackId] of this
      .pendingPersistenceBatches) {
      if (pendingBlockPackId === blockPackId) {
        this.pendingPersistenceBatches.delete(persistenceBatchId);
      }
    }

    if (room.projectionTimer !== null) {
      clearTimeout(room.projectionTimer);
    }
    if (room.persistenceDebounceTimer !== null) {
      clearTimeout(room.persistenceDebounceTimer);
    }
    if (room.persistenceMaximumWaitTimer !== null) {
      clearTimeout(room.persistenceMaximumWaitTimer);
    }
    if (room.persistenceRetryTimer !== null) {
      clearTimeout(room.persistenceRetryTimer);
    }

    const awarenessPayload = this.roomRegistry.clearAwareness(room);
    if (awarenessPayload !== null) {
      RealtimeGateway.broadcastInternalFrame(
        room,
        InternalFrameType.InternalFrameType_Awareness,
        blockPackId,
        awarenessPayload
      );
    }

    room.document?.destroy();
    room.document = null;
    room.validationDocument?.destroy();
    room.validationDocument = null;
    room.isLoading = false;
    room.dirtyUpdateCount = 0;
    room.lastUpdateSequence = 0;
    room.compactedUntilSequence = 0;
    room.projectedUntilSequence = -1;
    room.pendingYjsUpdates = [];
    room.pendingYjsPayloadBytes = 0;
    room.pendingPersistenceUpdates = [];
    room.pendingPersistencePayloadBytes = 0;
    room.pendingAwarenessUpdates.clear();
    room.persistenceDebounceTimer = null;
    room.persistenceMaximumWaitTimer = null;
    room.persistenceRetryTimer = null;
    room.inFlightPersistenceBatch = null;
    room.projectionTimer = null;
    room.inFlightProjection = null;

    RealtimeGateway.broadcastInternalFrame(
      room,
      InternalFrameType.InternalFrameType_ResyncRequired,
      blockPackId,
      Buffer.alloc(0)
    );

    this.roomRegistry.scheduleRoomEviction(blockPackId);
  }

  private logCoreCommandFailure(
    commandType: string,
    blockPackId: string,
    commandId: string | null,
    startedAt: number,
    error: unknown,
    persistenceBatchId?: string
  ): void {
    const commandError = error instanceof CoreCommandError ? error : null;
    this.logger.error("[YjsWorker] Core command failed", {
      blockPackId,
      commandId: commandError?.commandId ?? commandId,
      commandType,
      elapsedMilliseconds: performance.now() - startedAt,
      isTimeout: commandError?.isTimeout ?? false,
      errorCode: commandError?.code ?? null,
      retryable: commandError?.retryable ?? null,
      errorMessage: error instanceof Error ? error.message : String(error),
      errorStack: error instanceof Error ? error.stack : undefined,
      persistenceBatchId,
    });
  }

  private scheduleBlockProjection(
    room: Room,
    blockPackId: string,
    delayMilliseconds: number = 300
  ): void {
    if (
      room.document === null ||
      room.inFlightPersistenceBatch !== null ||
      room.pendingYjsUpdates.length > 0 ||
      room.pendingPersistenceUpdates.length > 0 ||
      room.inFlightProjection !== null ||
      room.projectionTimer !== null ||
      room.lastUpdateSequence <= room.projectedUntilSequence ||
      room.subscribers.size === 0
    ) {
      return;
    }

    room.projectionTimer = setTimeout(() => {
      room.projectionTimer = null;

      if (
        room.document === null ||
        room.inFlightPersistenceBatch !== null ||
        room.pendingYjsUpdates.length > 0 ||
        room.pendingPersistenceUpdates.length > 0 ||
        room.inFlightProjection !== null ||
        room.lastUpdateSequence <= room.projectedUntilSequence
      ) {
        return;
      }

      const subscriber = room.subscribers.values().next().value;
      if (subscriber === undefined) {
        return;
      }

      const projectedSequence = room.lastUpdateSequence;
      let payload: Buffer;
      try {
        payload = Buffer.from(
          JSON.stringify({
            schemaId: "notegic.blocknote",
            schemaVersion: 1,
            projectedSequence,
            blocks: this.blockPackProjector.projectYjsDocument(room.document),
          })
        );
      } catch (error) {
        this.logger.error("failed to project Yjs document", {
          blockPackId,
          error,
        });
        this.scheduleBlockProjection(room, blockPackId, 1_000);

        return;
      }

      room.inFlightProjection = {
        connectionId: subscriber.connectionId,
        connectorChannelId: subscriber.connectorChannelId,
        projectedSequence,
      };
      void this.coreCommandDispatcher
        .dispatch<
          { projection: string },
          { applied: boolean; projectedUntilSequence: number }
        >("ApplyBlockProjection", blockPackId, {
          projection: payload.toString("base64"),
        })
        .then(response => {
          if (
            room.inFlightProjection === null ||
            response.projectedUntilSequence < projectedSequence ||
            response.projectedUntilSequence > room.lastUpdateSequence
          ) {
            this.resyncRoom(room, blockPackId);

            return;
          }

          room.inFlightProjection = null;
          room.projectedUntilSequence = response.projectedUntilSequence;
          this.scheduleBlockProjection(room, blockPackId);
          this.roomRegistry.scheduleRoomEviction(blockPackId);
        })
        .catch(() => {
          room.inFlightProjection = null;
          this.scheduleBlockProjection(room, blockPackId, 1_000);
        });
    }, delayMilliseconds);
  }

  private requestYjsCompaction(room: Room, blockPackId: string): void {
    if (
      room.document === null ||
      room.isCompacting ||
      room.lastUpdateSequence - room.compactedUntilSequence <
        YjsCompactionUpdateThreshold ||
      room.isLoading ||
      room.pendingYjsUpdates.length > 0 ||
      room.pendingPersistenceUpdates.length > 0 ||
      room.inFlightPersistenceBatch !== null ||
      room.persistenceDebounceTimer !== null ||
      room.persistenceMaximumWaitTimer !== null ||
      room.persistenceRetryTimer !== null
    ) {
      return;
    }

    room.isCompacting = true;
    void this.coreCommandDispatcher
      .dispatch<Record<string, never>, { found: boolean; payload?: string }>(
        "LoadCompactableYjsDocument",
        blockPackId,
        {}
      )
      .then(async response => {
        if (!response.found || response.payload === undefined) {
          room.isCompacting = false;

          return;
        }
        const input = parseYjsCompactionInput(
          Buffer.from(response.payload, "base64")
        );
        if (input === null || !room.isCompacting) {
          room.isCompacting = false;

          return;
        }
        const compacted = this.yjsCompactionService.compact(input);
        const result = createYjsCompactionResult(
          input,
          compacted.snapshot,
          compacted.stateVector
        );
        const applied = await this.coreCommandDispatcher.dispatch<
          { payload: string },
          { applied: boolean }
        >("ApplyCompactedYjsDocument", blockPackId, {
          payload: result.toString("base64"),
        });
        room.isCompacting = false;
        if (applied.applied) room.compactedUntilSequence = input.cutoffSequence;
        this.roomRegistry.scheduleRoomEviction(blockPackId);
      })
      .catch(() => {
        room.isCompacting = false;
      });
  }

  private loadYjsDocument(room: Room, blockPackId: string): void {
    const startedAt = performance.now();
    void (async () => {
      let commandId: string | null = null;
      try {
        const dispatched = await this.coreCommandDispatcher.dispatchAsync<
          Record<string, never>,
          { found: boolean; payload?: string }
        >("LoadYjsDocument", blockPackId, {});
        commandId = dispatched.commandId;
        const response = await dispatched.reply;
        const state =
          response.found && response.payload !== undefined
            ? parseYjsDocumentState(Buffer.from(response.payload, "base64"))
            : null;
        if (state === null) {
          this.resyncRoom(room, blockPackId);

          return;
        }

        try {
          const document = new Y.Doc();
          if (state.snapshot.length > 0)
            Y.applyUpdate(document, state.snapshot);
          for (const update of state.updates)
            Y.applyUpdate(document, update.payload);

          this.roomRegistry.initializeAwareness(room, document);
          room.isLoading = false;
          room.lastUpdateSequence = state.lastUpdateSequence;
          room.compactedUntilSequence = state.compactedUntilSequence;
          room.projectedUntilSequence = state.projectedUntilSequence;
          const pendingYjsUpdates = room.pendingYjsUpdates;
          room.pendingYjsUpdates = [];
          room.pendingYjsPayloadBytes = 0;
          for (const pendingYjsUpdate of pendingYjsUpdates) {
            this.handleYjsDocumentUpdate(
              room,
              pendingYjsUpdate.webSocket,
              pendingYjsUpdate.frame
            );
          }
          if (
            room.inFlightPersistenceBatch === null &&
            room.pendingPersistenceUpdates.length === 0
          ) {
            this.sendRoomInitialState(room, blockPackId);
          }
          this.scheduleBlockProjection(room, blockPackId);
          this.roomRegistry.scheduleRoomEviction(blockPackId);
        } catch (error) {
          this.logCoreCommandFailure(
            "LoadYjsDocument",
            blockPackId,
            commandId,
            startedAt,
            error
          );
          this.resyncRoom(room, blockPackId);
        }
      } catch (error) {
        this.logCoreCommandFailure(
          "LoadYjsDocument",
          blockPackId,
          commandId,
          startedAt,
          error
        );
        this.resyncRoom(room, blockPackId);
      }
    })();
  }

  private handleYjsUpdatePersisted(
    room: Room,
    blockPackId: string,
    inFlightPersistenceBatch: {
      payload: Buffer;
      webSocket: WebSocket | null;
      connectionId: string;
      connectorChannelId: number;
    }
  ): void {
    RealtimeGateway.broadcastInternalFrame(
      room,
      InternalFrameType.InternalFrameType_YjsDocument,
      blockPackId,
      inFlightPersistenceBatch.payload
    );
    this.sendRoomInitialState(room, blockPackId);
    this.yjsDebouncer.flush(room, blockPackId);
    this.requestYjsCompaction(room, blockPackId);
    this.scheduleBlockProjection(room, blockPackId);
    this.roomRegistry.scheduleRoomEviction(blockPackId);
  }

  /* ============================== WebSocket connection ============================== */

  handleConnection(webSocket: WebSocket): void {
    this.webSockets.add(webSocket);
    this.telemetry.recordInternalSocket(1);

    webSocket.on("close", () => {
      this.webSockets.delete(webSocket);
      this.telemetry.recordInternalSocket(-1);
      for (const detachedRoom of this.roomRegistry.detachAll(webSocket)) {
        const { awarenessPayload, blockPackId, room } = detachedRoom;
        if (awarenessPayload !== null) {
          RealtimeGateway.broadcastInternalFrame(
            room,
            InternalFrameType.InternalFrameType_Awareness,
            blockPackId,
            awarenessPayload
          );
        }
        if (room.subscribers.size === 0) {
          this.yjsDebouncer.flush(room, blockPackId);
          this.roomRegistry.scheduleRoomEviction(blockPackId);
        }
      }
    });

    webSocket.on("message", (payload, isBinary) => {
      if (!isBinary) {
        webSocket.close(1003, "internal realtime frames must be binary");

        return;
      }

      let framePayload: Buffer;
      if (Buffer.isBuffer(payload)) {
        framePayload = payload;
      } else if (payload instanceof ArrayBuffer) {
        framePayload = Buffer.from(payload);
      } else if (Array.isArray(payload)) {
        framePayload = Buffer.concat(payload);
      } else {
        webSocket.close(1002, "invalid internal realtime frame");

        return;
      }

      const frame = parseInternalFrame(framePayload);
      if (frame === null || frame.version !== 1) {
        webSocket.close(1002, "invalid internal realtime frame");

        return;
      }

      switch (frame.type) {
        case InternalFrameType.InternalFrameType_Attach: {
          let quotaPolicy: Partial<BlockPackQuotaPolicy>;
          try {
            quotaPolicy = JSON.parse(
              frame.payload.toString("utf8")
            ) as Partial<BlockPackQuotaPolicy>;
          } catch {
            webSocket.send(
              createInternalFrame(
                InternalFrameType.InternalFrameType_ResyncRequired,
                frame.connectionId,
                frame.connectorChannelId,
                frame.blockPackId
              )
            );

            return;
          }
          if (
            quotaPolicy.version !== BlockPackDocumentQuotaPolicyVersion ||
            typeof quotaPolicy.maximumBlockCount !== "number" ||
            !Number.isSafeInteger(quotaPolicy.maximumBlockCount) ||
            quotaPolicy.maximumBlockCount <= 0
          ) {
            webSocket.send(
              createInternalFrame(
                InternalFrameType.InternalFrameType_ResyncRequired,
                frame.connectionId,
                frame.connectorChannelId,
                frame.blockPackId
              )
            );

            return;
          }

          const room = this.roomRegistry.attach(
            frame.blockPackId,
            webSocket,
            frame.connectionId,
            frame.connectorChannelId,
            quotaPolicy.version,
            quotaPolicy.maximumBlockCount
          );
          if (room.document !== null) {
            if (room.inFlightPersistenceBatch !== null) {
              this.yjsDebouncer.retryInFlight(
                room,
                frame.blockPackId,
                webSocket,
                frame.connectionId,
                frame.connectorChannelId
              );

              return;
            }

            if (room.pendingPersistenceUpdates.length > 0) {
              this.yjsDebouncer.flush(
                room,
                frame.blockPackId,
                webSocket,
                frame.connectionId,
                frame.connectorChannelId
              );

              return;
            }

            this.sendRoomInitialState(room, frame.blockPackId);
            this.scheduleBlockProjection(room, frame.blockPackId);

            return;
          }

          if (room.isLoading) {
            return;
          }

          room.isLoading = true;
          this.loadYjsDocument(room, frame.blockPackId);

          return;
        }
        case InternalFrameType.InternalFrameType_Detach: {
          const detachedRoom = this.roomRegistry.detach(
            frame.blockPackId,
            frame.connectionId,
            frame.connectorChannelId
          );
          if (detachedRoom === undefined) {
            return;
          }

          const { awarenessPayload, room } = detachedRoom;
          if (awarenessPayload !== null) {
            RealtimeGateway.broadcastInternalFrame(
              room,
              InternalFrameType.InternalFrameType_Awareness,
              frame.blockPackId,
              awarenessPayload
            );
          }
          if (room.subscribers.size === 0) {
            this.yjsDebouncer.flush(room, frame.blockPackId);
            this.requestYjsCompaction(room, frame.blockPackId);
            this.roomRegistry.scheduleRoomEviction(frame.blockPackId);
          }

          return;
        }
        case InternalFrameType.InternalFrameType_YjsDocument: {
          const room = this.roomRegistry.getSubscriber(
            frame.blockPackId,
            frame.connectionId,
            frame.connectorChannelId
          );
          if (room === undefined) {
            if (webSocket.readyState === WebSocket.OPEN) {
              webSocket.send(
                createInternalFrame(
                  InternalFrameType.InternalFrameType_ResyncRequired,
                  frame.connectionId,
                  frame.connectorChannelId,
                  frame.blockPackId
                )
              );
            }

            return;
          }

          this.handleYjsDocumentUpdate(room, webSocket, frame);

          return;
        }
        case InternalFrameType.InternalFrameType_Awareness: {
          const room = this.roomRegistry.getSubscriber(
            frame.blockPackId,
            frame.connectionId,
            frame.connectorChannelId
          );
          if (room === undefined) {
            if (webSocket.readyState === WebSocket.OPEN) {
              webSocket.send(
                createInternalFrame(
                  InternalFrameType.InternalFrameType_ResyncRequired,
                  frame.connectionId,
                  frame.connectorChannelId,
                  frame.blockPackId
                )
              );
            }

            return;
          }

          const awarenessPayload = this.roomRegistry.applyClientAwarenessUpdate(
            room,
            frame.connectionId,
            frame.connectorChannelId,
            frame.payload
          );
          if (awarenessPayload === null) {
            if (webSocket.readyState === WebSocket.OPEN) {
              webSocket.send(
                createInternalFrame(
                  InternalFrameType.InternalFrameType_ResyncRequired,
                  frame.connectionId,
                  frame.connectorChannelId,
                  frame.blockPackId
                )
              );
            }

            return;
          }

          if (awarenessPayload.length === 0) return;

          RealtimeGateway.broadcastInternalFrame(
            room,
            frame.type,
            frame.blockPackId,
            awarenessPayload
          );

          return;
        }
        case InternalFrameType.InternalFrameType_YjsDocumentLoaded: {
          const room = this.roomRegistry.get(frame.blockPackId);
          const state = parseYjsDocumentState(frame.payload);
          if (room === undefined || state === null) return;

          try {
            const document = new Y.Doc();
            if (state.snapshot.length > 0) {
              Y.applyUpdate(document, state.snapshot);
            }
            for (const update of state.updates) {
              Y.applyUpdate(document, update.payload);
            }

            this.roomRegistry.initializeAwareness(room, document);
            room.isLoading = false;
            room.lastUpdateSequence = state.lastUpdateSequence;
            room.compactedUntilSequence = state.compactedUntilSequence;
            room.projectedUntilSequence = state.projectedUntilSequence;
            const pendingYjsUpdates = room.pendingYjsUpdates;
            room.pendingYjsUpdates = [];
            room.pendingYjsPayloadBytes = 0;
            for (const pendingYjsUpdate of pendingYjsUpdates) {
              this.handleYjsDocumentUpdate(
                room,
                pendingYjsUpdate.webSocket,
                pendingYjsUpdate.frame
              );
            }
            if (
              room.inFlightPersistenceBatch === null &&
              room.pendingPersistenceUpdates.length === 0
            ) {
              this.sendRoomInitialState(room, frame.blockPackId);
            }
            this.scheduleBlockProjection(room, frame.blockPackId);
            this.roomRegistry.scheduleRoomEviction(frame.blockPackId);
          } catch {
            this.resyncRoom(room, frame.blockPackId);
          }

          return;
        }
        case InternalFrameType.InternalFrameType_YjsUpdatePersisted: {
          const room = this.roomRegistry.get(frame.blockPackId);
          if (room === undefined) {
            return;
          }

          const inFlightPersistenceBatch = this.yjsDebouncer.handlePersisted(
            room,
            frame.blockPackId,
            frame.connectionId,
            frame.connectorChannelId,
            parseYjsUpdateSequence(frame.payload)
          );
          if (inFlightPersistenceBatch === null) {
            return;
          }

          RealtimeGateway.broadcastInternalFrame(
            room,
            InternalFrameType.InternalFrameType_YjsDocument,
            frame.blockPackId,
            inFlightPersistenceBatch.payload
          );
          this.sendRoomInitialState(room, frame.blockPackId);
          this.yjsDebouncer.flush(room, frame.blockPackId);
          this.requestYjsCompaction(room, frame.blockPackId);
          this.scheduleBlockProjection(room, frame.blockPackId);
          this.roomRegistry.scheduleRoomEviction(frame.blockPackId);

          return;
        }
        case InternalFrameType.InternalFrameType_YjsPersistenceFailed: {
          const room = this.roomRegistry.get(frame.blockPackId);
          if (room !== undefined) {
            this.yjsDebouncer.handlePersistenceFailure(
              room,
              frame.blockPackId,
              frame.payload
            );
          }

          return;
        }
        case InternalFrameType.InternalFrameType_CompactableYjsDocumentLoaded: {
          const room = this.roomRegistry.get(frame.blockPackId);
          const input = parseYjsCompactionInput(frame.payload);
          if (input === null || (room !== undefined && !room.isCompacting)) {
            if (room !== undefined) room.isCompacting = false;

            return;
          }

          try {
            const compacted = this.yjsCompactionService.compact(input);

            if (webSocket.readyState === WebSocket.OPEN) {
              webSocket.send(
                createInternalFrame(
                  InternalFrameType.InternalFrameType_ApplyCompactedYjsDocument,
                  frame.connectionId,
                  frame.connectorChannelId,
                  frame.blockPackId,
                  createYjsCompactionResult(
                    input,
                    compacted.snapshot,
                    compacted.stateVector
                  )
                )
              );
            }
          } catch {
            if (room !== undefined) room.isCompacting = false;
          }

          return;
        }
        case InternalFrameType.InternalFrameType_YjsDocumentCompacted: {
          const room = this.roomRegistry.get(frame.blockPackId);
          const compactedUntilSequence = parseYjsUpdateSequence(frame.payload);
          if (room === undefined || compactedUntilSequence === null) return;

          room.isCompacting = false;
          room.compactedUntilSequence = compactedUntilSequence;
          this.roomRegistry.scheduleRoomEviction(frame.blockPackId);

          return;
        }
        case InternalFrameType.InternalFrameType_YjsDocumentCompactionFailed: {
          const room = this.roomRegistry.get(frame.blockPackId);
          if (room !== undefined) room.isCompacting = false;

          return;
        }
        case InternalFrameType.InternalFrameType_BlockProjectionApplied: {
          const room = this.roomRegistry.get(frame.blockPackId);
          let projectedUntilSequence: number | null = null;
          try {
            const value: unknown = JSON.parse(frame.payload.toString("utf8"));
            if (
              value !== null &&
              typeof value === "object" &&
              "projectedUntilSequence" in value &&
              typeof value.projectedUntilSequence === "number" &&
              Number.isSafeInteger(value.projectedUntilSequence) &&
              value.projectedUntilSequence >= -1
            ) {
              projectedUntilSequence = value.projectedUntilSequence;
            }
          } catch {}

          if (
            room === undefined ||
            projectedUntilSequence === null ||
            room.inFlightProjection === null ||
            room.inFlightProjection.connectionId !== frame.connectionId ||
            room.inFlightProjection.connectorChannelId !==
              frame.connectorChannelId ||
            projectedUntilSequence <
              room.inFlightProjection.projectedSequence ||
            projectedUntilSequence > room.lastUpdateSequence
          ) {
            if (room !== undefined) {
              this.resyncRoom(room, frame.blockPackId);
            }

            return;
          }

          room.inFlightProjection = null;
          room.projectedUntilSequence = projectedUntilSequence;
          this.scheduleBlockProjection(room, frame.blockPackId);
          this.roomRegistry.scheduleRoomEviction(frame.blockPackId);

          return;
        }
        case InternalFrameType.InternalFrameType_BlockProjectionFailed: {
          const room = this.roomRegistry.get(frame.blockPackId);
          if (
            room === undefined ||
            room.inFlightProjection === null ||
            room.inFlightProjection.connectionId !== frame.connectionId ||
            room.inFlightProjection.connectorChannelId !==
              frame.connectorChannelId
          ) {
            return;
          }

          room.inFlightProjection = null;
          this.scheduleBlockProjection(room, frame.blockPackId, 1_000);
          this.roomRegistry.scheduleRoomEviction(frame.blockPackId);

          return;
        }
        default:
          this.logger.warn(
            "received internal frame before its handler is enabled",
            {
              type: frame.type,
              blockPackId: frame.blockPackId,
            }
          );
      }
    });
  }

  getActiveRoomCount(): number {
    return this.roomRegistry.size;
  }

  async updateBlockPack(
    request: UpdateBlockPackYjsDocumentRequest
  ): Promise<UpdateBlockPackYjsDocumentResponse> {
    const room = this.roomRegistry.get(request.blockPackId);
    if (room === undefined || room.document === null) {
      if (this.yjsBlockPackUpdateService === null) {
        throw new Error("the Yjs block pack updater is not configured");
      }

      return this.yjsBlockPackUpdateService.update(request);
    }
    if (room.validationDocument === null) {
      throw new Error("the active Yjs room is not ready for updates");
    }

    const workingDocument = new Y.Doc();
    try {
      Y.applyUpdate(workingDocument, Y.encodeStateAsUpdate(room.document));
      if (this.yjsBlockPackUpdateService === null) {
        throw new Error("the Yjs block pack updater is not configured");
      }

      const result = this.yjsBlockPackUpdateService.apply(
        workingDocument,
        request.blocks
      );
      if (result.update.length === 0) {
        return result.response;
      }

      const validationDocument = new Y.Doc();
      try {
        Y.applyUpdate(
          validationDocument,
          Y.encodeStateAsUpdate(room.validationDocument)
        );
        Y.applyUpdate(validationDocument, result.update);
        if (
          this.blockPackProjector.countYjsDocumentBlocks(validationDocument) >
          room.maximumBlockCount
        ) {
          return {
            blocks: result.response.blocks.map(block => ({
              ...block,
              status: "failed",
              reason: "block_pack_quota_exceeded",
            })),
          };
        }
      } finally {
        validationDocument.destroy();
      }

      Y.applyUpdate(room.document, result.update);
      Y.applyUpdate(room.validationDocument, result.update);
      const subscriber = room.subscribers.values().next().value;
      this.yjsDebouncer.queueUpdate(room, {
        webSocket: subscriber?.webSocket ?? null,
        frame: {
          version: 1,
          type: InternalFrameType.InternalFrameType_YjsDocument,
          connectionId: randomUUID(),
          connectorChannelId: 1,
          blockPackId: request.blockPackId,
          payload: result.update,
        },
      });

      return result.response;
    } finally {
      workingDocument.destroy();
    }
  }

  async shutdown(): Promise<void> {
    for (const [blockPackId, room] of this.roomRegistry.entries()) {
      this.yjsDebouncer.flush(room, blockPackId);
    }

    const shutdownDeadline =
      Date.now() + YjsPersistenceBatchShutdownTimeoutMilliseconds;
    while (
      Date.now() < shutdownDeadline &&
      [...this.roomRegistry.entries()].some(
        ([, room]) =>
          room.inFlightPersistenceBatch !== null ||
          room.pendingPersistenceUpdates.length > 0
      )
    ) {
      await new Promise(resolve => setTimeout(resolve, 25));
    }

    this.webSockets.forEach(webSocket => {
      webSocket.close(1001, "server shutdown");
    });
  }
}
