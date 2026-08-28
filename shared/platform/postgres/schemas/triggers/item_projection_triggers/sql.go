package itemprojectiontriggersql

import (
	_ "embed"
)

const (
	ProjectSubShelvesToItemsTriggerFunctionName        = "trigger_function_project_sub_shelves_to_items"
	ProjectMaterialsToItemsTriggerFunctionName         = "trigger_function_project_materials_to_items_after_insert_or_update"
	DeleteMaterialItemsAfterDeleteTriggerFunctionName  = "trigger_function_delete_material_items_after_delete"
	ProjectBlockPacksToItemsTriggerFunctionName        = "trigger_function_project_block_packs_to_items_after_insert_or_update"
	DeleteBlockPackItemsAfterDeleteTriggerFunctionName = "trigger_function_delete_block_pack_items_after_delete"
)

var (
	//go:embed project_sub_shelves_to_items_trigger.sql
	ProjectSubShelvesToItemsTriggerSQL string

	//go:embed project_materials_to_items_trigger.sql
	ProjectMaterialsToItemsTriggerSQL string

	//go:embed project_block_packs_to_items_trigger.sql
	ProjectBlockPacksToItemsTriggerSQL string
)
