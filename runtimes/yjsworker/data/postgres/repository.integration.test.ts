import assert from "node:assert/strict";
import { randomUUID } from "node:crypto";
import { after, before, test } from "node:test";

import { blocksToYXmlFragment } from "@blocknote/core/yjs";
import { Pool } from "pg";
import * as Y from "yjs";

import { BlockPackProjector } from "../../services/block_pack_projector.js";
import { YjsCompactionService } from "../../services/yjs_compaction_service.js";
import { YjsProjectionService } from "../../services/yjs_projection_service.js";
import type { Telemetry } from "../../telemetry.js";
import {
  type NotegicBlock,
  notegicBlockNoteEditor,
} from "../../types/blocknote_schema.js";
import type { ProjectedBlock } from "./models.js";
import type { YjsPostgresRepository } from "./repository.js";

const runIntegration = process.env.NOTEGIC_RUN_INTEGRATION === "1";
const yjsRole = "notegic_yjsworker_test";
const yjsRolePassword = "notegic_yjsworker_test_password";
const tables = [
  '"BlockTable"',
  '"BlockPackTable"',
  '"BlockPackYjsDocumentTable"',
  '"BlockPackYjsUpdateTable"',
  '"UserTable"',
];

let adminPool: Pool | undefined;
let globalPostgresPool: Pool | undefined;
let repository: YjsPostgresRepository | undefined;
let YjsPostgresRepositoryConstructor: typeof import("./repository.js").YjsPostgresRepository;

function databaseConfig(): {
  host: string;
  port: number;
  user: string;
  password: string;
  database: string;
} {
  return {
    host: process.env.YJS_DB_HOST ?? "127.0.0.1",
    port: Number(process.env.YJS_DB_PORT ?? "15432"),
    user: process.env.YJS_DB_USER ?? "notegic",
    password: process.env.YJS_DB_PASSWORD ?? "notegic",
    database: process.env.YJS_DB_NAME ?? "notegic_integration",
  };
}

async function resetDatabase(): Promise<void> {
  if (adminPool === undefined) return;

  for (const table of tables) {
    await adminPool.query(`DROP TABLE IF EXISTS ${table} CASCADE`);
  }

  await adminPool.query(`
    CREATE TABLE "BlockPackTable" (
      id uuid PRIMARY KEY,
      deleted_at timestamptz
    )
  `);
  await adminPool.query(`
    CREATE TABLE "BlockPackYjsDocumentTable" (
      id uuid PRIMARY KEY,
      block_pack_id uuid NOT NULL,
      snapshot bytea NOT NULL,
      state_vector bytea NOT NULL,
      last_update_sequence bigint NOT NULL,
      compacted_until_sequence bigint NOT NULL,
      last_compacted_at timestamptz,
      projected_until_sequence bigint NOT NULL,
      deleted_at timestamptz,
      updated_at timestamptz NOT NULL,
      created_at timestamptz NOT NULL
    )
  `);
  await adminPool.query(`
    CREATE TABLE "BlockPackYjsUpdateTable" (
      id uuid PRIMARY KEY,
      block_pack_id uuid NOT NULL,
      update_sequence bigint NOT NULL,
      payload bytea NOT NULL,
      compacted_at timestamptz,
      created_at timestamptz NOT NULL
    )
  `);
  await adminPool.query(`
    CREATE TABLE "BlockTable" (
      id uuid PRIMARY KEY,
      block_pack_id uuid NOT NULL,
      parent_block_id uuid,
      prev_block_id uuid,
      next_block_id uuid,
      type text NOT NULL CHECK (type <> 'forbidden'),
      props jsonb NOT NULL,
      content jsonb NOT NULL,
      updated_at timestamptz NOT NULL,
      created_at timestamptz NOT NULL
    )
  `);
  await adminPool.query('CREATE TABLE "UserTable" (id uuid PRIMARY KEY)');

  await dropTestRole();
  await adminPool.query(
    `CREATE ROLE "${yjsRole}" LOGIN PASSWORD '${yjsRolePassword}' NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS NOINHERIT`
  );
  await adminPool.query(`GRANT USAGE ON SCHEMA public TO "${yjsRole}"`);
  await adminPool.query(
    `GRANT SELECT ON TABLE "BlockPackTable" TO "${yjsRole}"`
  );
  await adminPool.query(
    `GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE "BlockTable", "BlockPackYjsDocumentTable", "BlockPackYjsUpdateTable" TO "${yjsRole}"`
  );
  await adminPool.query(`REVOKE ALL ON TABLE "UserTable" FROM "${yjsRole}"`);
}

async function dropTestRole(): Promise<void> {
  if (adminPool === undefined) return;

  const role = await adminPool.query(
    "SELECT 1 FROM pg_roles WHERE rolname = $1",
    [yjsRole]
  );
  if (role.rowCount === 0) return;

  await adminPool.query(`DROP OWNED BY "${yjsRole}"`);
  await adminPool.query(`DROP ROLE "${yjsRole}"`);
}

