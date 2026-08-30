FROM node:22.23.1-slim AS base

WORKDIR /app

RUN npm install --global pnpm@11.7.0

FROM base AS production-dependencies

WORKDIR /app/runtimes/yjsworker

COPY runtimes/yjsworker/package.json runtimes/yjsworker/pnpm-lock.yaml runtimes/yjsworker/pnpm-workspace.yaml ./
RUN pnpm install --prod --frozen-lockfile

FROM base AS build

WORKDIR /app/runtimes/yjsworker

COPY runtimes/yjsworker/package.json runtimes/yjsworker/pnpm-lock.yaml runtimes/yjsworker/pnpm-workspace.yaml runtimes/yjsworker/tsconfig.json ./
RUN pnpm install --frozen-lockfile

COPY runtimes/yjsworker ./
COPY contracts/yjs-worker/v1/package.json contracts/yjs-worker/v1/yjsworker_contract.ts /app/contracts/yjs-worker/v1/
RUN pnpm build

FROM build AS dev

WORKDIR /app/runtimes/yjsworker

ENV NODE_ENV=development

CMD ["sh", "-ec", "pnpm --config.confirmModulesPurge=false install --frozen-lockfile && exec pnpm dev"]

FROM node:22.23.1-slim AS runtime

WORKDIR /app/runtimes/yjsworker

ENV NODE_ENV=production
ENV YJS_WORKER_HOST=0.0.0.0
ENV YJS_WORKER_PORT=8787

COPY --from=production-dependencies /app/runtimes/yjsworker/node_modules ./node_modules
COPY --from=build /app/runtimes/yjsworker/dist ./dist

USER node

EXPOSE 8787

HEALTHCHECK --interval=60s --timeout=3s --start-period=5s --retries=3 \
  CMD node --eval "fetch('http://127.0.0.1:8787/healthz').then((response) => process.exit(response.ok ? 0 : 1)).catch(() => process.exit(1))"

CMD ["node", "dist/runtimes/yjsworker/main.js"]
