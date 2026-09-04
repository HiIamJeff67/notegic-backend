package apicontract

const (
	GetMyMaterialByIdOperation                = "material.get-by-id"
	GetMyMaterialAndItsParentByIdOperation    = "material.get-and-parent-by-id"
	GetMyMaterialsByParentSubShelfIdOperation = "material.get-by-parent-sub-shelf-id"
	GetMyMaterialsByRootShelfIdOperation      = "material.get-by-root-shelf-id"
	CreateMyMaterialOperation                 = "material.create"
	UpdateMyMaterialByIdOperation             = "material.update"
	SaveMyMaterialByIdOperation               = "material.save"
	MoveMyMaterialByIdOperation               = "material.move"
	MoveMyMaterialsByIdsOperation             = "material.move-many"
	RestoreMyMaterialByIdOperation            = "material.restore"
	RestoreMyMaterialsByIdsOperation          = "material.restore-many"
	DeleteMyMaterialByIdOperation             = "material.delete"
	DeleteMyMaterialsByIdsOperation           = "material.delete-many"
	SearchMaterialsOperation                  = "graphql.search-materials"
)
