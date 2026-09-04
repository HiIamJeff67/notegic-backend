package apicontract

const (
	GetMyBlockPackByIdOperation                  = "block-pack.get-by-id"
	GetMyBlockPackAndItsParentByIdOperation      = "block-pack.get-and-parent-by-id"
	GetMyBlockPacksByParentSubShelfIdOperation   = "block-pack.get-by-parent-sub-shelf-id"
	GetMyBlockPacksByRootShelfIdOperation        = "block-pack.get-by-root-shelf-id"
	CreateBlockPackOperation                     = "block-pack.create"
	CreateBlockPacksOperation                    = "block-pack.create-many"
	UpdateMyBlockPackByIdOperation               = "block-pack.update"
	UpdateMyBlockPacksByIdsOperation             = "block-pack.update-many"
	MoveMyBlockPackByParentSubShelfIdOperation   = "block-pack.move"
	MoveMyBlockPacksByParentSubShelfIdOperation  = "block-pack.move-many"
	MoveMyBlockPacksByParentSubShelfIdsOperation = "block-pack.move-many-by-parent-sub-shelves"
	RestoreMyBlockPackByIdOperation              = "block-pack.restore"
	RestoreMyBlockPacksByIdsOperation            = "block-pack.restore-many"
	DeleteMyBlockPackByIdOperation               = "block-pack.delete"
	DeleteMyBlockPacksByIdsOperation             = "block-pack.delete-many"
	SearchBlockPacksOperation                    = "graphql.search-block-packs"
)
