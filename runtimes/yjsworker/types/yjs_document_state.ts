import { YjsDocumentMaximumLoadPayloadBytes } from "../../../contracts/yjs-worker/v1/yjsworker_contract.js";

export type YjsDocumentUpdate = {
  updateSequence: number;
  payload: Buffer;
};

export type YjsDocumentState = {
  snapshot: Buffer;
  stateVector: Buffer;
  lastUpdateSequence: number;
  compactedUntilSequence: number;
  projectedUntilSequence: number;
  updates: YjsDocumentUpdate[];
};

export function parseYjsDocumentState(
  payload: Buffer
): YjsDocumentState | null {
  if (payload.length > YjsDocumentMaximumLoadPayloadBytes) {
    return null;
  }
  if (payload.length < 36) {
    return null;
  }

  const lastUpdateSequence = Number(payload.readBigInt64BE(0));
  const compactedUntilSequence = Number(payload.readBigInt64BE(8));
  const projectedUntilSequence = Number(payload.readBigInt64BE(16));
  const snapshotLength = payload.readUInt32BE(24);
  const stateVectorLength = payload.readUInt32BE(28);
  const updateCount = payload.readUInt32BE(32);
  if (
    !Number.isSafeInteger(lastUpdateSequence) ||
    !Number.isSafeInteger(compactedUntilSequence) ||
    !Number.isSafeInteger(projectedUntilSequence) ||
    lastUpdateSequence < 0 ||
    compactedUntilSequence < 0 ||
    compactedUntilSequence > lastUpdateSequence ||
    projectedUntilSequence < -1 ||
    projectedUntilSequence > lastUpdateSequence
  ) {
    return null;
  }

  let offset = 36;
  if (snapshotLength > payload.length - offset) {
    return null;
  }
  const snapshot = payload.subarray(offset, offset + snapshotLength);
  offset += snapshotLength;

  if (stateVectorLength > payload.length - offset) {
    return null;
  }
  const stateVector = payload.subarray(offset, offset + stateVectorLength);
  offset += stateVectorLength;

  const updates: YjsDocumentState["updates"] = [];
  let expectedUpdateSequence = compactedUntilSequence + 1;
  for (let index = 0; index < updateCount; index += 1) {
    if (payload.length - offset < 12) {
      return null;
    }

    const updateSequence = Number(payload.readBigInt64BE(offset));
    const updateLength = payload.readUInt32BE(offset + 8);
    offset += 12;
    if (
      !Number.isSafeInteger(updateSequence) ||
      updateSequence !== expectedUpdateSequence ||
      updateLength > payload.length - offset
    ) {
      return null;
    }

    updates.push({
      updateSequence,
      payload: payload.subarray(offset, offset + updateLength),
    });
    offset += updateLength;
    expectedUpdateSequence += 1;
  }

  if (
    offset !== payload.length ||
    expectedUpdateSequence - 1 !== lastUpdateSequence
  ) {
    return null;
  }

  return {
    snapshot,
    stateVector,
    lastUpdateSequence,
    compactedUntilSequence,
    projectedUntilSequence,
    updates,
  };
}

export function parseYjsUpdateSequence(payload: Buffer): number | null {
  if (payload.length !== 8) {
    return null;
  }

  const updateSequence = Number(payload.readBigInt64BE());
  if (!Number.isSafeInteger(updateSequence) || updateSequence < 1) {
    return null;
  }

  return updateSequence;
}
