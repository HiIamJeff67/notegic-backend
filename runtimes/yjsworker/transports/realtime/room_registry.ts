import * as decoding from "lib0/decoding";
import type WebSocket from "ws";
import {
  Awareness,
  applyAwarenessUpdate,
  encodeAwarenessUpdate,
  removeAwarenessStates,
} from "y-protocols/awareness";
import * as Y from "yjs";

import { YjsRoomIdleEvictionMilliseconds } from "../../configs/realtime_config.js";
import type { Telemetry } from "../../telemetry.js";
import type { AwarenessUpdateEntry } from "../../types/awareness.js";
import type { Room } from "../../types/room.js";

export type DetachedRoom = {
  blockPackId: string;
  room: Room;
  awarenessPayload: Buffer | null;
};

export class RoomRegistry {
  private readonly rooms = new Map<string, Room>();
  private readonly telemetry: Telemetry;

  constructor(telemetry: Telemetry) {
    this.telemetry = telemetry;
  }

  getOrCreate(
    blockPackId: string,
    documentQuotaPolicyVersion: number,
    maximumBlockCount: number
  ): Room {
    const existingRoom = this.rooms.get(blockPackId);
    if (existingRoom !== undefined) {
      existingRoom.lastActiveAt = new Date();
      existingRoom.maximumBlockCount = Math.min(
        existingRoom.maximumBlockCount,
        maximumBlockCount
      );

      return existingRoom;
    }

    const room: Room = {
      document: null,
      validationDocument: null,
      maximumBlockCount,
      documentQuotaPolicyVersion,
      awareness: null,
      awarenessClientOwners: new Map(),
      pendingAwarenessUpdates: new Map(),
      dirtyUpdateCount: 0,
      lastActiveAt: new Date(),
      subscribers: new Map(),
      isLoading: false,
      lastUpdateSequence: 0,
      compactedUntilSequence: 0,
      projectedUntilSequence: -1,
      pendingYjsUpdates: [],
      pendingYjsPayloadBytes: 0,
      pendingPersistenceUpdates: [],
      pendingPersistencePayloadBytes: 0,
      idleEvictionTimer: null,
      persistenceDebounceTimer: null,
      persistenceMaximumWaitTimer: null,
      persistenceRetryTimer: null,
      inFlightPersistenceBatch: null,
      isCompacting: false,
      projectionTimer: null,
      inFlightProjection: null,
    };
    this.rooms.set(blockPackId, room);
    this.telemetry.recordOperation({
      operation: "room.created",
      outcome: "success",
      durationMilliseconds: 0,
    });

    return room;
  }

  attach(
    blockPackId: string,
    webSocket: WebSocket,
    connectionId: string,
    connectorChannelId: number,
    documentQuotaPolicyVersion: number,
    maximumBlockCount: number
  ): Room {
    const room = this.getOrCreate(
      blockPackId,
      documentQuotaPolicyVersion,
      maximumBlockCount
    );
    this.cancelEviction(room);
    const subscriberKey = this.getSubscriberKey(
      connectionId,
      connectorChannelId
    );
    const existingSubscriber = room.subscribers.get(subscriberKey);
    if (existingSubscriber !== undefined) {
      existingSubscriber.webSocket = webSocket;

      return room;
    }

    room.subscribers.set(subscriberKey, {
      webSocket,
      connectionId,
      connectorChannelId,
      isReady: false,
      quotaRecoveryRequired: false,
      awarenessClientIds: new Set(),
    });

    return room;
  }

  detach(
    blockPackId: string,
    connectionId: string,
    connectorChannelId: number
  ): DetachedRoom | undefined {
    const room = this.rooms.get(blockPackId);
    if (room === undefined) {
      return undefined;
    }

    const awarenessClientIds = this.detachSubscriber(
      room,
      this.getSubscriberKey(connectionId, connectorChannelId)
    );
    room.lastActiveAt = new Date();

    return {
      blockPackId,
      room,
      awarenessPayload: this.removeAwarenessClientStates(
        room,
        awarenessClientIds
      ),
    };
  }

