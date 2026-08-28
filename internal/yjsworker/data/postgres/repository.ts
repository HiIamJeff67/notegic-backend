import {
  and,
  asc,
  eq,
  gt,
  inArray,
  isNull,
  lte,
  notInArray,
  sql,
} from "drizzle-orm";
import { drizzle, type NodePgDatabase } from "drizzle-orm/node-postgres";
import { type Pool } from "pg";

import type {
  ProjectedBlock,
  YjsDocumentLoad,
  YjsDocumentRow,
} from "./models.js";
import { postgresPool } from "./pool.js";
import {
  blockPackTable,
  blockPackYjsDocumentTable,
  blockPackYjsUpdateTable,
  blockTable,
  schema,
} from "./schema.js";

type YjsDatabase = NodePgDatabase<typeof schema>;
type YjsQueryDatabase = Pick<YjsDatabase, "select">;

type DocumentResult = {
  document: YjsDocumentRow | null;
  updates: YjsDocumentLoad["updates"];
};

export class YjsPostgresRepository {
  private readonly pool: Pool;
  private readonly db: YjsDatabase;

  constructor(pool: Pool = postgresPool) {
    this.pool = pool;
    this.db = drizzle({ client: pool, schema });
  }

  private async loadDocument(
    db: YjsQueryDatabase,
    blockPackId: string,
    targetSequence: number
  ): Promise<DocumentResult> {
    const [document] = await db
      .select({
        id: blockPackYjsDocumentTable.id,
        blockPackId: blockPackYjsDocumentTable.blockPackId,
        snapshot: blockPackYjsDocumentTable.snapshot,
        stateVector: blockPackYjsDocumentTable.stateVector,
        lastUpdateSequence: blockPackYjsDocumentTable.lastUpdateSequence,
        compactedUntilSequence:
          blockPackYjsDocumentTable.compactedUntilSequence,
        projectedUntilSequence:
          blockPackYjsDocumentTable.projectedUntilSequence,
      })
      .from(blockPackYjsDocumentTable)
      .innerJoin(
        blockPackTable,
        eq(blockPackTable.id, blockPackYjsDocumentTable.blockPackId)
      )
      .where(
        and(
          eq(blockPackYjsDocumentTable.blockPackId, blockPackId),
          isNull(blockPackYjsDocumentTable.deletedAt),
          isNull(blockPackTable.deletedAt)
        )
      )
      .limit(1)
      .for("update");
    if (document === undefined) {
      return { document: null, updates: [] };
    }

    const boundedTarget = Math.min(targetSequence, document.lastUpdateSequence);
    const updates = await db
      .select({
        updateSequence: blockPackYjsUpdateTable.updateSequence,
        payload: blockPackYjsUpdateTable.payload,
      })
      .from(blockPackYjsUpdateTable)
      .where(
        and(
          eq(blockPackYjsUpdateTable.blockPackId, blockPackId),
          gt(
            blockPackYjsUpdateTable.updateSequence,
            document.compactedUntilSequence
          ),
          lte(blockPackYjsUpdateTable.updateSequence, boundedTarget)
        )
      )
      .orderBy(asc(blockPackYjsUpdateTable.updateSequence));

    return {
      document: {
        id: document.id,
        blockPackId: document.blockPackId,
        snapshot: document.snapshot,
        stateVector: document.stateVector,
        lastUpdateSequence: document.lastUpdateSequence,
        compactedUntilSequence: document.compactedUntilSequence,
        projectedUntilSequence: document.projectedUntilSequence,
      },
      updates,
    };
  }

  async loadCompactable(
    blockPackId: string,
    targetSequence: number
  ): Promise<YjsDocumentLoad | null> {
    return this.db.transaction(async tx => {
      const loaded = await this.loadDocument(tx, blockPackId, targetSequence);
      if (loaded.document === null) return null;

      return { document: loaded.document, updates: loaded.updates };
    });
  }

