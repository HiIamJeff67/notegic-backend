import { yXmlFragmentToBlocks } from "@blocknote/core/yjs";
import type * as Y from "yjs";
import { YjsBlockPackFragmentName } from "../../../contracts/yjs-worker/v1/yjsworker_contract.js";
import {
  type NotegicBlock,
  notegicBlockNoteEditor,
} from "../types/blocknote_schema.js";

export class BlockPackProjector {
  projectYjsDocument(document: Y.Doc): NotegicBlock[] {
    return yXmlFragmentToBlocks(
      notegicBlockNoteEditor,
      document.getXmlFragment(YjsBlockPackFragmentName)
    ) as NotegicBlock[];
  }

  countYjsDocumentBlocks(document: Y.Doc): number {
    const pendingBlocks = [...this.projectYjsDocument(document)];
    let blockCount = 0;
    while (pendingBlocks.length > 0) {
      const block = pendingBlocks.pop();
      if (block === undefined) continue;

      blockCount += 1;
      pendingBlocks.push(...block.children);
    }

    return blockCount;
  }
}
