package exceptions

type EmailException struct {
	Domain string
}

func NewEmailException() EmailException {
	return EmailException{
		Domain: "Email",
	}
}