async function seedDocument(
  blockPackId: string,
  blocks: ProjectedBlock[]
): Promise<void> {
  assert.ok(adminPool);

  const sourceDocument = new Y.Doc();
  blocksToYXmlFragment(
    notegicBlockNoteEditor,
    blocks as unknown as NotegicBlock[],
    sourceDocument.getXmlFragment("document-store")
  );
  const payload = Buffer.from(Y.encodeStateAsUpdate(sourceDocument));
  const emptyDocument = new Y.Doc();
  const stateVector = Buffer.from(Y.encodeStateVector(emptyDocument));
  sourceDocument.destroy();
  emptyDocument.destroy();

  const now = new Date();
  await adminPool.query(
    `INSERT INTO "BlockPackTable" (id, deleted_at) VALUES ($1, NULL)`,
    [blockPackId]
  );
  await adminPool.query(
    `INSERT INTO "BlockPackYjsDocumentTable"
      (id, block_pack_id, snapshot, state_vector, last_update_sequence,
       compacted_until_sequence, projected_until_sequence, deleted_at,
       updated_at, created_at)
     VALUES ($1, $2, $3, $4, 1, 0, -1, NULL, $5, $5)`,
    [randomUUID(), blockPackId, Buffer.alloc(0), stateVector, now]
  );
  await adminPool.query(
    `INSERT INTO "BlockPackYjsUpdateTable"
      (id, block_pack_id, update_sequence, payload, created_at)
     VALUES ($1, $2, 1, $3, $4)`,
    [randomUUID(), blockPackId, payload, now]
  );
}

function telemetryRecorder(): {
  telemetry: Telemetry;
  operations: Array<{ operation: string; outcome: "success" | "error" }>;
} {
  const operations: Array<{
    operation: string;
    outcome: "success" | "error";
  }> = [];
  const span = {
    end(): void {},
    recordException(): void {},
    setStatus(): void {},
  };

  return {
    telemetry: {
      startSpan: () => span,
      recordOperation: operation => {
        operations.push({
          operation: operation.operation,
          outcome: operation.outcome,
        });
      },
    } as unknown as Telemetry,
    operations,
  };
}

before(async () => {
  if (!runIntegration) return;

  adminPool = new Pool(databaseConfig());
  const repositoryModule = await import("./repository.js");
  const poolModule = await import("./pool.js");
  YjsPostgresRepositoryConstructor = repositoryModule.YjsPostgresRepository;
  globalPostgresPool = poolModule.postgresPool;
  repository = new YjsPostgresRepositoryConstructor(adminPool);
  await resetDatabase();
});

after(async () => {
  if (adminPool !== undefined) {
    for (const table of tables) {
      await adminPool.query(`DROP TABLE IF EXISTS ${table} CASCADE`);
    }
    await dropTestRole();
    await adminPool.end();
  }
  await globalPostgresPool?.end();
});

test("runs compaction and projection through the Yjs PostgreSQL repository", async t => {
  if (!runIntegration) {
    t.skip("set NOTEGIC_RUN_INTEGRATION=1 to run database tests");
    return;
  }
  assert.ok(repository);
  assert.ok(adminPool);

  const { telemetry, operations } = telemetryRecorder();
  const blocks: ProjectedBlock[] = [
    {
      id: randomUUID(),
      type: "paragraph",
      props: {},
      content: [{ type: "text", text: "integration", styles: {} }],
      children: [],
    },
  ];
  const blockPackId = randomUUID();
  await seedDocument(blockPackId, blocks);

  const compactionService = new YjsCompactionService(telemetry);
  const projectionService = new YjsProjectionService(
    new BlockPackProjector(),
    telemetry
  );
  const loaded = await repository.loadCompactable(blockPackId, 1);
  assert.ok(loaded);
  const compacted = compactionService.compact({
    snapshot: loaded.document.snapshot,
    stateVector: loaded.document.stateVector,
    baseCompactedUntilSequence: loaded.document.compactedUntilSequence,
    cutoffSequence: 1,
    updates: loaded.updates,
  });
  assert.deepEqual(
    await repository.applyCompaction({ blockPackId, ...compacted }),
    { applied: true, compactedUntilSequence: 1 }
  );

  const projectable = await repository.loadProjectable(blockPackId, 1);
  assert.ok(projectable);
  const projection = projectionService.project({
    blockPackId,
    state: {
      snapshot: projectable.document.snapshot,
      stateVector: projectable.document.stateVector,
      lastUpdateSequence: 1,
      compactedUntilSequence: projectable.document.compactedUntilSequence,
      projectedUntilSequence: projectable.document.projectedUntilSequence,
      updates: projectable.updates,
    },
  });
  assert.deepEqual(
    await repository.applyProjection({
      blockPackId,
      projectedSequence: 1,
      blocks: projection.blocks,
    }),
    { applied: true, projectedUntilSequence: 1 }
  );

  const result = await adminPool.query(
    `SELECT type, content FROM "BlockTable" WHERE block_pack_id = $1`,
    [blockPackId]
  );
  assert.equal(result.rowCount, 1);
  assert.equal(result.rows[0].type, "paragraph");
  assert.deepEqual(result.rows[0].content, [
    { type: "text", text: "integration", styles: {} },
  ]);
  assert.deepEqual(operations, [
    { operation: "compaction", outcome: "success" },
    { operation: "projection", outcome: "success" },
  ]);
});

