package exceptions

type MaterialException struct {
	CoreException
}

func NewMaterialException() MaterialException {
	return MaterialException{
		CoreException: CoreException{
			Domain: "Material",
		},
	}
}
