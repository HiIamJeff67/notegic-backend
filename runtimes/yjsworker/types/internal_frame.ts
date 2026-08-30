import { InternalFrameHeaderSize } from "../../../contracts/yjs-worker/v1/yjsworker_contract.js";
import { convertBytesToUUIDString, convertUUIDStringToBytes } from "../util/uuid.js";
import { InternalChannelType } from "./internal_channel_type.js";
import { InternalFrameType } from "./internal_frame_type.js";

export type InternalFrame = {
  version: number;
  type: InternalFrameType;
  connectionId: string;
  connectorChannelId: number;
  blockPackId: string;
  payload: Buffer;
};

export function parseInternalFrame(payload: Buffer): InternalFrame | null {
  // [version:1][type:1][channelType:1][connectionId:16][connectorChannelId:4][channelId:16][raw payload:n]
  if (payload.length < InternalFrameHeaderSize) {
    return null;
  }
  if (payload[2] !== InternalChannelType.InternalChannelType_BlockPack) {
    return null;
  }

  const type = payload[1] as InternalFrameType;
  if (!Object.values(InternalFrameType).includes(type)) {
    return null;
  }

  const connectorChannelId = payload.readUInt32BE(19);
  if (connectorChannelId === 0) {
    return null;
  }

  const connectionId = convertBytesToUUIDString(payload.subarray(3, 19));
  const blockPackId = convertBytesToUUIDString(payload.subarray(23, 39));
  if (connectionId === null || blockPackId === null) {
    return null;
  }

  return {
    version: payload[0],
    type,
    connectionId,
    connectorChannelId,
    blockPackId,
    payload: payload.subarray(InternalFrameHeaderSize),
  };
}

export function createInternalFrame(
  type: InternalFrameType,
  connectionId: string,
  connectorChannelId: number,
  blockPackId: string,
  payload: Buffer = Buffer.alloc(0)
): Buffer {
  const frame = Buffer.alloc(InternalFrameHeaderSize + payload.length);

  frame[0] = 1;
  frame[1] = type;
  frame[2] = InternalChannelType.InternalChannelType_BlockPack;
  convertUUIDStringToBytes(connectionId).copy(frame, 3);
  frame.writeUInt32BE(connectorChannelId, 19);
  convertUUIDStringToBytes(blockPackId).copy(frame, 23);
  payload.copy(frame, InternalFrameHeaderSize);

  return frame;
}
