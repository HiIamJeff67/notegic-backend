package itemprojectiontriggersql

import (
	_ "embed"
)

var (
	//go:embed project_sub_shelves_to_items_trigger.sql
	ProjectSubShelvesToItemsTriggerSQL string

	//go:embed project_materials_to_items_trigger.sql
	ProjectMaterialsToItemsTriggerSQL string

	//go:embed project_block_packs_to_items_trigger.sql
	ProjectBlockPacksToItemsTriggerSQL string
)
