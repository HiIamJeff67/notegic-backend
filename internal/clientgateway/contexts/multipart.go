package contexts

import (
	"mime/multipart"
	"net/http"

	"github.com/gin-gonic/gin"

	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	sharedcontexts "github.com/HiIamJeff67/notegic-backend/shared/lib/contexts"
)

func GetAndConvertContextToMultipartFileHeaders(ctx *gin.Context) ([]*multipart.FileHeader, *cexceptions.Exception) {
	value, exists := ctx.Get(sharedcontexts.ContextFieldName_FormDataFileHeaders.String())
	if !exists {
		return nil, cexceptions.New(
			"ContextFieldMissing",
			"Gateway",
			"ReadFormData",
			"The request context does not contain multipart file headers",
			http.StatusInternalServerError,
			true,
		)
	}

	fileHeaders, ok := value.([]*multipart.FileHeader)
	if !ok {
		return nil, cexceptions.New(
			"ContextFieldInvalid",
			"Gateway",
			"ReadFormData",
			"The multipart file headers have an invalid type",
			http.StatusInternalServerError,
			true,
		)
	}

	return fileHeaders, nil
}
