import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";

import { blocksToYXmlFragment } from "@blocknote/core/yjs";
import * as Y from "yjs";
import { Telemetry } from "../telemetry.js";
import {
  type NotegicBlock,
  notegicBlockNoteEditor,
} from "../types/blocknote_schema.js";
import { convertUUIDStringToBytes } from "../util/uuid.js";
import { BlockPackProjector } from "./block_pack_projector.js";
import { YjsProjectionService } from "./yjs_projection_service.js";

const telemetry = Telemetry.initialize();

test.after(async () => {
  await telemetry.shutdown();
});

test("YjsProjectionService projects a snapshot without a durable update tail", async () => {
  const sourceBlocks = JSON.parse(
    await readFile(
      new URL("../../../tmp/temp_wide_block_contents.json", import.meta.url),
      "utf8"
    )
  ) as NotegicBlock[];
  const sourceDocument = new Y.Doc();
  blocksToYXmlFragment(
    notegicBlockNoteEditor,
    sourceBlocks,
    sourceDocument.getXmlFragment("document-store")
  );

  const snapshot = Buffer.from(Y.encodeStateAsUpdate(sourceDocument));
  const stateVector = Buffer.from(Y.encodeStateVector(sourceDocument));
  const state = Buffer.alloc(36 + snapshot.length + stateVector.length);
  state.writeBigInt64BE(0n, 0);
  state.writeBigInt64BE(0n, 8);
  state.writeBigInt64BE(-1n, 16);
  state.writeUInt32BE(snapshot.length, 24);
  state.writeUInt32BE(stateVector.length, 28);
  state.writeUInt32BE(0, 32);
  snapshot.copy(state, 36);
  stateVector.copy(state, 36 + snapshot.length);

  const blockPackId = "2987fdbe-ca4b-4a70-81b7-058f0bb81c0e";
  const input = Buffer.alloc(20 + state.length);
  convertUUIDStringToBytes(blockPackId).copy(input, 0);
  input.writeUInt32BE(state.length, 16);
  state.copy(input, 20);

  const payload = Buffer.alloc(8 + input.length);
  payload.writeUInt32BE(1, 0);
  payload.writeUInt32BE(input.length, 4);
  input.copy(payload, 8);

  const result = new YjsProjectionService(
    new BlockPackProjector(),
    telemetry
  ).projectBatch(payload);
  assert.equal(result.readUInt32BE(0), 1);

  const resultLength = result.readUInt32BE(4);
  const resultPayload = result.subarray(8, 8 + resultLength);
  assert.equal(
    resultPayload.subarray(0, 16).equals(input.subarray(0, 16)),
    true
  );
  const projectionLength = resultPayload.readUInt32BE(16);
  const projection = JSON.parse(
    resultPayload.subarray(20, 20 + projectionLength).toString("utf8")
  ) as {
    schemaId: string;
    schemaVersion: number;
    projectedSequence: number;
    blocks: NotegicBlock[];
  };
  assert.equal(projection.schemaId, "notegic.blocknote");
  assert.equal(projection.schemaVersion, 1);
  assert.equal(projection.projectedSequence, 0);
  assert.deepEqual(projection.blocks, sourceBlocks);

  sourceDocument.destroy();
});
