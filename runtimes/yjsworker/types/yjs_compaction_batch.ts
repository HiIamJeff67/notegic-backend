import { convertBytesToUUIDString, convertUUIDStringToBytes } from "../util/uuid.js";
import {
  createYjsCompactionResult,
  parseYjsCompactionInput,
  parseYjsCompactionResult,
  type YjsCompactionInput,
  type YjsCompactionResult,
} from "./yjs_compaction.js";

export type YjsCompactionBatchInput = {
  blockPackId: string;
  input: YjsCompactionInput;
};

export function parseYjsCompactionBatchInput(
  payload: Buffer
): YjsCompactionBatchInput | null {
  if (payload.length < 20) return null;

  const blockPackId = convertBytesToUUIDString(payload.subarray(0, 16));
  if (blockPackId === null) return null;

  const inputLength = payload.readUInt32BE(16);
  if (inputLength !== payload.length - 20) return null;

  const input = parseYjsCompactionInput(payload.subarray(20));
  return input === null ? null : { blockPackId, input };
}

export type YjsCompactionBatchResult = {
  blockPackId: string;
  result: YjsCompactionResult;
};

export function createYjsCompactionBatchResult(
  blockPackId: string,
  input: YjsCompactionInput,
  snapshot: Uint8Array,
  stateVector: Uint8Array
): Buffer {
  const resultPayload = createYjsCompactionResult(input, snapshot, stateVector);
  const payload = Buffer.alloc(20 + resultPayload.length);
  convertUUIDStringToBytes(blockPackId).copy(payload, 0);
  payload.writeUInt32BE(resultPayload.length, 16);
  resultPayload.copy(payload, 20);

  return payload;
}

export function parseYjsCompactionBatchResult(
  payload: Buffer
): YjsCompactionBatchResult | null {
  if (payload.length < 20) return null;

  const blockPackId = convertBytesToUUIDString(payload.subarray(0, 16));
  if (blockPackId === null) return null;

  const resultLength = payload.readUInt32BE(16);
  if (resultLength !== payload.length - 20) return null;

  const result = parseYjsCompactionResult(payload.subarray(20));
  return result === null ? null : { blockPackId, result };
}
