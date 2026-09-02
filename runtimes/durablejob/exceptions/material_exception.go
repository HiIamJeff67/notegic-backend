package exceptions

type MaterialException struct {
	DurableJobException
}

func NewMaterialException() MaterialException {
	return MaterialException{
		DurableJobException: DurableJobException{
			Domain: "Material",
		},
	}
}
