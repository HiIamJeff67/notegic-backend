# Beta Yjs Data Reset

Beta does not migrate legacy `BlockTable` content into an active Yjs document. The old linked-list state is discarded and a reset creates an empty database with the current Yjs schema.

`BlockTable` is a projection read model. It must never be used to rehydrate an active `Y.Doc` after this cutover.

## Reset

Run this only against the beta/development database. It drops the PostgreSQL `public` schema and all data in it.

```bash
make -C runtimes/core remigrate-db
make -C runtimes/core seed
```

After reset, create BlockPacks through the normal BlockPack APIs or durable system workflow. Both paths create the `BlockPack` and its empty `BlockPackYjsDocument` in the same transaction. The first BlockPack channel subscription loads that empty Yjs document; editor changes then produce the update log and asynchronous `BlockTable` projection.

## Verify

Every BlockPack must have exactly one Yjs document:

```sql
SELECT bp.id
FROM "BlockPackTable" AS bp
LEFT JOIN "BlockPackYjsDocumentTable" AS ydoc
  ON ydoc.block_pack_id = bp.id
WHERE ydoc.id IS NULL;
```

The result must be empty. For a specific active BlockPack, inspect persistence and projection progress with:

```sql
SELECT
  ydoc.last_update_sequence,
  ydoc.compacted_until_sequence,
  ydoc.projected_until_sequence,
  COUNT(block.id) AS projected_block_count
FROM "BlockPackYjsDocumentTable" AS ydoc
INNER JOIN "BlockPackTable" AS bp ON bp.id = ydoc.block_pack_id
LEFT JOIN "BlockTable" AS block ON block.block_pack_id = bp.id
WHERE bp.id = '<BLOCK_PACK_ID>'
  AND bp.deleted_at IS NULL
GROUP BY
  ydoc.last_update_sequence,
  ydoc.compacted_until_sequence,
  ydoc.projected_until_sequence;
```

An empty Yjs document legitimately has zero projected blocks. Once updates are persisted, `projected_until_sequence` eventually catches up to `last_update_sequence`; this is intentionally asynchronous.

## Recovery Boundary

Do not reconstruct a Yjs document from `BlockTable`. If beta data must be discarded, repeat the reset. Runtime recovery uses `BlockPackYjsDocument.snapshot` plus the update log after `compacted_until_sequence`, not projected Blocks.
