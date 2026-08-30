import {
  convertBytesToUUIDString,
  convertUUIDStringToBytes,
} from "../util/uuid.js";
import type { NotegicBlock } from "./blocknote_schema.js";
import {
  parseYjsDocumentState,
  type YjsDocumentState,
} from "./yjs_document_state.js";

export type YjsProjectionBatchInput = {
  blockPackId: string;
  state: YjsDocumentState;
};

export function parseYjsProjectionBatchInput(
  payload: Buffer
): YjsProjectionBatchInput | null {
  if (payload.length < 20) {
    return null;
  }

  const blockPackId = convertBytesToUUIDString(payload.subarray(0, 16));
  if (blockPackId === null) {
    return null;
  }

  const stateLength = payload.readUInt32BE(16);
  if (stateLength !== payload.length - 20) {
    return null;
  }

  const state = parseYjsDocumentState(payload.subarray(20));

  return state === null ? null : { blockPackId, state };
}

export type YjsProjectionBatchResult = {
  blockPackId: string;
  schemaId: "notegic.blocknote";
  schemaVersion: 1;
  projectedSequence: number;
  blocks: NotegicBlock[];
};

export function createYjsProjectionBatchResult(
  result: YjsProjectionBatchResult
): Buffer {
  const resultPayload = Buffer.from(JSON.stringify(result));
  const payload = Buffer.alloc(20 + resultPayload.length);
  convertUUIDStringToBytes(result.blockPackId).copy(payload, 0);
  payload.writeUInt32BE(resultPayload.length, 16);
  resultPayload.copy(payload, 20);

  return payload;
}