  async loadProjectable(
    blockPackId: string,
    targetSequence: number
  ): Promise<YjsDocumentLoad | null> {
    return this.db.transaction(async tx => {
      const loaded = await this.loadDocument(tx, blockPackId, targetSequence);
      if (loaded.document === null) return null;

      return { document: loaded.document, updates: loaded.updates };
    });
  }

  async applyCompaction(input: {
    blockPackId: string;
    baseCompactedUntilSequence: number;
    cutoffSequence: number;
    snapshot: Buffer;
    stateVector: Buffer;
  }): Promise<{ applied: boolean; compactedUntilSequence: number }> {
    return this.db.transaction(async tx => {
      const [document] = await tx
        .select({
          id: blockPackYjsDocumentTable.id,
          compactedUntilSequence:
            blockPackYjsDocumentTable.compactedUntilSequence,
          lastUpdateSequence: blockPackYjsDocumentTable.lastUpdateSequence,
        })
        .from(blockPackYjsDocumentTable)
        .where(
          and(
            eq(blockPackYjsDocumentTable.blockPackId, input.blockPackId),
            isNull(blockPackYjsDocumentTable.deletedAt)
          )
        )
        .for("update");
      if (document === undefined) {
        return { applied: false, compactedUntilSequence: 0 };
      }

      const currentCompacted = document.compactedUntilSequence;
      const lastUpdate = document.lastUpdateSequence;
      if (currentCompacted !== input.baseCompactedUntilSequence) {
        return { applied: false, compactedUntilSequence: currentCompacted };
      }
      if (input.cutoffSequence > lastUpdate) {
        throw new Error(
          "yjs compaction cutoff exceeds durable update sequence"
        );
      }

      await tx
        .update(blockPackYjsDocumentTable)
        .set({
          snapshot: input.snapshot,
          stateVector: input.stateVector,
          compactedUntilSequence: input.cutoffSequence,
          lastCompactedAt: sql`NOW()`,
          updatedAt: sql`NOW()`,
        })
        .where(eq(blockPackYjsDocumentTable.id, document.id));
      await tx
        .update(blockPackYjsUpdateTable)
        .set({ compactedAt: sql`NOW()` })
        .where(
          and(
            eq(blockPackYjsUpdateTable.blockPackId, input.blockPackId),
            gt(
              blockPackYjsUpdateTable.updateSequence,
              input.baseCompactedUntilSequence
            ),
            lte(blockPackYjsUpdateTable.updateSequence, input.cutoffSequence),
            isNull(blockPackYjsUpdateTable.compactedAt)
          )
        );

      return {
        applied: true,
        compactedUntilSequence: input.cutoffSequence,
      };
    });
  }

