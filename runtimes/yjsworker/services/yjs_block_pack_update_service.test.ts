import assert from "node:assert/strict";
import test from "node:test";

import { blocksToYXmlFragment } from "@blocknote/core/yjs";
import * as Y from "yjs";

import type { UpdateBlockPackYjsDocumentBlockRequest } from "../../../contracts/yjs-worker/v1/update_block_pack.js";
import {
  type NotegicBlock,
  notegicBlockNoteEditor,
} from "../types/blocknote_schema.js";
import { BlockPackProjector } from "./block_pack_projector.js";
import { YjsBlockPackUpdateService } from "./yjs_block_pack_update_service.js";

test("YjsBlockPackUpdateService updates existing blocks and skips missing blocks", () => {
  const block: NotegicBlock = {
    id: "c58c8cba-74b3-46e6-a758-16530edc9a01",
    type: "paragraph",
    props: {
      backgroundColor: "default",
      textAlignment: "left",
      textColor: "default",
    },
    content: [{ styles: {}, text: "before", type: "text" }],
    children: [],
  };
  const sourceDocument = new Y.Doc();
  blocksToYXmlFragment(
    notegicBlockNoteEditor,
    [block],
    sourceDocument.getXmlFragment("document-store")
  );
  const sourceSnapshot = Y.encodeStateAsUpdate(sourceDocument);
  const document = new Y.Doc();
  Y.applyUpdate(document, sourceSnapshot);

  const update: UpdateBlockPackYjsDocumentBlockRequest = {
    blockId: block.id,
    block: {
      ...block,
      content: [{ styles: {}, text: "after", type: "text" }],
    },
  };
  const result = new YjsBlockPackUpdateService(
    {} as ConstructorParameters<typeof YjsBlockPackUpdateService>[0]
  ).apply(document, [
    update,
    {
      blockId: "00000000-0000-0000-0000-000000000001",
      block: update.block,
    },
  ]);

  assert.deepEqual(result.response.blocks, [
    { blockId: block.id, status: "updated" },
    {
      blockId: "00000000-0000-0000-0000-000000000001",
      status: "skipped",
      reason: "block_not_found",
    },
  ]);
  assert.notEqual(result.update.length, 0);
  assert.deepEqual(new BlockPackProjector().projectYjsDocument(document), [
    {
      ...block,
      content: [{ styles: {}, text: "after", type: "text" }],
    },
  ]);

  sourceDocument.destroy();
  document.destroy();
});
