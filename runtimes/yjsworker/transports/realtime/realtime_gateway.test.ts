import assert from "node:assert/strict";
import { EventEmitter } from "node:events";
import { readFile } from "node:fs/promises";
import test from "node:test";

import { blocksToYXmlFragment } from "@blocknote/core/yjs";
import { WebSocket } from "ws";
import {
  Awareness,
  applyAwarenessUpdate,
  encodeAwarenessUpdate,
} from "y-protocols/awareness";
import * as Y from "yjs";
import { BlockPackProjector } from "../../services/block_pack_projector.js";
import { YjsCompactionService } from "../../services/yjs_compaction_service.js";
import { Telemetry } from "../../telemetry.js";
import {
  type NotegicBlock,
  notegicBlockNoteEditor,
} from "../../types/blocknote_schema.js";
import {
  createInternalFrame,
  parseInternalFrame,
} from "../../types/internal_frame.js";
import { InternalFrameType } from "../../types/internal_frame_type.js";
import { CoreCommandDispatcher } from "../core/dispatchers/core_command_dispatcher.js";
import { RealtimeGateway } from "./realtime_gateway.js";
import { RoomRegistry } from "./room_registry.js";

class TestWebSocket extends EventEmitter {
  readonly sentFrames: Buffer[] = [];
  readyState = WebSocket.OPEN;

  send(frame: Buffer): void {
    this.sentFrames.push(frame);
  }

  close(): void {
    this.readyState = WebSocket.CLOSED;
  }
}

const telemetry = Telemetry.initialize();

test.after(async () => {
  await telemetry.shutdown();
});

test("RealtimeGateway isolates awareness client IDs and removes them on detach", () => {
  const roomRegistry = new RoomRegistry(telemetry);
  const gateway = new RealtimeGateway(
    roomRegistry,
    new YjsCompactionService(telemetry),
    new CoreCommandDispatcher(),
    telemetry
  );
  const firstSocket = new TestWebSocket();
  const secondSocket = new TestWebSocket();
  const firstConnectionId = "0f5e3ec2-9211-4f62-8e57-25fd5d8104ec";
  const secondConnectionId = "97553d3a-c805-4372-b624-c7c30aad5f10";
  const blockPackId = "4bb4cc0e-44e5-4c2f-a1a1-26e7b5150da6";

  gateway.handleConnection(firstSocket as unknown as WebSocket);
  gateway.handleConnection(secondSocket as unknown as WebSocket);

  const room = roomRegistry.attach(
    blockPackId,
    firstSocket as unknown as WebSocket,
    firstConnectionId,
    1,
    1,
    1000
  );
  roomRegistry.attach(
    blockPackId,
    secondSocket as unknown as WebSocket,
    secondConnectionId,
    2,
    1,
    1000
  );
  roomRegistry.initializeAwareness(room, new Y.Doc());
  for (const subscriber of room.subscribers.values()) {
    subscriber.isReady = true;
  }

  const sourceDocument = new Y.Doc();
  const sourceAwareness = new Awareness(sourceDocument);
  sourceAwareness.setLocalState({ user: "first" });
  const clientId = sourceAwareness.clientID;
  const update = Buffer.from(
    encodeAwarenessUpdate(sourceAwareness, [clientId])
  );

  firstSocket.emit(
    "message",
    createInternalFrame(
      InternalFrameType.InternalFrameType_Awareness,
      firstConnectionId,
      1,
      blockPackId,
      update
    ),
    true
  );

  assert.deepEqual(room.awareness?.getStates().get(clientId), {
    user: "first",
  });
  const awarenessFrame = parseInternalFrame(secondSocket.sentFrames.at(-1)!);
  assert.notEqual(awarenessFrame, null);
  assert.equal(
    awarenessFrame.type,
    InternalFrameType.InternalFrameType_Awareness
  );

  secondSocket.emit(
    "message",
    createInternalFrame(
      InternalFrameType.InternalFrameType_Awareness,
      secondConnectionId,
      2,
      blockPackId,
      update
    ),
    true
  );

  const resyncFrame = parseInternalFrame(secondSocket.sentFrames.at(-1)!);
  assert.notEqual(resyncFrame, null);
  assert.equal(
    resyncFrame.type,
    InternalFrameType.InternalFrameType_ResyncRequired
  );
  assert.deepEqual(room.awareness?.getStates().get(clientId), {
    user: "first",
  });

  const observedDocument = new Y.Doc();
  const observedAwareness = new Awareness(observedDocument);
  observedAwareness.setLocalState(null);
  applyAwarenessUpdate(observedAwareness, awarenessFrame.payload, gateway);
  assert.deepEqual(observedAwareness.getStates().get(clientId), {
    user: "first",
  });

  firstSocket.emit(
    "message",
    createInternalFrame(
      InternalFrameType.InternalFrameType_Detach,
      firstConnectionId,
      1,
      blockPackId
    ),
    true
  );

  assert.equal(room.awareness?.getStates().has(clientId), false);
  const removalFrame = parseInternalFrame(secondSocket.sentFrames.at(-1)!);
  assert.notEqual(removalFrame, null);
  assert.equal(
    removalFrame.type,
    InternalFrameType.InternalFrameType_Awareness
  );
  applyAwarenessUpdate(observedAwareness, removalFrame.payload, gateway);
  assert.equal(observedAwareness.getStates().has(clientId), false);

  sourceAwareness.destroy();
  sourceDocument.destroy();
  observedAwareness.destroy();
  observedDocument.destroy();
  roomRegistry.clearAwareness(room);
  room.document?.destroy();
});

