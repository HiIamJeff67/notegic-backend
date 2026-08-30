export type YjsDocumentRow = {
  id: string;
  blockPackId: string;
  snapshot: Buffer;
  stateVector: Buffer;
  lastUpdateSequence: number;
  compactedUntilSequence: number;
  projectedUntilSequence: number;
};

export type YjsUpdateRow = {
  updateSequence: number;
  payload: Buffer;
};

export type YjsDocumentLoad = {
  document: YjsDocumentRow;
  updates: YjsUpdateRow[];
};

export type ProjectedBlock = {
  id: string;
  parentBlockId?: string | null;
  prevBlockId?: string | null;
  nextBlockId?: string | null;
  type: string;
  props: unknown;
  content: unknown;
  children?: ProjectedBlock[];
};
