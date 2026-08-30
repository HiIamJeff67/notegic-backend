package exceptions

type EmailException struct {
	Domain string
}

func NewEmailException(domain string) EmailException {
	return EmailException{Domain: domain}
}
