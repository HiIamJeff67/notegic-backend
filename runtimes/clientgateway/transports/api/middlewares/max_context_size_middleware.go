package middlewares

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	stypes "github.com/HiIamJeff67/notegic-backend/shared/types"
	sexceptionwriter "github.com/HiIamJeff67/notegic-backend/shared/util/exceptionwriter"
)

func MaxContextSizeMiddleware(limitBytes int64, unit stypes.ByteType) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if ctx.Request.ContentLength > limitBytes*int64(unit) {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.New(
				"MaxContextBodySizeExceeded",
				"Context",
				"Validate",
				fmt.Sprintf("The request body size of %d bytes exceeds the maximum of %d bytes", ctx.Request.ContentLength, limitBytes*unit.ToInt64()),
				http.StatusRequestEntityTooLarge,
			), ctx)
			return
		}

		ctx.Request.Body = http.MaxBytesReader(ctx.Writer, ctx.Request.Body, limitBytes*int64(unit))
		ctx.Next()
	}
}
