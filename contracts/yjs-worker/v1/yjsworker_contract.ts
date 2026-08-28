// Versioned TypeScript contract corresponding to constants.go in this package.
// The planned contract generator will own this file once it is introduced.

export const YjsBlockPackRoomPrefix = "block-pack";
export const YjsBlockPackFragmentName = "document-store";
export const YjsBlockPackSchemaId = "notegic.blocknote";
export const YjsBlockPackSchemaVersion = 1;
export const BlockPackDocumentQuotaPolicyVersion = 1;
export const YjsCompactionUpdateThreshold = 500;
export const YjsMaintenanceMaximumPayloadBytes = 64 * 1024 * 1024;
export const YjsDocumentMaximumLoadPayloadBytes = 64 * 1024 * 1024;
export const InternalFrameHeaderSize = 39;

export type BlockPackQuotaPolicy = {
  version: number;
  maximumBlockCount: number;
};

export const YjsWorkerCoreCommandTopic = "notegic.adapters.core.command.v1";
export const CoreYjsWorkerReplyTopic = "notegic.core.adapters.reply.v1";
export const CoreYjsMaintenanceHintTopic =
  "notegic.core.yjs-maintenance-hint.v1";
