import type { Hono } from "hono";

export function configureStartedRouter(
  app: Hono,
  isStarted: () => boolean
): void {
  app.get("/startedz", context => {
    if (!isStarted()) {
      return context.body(null, 503);
    }

    return context.body(null, 200);
  });
}