  detachAll(webSocket: WebSocket): DetachedRoom[] {
    const detachedRooms: DetachedRoom[] = [];
    for (const [blockPackId, room] of this.rooms) {
      let didDetach = false;
      const awarenessClientIds: number[] = [];
      for (const [subscriberKey, subscriber] of room.subscribers) {
        if (subscriber.webSocket === webSocket) {
          awarenessClientIds.push(
            ...this.detachSubscriber(room, subscriberKey)
          );
          didDetach = true;
        }
      }

      if (didDetach) {
        room.lastActiveAt = new Date();
        detachedRooms.push({
          blockPackId,
          room,
          awarenessPayload: this.removeAwarenessClientStates(
            room,
            awarenessClientIds
          ),
        });
      }
    }

    return detachedRooms;
  }

  getSubscriber(
    blockPackId: string,
    connectionId: string,
    connectorChannelId: number
  ): Room | undefined {
    const room = this.rooms.get(blockPackId);
    if (
      room === undefined ||
      !room.subscribers.has(
        this.getSubscriberKey(connectionId, connectorChannelId)
      )
    ) {
      return undefined;
    }

    room.lastActiveAt = new Date();

    return room;
  }

  initializeAwareness(room: Room, document: Y.Doc): void {
    room.awareness?.destroy();
    room.validationDocument?.destroy();
    room.document = document;
    room.validationDocument = new Y.Doc();
    Y.applyUpdate(room.validationDocument, Y.encodeStateAsUpdate(document));
    room.awareness = new Awareness(document);
    room.awareness.setLocalState(null);

    for (const [subscriberKey, updates] of room.pendingAwarenessUpdates) {
      const subscriber = room.subscribers.get(subscriberKey);
      if (subscriber === undefined) continue;

      for (const payload of updates) {
        const entries = this.parseAwarenessUpdateEntries(payload);
        if (
          entries === null ||
          entries.some(entry => entry.clientId === room.awareness?.clientID) ||
          !this.validateAwarenessUpdateEntries(
            room,
            subscriber.connectionId,
            subscriber.connectorChannelId,
            entries
          )
        ) {
          continue;
        }

        try {
          applyAwarenessUpdate(room.awareness, payload, this);
        } catch {
          continue;
        }

        this.registerAwarenessUpdateEntries(
          room,
          subscriber.connectionId,
          subscriber.connectorChannelId,
          entries
        );
      }
    }
    room.pendingAwarenessUpdates.clear();
  }

  getAwarenessSnapshot(room: Room): Buffer | null {
    if (room.awareness === null) {
      return null;
    }

    const clientIds = [...room.awareness.getStates().keys()];
    if (clientIds.length === 0) {
      return null;
    }

    return Buffer.from(encodeAwarenessUpdate(room.awareness, clientIds));
  }

  applyClientAwarenessUpdate(
    room: Room,
    connectionId: string,
    connectorChannelId: number,
    payload: Buffer
  ): Buffer | null {
    const awareness = room.awareness;
    const entries = this.parseAwarenessUpdateEntries(payload);
    if (entries === null) {
      return null;
    }

    const subscriberKey = this.getSubscriberKey(
      connectionId,
      connectorChannelId
    );
    if (!room.subscribers.has(subscriberKey)) return null;

    if (awareness === null) {
      const updates = room.pendingAwarenessUpdates.get(subscriberKey) ?? [];
      updates.push(payload);
      room.pendingAwarenessUpdates.set(subscriberKey, updates);

      return Buffer.alloc(0);
    }

    if (
      entries.some(entry => entry.clientId === awareness.clientID) ||
      !this.validateAwarenessUpdateEntries(
        room,
        connectionId,
        connectorChannelId,
        entries
      )
    ) {
      return null;
    }

    try {
      applyAwarenessUpdate(awareness, payload, this);
    } catch {
      return null;
    }

    this.registerAwarenessUpdateEntries(
      room,
      connectionId,
      connectorChannelId,
      entries
    );

    return Buffer.from(
      encodeAwarenessUpdate(
        awareness,
        entries.map(entry => entry.clientId)
      )
    );
  }

