import assert from "node:assert/strict";
import test from "node:test";

import {
  createYjsCompactionBatchResult,
  parseYjsCompactionBatchInput,
  parseYjsCompactionBatchResult,
} from "./yjs_compaction_batch.js";
import { convertUUIDStringToBytes } from "../util/uuid.js";

function createCompactionInputPayload(
  baseCompactedUntilSequence: number,
  cutoffSequence: number,
  updateSequence: number
): Buffer {
  const snapshot = Buffer.from([1]);
  const stateVector = Buffer.from([2]);
  const update = Buffer.from([3]);
  const payload = Buffer.alloc(
    28 + snapshot.length + stateVector.length + 12 + update.length
  );

  payload.writeBigInt64BE(BigInt(baseCompactedUntilSequence), 0);
  payload.writeBigInt64BE(BigInt(cutoffSequence), 8);
  payload.writeUInt32BE(snapshot.length, 16);
  payload.writeUInt32BE(stateVector.length, 20);
  payload.writeUInt32BE(1, 24);
  snapshot.copy(payload, 28);
  stateVector.copy(payload, 29);
  payload.writeBigInt64BE(BigInt(updateSequence), 30);
  payload.writeUInt32BE(update.length, 38);
  update.copy(payload, 42);

  return payload;
}

test("parses and returns one Yjs compaction batch item", () => {
  const blockPackId = "6c6a5f1f-5f9f-4b05-b3c0-3ab7a3d7a4e0";
  const inputPayload = createCompactionInputPayload(0, 1, 1);
  const payload = Buffer.alloc(20 + inputPayload.length);
  convertUUIDStringToBytes(blockPackId).copy(payload, 0);
  payload.writeUInt32BE(inputPayload.length, 16);
  inputPayload.copy(payload, 20);

  const input = parseYjsCompactionBatchInput(payload);
  assert.notEqual(input, null);
  assert.equal(input.blockPackId, blockPackId);

  const resultPayload = createYjsCompactionBatchResult(
    input.blockPackId,
    input.input,
    Buffer.from([4]),
    Buffer.from([5])
  );
  const result = parseYjsCompactionBatchResult(resultPayload);
  assert.notEqual(result, null);
  assert.equal(result.blockPackId, blockPackId);
  assert.equal(result.result.cutoffSequence, 1);
});
