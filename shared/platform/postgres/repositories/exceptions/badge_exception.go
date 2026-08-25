package exceptions

type BadgeException struct {
	RepositoryException
}

func NewBadgeException() BadgeException {
	return BadgeException{RepositoryException: NewRepositoryException("Badge")}
}