  async applyProjection(input: {
    blockPackId: string;
    projectedSequence: number;
    blocks: ProjectedBlock[];
  }): Promise<{ applied: boolean; projectedUntilSequence: number }> {
    return this.db.transaction(async tx => {
      const [document] = await tx
        .select({
          id: blockPackYjsDocumentTable.id,
          projectedUntilSequence:
            blockPackYjsDocumentTable.projectedUntilSequence,
          lastUpdateSequence: blockPackYjsDocumentTable.lastUpdateSequence,
        })
        .from(blockPackYjsDocumentTable)
        .innerJoin(
          blockPackTable,
          eq(blockPackTable.id, blockPackYjsDocumentTable.blockPackId)
        )
        .where(
          and(
            eq(blockPackYjsDocumentTable.blockPackId, input.blockPackId),
            isNull(blockPackYjsDocumentTable.deletedAt),
            isNull(blockPackTable.deletedAt)
          )
        )
        .for("update");
      if (document === undefined) {
        return { applied: false, projectedUntilSequence: -1 };
      }

      const currentProjected = document.projectedUntilSequence;
      const lastUpdate = document.lastUpdateSequence;
      if (input.projectedSequence <= currentProjected) {
        return { applied: false, projectedUntilSequence: currentProjected };
      }
      if (input.projectedSequence > lastUpdate) {
        throw new Error("block projection target exceeds durable yjs state");
      }

      const blocks: ProjectedBlock[] = [];
      const pendingBlocks: Array<{
        block: ProjectedBlock;
        parentBlockId: string | null;
        previousBlockId: string | null;
        nextBlockId: string | null;
      }> = [];
      for (let index = input.blocks.length - 1; index >= 0; index -= 1) {
        pendingBlocks.push({
          block: input.blocks[index],
          parentBlockId: null,
          previousBlockId: index > 0 ? input.blocks[index - 1].id : null,
          nextBlockId:
            index + 1 < input.blocks.length ? input.blocks[index + 1].id : null,
        });
      }
      while (pendingBlocks.length > 0) {
        const pendingBlock = pendingBlocks.pop();
        if (pendingBlock === undefined) continue;

        const block = pendingBlock.block;
        blocks.push({
          ...block,
          parentBlockId: block.parentBlockId ?? pendingBlock.parentBlockId,
          prevBlockId: block.prevBlockId ?? pendingBlock.previousBlockId,
          nextBlockId: block.nextBlockId ?? pendingBlock.nextBlockId,
          children: undefined,
        });

        if (block.children === undefined) continue;

        for (let index = block.children.length - 1; index >= 0; index -= 1) {
          pendingBlocks.push({
            block: block.children[index],
            parentBlockId: block.id,
            previousBlockId: index > 0 ? block.children[index - 1].id : null,
            nextBlockId:
              index + 1 < block.children.length
                ? block.children[index + 1].id
                : null,
          });
        }
      }
      const blockIds = new Set<string>();
      for (const block of blocks) {
        if (blockIds.has(block.id)) {
          throw new Error("block projection contains duplicate block ids");
        }
        blockIds.add(block.id);
      }

      if (blockIds.size > 0) {
        const existingBlocks = await tx
          .select({ blockPackId: blockTable.blockPackId })
          .from(blockTable)
          .where(inArray(blockTable.id, [...blockIds]))
          .for("update");
        for (const existingBlock of existingBlocks) {
          if (existingBlock.blockPackId !== input.blockPackId) {
            throw new Error(
              "block projection contains an id owned by another block pack"
            );
          }
        }
      }

      if (blocks.length === 0) {
        await tx
          .delete(blockTable)
          .where(eq(blockTable.blockPackId, input.blockPackId));
      } else {
        await tx
          .delete(blockTable)
          .where(
            and(
              eq(blockTable.blockPackId, input.blockPackId),
              notInArray(blockTable.id, [...blockIds])
            )
          );
        await tx
          .insert(blockTable)
          .values(
            blocks.map(block => ({
              id: block.id,
              blockPackId: input.blockPackId,
              parentBlockId: block.parentBlockId,
              prevBlockId: block.prevBlockId,
              nextBlockId: block.nextBlockId,
              type: block.type,
              props: block.props ?? {},
              content: block.content ?? [],
              updatedAt: new Date(),
              createdAt: new Date(),
            }))
          )
          .onConflictDoUpdate({
            target: blockTable.id,
            set: {
              blockPackId: sql`excluded.block_pack_id`,
              parentBlockId: sql`excluded.parent_block_id`,
              prevBlockId: sql`excluded.prev_block_id`,
              nextBlockId: sql`excluded.next_block_id`,
              type: sql`excluded.type`,
              props: sql`excluded.props`,
              content: sql`excluded.content`,
              updatedAt: sql`NOW()`,
            },
          });
      }

      await tx
        .update(blockPackYjsDocumentTable)
        .set({
          projectedUntilSequence: input.projectedSequence,
          updatedAt: sql`NOW()`,
        })
        .where(eq(blockPackYjsDocumentTable.id, document.id));

      return {
        applied: true,
        projectedUntilSequence: input.projectedSequence,
      };
    });
  }

  async close(): Promise<void> {
    await this.pool.end();
  }
}
