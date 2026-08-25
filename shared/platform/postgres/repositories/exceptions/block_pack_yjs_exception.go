package exceptions

type BlockPackYjsException struct {
	RepositoryException
}

func NewBlockPackYjsException() BlockPackYjsException {
	return BlockPackYjsException{RepositoryException: NewRepositoryException("BlockPackYjs")}
}
