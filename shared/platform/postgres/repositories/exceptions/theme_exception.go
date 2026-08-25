package exceptions

type ThemeException struct {
	RepositoryException
}

func NewThemeException() ThemeException {
	return ThemeException{RepositoryException: NewRepositoryException("Theme")}
}
