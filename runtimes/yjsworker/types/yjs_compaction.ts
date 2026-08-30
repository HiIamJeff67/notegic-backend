import type { YjsDocumentUpdate } from "./yjs_document_state.js";

export type YjsCompactionInput = {
  snapshot: Buffer;
  stateVector: Buffer;
  baseCompactedUntilSequence: number;
  cutoffSequence: number;
  updates: YjsDocumentUpdate[];
};

export function parseYjsCompactionInput(
  payload: Buffer
): YjsCompactionInput | null {
  if (payload.length < 28) {
    return null;
  }

  const baseCompactedUntilSequence = Number(payload.readBigInt64BE(0));
  const cutoffSequence = Number(payload.readBigInt64BE(8));
  const snapshotLength = payload.readUInt32BE(16);
  const stateVectorLength = payload.readUInt32BE(20);
  const updateCount = payload.readUInt32BE(24);
  if (
    !Number.isSafeInteger(baseCompactedUntilSequence) ||
    !Number.isSafeInteger(cutoffSequence) ||
    baseCompactedUntilSequence < 0 ||
    cutoffSequence <= baseCompactedUntilSequence
  ) {
    return null;
  }

  let offset = 28;
  if (snapshotLength > payload.length - offset) return null;
  const snapshot = payload.subarray(offset, offset + snapshotLength);
  offset += snapshotLength;
  if (stateVectorLength > payload.length - offset) return null;
  const stateVector = payload.subarray(offset, offset + stateVectorLength);
  offset += stateVectorLength;

  const updates: YjsDocumentUpdate[] = [];
  let expectedSequence = baseCompactedUntilSequence + 1;
  for (let index = 0; index < updateCount; index += 1) {
    if (payload.length - offset < 12) return null;
    const updateSequence = Number(payload.readBigInt64BE(offset));
    const updateLength = payload.readUInt32BE(offset + 8);
    offset += 12;
    if (
      !Number.isSafeInteger(updateSequence) ||
      updateSequence !== expectedSequence ||
      updateLength > payload.length - offset
    )
      return null;
    updates.push({
      updateSequence,
      payload: payload.subarray(offset, offset + updateLength),
    });
    offset += updateLength;
    expectedSequence += 1;
  }

  if (offset !== payload.length || expectedSequence - 1 !== cutoffSequence)
    return null;

  return {
    snapshot,
    stateVector,
    baseCompactedUntilSequence,
    cutoffSequence,
    updates,
  };
}

export function createYjsCompactionResult(
  input: YjsCompactionInput,
  snapshot: Uint8Array,
  stateVector: Uint8Array
): Buffer {
  const payload = Buffer.alloc(24 + snapshot.length + stateVector.length);
  payload.writeBigInt64BE(BigInt(input.baseCompactedUntilSequence), 0);
  payload.writeBigInt64BE(BigInt(input.cutoffSequence), 8);
  payload.writeUInt32BE(snapshot.length, 16);
  payload.writeUInt32BE(stateVector.length, 20);
  Buffer.from(snapshot).copy(payload, 24);
  Buffer.from(stateVector).copy(payload, 24 + snapshot.length);

  return payload;
}

export type YjsCompactionResult = {
  baseCompactedUntilSequence: number;
  cutoffSequence: number;
  snapshot: Buffer;
  stateVector: Buffer;
};

export function parseYjsCompactionResult(
  payload: Buffer
): YjsCompactionResult | null {
  if (payload.length < 24) return null;

  const baseCompactedUntilSequence = Number(payload.readBigInt64BE(0));
  const cutoffSequence = Number(payload.readBigInt64BE(8));
  const snapshotLength = payload.readUInt32BE(16);
  const stateVectorLength = payload.readUInt32BE(20);
  if (
    !Number.isSafeInteger(baseCompactedUntilSequence) ||
    !Number.isSafeInteger(cutoffSequence) ||
    baseCompactedUntilSequence < 0 ||
    cutoffSequence <= baseCompactedUntilSequence ||
    snapshotLength + stateVectorLength !== payload.length - 24
  )
    return null;

  return {
    baseCompactedUntilSequence,
    cutoffSequence,
    snapshot: payload.subarray(24, 24 + snapshotLength),
    stateVector: payload.subarray(24 + snapshotLength),
  };
}
