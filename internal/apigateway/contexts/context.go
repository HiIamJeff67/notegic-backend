package contexts

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	sharedcontexts "github.com/HiIamJeff67/notegic-backend/shared/lib/contexts"
)

func SetGatewaySource(ctx *gin.Context, source string) {
	ctx.Set(sharedcontexts.ContextFieldName_Gateway_Source.String(), source)
}

func GetOptionalGatewaySource(ctx *gin.Context) string {
	value, exists := ctx.Get(sharedcontexts.ContextFieldName_Gateway_Source.String())
	if !exists {
		return ""
	}
	source, _ := value.(string)
	return source
}

func GetAndConvertContextFieldToBoolean(ctx *gin.Context, name sharedcontexts.ContextFieldName) (*bool, *cexceptions.Exception) {
	value, exists := ctx.Get(name.String())
	if !exists {
		return nil, cexceptions.New(
			"ContextFieldMissing",
			"Gateway",
			"ReadContextField",
			"The required request context field is missing",
			http.StatusInternalServerError,
			true,
		)
	}

	valueBoolean, ok := value.(bool)
	if !ok {
		return nil, cexceptions.New(
			"ContextFieldInvalid",
			"Gateway",
			"ReadContextField",
			"The request context field has an invalid type",
			http.StatusInternalServerError,
			true,
		)
	}

	return &valueBoolean, nil
}

func GetAndConvertContextFieldToString(ctx *gin.Context, name sharedcontexts.ContextFieldName) (*string, *cexceptions.Exception) {
	value, exists := ctx.Get(name.String())
	if !exists {
		return nil, cexceptions.New(
			"ContextFieldMissing",
			"Gateway",
			"ReadContextField",
			"The required request context field is missing",
			http.StatusInternalServerError,
			true,
		)
	}

	valueString, ok := value.(string)
	if !ok {
		return nil, cexceptions.New(
			"ContextFieldInvalid",
			"Gateway",
			"ReadContextField",
			"The request context field has an invalid type",
			http.StatusInternalServerError,
			true,
		)
	}

	return &valueString, nil
}

func GetAndConvertContextFieldToUUID(ctx *gin.Context, name sharedcontexts.ContextFieldName) (*uuid.UUID, *cexceptions.Exception) {
	value, exists := ctx.Get(name.String())
	if !exists {
		return nil, cexceptions.New(
			"ContextFieldMissing",
			"Gateway",
			"ReadContextField",
			"The required request context field is missing",
			http.StatusInternalServerError,
			true,
		)
	}

	if valueUUID, ok := value.(uuid.UUID); ok {
		return &valueUUID, nil
	}

	valueString, ok := value.(string)
	if !ok {
		return nil, cexceptions.New(
			"ContextFieldInvalid",
			"Gateway",
			"ReadContextField",
			"The request context field has an invalid type",
			http.StatusInternalServerError,
			true,
		)
	}

	id, err := uuid.Parse(valueString)
	if err != nil {
		return nil, cexceptions.New(
			"InvalidInput",
			"Gateway",
			"ReadContextField",
			"The request context field contains an invalid UUID",
			http.StatusBadRequest,
		).WithOrigin(err)
	}

	return &id, nil
}

func GetAndConvertContextToGinContext(ctx context.Context) (*gin.Context, *cexceptions.Exception) {
	ginCtx, ok := ctx.Value(sharedcontexts.ContextFieldName_GinContext).(*gin.Context)
	if !ok {
		return nil, cexceptions.New(
			"ContextFieldInvalid",
			"Gateway",
			"ReadGinContext",
			"The request context does not contain a Gin context",
			http.StatusInternalServerError,
			true,
		)
	}

	return ginCtx, nil
}
