package exceptions

type UserSettingException struct {
	RepositoryException
}

func NewUserSettingException() UserSettingException {
	return UserSettingException{RepositoryException: NewRepositoryException("UserSetting")}
}
