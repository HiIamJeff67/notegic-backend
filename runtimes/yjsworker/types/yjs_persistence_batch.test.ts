import assert from "node:assert/strict";
import test from "node:test";

import * as Y from "yjs";

import {
  createYjsPersistenceBatch,
  parseYjsPersistenceBatch,
} from "./yjs_persistence_batch.js";

test("encodes a mixed-origin Yjs persistence batch", () => {
  const persistenceBatchId = "719ea8f4-4fcb-4cee-b2f2-8652c52c343f";
  const payload = Buffer.from([1, 2, 3]);

  assert.deepEqual(
    parseYjsPersistenceBatch(
      createYjsPersistenceBatch(persistenceBatchId, null, payload)
    ),
    {
      persistenceBatchId,
      originConnectionId: null,
      payload,
    }
  );
});

test("merges a Yjs update burst into the same document state", () => {
  const sourceDocument = new Y.Doc();
  const updates: Uint8Array[] = [];
  sourceDocument.on("update", (update: Uint8Array) => updates.push(update));

  const map = sourceDocument.getMap<string>("document-store");
  map.set("title", "Notegic");
  map.set("title", "Notegic realtime");
  map.set("status", "draft");

  const sequentialDocument = new Y.Doc();
  for (const update of updates) {
    Y.applyUpdate(sequentialDocument, update);
  }

  const mergedDocument = new Y.Doc();
  Y.applyUpdate(mergedDocument, Y.mergeUpdates(updates));

  assert.deepEqual(
    Buffer.from(Y.encodeStateAsUpdate(mergedDocument)),
    Buffer.from(Y.encodeStateAsUpdate(sequentialDocument))
  );
});
