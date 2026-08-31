import { randomUUID } from "node:crypto";
import type WebSocket from "ws";
import { WebSocket as WebSocketState } from "ws";
import * as Y from "yjs";

import {
  YjsPersistenceBatchDebounceMilliseconds,
  YjsPersistenceBatchMaximumPayloadBytes,
  YjsPersistenceBatchMaximumUpdateCount,
  YjsPersistenceBatchMaximumWaitMilliseconds,
  YjsPersistenceBatchRetryMilliseconds,
} from "../../configs/persistence_config.js";
import type { Telemetry } from "../../telemetry.js";
import { createInternalFrame } from "../../types/internal_frame.js";
import { InternalFrameType } from "../../types/internal_frame_type.js";
import type { Room } from "../../types/room.js";
import { type InFlightYjsPersistenceBatch } from "../../types/yjs_persistence_batch.js";
import { YjsPersistenceFailureType } from "../../types/yjs_persistence_failure_type.js";
import type { PendingYjsUpdate } from "../../types/yjs_update.js";

export class YjsDebouncer {
  private resyncRoom!: (room: Room, blockPackId: string) => void;
  private readonly appendYjsUpdate: (
    blockPackId: string,
    persistenceBatchId: string,
    originConnectionId: string | null,
    payload: Buffer
  ) => Promise<number>;
  private readonly persisted: (
    room: Room,
    blockPackId: string,
    batch: InFlightYjsPersistenceBatch
  ) => void;
  private readonly telemetry: Telemetry;

  constructor(
    telemetry: Telemetry,
    resyncRoom: (room: Room, blockPackId: string) => void,
    appendYjsUpdate: (
      blockPackId: string,
      persistenceBatchId: string,
      originConnectionId: string | null,
      payload: Buffer
    ) => Promise<number>,
    persisted: (
      room: Room,
      blockPackId: string,
      batch: InFlightYjsPersistenceBatch
    ) => void
  ) {
    this.telemetry = telemetry;
    this.resyncRoom = resyncRoom;
    this.appendYjsUpdate = appendYjsUpdate;
    this.persisted = persisted;
  }

  scheduleFlush(room: Room, blockPackId: string): void {
    if (room.document === null || room.pendingPersistenceUpdates.length === 0) {
      return;
    }

    if (room.persistenceDebounceTimer !== null) {
      clearTimeout(room.persistenceDebounceTimer);
    }
    room.persistenceDebounceTimer = setTimeout(() => {
      room.persistenceDebounceTimer = null;
      this.flush(room, blockPackId);
    }, YjsPersistenceBatchDebounceMilliseconds);

    if (room.persistenceMaximumWaitTimer !== null) {
      return;
    }

    room.persistenceMaximumWaitTimer = setTimeout(() => {
      room.persistenceMaximumWaitTimer = null;
      this.flush(room, blockPackId);
    }, YjsPersistenceBatchMaximumWaitMilliseconds);
  }

  flush(
    room: Room,
    blockPackId: string,
    webSocket: WebSocket | null = null,
    connectionId: string | null = null,
    connectorChannelId: number | null = null
  ): void {
    const startedAt = performance.now();
    if (
      room.document === null ||
      room.inFlightPersistenceBatch !== null ||
      room.pendingPersistenceUpdates.length === 0
    ) {
      return;
    }

    if (room.persistenceDebounceTimer !== null) {
      clearTimeout(room.persistenceDebounceTimer);
      room.persistenceDebounceTimer = null;
    }
    if (room.persistenceMaximumWaitTimer !== null) {
      clearTimeout(room.persistenceMaximumWaitTimer);
      room.persistenceMaximumWaitTimer = null;
    }

    const pendingPersistenceUpdates = room.pendingPersistenceUpdates;
    room.pendingPersistenceUpdates = [];
    room.pendingPersistencePayloadBytes = 0;

    let payload: Buffer;
    try {
      payload = Buffer.from(
        Y.mergeUpdates(
          pendingPersistenceUpdates.map(update => update.frame.payload)
        )
      );
    } catch {
      this.telemetry.recordOperation({
        operation: "persistence.merge",
        outcome: "error",
        durationMilliseconds: performance.now() - startedAt,
      });
      this.resyncRoom(room, blockPackId);

      return;
    }

    const firstUpdate = pendingPersistenceUpdates[0];
    const originConnectionId = pendingPersistenceUpdates.every(
      update => update.frame.connectionId === firstUpdate.frame.connectionId
    )
      ? firstUpdate.frame.connectionId
      : null;
    room.inFlightPersistenceBatch = {
      persistenceBatchId: randomUUID(),
      originConnectionId,
      payload,
      webSocket: webSocket ?? firstUpdate.webSocket,
      connectionId: connectionId ?? firstUpdate.frame.connectionId,
      connectorChannelId:
        connectorChannelId ?? firstUpdate.frame.connectorChannelId,
      updateCount: pendingPersistenceUpdates.length,
    };
    this.telemetry.recordOperation({
      operation: "persistence.batch_created",
      outcome: "success",
      durationMilliseconds: performance.now() - startedAt,
      payloadBytes: payload.length,
    });

    if (!this.sendInFlight(room, blockPackId)) {
      this.scheduleRetry(room, blockPackId);
    }
  }

