package exceptions

type BlockException struct {
	RepositoryException
}

func NewBlockException() BlockException {
	return BlockException{RepositoryException: NewRepositoryException("Block")}
}
