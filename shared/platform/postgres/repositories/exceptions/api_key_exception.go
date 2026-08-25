package exceptions

type APIKeyException struct {
	RepositoryException
}

func NewAPIKeyException() APIKeyException {
	return APIKeyException{RepositoryException: NewRepositoryException("APIKey")}
}
