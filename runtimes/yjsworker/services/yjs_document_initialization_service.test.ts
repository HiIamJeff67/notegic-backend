import assert from "node:assert/strict";
import test from "node:test";

import * as Y from "yjs";

import { BlockPackProjector } from "./block_pack_projector.js";
import { YjsDocumentInitializationService } from "./yjs_document_initialization_service.js";
import type { NotegicBlock } from "../types/blocknote_schema.js";

test("YjsDocumentInitializationService preserves initial BlockNote content", () => {
  const blocks: NotegicBlock[] = [
    {
      id: "c58c8cba-74b3-46e6-a758-16530edc9a01",
      type: "paragraph",
      props: {
        backgroundColor: "default",
        textAlignment: "left",
        textColor: "default",
      },
      content: [
        {
          styles: {},
          text: "CI daily note - 2026-07-27",
          type: "text",
        },
      ],
      children: [],
    },
  ];
  const result = new YjsDocumentInitializationService().initialize(blocks);
  const document = new Y.Doc();

  Y.applyUpdate(document, result.snapshot);

  assert.deepEqual(
    new BlockPackProjector().projectYjsDocument(document),
    blocks
  );
  assert.notEqual(result.stateVector.length, 0);

  document.destroy();
});
