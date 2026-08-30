import { upgradeWebSocket } from "@hono/node-server";
import type { Hono } from "hono";
import type WebSocket from "ws";

export function configureRealtimeRouter(
  app: Hono,
  handleConnection: (webSocket: WebSocket) => void
): void {
  app.get(
    "/core/realtime/v1",
    upgradeWebSocket(() => ({
      onOpen(_event, webSocket) {
        handleConnection(webSocket.raw as WebSocket);
      },
    }))
  );
}
