import assert from "node:assert/strict";
import test from "node:test";

import * as Y from "yjs";

import {
  createYjsCompactionResult,
  parseYjsCompactionInput,
} from "./yjs_compaction.js";

test("materializes a compaction range into an equivalent Yjs snapshot", () => {
  const sourceDocument = new Y.Doc();
  const documentMap = sourceDocument.getMap<string>("document-store");
  documentMap.set("title", "Before");
  const snapshot = Buffer.from(Y.encodeStateAsUpdate(sourceDocument));
  const stateVector = Buffer.from(Y.encodeStateVector(sourceDocument));

  documentMap.set("title", "After");
  const update = Buffer.from(
    Y.encodeStateAsUpdate(sourceDocument, stateVector)
  );
  const payload = Buffer.alloc(
    28 + snapshot.length + stateVector.length + 12 + update.length
  );
  payload.writeBigInt64BE(0n, 0);
  payload.writeBigInt64BE(1n, 8);
  payload.writeUInt32BE(snapshot.length, 16);
  payload.writeUInt32BE(stateVector.length, 20);
  payload.writeUInt32BE(1, 24);
  snapshot.copy(payload, 28);
  stateVector.copy(payload, 28 + snapshot.length);
  payload.writeBigInt64BE(1n, 28 + snapshot.length + stateVector.length);
  payload.writeUInt32BE(
    update.length,
    36 + snapshot.length + stateVector.length
  );
  update.copy(payload, 40 + snapshot.length + stateVector.length);

  const input = parseYjsCompactionInput(payload);
  assert.notEqual(input, null);

  const compactedDocument = new Y.Doc();
  Y.applyUpdate(compactedDocument, input.snapshot);
  for (const tailUpdate of input.updates) {
    Y.applyUpdate(compactedDocument, tailUpdate.payload);
  }

  const result = createYjsCompactionResult(
    input,
    Y.encodeStateAsUpdate(compactedDocument),
    Y.encodeStateVector(compactedDocument)
  );
  const restoredDocument = new Y.Doc();
  Y.applyUpdate(
    restoredDocument,
    result.subarray(24, 24 + result.readUInt32BE(16))
  );

  assert.equal(restoredDocument.getMap("document-store").get("title"), "After");
});
