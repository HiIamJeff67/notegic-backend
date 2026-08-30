import assert from "node:assert/strict";
import test from "node:test";

import type WebSocket from "ws";
import * as Y from "yjs";
import { Telemetry } from "../../telemetry.js";
import { RoomRegistry } from "./room_registry.js";

const telemetry = Telemetry.initialize();

test.after(async () => {
  await telemetry.shutdown();
});

test("RoomRegistry tracks subscribers before a Yjs document is materialized", () => {
  const registry = new RoomRegistry(telemetry);
  const blockPackId = "7bc6ae1a-b1b3-47a7-9fab-42f34f48f7ca";
  const connectionId = "a7577a40-a86d-4fa9-9233-49c0b3e80385";
  const room = registry.attach(
    blockPackId,
    {} as WebSocket,
    connectionId,
    1,
    1,
    1000
  );
  assert.equal(registry.size, 1);
  assert.equal(room.subscribers.size, 1);
  assert.equal(room.document, null);
  assert.equal(registry.getSubscriber(blockPackId, connectionId, 1), room);

  const document = new Y.Doc();
  document.getMap("document-store").set("title", "Notegic");
  room.document = document;

  assert.equal(room.document.getMap("document-store").get("title"), "Notegic");

  registry.detach(blockPackId, connectionId, 1);

  assert.equal(room.subscribers.size, 0);
  assert.equal(registry.getSubscriber(blockPackId, connectionId, 1), undefined);
});

test("RoomRegistry cancels an idle eviction when a subscriber reattaches", () => {
  const registry = new RoomRegistry(telemetry);
  const blockPackId = "3dcbaa6f-6af6-4c09-90c1-4c5eaf4fda0f";
  const connectionId = "e42803f9-220d-4e35-8cc6-58faa2f9b0a6";
  const room = registry.attach(
    blockPackId,
    {} as WebSocket,
    connectionId,
    1,
    1,
    1000
  );

  registry.detach(blockPackId, connectionId, 1);

  registry.scheduleRoomEviction(blockPackId);
  assert.notEqual(room.idleEvictionTimer, null);

  registry.attach(blockPackId, {} as WebSocket, connectionId, 2, 1, 1000);

  assert.equal(registry.get(blockPackId), room);
  assert.equal(room.subscribers.size, 1);
  assert.equal(room.idleEvictionTimer, null);
});

test("RoomRegistry destroys a detached Yjs document when evicting a room", () => {
  const registry = new RoomRegistry(telemetry);
  const blockPackId = "fad8f69d-44f0-4893-b7b0-1015d64e7fc4";
  const connectionId = "e112b738-91f9-4a10-ae35-ae634a9b2c50";
  const room = registry.attach(
    blockPackId,
    {} as WebSocket,
    connectionId,
    1,
    1,
    1000
  );
  const document = new Y.Doc();
  let didDestroy = false;
  document.on("destroy", () => {
    didDestroy = true;
  });
  room.document = document;

  registry.detach(blockPackId, connectionId, 1);

  assert.equal(registry.evict(blockPackId, room), true);
  assert.equal(didDestroy, true);
  assert.equal(room.document, null);
  assert.equal(registry.get(blockPackId), undefined);
});
