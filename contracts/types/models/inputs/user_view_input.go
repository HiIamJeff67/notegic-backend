package inputs

import "github.com/google/uuid"

type GetUserViewByPublicIdInput struct {
	PublicId uuid.UUID
}
