package exceptions

type ItemException struct {
	RepositoryException
}

func NewItemException() ItemException {
	return ItemException{RepositoryException: NewRepositoryException("Item")}
}
