package exceptions

type UsersToBillingPlansException struct {
	RepositoryException
}

func NewUsersToBillingPlansException() UsersToBillingPlansException {
	return UsersToBillingPlansException{RepositoryException: NewRepositoryException("UsersToBillingPlans")}
}
