package blockpacksql

import (
	_ "embed"
)

//go:embed get_my_block_pack_and_its_parent_by_id.sql
var GetMyBlockPackAndItsParentByIdSQL string

//go:embed move_my_block_pack_by_id.sql
var MoveMyBlockPackByIdSQL string

//go:embed move_my_block_packs_by_ids.sql
var MoveMyBlockPacksByParentSubShelfIdSQL string
