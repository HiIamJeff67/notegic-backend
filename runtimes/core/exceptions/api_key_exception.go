package exceptions

type APIKeyException struct {
	CoreException
}

func NewAPIKeyException() APIKeyException {
	return APIKeyException{
		CoreException: CoreException{
			Domain: "APIKey",
		},
	}
}
