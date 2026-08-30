package exceptions

type MaterialException struct {
	Exception
}

func NewMaterialException() MaterialException {
	return MaterialException{
		Exception: NewException("Material"),
	}
}
