import assert from "node:assert/strict";
import test from "node:test";

import * as Y from "yjs";

import { YjsCompactionService } from "./yjs_compaction_service.js";
import { Telemetry } from "../telemetry.js";
import { parseYjsCompactionBatchResult } from "../types/yjs_compaction_batch.js";
import { convertUUIDStringToBytes } from "../util/uuid.js";

const telemetry = Telemetry.initialize();

test.after(async () => {
  await telemetry.shutdown();
});

test("YjsCompactionService compacts one binary batch without changing its document", () => {
  const sourceDocument = new Y.Doc();
  const documentMap = sourceDocument.getMap<string>("document-store");
  documentMap.set("title", "Before");
  const snapshot = Buffer.from(Y.encodeStateAsUpdate(sourceDocument));
  const stateVector = Buffer.from(Y.encodeStateVector(sourceDocument));

  documentMap.set("title", "After");
  const update = Buffer.from(
    Y.encodeStateAsUpdate(sourceDocument, stateVector)
  );
  const compactionInput = Buffer.alloc(
    28 + snapshot.length + stateVector.length + 12 + update.length
  );
  compactionInput.writeBigInt64BE(0n, 0);
  compactionInput.writeBigInt64BE(1n, 8);
  compactionInput.writeUInt32BE(snapshot.length, 16);
  compactionInput.writeUInt32BE(stateVector.length, 20);
  compactionInput.writeUInt32BE(1, 24);
  snapshot.copy(compactionInput, 28);
  stateVector.copy(compactionInput, 28 + snapshot.length);
  compactionInput.writeBigInt64BE(
    1n,
    28 + snapshot.length + stateVector.length
  );
  compactionInput.writeUInt32BE(
    update.length,
    36 + snapshot.length + stateVector.length
  );
  update.copy(compactionInput, 40 + snapshot.length + stateVector.length);

  const blockPackId = "fc69a0e4-d022-488a-bd74-d610ccd2a84d";
  const batchItem = Buffer.alloc(20 + compactionInput.length);
  convertUUIDStringToBytes(blockPackId).copy(batchItem, 0);
  batchItem.writeUInt32BE(compactionInput.length, 16);
  compactionInput.copy(batchItem, 20);

  const batch = Buffer.alloc(8 + batchItem.length);
  batch.writeUInt32BE(1, 0);
  batch.writeUInt32BE(batchItem.length, 4);
  batchItem.copy(batch, 8);

  const response = new YjsCompactionService(telemetry).compactBatch(batch);
  assert.equal(response.readUInt32BE(0), 1);

  const resultLength = response.readUInt32BE(4);
  const result = parseYjsCompactionBatchResult(
    response.subarray(8, 8 + resultLength)
  );
  assert.notEqual(result, null);
  assert.equal(result.blockPackId, blockPackId);

  const restoredDocument = new Y.Doc();
  Y.applyUpdate(restoredDocument, result.result.snapshot);
  assert.equal(restoredDocument.getMap("document-store").get("title"), "After");
});