  clearAwareness(room: Room): Buffer | null {
    const awarenessPayload = this.removeAwarenessClientStates(
      room,
      room.awareness === null ? [] : [...room.awareness.getStates().keys()]
    );
    room.awareness?.destroy();
    room.awareness = null;
    this.clearAwarenessClientIds(room);

    return awarenessPayload;
  }

  private parseAwarenessUpdateEntries(
    payload: Buffer
  ): AwarenessUpdateEntry[] | null {
    try {
      const decoder = decoding.createDecoder(payload);
      const updateCount = decoding.readVarUint(decoder);
      if (!Number.isSafeInteger(updateCount) || updateCount <= 0) {
        return null;
      }

      const clientIds = new Set<number>();
      const entries: AwarenessUpdateEntry[] = [];
      for (let index = 0; index < updateCount; index += 1) {
        const clientId = decoding.readVarUint(decoder);
        const clock = decoding.readVarUint(decoder);
        const state: unknown = JSON.parse(decoding.readVarString(decoder));
        if (
          !Number.isSafeInteger(clientId) ||
          clientId < 0 ||
          !Number.isSafeInteger(clock) ||
          clock < 0 ||
          clientIds.has(clientId) ||
          (state !== null &&
            (typeof state !== "object" || Array.isArray(state)))
        ) {
          return null;
        }

        clientIds.add(clientId);
        entries.push({
          clientId,
          state: state as Record<string, unknown> | null,
        });
      }

      if (decoding.hasContent(decoder)) {
        return null;
      }

      return entries;
    } catch {
      return null;
    }
  }

  private validateAwarenessUpdateEntries(
    room: Room,
    connectionId: string,
    connectorChannelId: number,
    entries: AwarenessUpdateEntry[]
  ): boolean {
    const subscriberKey = this.getSubscriberKey(
      connectionId,
      connectorChannelId
    );
    const subscriber = room.subscribers.get(subscriberKey);
    if (subscriber === undefined) {
      return false;
    }

    for (const entry of entries) {
      const owner = room.awarenessClientOwners.get(entry.clientId);
      if (owner !== undefined && owner !== subscriberKey) {
        return false;
      }
      if (entry.state === null && owner !== subscriberKey) {
        return false;
      }
    }

    return true;
  }

  private registerAwarenessUpdateEntries(
    room: Room,
    connectionId: string,
    connectorChannelId: number,
    entries: AwarenessUpdateEntry[]
  ): void {
    const subscriberKey = this.getSubscriberKey(
      connectionId,
      connectorChannelId
    );
    const subscriber = room.subscribers.get(subscriberKey);
    if (subscriber === undefined) {
      return;
    }

    for (const entry of entries) {
      if (
        entry.state !== null &&
        !room.awarenessClientOwners.has(entry.clientId)
      ) {
        room.awarenessClientOwners.set(entry.clientId, subscriberKey);
        subscriber.awarenessClientIds.add(entry.clientId);
      }
    }
  }

  private clearAwarenessClientIds(room: Room): void {
    room.awarenessClientOwners.clear();
    for (const subscriber of room.subscribers.values()) {
      subscriber.awarenessClientIds.clear();
    }
  }

  private removeAwarenessClientStates(
    room: Room,
    clientIds: number[]
  ): Buffer | null {
    if (room.awareness === null || clientIds.length === 0) {
      return null;
    }

    removeAwarenessStates(room.awareness, clientIds, this);

    return Buffer.from(encodeAwarenessUpdate(room.awareness, clientIds));
  }

  get(blockPackId: string): Room | undefined {
    return this.rooms.get(blockPackId);
  }

