import {
  bigint,
  customType,
  jsonb,
  pgTable,
  text,
  timestamp,
  uuid,
} from "drizzle-orm/pg-core";

const bytea = customType<{ data: Buffer }>({
  dataType: () => "bytea",
});

export const blockPackTable = pgTable("BlockPackTable", {
  id: uuid("id").notNull(),
  deletedAt: timestamp("deleted_at", {
    withTimezone: true,
    mode: "date",
  }),
});

export const blockPackYjsDocumentTable = pgTable("BlockPackYjsDocumentTable", {
  id: uuid("id").notNull(),
  blockPackId: uuid("block_pack_id").notNull(),
  snapshot: bytea("snapshot").notNull(),
  stateVector: bytea("state_vector").notNull(),
  lastUpdateSequence: bigint("last_update_sequence", {
    mode: "number",
  }).notNull(),
  compactedUntilSequence: bigint("compacted_until_sequence", {
    mode: "number",
  }).notNull(),
  lastCompactedAt: timestamp("last_compacted_at", {
    withTimezone: true,
    mode: "date",
  }),
  projectedUntilSequence: bigint("projected_until_sequence", {
    mode: "number",
  }).notNull(),
  deletedAt: timestamp("deleted_at", {
    withTimezone: true,
    mode: "date",
  }),
  updatedAt: timestamp("updated_at", {
    withTimezone: true,
    mode: "date",
  }).notNull(),
  createdAt: timestamp("created_at", {
    withTimezone: true,
    mode: "date",
  }).notNull(),
});

export const blockPackYjsUpdateTable = pgTable("BlockPackYjsUpdateTable", {
  id: uuid("id").notNull(),
  blockPackId: uuid("block_pack_id").notNull(),
  updateSequence: bigint("update_sequence", { mode: "number" }).notNull(),
  payload: bytea("payload").notNull(),
  compactedAt: timestamp("compacted_at", {
    withTimezone: true,
    mode: "date",
  }),
  createdAt: timestamp("created_at", {
    withTimezone: true,
    mode: "date",
  }).notNull(),
});

export const blockTable = pgTable("BlockTable", {
  id: uuid("id").notNull(),
  blockPackId: uuid("block_pack_id").notNull(),
  parentBlockId: uuid("parent_block_id"),
  prevBlockId: uuid("prev_block_id"),
  nextBlockId: uuid("next_block_id"),
  type: text("type").notNull(),
  props: jsonb("props").$type<unknown>().notNull(),
  content: jsonb("content").$type<unknown>().notNull(),
  updatedAt: timestamp("updated_at", {
    withTimezone: true,
    mode: "date",
  }).notNull(),
  createdAt: timestamp("created_at", {
    withTimezone: true,
    mode: "date",
  }).notNull(),
});

export const schema = {
  blockPackTable,
  blockPackYjsDocumentTable,
  blockPackYjsUpdateTable,
  blockTable,
};
