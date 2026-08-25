package exceptions

type ShelfException struct {
	RepositoryException
}

func NewShelfException() ShelfException {
	return ShelfException{RepositoryException: NewRepositoryException("Shelf")}
}
