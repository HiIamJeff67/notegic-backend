import type WebSocket from "ws";
import type { Awareness } from "y-protocols/awareness";
import type { Doc } from "yjs";

import type { InFlightProjection } from "./projection.js";
import type { InFlightYjsPersistenceBatch } from "./yjs_persistence_batch.js";
import type { PendingYjsUpdate } from "./yjs_update.js";

// RoomSubscriber identifies one Go Gateway channel currently attached to a worker room
export type RoomSubscriber = {
  webSocket: WebSocket;
  connectionId: string;
  connectorChannelId: number;
  isReady: boolean;
  quotaRecoveryRequired: boolean;
  awarenessClientIds: Set<number>;
};

// Room owns the active in-memory Y.Doc and all transient state for one BlockPack collaboration room.
export type Room = {
  document: Doc | null;
  validationDocument: Doc | null;
  maximumBlockCount: number;
  documentQuotaPolicyVersion: number;
  awareness: Awareness | null;
  awarenessClientOwners: Map<number, string>;
  pendingAwarenessUpdates: Map<string, Buffer[]>;
  dirtyUpdateCount: number;
  lastActiveAt: Date;
  subscribers: Map<string, RoomSubscriber>;
  isLoading: boolean;
  lastUpdateSequence: number;
  compactedUntilSequence: number;
  projectedUntilSequence: number;
  pendingYjsUpdates: PendingYjsUpdate[];
  pendingYjsPayloadBytes: number;
  pendingPersistenceUpdates: PendingYjsUpdate[];
  pendingPersistencePayloadBytes: number;
  idleEvictionTimer: NodeJS.Timeout | null;
  persistenceDebounceTimer: NodeJS.Timeout | null;
  persistenceMaximumWaitTimer: NodeJS.Timeout | null;
  persistenceRetryTimer: NodeJS.Timeout | null;
  inFlightPersistenceBatch: InFlightYjsPersistenceBatch | null;
  isCompacting: boolean;
  projectionTimer: NodeJS.Timeout | null;
  inFlightProjection: InFlightProjection | null;
};
