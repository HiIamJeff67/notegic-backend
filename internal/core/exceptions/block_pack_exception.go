package apiexceptions

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"

	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"
)

type BlockPackException struct {
	CoreException
}

func NewBlockPackException() BlockPackException {
	return BlockPackException{
		CoreException: NewCoreException("BlockPack"),
	}
}

func (BlockPackException) NoRootBlockInBlockPack(blockPackId uuid.UUID) *cexceptions.Exception {
	return cexceptions.New(
		"NoRootBlockInBlockPack",
		"BlockPack",
		"Project",
		fmt.Sprintf("No root block exists in block pack %s", blockPackId),
		http.StatusInternalServerError,
		true,
	)
}