test("RealtimeGateway rejects an entire update before it exceeds the BlockPack quota", async () => {
  const roomRegistry = new RoomRegistry(telemetry);
  const gateway = new RealtimeGateway(
    roomRegistry,
    new YjsCompactionService(telemetry),
    new CoreCommandDispatcher(),
    telemetry
  );
  const webSocket = new TestWebSocket();
  const connectionId = "84e53484-6773-421d-941f-ff337a0c85a2";
  const blockPackId = "f8a81806-c623-4388-a4c4-0c30d5cab42e";
  gateway.handleConnection(webSocket as unknown as WebSocket);

  const room = roomRegistry.attach(
    blockPackId,
    webSocket as unknown as WebSocket,
    connectionId,
    1,
    1,
    1
  );
  roomRegistry.initializeAwareness(room, new Y.Doc());
  const subscriber = room.subscribers.get(`${connectionId}:1`);
  assert.notEqual(subscriber, undefined);
  subscriber.isReady = true;

  const sourceBlocks = JSON.parse(
    await readFile(
      new URL("../../../../tmp/temp_wide_block_contents.json", import.meta.url),
      "utf8"
    )
  ) as NotegicBlock[];
  const sourceDocument = new Y.Doc();
  blocksToYXmlFragment(
    notegicBlockNoteEditor,
    sourceBlocks.slice(0, 2),
    sourceDocument.getXmlFragment("document-store")
  );
  webSocket.emit(
    "message",
    createInternalFrame(
      InternalFrameType.InternalFrameType_YjsDocument,
      connectionId,
      1,
      blockPackId,
      Buffer.from(Y.encodeStateAsUpdate(sourceDocument))
    ),
    true
  );

  const quotaFrame = parseInternalFrame(webSocket.sentFrames.at(-1)!);
  assert.notEqual(quotaFrame, null);
  assert.equal(
    quotaFrame.type,
    InternalFrameType.InternalFrameType_BlockPackQuotaExceeded
  );
  assert.equal(subscriber.quotaRecoveryRequired, true);
  assert.equal(room.pendingPersistenceUpdates.length, 0);
  assert.equal(
    new BlockPackProjector().countYjsDocumentBlocks(room.document!),
    0
  );
  assert.equal(
    new BlockPackProjector().countYjsDocumentBlocks(room.validationDocument!),
    0
  );

  sourceDocument.destroy();
  room.validationDocument?.destroy();
  room.document?.destroy();
});

test("RealtimeGateway bounds deferred updates while a room is loading", () => {
  const roomRegistry = new RoomRegistry(telemetry);
  const gateway = new RealtimeGateway(
    roomRegistry,
    new YjsCompactionService(telemetry),
    new CoreCommandDispatcher(),
    telemetry
  );
  const webSocket = new TestWebSocket();
  const connectionId = "d8b50393-f658-46f0-972d-f0d3e74c4636";
  const blockPackId = "e5b0a61f-08f0-4ae5-8e0b-439c7103302f";
  gateway.handleConnection(webSocket as unknown as WebSocket);
  const room = roomRegistry.attach(
    blockPackId,
    webSocket as unknown as WebSocket,
    connectionId,
    1,
    1,
    1000
  );

  for (let index = 0; index < 65; index += 1) {
    webSocket.emit(
      "message",
      createInternalFrame(
        InternalFrameType.InternalFrameType_YjsDocument,
        connectionId,
        1,
        blockPackId,
        Buffer.from([index])
      ),
      true
    );
  }

  assert.equal(room.pendingYjsUpdates.length, 64);
  assert.equal(room.pendingYjsPayloadBytes, 64);
  const resyncFrame = parseInternalFrame(webSocket.sentFrames.at(-1)!);
  assert.notEqual(resyncFrame, null);
  assert.equal(
    resyncFrame.type,
    InternalFrameType.InternalFrameType_ResyncRequired
  );
});
