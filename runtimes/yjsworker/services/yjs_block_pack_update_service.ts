import { randomUUID } from "node:crypto";
import { blocksToYXmlFragment } from "@blocknote/core/yjs";
import * as Y from "yjs";

import type {
  UpdateBlockPackYjsDocumentBlockRequest,
  UpdateBlockPackYjsDocumentRequest,
  UpdateBlockPackYjsDocumentResponse,
} from "../../../contracts/yjs-worker/v1/update_block_pack.js";
import { YjsBlockPackFragmentName } from "../../../contracts/yjs-worker/v1/yjsworker_contract.js";
import { CoreCommandDispatcher } from "../transports/core/dispatchers/core_command_dispatcher.js";
import type { NotegicBlock } from "../types/blocknote_schema.js";
import { notegicBlockNoteEditor } from "../types/blocknote_schema.js";
import { parseYjsDocumentState } from "../types/yjs_document_state.js";
import { BlockPackProjector } from "./block_pack_projector.js";

export class YjsBlockPackUpdateService {
  private readonly coreCommandDispatcher: CoreCommandDispatcher;
  private readonly blockPackProjector: BlockPackProjector;

  constructor(
    coreCommandDispatcher: CoreCommandDispatcher,
    blockPackProjector = new BlockPackProjector()
  ) {
    this.coreCommandDispatcher = coreCommandDispatcher;
    this.blockPackProjector = blockPackProjector;
  }

  apply(
    document: Y.Doc,
    updates: UpdateBlockPackYjsDocumentBlockRequest[]
  ): {
    response: UpdateBlockPackYjsDocumentResponse;
    update: Buffer;
  } {
    const stateVector = Y.encodeStateVector(document);
    const blocks = this.blockPackProjector.projectYjsDocument(document);
    const response: UpdateBlockPackYjsDocumentResponse = { blocks: [] };
    let didUpdate = false;

    for (const update of updates) {
      const target = this.findBlock(blocks, update.blockId);
      if (target === null) {
        response.blocks.push({
          blockId: update.blockId,
          status: "skipped",
          reason: "block_not_found",
        });
        continue;
      }

      target.id = update.blockId;
      target.type = update.block.type as typeof target.type;
      target.props = update.block.props as typeof target.props;
      target.content = update.block.content as typeof target.content;
      didUpdate = true;
      response.blocks.push({ blockId: update.blockId, status: "updated" });
    }

    if (didUpdate) {
      blocksToYXmlFragment(
        notegicBlockNoteEditor,
        blocks,
        document.getXmlFragment(YjsBlockPackFragmentName)
      );
    }

    return {
      response,
      update: didUpdate
        ? Buffer.from(Y.encodeStateAsUpdate(document, stateVector))
        : Buffer.alloc(0),
    };
  }

  async update(
    request: UpdateBlockPackYjsDocumentRequest
  ): Promise<UpdateBlockPackYjsDocumentResponse> {
    const loaded = await this.coreCommandDispatcher.dispatchAsync<
      Record<string, never>,
      { found: boolean; payload?: string }
    >("LoadYjsDocument", request.blockPackId, {});
    const loadedResponse = await loaded.reply;
    if (!loadedResponse.found || loadedResponse.payload === undefined) {
      throw new Error("the Yjs document was not found");
    }

    const state = parseYjsDocumentState(
      Buffer.from(loadedResponse.payload, "base64")
    );
    if (state === null) {
      throw new Error("the Yjs document state is invalid");
    }

    const document = new Y.Doc();
    try {
      if (state.snapshot.length > 0) Y.applyUpdate(document, state.snapshot);
      for (const update of state.updates) {
        Y.applyUpdate(document, update.payload);
      }

      const result = this.apply(document, request.blocks);
      if (result.update.length > 0) {
        const persisted = await this.coreCommandDispatcher.dispatchAsync<
          {
            persistenceBatchId: string;
            originConnectionId: null;
            payload: string;
          },
          { updateSequence: number }
        >("AppendYjsUpdate", request.blockPackId, {
          persistenceBatchId: randomUUID(),
          originConnectionId: null,
          payload: result.update.toString("base64"),
        });
        await persisted.reply;
      }

      return result.response;
    } finally {
      document.destroy();
    }
  }

  private findBlock(
    blocks: NotegicBlock[],
    blockId: string
  ): NotegicBlock | null {
    const pendingBlocks = [...blocks];
    while (pendingBlocks.length > 0) {
      const block = pendingBlocks.pop();
      if (block === undefined) continue;
      if (block.id === blockId) return block;
      pendingBlocks.push(...block.children);
    }

    return null;
  }
}
