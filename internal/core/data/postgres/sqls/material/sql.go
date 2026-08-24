package materialsql

import (
	_ "embed"
)

//go:embed get_my_material_and_its_parent_by_id.sql
var GetMyMaterialAndItsParentByIdSQL string

//go:embed move_my_material_by_id.sql
var MoveMyMaterialByIdSQL string

//go:embed move_my_materials_by_ids.sql
var MoveMyMaterialsByIdsSQL string
