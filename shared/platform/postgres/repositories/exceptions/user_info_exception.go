package exceptions

type UserInfoException struct {
	RepositoryException
}

func NewUserInfoException() UserInfoException {
	return UserInfoException{RepositoryException: NewRepositoryException("UserInfo")}
}
