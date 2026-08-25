package exceptions

type MaterialException struct {
	CoreException
}

func NewMaterialException() MaterialException {
	return MaterialException{
		CoreException: NewCoreException("Material"),
	}
}