  private isEvictable(blockPackId: string, room: Room): boolean {
    return (
      this.rooms.get(blockPackId) === room &&
      room.subscribers.size === 0 &&
      !room.isLoading &&
      room.pendingYjsUpdates.length === 0 &&
      room.pendingPersistenceUpdates.length === 0 &&
      room.inFlightPersistenceBatch === null &&
      !room.isCompacting &&
      room.persistenceDebounceTimer === null &&
      room.persistenceMaximumWaitTimer === null &&
      room.persistenceRetryTimer === null &&
      room.inFlightProjection === null
    );
  }

  scheduleRoomEviction(blockPackId: string): void {
    const room = this.rooms.get(blockPackId);
    if (room === undefined) {
      return;
    }

    if (room.subscribers.size > 0) {
      this.cancelEviction(room);

      return;
    }

    if (room.projectionTimer !== null) {
      clearTimeout(room.projectionTimer);
      room.projectionTimer = null;
    }

    if (
      !this.isEvictable(blockPackId, room) ||
      room.idleEvictionTimer !== null
    ) {
      return;
    }

    room.idleEvictionTimer = setTimeout(() => {
      room.idleEvictionTimer = null;
      if (this.isEvictable(blockPackId, room)) {
        this.evict(blockPackId, room);
      }
    }, YjsRoomIdleEvictionMilliseconds);
  }

  cancelEviction(room: Room): void {
    if (room.idleEvictionTimer === null) {
      return;
    }

    clearTimeout(room.idleEvictionTimer);
    room.idleEvictionTimer = null;
  }

  evict(blockPackId: string, room: Room): boolean {
    if (!this.isEvictable(blockPackId, room)) {
      return false;
    }

    this.cancelEviction(room);
    if (room.persistenceDebounceTimer !== null) {
      clearTimeout(room.persistenceDebounceTimer);
    }
    if (room.persistenceMaximumWaitTimer !== null) {
      clearTimeout(room.persistenceMaximumWaitTimer);
    }
    if (room.persistenceRetryTimer !== null) {
      clearTimeout(room.persistenceRetryTimer);
    }
    if (room.projectionTimer !== null) {
      clearTimeout(room.projectionTimer);
    }

    this.clearAwareness(room);
    room.document?.destroy();
    room.document = null;
    room.validationDocument?.destroy();
    room.validationDocument = null;
    room.pendingYjsUpdates = [];
    room.pendingYjsPayloadBytes = 0;
    room.pendingPersistenceUpdates = [];
    room.pendingPersistencePayloadBytes = 0;
    room.persistenceDebounceTimer = null;
    room.persistenceMaximumWaitTimer = null;
    room.persistenceRetryTimer = null;
    room.projectionTimer = null;
    this.rooms.delete(blockPackId);
    this.telemetry.recordOperation({
      operation: "room.evicted",
      outcome: "success",
      durationMilliseconds: 0,
    });

    return true;
  }

  private getSubscriberKey(
    connectionId: string,
    connectorChannelId: number
  ): string {
    return `${connectionId}:${connectorChannelId}`;
  }

  private detachSubscriber(room: Room, subscriberKey: string): number[] {
    const subscriber = room.subscribers.get(subscriberKey);
    if (subscriber === undefined) {
      return [];
    }

    const awarenessClientIds = [...subscriber.awarenessClientIds];
    for (const clientId of awarenessClientIds) {
      if (room.awarenessClientOwners.get(clientId) === subscriberKey) {
        room.awarenessClientOwners.delete(clientId);
      }
    }
    room.pendingAwarenessUpdates.delete(subscriberKey);
    room.subscribers.delete(subscriberKey);

    return awarenessClientIds;
  }

  get size(): number {
    return this.rooms.size;
  }

  get subscriberCount(): number {
    let count = 0;
    for (const room of this.rooms.values()) {
      count += room.subscribers.size;
    }

    return count;
  }

  entries(): IterableIterator<[string, Room]> {
    return this.rooms.entries();
  }
}
