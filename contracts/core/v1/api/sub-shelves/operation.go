package apicontract

const (
	GetMySubShelfByIdOperation                       = "sub-shelf.get-by-id"
	GetMySubShelvesByPrevSubShelfIdOperation         = "sub-shelf.get-by-prev-sub-shelf-id"
	GetMySubShelvesByRootShelfIdOperation            = "sub-shelf.get-by-root-shelf-id"
	GetMySubShelvesAndItemsByPrevSubShelfIdOperation = "sub-shelf.get-and-items-by-prev-sub-shelf-id"
	CreateSubShelfByRootShelfIdOperation             = "sub-shelf.create"
	CreateSubShelvesByRootShelfIdsOperation          = "sub-shelf.create-many"
	UpdateMySubShelfByIdOperation                    = "sub-shelf.update"
	UpdateMySubShelvesByIdsOperation                 = "sub-shelf.update-many"
	MoveMySubShelfByRootShelfIdOperation             = "sub-shelf.move"
	MoveMySubShelvesByRootShelfIdOperation           = "sub-shelf.move-many"
	MoveMySubShelvesByRootShelfIdsOperation          = "sub-shelf.move-many-by-root-shelves"
	RestoreMySubShelfByIdOperation                   = "sub-shelf.restore"
	RestoreMySubShelvesByIdsOperation                = "sub-shelf.restore-many"
	DeleteMySubShelfByIdOperation                    = "sub-shelf.delete"
	DeleteMySubShelvesByIdsOperation                 = "sub-shelf.delete-many"
	SearchSubShelvesOperation                        = "graphql.search-sub-shelves"
)
