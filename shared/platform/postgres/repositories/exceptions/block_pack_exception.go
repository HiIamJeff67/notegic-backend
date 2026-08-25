package exceptions

type BlockPackException struct {
	RepositoryException
}

func NewBlockPackException() BlockPackException {
	return BlockPackException{RepositoryException: NewRepositoryException("BlockPack")}
}
