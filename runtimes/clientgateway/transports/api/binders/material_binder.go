package binders

import (
	"io"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	capi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/materials"
	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	sexceptionwriter "github.com/HiIamJeff67/notegic-backend/shared/util/exceptionwriter"

	controllers "github.com/HiIamJeff67/notegic-backend/runtimes/clientgateway/transports/api/controllers"
)

type MaterialBinderInterface interface {
	BindGetMyMaterialById(controllerFunc controllers.Func[*capi.GetMyMaterialByIdRequestDto]) gin.HandlerFunc
	BindGetMyMaterialAndItsParentById(controllerFunc controllers.Func[*capi.GetMyMaterialAndItsParentByIdRequestDto]) gin.HandlerFunc
	BindGetMyMaterialsByParentSubShelfId(controllerFunc controllers.Func[*capi.GetMyMaterialsByParentSubShelfIdRequestDto]) gin.HandlerFunc
	BindGetAllMyMaterialsByRootShelfId(controllerFunc controllers.Func[*capi.GetAllMyMaterialsByRootShelfIdRequestDto]) gin.HandlerFunc
	BindCreateMyMaterial(controllerFunc controllers.Func[*capi.CreateMyMaterialRequestDto]) gin.HandlerFunc
	BindUpdateMyMaterialById(controllerFunc controllers.Func[*capi.UpdateMyMaterialByIdRequestDto]) gin.HandlerFunc
	BindSaveMyMaterialById(controllerFunc controllers.Func[*capi.SaveMyMaterialByIdRequestDto]) gin.HandlerFunc
	BindMoveMyMaterialById(controllerFunc controllers.Func[*capi.MoveMyMaterialByIdRequestDto]) gin.HandlerFunc
	BindMoveMyMaterialsByIds(controllerFunc controllers.Func[*capi.MoveMyMaterialsByIdsRequestDto]) gin.HandlerFunc
	BindRestoreMyMaterialById(controllerFunc controllers.Func[*capi.RestoreMyMaterialByIdRequestDto]) gin.HandlerFunc
	BindRestoreMyMaterialsByIds(controllerFunc controllers.Func[*capi.RestoreMyMaterialsByIdsRequestDto]) gin.HandlerFunc
	BindDeleteMyMaterialById(controllerFunc controllers.Func[*capi.DeleteMyMaterialByIdRequestDto]) gin.HandlerFunc
	BindDeleteMyMaterialsByIds(controllerFunc controllers.Func[*capi.DeleteMyMaterialsByIdsRequestDto]) gin.HandlerFunc
}

type MaterialBinder struct{}

func NewMaterialBinder() MaterialBinderInterface {
	return &MaterialBinder{}
}

func (b *MaterialBinder) BindGetMyMaterialById(controllerFunc controllers.Func[*capi.GetMyMaterialByIdRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var requestDto capi.GetMyMaterialByIdRequestDto

		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")

		valueString := ctx.Query("isDeleted")
		if valueString != "" {
			value, err := strconv.ParseBool(valueString)
			if err != nil {
				sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidInput("Material").WithOrigin(err), ctx)
				return
			}
			requestDto.Param.IsDeleted = &value
		}

		value, err := uuid.Parse(ctx.Param("material-id"))
		if err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidInput("Material").WithOrigin(err), ctx)
			return
		}
		requestDto.Param.MaterialId = value

		controllerFunc(ctx, &requestDto)
	}
}

func (b *MaterialBinder) BindGetMyMaterialAndItsParentById(controllerFunc controllers.Func[*capi.GetMyMaterialAndItsParentByIdRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var requestDto capi.GetMyMaterialAndItsParentByIdRequestDto

		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")

		valueString := ctx.Query("isDeleted")
		if valueString != "" {
			value, err := strconv.ParseBool(valueString)
			if err != nil {
				sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidInput("Material").WithOrigin(err), ctx)
				return
			}
			requestDto.Param.IsDeleted = &value
		}

		value, err := uuid.Parse(ctx.Param("material-id"))
		if err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidInput("Material").WithOrigin(err), ctx)
			return
		}
		requestDto.Param.MaterialId = value

		controllerFunc(ctx, &requestDto)
	}
}

func (b *MaterialBinder) BindGetMyMaterialsByParentSubShelfId(controllerFunc controllers.Func[*capi.GetMyMaterialsByParentSubShelfIdRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var requestDto capi.GetMyMaterialsByParentSubShelfIdRequestDto

		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")

		valueString := ctx.Query("areDeleted")
		if valueString != "" {
			value, err := strconv.ParseBool(valueString)
			if err != nil {
				sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidInput("Material").WithOrigin(err), ctx)
				return
			}
			requestDto.Param.AreDeleted = &value
		}

		value, err := uuid.Parse(ctx.Param("parent-sub-shelf-id"))
		if err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidInput("Material").WithOrigin(err), ctx)
			return
		}
		requestDto.Param.ParentSubShelfId = value

		controllerFunc(ctx, &requestDto)
	}
}

