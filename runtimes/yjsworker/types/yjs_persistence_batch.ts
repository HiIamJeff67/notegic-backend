import type WebSocket from "ws";

import { convertBytesToUUIDString, convertUUIDStringToBytes } from "../util/uuid.js";

// YjsPersistenceBatch is the internal Go/worker envelope for one merged raw Yjs update and its idempotency key
export type YjsPersistenceBatch = {
  persistenceBatchId: string;
  originConnectionId: string | null;
  payload: Buffer;
};

// InFlightYjsPersistenceBatch is one merged update awaiting its durable Go/PostgreSQL acknowledgement
export type InFlightYjsPersistenceBatch = {
  persistenceBatchId: string;
  originConnectionId: string | null;
  payload: Buffer;
  webSocket: WebSocket | null;
  connectionId: string;
  connectorChannelId: number;
  updateCount: number;
};

export function createYjsPersistenceBatch(
  persistenceBatchId: string,
  originConnectionId: string | null,
  payload: Buffer
): Buffer {
  // [persistenceBatchId:16][originConnectionId:16, zero UUID when mixed][raw Yjs update:n]
  const batchPayload = Buffer.alloc(32 + payload.length);

  convertUUIDStringToBytes(persistenceBatchId).copy(batchPayload, 0);
  if (originConnectionId !== null) {
    convertUUIDStringToBytes(originConnectionId).copy(batchPayload, 16);
  }
  payload.copy(batchPayload, 32);

  return batchPayload;
}

export function parseYjsPersistenceBatch(
  payload: Buffer
): YjsPersistenceBatch | null {
  if (payload.length <= 32) {
    return null;
  }

  const persistenceBatchId = convertBytesToUUIDString(payload.subarray(0, 16));
  const originConnectionId = convertBytesToUUIDString(payload.subarray(16, 32));
  if (persistenceBatchId === null) {
    return null;
  }

  return {
    persistenceBatchId,
    originConnectionId,
    payload: payload.subarray(32),
  };
}
