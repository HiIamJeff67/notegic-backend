import type WebSocket from "ws";

import type { InternalFrame } from "./internal_frame.js";

// PendingYjsUpdate is a raw public update retained until its Y.Doc is available or persistence batch flush runs
export type PendingYjsUpdate = {
  webSocket: WebSocket;
  frame: InternalFrame;
};
