package apiexceptions

type UserInfoException struct {
	Exception
}

func NewUserInfoException() UserInfoException {
	return UserInfoException{
		Exception: NewException("UserInfo"),
	}
}