func (b *MaterialBinder) BindGetAllMyMaterialsByRootShelfId(controllerFunc controllers.Func[*capi.GetAllMyMaterialsByRootShelfIdRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var requestDto capi.GetAllMyMaterialsByRootShelfIdRequestDto

		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")

		valueString := ctx.Query("areDeleted")
		if valueString != "" {
			value, err := strconv.ParseBool(valueString)
			if err != nil {
				sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidInput("Material").WithOrigin(err), ctx)
				return
			}
			requestDto.Param.AreDeleted = &value
		}

		value, err := uuid.Parse(ctx.Param("root-shelf-id"))
		if err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidInput("Material").WithOrigin(err), ctx)
			return
		}
		requestDto.Param.RootShelfId = value

		controllerFunc(ctx, &requestDto)
	}
}

func (b *MaterialBinder) BindCreateMyMaterial(controllerFunc controllers.Func[*capi.CreateMyMaterialRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var requestDto capi.CreateMyMaterialRequestDto

		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")

		if err := ctx.ShouldBindJSON(&requestDto.Body); err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidDto("Material").WithOrigin(err), ctx)
			return
		}

		value, err := uuid.Parse(ctx.Param("parent-sub-shelf-id"))
		if err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidInput("Material").WithOrigin(err), ctx)
			return
		}
		requestDto.Body.ParentSubShelfId = value

		controllerFunc(ctx, &requestDto)
	}
}

func (b *MaterialBinder) BindUpdateMyMaterialById(controllerFunc controllers.Func[*capi.UpdateMyMaterialByIdRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var requestDto capi.UpdateMyMaterialByIdRequestDto

		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")

		if err := ctx.ShouldBindJSON(&requestDto.Body); err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidDto("Material").WithOrigin(err), ctx)
			return
		}

		value, err := uuid.Parse(ctx.Param("material-id"))
		if err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidInput("Material").WithOrigin(err), ctx)
			return
		}
		requestDto.Param.MaterialId = value

		controllerFunc(ctx, &requestDto)
	}
}

func (b *MaterialBinder) BindSaveMyMaterialById(controllerFunc controllers.Func[*capi.SaveMyMaterialByIdRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var requestDto capi.SaveMyMaterialByIdRequestDto

		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")

		fileHeader, err := ctx.FormFile("contentFile")
		if err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidDto("Material").WithOrigin(err), ctx)
			return
		}
		file, err := fileHeader.Open()
		if err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidDto("Material").WithOrigin(err), ctx)
			return
		}
		defer file.Close()

		contentFile, err := io.ReadAll(file)
		if err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidDto("Material").WithOrigin(err), ctx)
			return
		}
		requestDto.Body.ContentFile = contentFile

		value, err := uuid.Parse(ctx.Param("material-id"))
		if err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidInput("Material").WithOrigin(err), ctx)
			return
		}
		requestDto.Param.MaterialId = value

		controllerFunc(ctx, &requestDto)
	}
}

func (b *MaterialBinder) BindMoveMyMaterialById(controllerFunc controllers.Func[*capi.MoveMyMaterialByIdRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var requestDto capi.MoveMyMaterialByIdRequestDto

		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")

		if err := ctx.ShouldBindJSON(&requestDto.Body); err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidDto("Material").WithOrigin(err), ctx)
			return
		}

		controllerFunc(ctx, &requestDto)
	}
}

func (b *MaterialBinder) BindMoveMyMaterialsByIds(controllerFunc controllers.Func[*capi.MoveMyMaterialsByIdsRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var requestDto capi.MoveMyMaterialsByIdsRequestDto

		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")

		if err := ctx.ShouldBindJSON(&requestDto.Body); err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidDto("Material").WithOrigin(err), ctx)
			return
		}

		controllerFunc(ctx, &requestDto)
	}
}

func (b *MaterialBinder) BindRestoreMyMaterialById(controllerFunc controllers.Func[*capi.RestoreMyMaterialByIdRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var requestDto capi.RestoreMyMaterialByIdRequestDto

		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")

		value, err := uuid.Parse(ctx.Param("material-id"))
		if err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidInput("Material").WithOrigin(err), ctx)
			return
		}
		requestDto.Param.MaterialId = value

		controllerFunc(ctx, &requestDto)
	}
}

func (b *MaterialBinder) BindRestoreMyMaterialsByIds(controllerFunc controllers.Func[*capi.RestoreMyMaterialsByIdsRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var requestDto capi.RestoreMyMaterialsByIdsRequestDto

		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")

		if err := ctx.ShouldBindJSON(&requestDto.Body); err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidDto("Material").WithOrigin(err), ctx)
			return
		}

		controllerFunc(ctx, &requestDto)
	}
}

func (b *MaterialBinder) BindDeleteMyMaterialById(controllerFunc controllers.Func[*capi.DeleteMyMaterialByIdRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var requestDto capi.DeleteMyMaterialByIdRequestDto

		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")

		value, err := uuid.Parse(ctx.Param("material-id"))
		if err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidInput("Material").WithOrigin(err), ctx)
			return
		}
		requestDto.Param.MaterialId = value

		controllerFunc(ctx, &requestDto)
	}
}

func (b *MaterialBinder) BindDeleteMyMaterialsByIds(controllerFunc controllers.Func[*capi.DeleteMyMaterialsByIdsRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var requestDto capi.DeleteMyMaterialsByIdsRequestDto

		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")

		if err := ctx.ShouldBindJSON(&requestDto.Body); err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidDto("Material").WithOrigin(err), ctx)
			return
		}

		controllerFunc(ctx, &requestDto)
	}
}
