import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";

import { blocksToYXmlFragment } from "@blocknote/core/yjs";
import * as Y from "yjs";
import {
  type NotegicBlock,
  notegicBlockNoteEditor,
} from "../types/blocknote_schema.js";
import { BlockPackProjector } from "./block_pack_projector.js";

const blockPackProjector = new BlockPackProjector();

async function readFixture(name: string): Promise<NotegicBlock[]> {
  const fixture = await readFile(
    new URL(`../../../tmp/${name}`, import.meta.url),
    "utf8"
  );

  return JSON.parse(fixture) as NotegicBlock[];
}

for (const fixtureName of [
  "temp_deep_block_contents.json",
  "temp_wide_block_contents.json",
]) {
  test(`projects ${fixtureName} through the canonical BlockNote Y.XmlFragment`, async () => {
    const sourceBlocks = await readFixture(fixtureName);
    const document = new Y.Doc();
    blocksToYXmlFragment(
      notegicBlockNoteEditor,
      sourceBlocks,
      document.getXmlFragment("document-store")
    );

    const projectedBlocks = blockPackProjector.projectYjsDocument(document);
    const rematerializedDocument = new Y.Doc();
    blocksToYXmlFragment(
      notegicBlockNoteEditor,
      projectedBlocks,
      rematerializedDocument.getXmlFragment("document-store")
    );

    assert.deepEqual(
      blockPackProjector.projectYjsDocument(rematerializedDocument),
      projectedBlocks
    );
    assert.equal(projectedBlocks.length, sourceBlocks.length);
    if (fixtureName === "temp_wide_block_contents.json") {
      assert.equal(
        blockPackProjector.countYjsDocumentBlocks(document),
        sourceBlocks.length
      );
    }
  });
}

test("projects Math, Diagram, Calendar, and inline Math blocks", async () => {
  const sourceBlocks = [
    {
      id: "11111111-1111-4111-8111-111111111111",
      type: "mathBlock",
      props: {},
      content: "x^2 + y^2",
      children: [],
    },
    {
      id: "22222222-2222-4222-8222-222222222222",
      type: "diagram",
      props: {},
      content: "graph TD\n    A[Start] --> B[Stop]",
      children: [],
    },
    {
      id: "33333333-3333-4333-8333-333333333333",
      type: "calendar",
      props: {
        calendarId: "work",
        anchorDate: "2026-08-01",
        timezone: "Asia/Taipei",
        view: "month",
      },
      children: [],
    },
    {
      id: "44444444-4444-4444-8444-444444444444",
      type: "paragraph",
      props: {},
      content: [{ type: "math", props: {}, content: "a+b" }],
      children: [],
    },
  ] as unknown as NotegicBlock[];

  const document = new Y.Doc();
  blocksToYXmlFragment(
    notegicBlockNoteEditor,
    sourceBlocks,
    document.getXmlFragment("document-store")
  );

  const projectedBlocks = new BlockPackProjector().projectYjsDocument(document);
  assert.equal(projectedBlocks[0]?.type, "mathBlock");
  assert.deepEqual(projectedBlocks[0]?.content, [
    { type: "text", text: "x^2 + y^2", styles: {} },
  ]);
  assert.equal(projectedBlocks[1]?.type, "diagram");
  assert.deepEqual(projectedBlocks[1]?.content, [
    { type: "text", text: "graph TD\n    A[Start] --> B[Stop]", styles: {} },
  ]);
  assert.equal(projectedBlocks[2]?.type, "calendar");
  assert.deepEqual(projectedBlocks[2]?.props, sourceBlocks[2]?.props);
  assert.equal(projectedBlocks[2]?.content, undefined);
  assert.deepEqual(projectedBlocks[3]?.content, sourceBlocks[3]?.content);
  document.destroy();
});
