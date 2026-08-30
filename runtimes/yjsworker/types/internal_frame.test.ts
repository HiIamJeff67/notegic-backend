import assert from "node:assert/strict";
import test from "node:test";

import {
  createInternalFrame,
  parseInternalFrame,
} from "./internal_frame.js";
import { InternalFrameType } from "./internal_frame_type.js";

test("parses an attach frame", () => {
  const connectionId = "0c65c336-5da0-4dfc-b6f9-524599155733";
  const blockPackId = "1d18feb5-3b73-451c-b7ea-e3bdcbbf2fea";
  const frame = createInternalFrame(
    InternalFrameType.InternalFrameType_Attach,
    connectionId,
    7,
    blockPackId
  );

  assert.deepEqual(parseInternalFrame(frame), {
    version: 1,
    type: InternalFrameType.InternalFrameType_Attach,
    connectionId,
    connectorChannelId: 7,
    blockPackId,
    payload: Buffer.alloc(0),
  });
});
