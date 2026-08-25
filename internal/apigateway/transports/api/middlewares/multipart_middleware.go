package middlewares

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"
	constants "github.com/HiIamJeff67/notegic-backend/shared/constants"

	sharedcontexts "github.com/HiIamJeff67/notegic-backend/shared/lib/contexts"

	exceptionwriter "github.com/HiIamJeff67/notegic-backend/shared/util/exceptionwriter"
)

func MultipartMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		form, err := ctx.MultipartForm()
		if err != nil {
			exceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.New(
				"InvalidMultipartForm",
				"Multipart",
				"Bind",
				"Multipart form is missing or invalid",
				http.StatusBadRequest,
			).WithOrigin(err), ctx)
			ctx.Abort()
			return
		}

		jsonData := make(map[string]interface{})
		var fileHeaders []*multipart.FileHeader

		for key, values := range form.Value {
			if len(values) > 0 {
				valueStr := values[0]
				if intVal, err := strconv.Atoi(valueStr); err == nil {
					jsonData[key] = intVal
				} else if boolVal, err := strconv.ParseBool(valueStr); err == nil {
					jsonData[key] = boolVal
				} else {
					jsonData[key] = valueStr
				}
			}
		}

		for _, fileHeadersSlice := range form.File {
			for _, fileHeader := range fileHeadersSlice {
				if fileHeader.Size > constants.MaxNonVideoFileSize.ToInt64() {
					exceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.New(
						"FileTooLarge",
						"Multipart",
						"Bind",
						"Uploaded file is too large",
						http.StatusRequestEntityTooLarge,
					), ctx)
					ctx.Abort()
					return
				}
				fileHeaders = append(fileHeaders, fileHeader)
			}
		}

		if len(jsonData) > 0 {
			jsonBytes, _ := json.Marshal(jsonData)
			ctx.Request.Body = io.NopCloser(bytes.NewReader(jsonBytes))
		}

		if len(fileHeaders) > 0 {
			ctx.Set(sharedcontexts.ContextFieldName_FormDataFileHeaders.String(), fileHeaders)
		}

		ctx.Next()
	}
}
