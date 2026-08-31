export type UpdateBlockPackYjsDocumentBlockRequest = {
  blockId: string;
  block: {
    id: string;
    type: string;
    props: unknown;
    content: unknown;
    children: Array<unknown>;
  };
};

export type UpdateBlockPackYjsDocumentRequest = {
  blockPackId: string;
  blocks: Array<UpdateBlockPackYjsDocumentBlockRequest>;
};

export type UpdateBlockPackYjsDocumentBlockResult = {
  blockId: string;
  status: "updated" | "skipped" | "failed";
  reason?: string;
};

export type UpdateBlockPackYjsDocumentResponse = {
  blocks: Array<UpdateBlockPackYjsDocumentBlockResult>;
};