test("allows only one concurrent compaction to win the row lock and CAS", async t => {
  if (!runIntegration) {
    t.skip("set NOTEGIC_RUN_INTEGRATION=1 to run database tests");
    return;
  }
  assert.ok(adminPool);

  const firstPool = new Pool(databaseConfig());
  const secondPool = new Pool(databaseConfig());
  const firstRepository = new YjsPostgresRepositoryConstructor(firstPool);
  const secondRepository = new YjsPostgresRepositoryConstructor(secondPool);
  try {
    const blockPackId = randomUUID();
    await seedDocument(blockPackId, []);
    const [first, second] = await Promise.all([
      firstRepository.applyCompaction({
        blockPackId,
        baseCompactedUntilSequence: 0,
        cutoffSequence: 1,
        snapshot: Buffer.from("first"),
        stateVector: Buffer.alloc(0),
      }),
      secondRepository.applyCompaction({
        blockPackId,
        baseCompactedUntilSequence: 0,
        cutoffSequence: 1,
        snapshot: Buffer.from("second"),
        stateVector: Buffer.alloc(0),
      }),
    ]);

    assert.equal(Number(first.applied) + Number(second.applied), 1);
    const result = await adminPool.query(
      `SELECT compacted_until_sequence FROM "BlockPackYjsDocumentTable" WHERE block_pack_id = $1`,
      [blockPackId]
    );
    assert.equal(result.rows[0].compacted_until_sequence, "1");
  } finally {
    await firstRepository.close();
    await secondRepository.close();
  }
});

test("rolls back a failed projection without leaving partial blocks", async t => {
  if (!runIntegration) {
    t.skip("set NOTEGIC_RUN_INTEGRATION=1 to run database tests");
    return;
  }
  assert.ok(repository);
  assert.ok(adminPool);

  const blockPackId = randomUUID();
  await seedDocument(blockPackId, []);
  await assert.rejects(
    repository.applyProjection({
      blockPackId,
      projectedSequence: 1,
      blocks: [
        {
          id: randomUUID(),
          type: "paragraph",
          props: {},
          content: [],
        },
        {
          id: randomUUID(),
          type: "forbidden",
          props: {},
          content: [],
        },
      ],
    })
  );

  const blocks = await adminPool.query(
    `SELECT COUNT(*)::int AS count FROM "BlockTable" WHERE block_pack_id = $1`,
    [blockPackId]
  );
  const document = await adminPool.query(
    `SELECT projected_until_sequence FROM "BlockPackYjsDocumentTable" WHERE block_pack_id = $1`,
    [blockPackId]
  );
  assert.equal(blocks.rows[0].count, 0);
  assert.equal(Number(document.rows[0].projected_until_sequence), -1);
});

test("enforces the Yjs database permission boundary", async t => {
  if (!runIntegration) {
    t.skip("set NOTEGIC_RUN_INTEGRATION=1 to run database tests");
    return;
  }
  assert.ok(adminPool);

  const restrictedPool = new Pool({
    ...databaseConfig(),
    user: yjsRole,
    password: yjsRolePassword,
  });
  const restrictedRepository = new YjsPostgresRepositoryConstructor(
    restrictedPool
  );
  try {
    const blockPackId = randomUUID();
    await seedDocument(blockPackId, []);
    const loaded = await restrictedRepository.loadProjectable(blockPackId, 1);
    assert.ok(loaded);
    assert.equal(loaded.document.blockPackId, blockPackId);

    await assert.rejects(
      restrictedPool.query('SELECT * FROM "UserTable"'),
      error => (error as { code?: string }).code === "42501"
    );
  } finally {
    await restrictedRepository.close();
  }
});

test("records a failed compaction operation for observability", async t => {
  if (!runIntegration) {
    t.skip("set NOTEGIC_RUN_INTEGRATION=1 to run database tests");
    return;
  }

  const { telemetry, operations } = telemetryRecorder();
  const service = new YjsCompactionService(telemetry);
  assert.throws(() =>
    service.compact({
      snapshot: Buffer.from([1]),
      stateVector: Buffer.alloc(0),
      baseCompactedUntilSequence: 0,
      cutoffSequence: 1,
      updates: [],
    })
  );
  assert.deepEqual(operations, [{ operation: "compaction", outcome: "error" }]);
});
