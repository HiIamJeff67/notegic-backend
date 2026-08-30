package apiexceptions

type MaterialException struct {
	Exception
}

func NewMaterialException() MaterialException {
	return MaterialException{
		Exception: NewException("Material"),
	}
}
