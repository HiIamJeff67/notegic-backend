ALTER TABLE "BlockTable" DROP CONSTRAINT IF EXISTS block_unique_sibling_prev;
ALTER TABLE "BlockTable" DROP CONSTRAINT IF EXISTS block_unique_sibling_next;

-- ============================== SQL Separator ==============================

-- BlockTable is a full Yjs projection. Pointer changes are valid only as a
-- complete document state, so validate sibling uniqueness at transaction commit.
ALTER TABLE "BlockTable"
ADD CONSTRAINT block_unique_sibling_prev
UNIQUE NULLS NOT DISTINCT (block_pack_id, parent_block_id, prev_block_id)
DEFERRABLE INITIALLY DEFERRED;

-- ============================== SQL Separator ==============================

ALTER TABLE "BlockTable"
ADD CONSTRAINT block_unique_sibling_next
UNIQUE NULLS NOT DISTINCT (block_pack_id, parent_block_id, next_block_id)
DEFERRABLE INITIALLY DEFERRED;