  queueUpdate(room: Room, pendingYjsUpdate: PendingYjsUpdate): void {
    if (room.document === null) {
      room.pendingYjsUpdates.push(pendingYjsUpdate);

      return;
    }

    try {
      Y.applyUpdate(room.document, pendingYjsUpdate.frame.payload);
    } catch {
      if (
        pendingYjsUpdate.webSocket !== null &&
        pendingYjsUpdate.webSocket.readyState === WebSocketState.OPEN
      ) {
        pendingYjsUpdate.webSocket.send(
          createInternalFrame(
            InternalFrameType.InternalFrameType_ResyncRequired,
            pendingYjsUpdate.frame.connectionId,
            pendingYjsUpdate.frame.connectorChannelId,
            pendingYjsUpdate.frame.blockPackId
          )
        );
      }

      return;
    }

    room.pendingPersistenceUpdates.push(pendingYjsUpdate);
    room.pendingPersistencePayloadBytes +=
      pendingYjsUpdate.frame.payload.length;
    if (
      room.pendingPersistenceUpdates.length >=
        YjsPersistenceBatchMaximumUpdateCount ||
      room.pendingPersistencePayloadBytes >=
        YjsPersistenceBatchMaximumPayloadBytes
    ) {
      this.flush(room, pendingYjsUpdate.frame.blockPackId);

      return;
    }

    this.scheduleFlush(room, pendingYjsUpdate.frame.blockPackId);
  }

  retryInFlight(
    room: Room,
    blockPackId: string,
    webSocket: WebSocket,
    connectionId: string,
    connectorChannelId: number
  ): void {
    if (room.inFlightPersistenceBatch === null) {
      return;
    }

    if (room.persistenceRetryTimer !== null) {
      clearTimeout(room.persistenceRetryTimer);
      room.persistenceRetryTimer = null;
    }

    room.inFlightPersistenceBatch.webSocket = webSocket;
    room.inFlightPersistenceBatch.connectionId = connectionId;
    room.inFlightPersistenceBatch.connectorChannelId = connectorChannelId;
    if (!this.sendInFlight(room, blockPackId)) {
      this.scheduleRetry(room, blockPackId);
    }
  }

  handlePersisted(
    room: Room,
    blockPackId: string,
    connectionId: string,
    connectorChannelId: number,
    updateSequence: number | null
  ): InFlightYjsPersistenceBatch | null {
    if (
      updateSequence === null ||
      room.inFlightPersistenceBatch === null ||
      room.inFlightPersistenceBatch.connectionId !== connectionId ||
      room.inFlightPersistenceBatch.connectorChannelId !== connectorChannelId ||
      updateSequence !== room.lastUpdateSequence + 1
    ) {
      this.resyncRoom(room, blockPackId);

      return null;
    }

    return this.completePersistence(room, updateSequence);
  }

  private completePersistence(
    room: Room,
    updateSequence: number | null
  ): InFlightYjsPersistenceBatch | null {
    const inFlightPersistenceBatch = room.inFlightPersistenceBatch;
    if (inFlightPersistenceBatch === null) {
      return null;
    }

    room.inFlightPersistenceBatch = null;
    if (room.persistenceRetryTimer !== null) {
      clearTimeout(room.persistenceRetryTimer);
      room.persistenceRetryTimer = null;
    }
    room.lastUpdateSequence = updateSequence ?? room.lastUpdateSequence + 1;
    room.dirtyUpdateCount += inFlightPersistenceBatch.updateCount;
    this.telemetry.recordOperation({
      operation: "persistence.batch_persisted",
      outcome: "success",
      durationMilliseconds: 0,
      payloadBytes: inFlightPersistenceBatch.payload.length,
    });

    return inFlightPersistenceBatch;
  }

  handlePersistenceFailure(
    room: Room,
    blockPackId: string,
    failure: unknown
  ): void {
    if (room.inFlightPersistenceBatch === null) {
      this.resyncRoom(room, blockPackId);

      return;
    }

    const isRetryable = Buffer.isBuffer(failure)
      ? failure[0] ===
        YjsPersistenceFailureType.YjsPersistenceFailureType_Retryable
      : !(
          failure !== null &&
          typeof failure === "object" &&
          "retryable" in failure &&
          (failure as { retryable?: unknown }).retryable === false
        );

    if (isRetryable) {
      this.telemetry.recordOperation({
        operation: "persistence.batch_persisted",
        outcome: "error",
        durationMilliseconds: 0,
      });
      this.scheduleRetry(room, blockPackId);

      return;
    }

    this.telemetry.recordOperation({
      operation: "persistence.batch_persisted",
      outcome: "error",
      durationMilliseconds: 0,
    });
    this.resyncRoom(room, blockPackId);
  }

  private sendInFlight(room: Room, blockPackId: string): boolean {
    const batch = room.inFlightPersistenceBatch;
    if (batch === null) {
      return false;
    }

    void this.appendYjsUpdate(
      blockPackId,
      batch.persistenceBatchId,
      batch.originConnectionId,
      batch.payload
    )
      .then(updateSequence => {
        if (room.inFlightPersistenceBatch !== batch) return;

        const persisted = this.handlePersisted(
          room,
          blockPackId,
          batch.connectionId,
          batch.connectorChannelId,
          updateSequence
        );
        if (persisted !== null) this.persisted(room, blockPackId, persisted);
      })
      .catch(error => {
        if (room.inFlightPersistenceBatch !== batch) return;

        this.handlePersistenceFailure(room, blockPackId, error);
      });

    return true;
  }

  private scheduleRetry(room: Room, blockPackId: string): void {
    if (
      room.inFlightPersistenceBatch === null ||
      room.persistenceRetryTimer !== null ||
      (room.inFlightPersistenceBatch.webSocket !== null &&
        room.inFlightPersistenceBatch.webSocket.readyState !==
          WebSocketState.OPEN)
    ) {
      return;
    }

    room.persistenceRetryTimer = setTimeout(() => {
      room.persistenceRetryTimer = null;

      if (!this.sendInFlight(room, blockPackId)) {
        this.scheduleRetry(room, blockPackId);
      }
    }, YjsPersistenceBatchRetryMilliseconds);
  }
}
