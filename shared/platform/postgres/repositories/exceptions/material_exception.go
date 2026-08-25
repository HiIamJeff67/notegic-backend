package exceptions

type MaterialException struct {
	RepositoryException
}

func NewMaterialException() MaterialException {
	return MaterialException{RepositoryException: NewRepositoryException("Material")}
}
