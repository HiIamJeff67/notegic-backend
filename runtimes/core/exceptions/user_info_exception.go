package exceptions

type UserInfoException struct {
	CoreException
}

func NewUserInfoException() UserInfoException {
	return UserInfoException{
		CoreException: CoreException{
			Domain: "UserInfo",
		},
	}
}
