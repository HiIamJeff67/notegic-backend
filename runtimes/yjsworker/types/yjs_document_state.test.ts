import assert from "node:assert/strict";
import test from "node:test";

import * as Y from "yjs";

import {
  parseYjsDocumentState,
  parseYjsUpdateSequence,
} from "./yjs_document_state.js";

test("parses a snapshot and contiguous Yjs update tail", () => {
  const snapshot = Buffer.from([1, 2]);
  const stateVector = Buffer.from([3]);
  const update = Buffer.from([4, 5, 6]);
  const payload = Buffer.alloc(
    36 + snapshot.length + stateVector.length + 12 + update.length
  );

  payload.writeBigInt64BE(4n, 0);
  payload.writeBigInt64BE(3n, 8);
  payload.writeBigInt64BE(2n, 16);
  payload.writeUInt32BE(snapshot.length, 24);
  payload.writeUInt32BE(stateVector.length, 28);
  payload.writeUInt32BE(1, 32);
  snapshot.copy(payload, 36);
  stateVector.copy(payload, 38);
  payload.writeBigInt64BE(4n, 39);
  payload.writeUInt32BE(update.length, 47);
  update.copy(payload, 51);

  assert.deepEqual(parseYjsDocumentState(payload), {
    snapshot,
    stateVector,
    lastUpdateSequence: 4,
    compactedUntilSequence: 3,
    projectedUntilSequence: 2,
    updates: [{ updateSequence: 4, payload: update }],
  });
});

test("parses an acknowledged Yjs update sequence", () => {
  const payload = Buffer.alloc(8);
  payload.writeBigInt64BE(42n);

  assert.equal(parseYjsUpdateSequence(payload), 42);
});

test("materializes the same Yjs document from a snapshot and update tail", () => {
  const sourceDocument = new Y.Doc();
  sourceDocument.getMap("document-store").set("title", "Before");
  const snapshot = Buffer.from(Y.encodeStateAsUpdate(sourceDocument));
  const stateVector = Buffer.from(Y.encodeStateVector(sourceDocument));

  sourceDocument.getMap("document-store").set("title", "After");
  const update = Buffer.from(
    Y.encodeStateAsUpdate(sourceDocument, stateVector)
  );
  const payload = Buffer.alloc(
    36 + snapshot.length + stateVector.length + 12 + update.length
  );

  payload.writeBigInt64BE(1n, 0);
  payload.writeBigInt64BE(0n, 8);
  payload.writeBigInt64BE(-1n, 16);
  payload.writeUInt32BE(snapshot.length, 24);
  payload.writeUInt32BE(stateVector.length, 28);
  payload.writeUInt32BE(1, 32);
  snapshot.copy(payload, 36);
  stateVector.copy(payload, 36 + snapshot.length);
  payload.writeBigInt64BE(1n, 36 + snapshot.length + stateVector.length);
  payload.writeUInt32BE(
    update.length,
    44 + snapshot.length + stateVector.length
  );
  update.copy(payload, 48 + snapshot.length + stateVector.length);

  const state = parseYjsDocumentState(payload);
  assert.notEqual(state, null);

  const restoredDocument = new Y.Doc();
  Y.applyUpdate(restoredDocument, state.snapshot);
  for (const tailUpdate of state.updates) {
    Y.applyUpdate(restoredDocument, tailUpdate.payload);
  }

  assert.equal(restoredDocument.getMap("document-store").get("title"), "After");
});
