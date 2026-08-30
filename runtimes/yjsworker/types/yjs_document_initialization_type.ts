import type { NotegicBlock } from "./blocknote_schema.js";

export type YjsDocumentInitializationRequest = {
  documents: Array<{
    blocks: NotegicBlock[];
  }>;
};
