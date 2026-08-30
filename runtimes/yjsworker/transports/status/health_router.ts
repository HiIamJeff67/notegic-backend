import type { Hono } from "hono";

export function configureHealthRouter(
  app: Hono,
  isHealthy: () => boolean
): void {
  app.get("/healthz", context => {
    if (!isHealthy()) {
      return context.body(null, 503);
    }

    return context.body(null, 200);
  });
}
