package apiexceptions

type APIKeyException struct {
	Exception
}

func NewAPIKeyException() APIKeyException {
	return APIKeyException{Exception: NewException("APIKey")}
}
