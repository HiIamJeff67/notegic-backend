package controllers

import "github.com/gin-gonic/gin"

type Func[RequestDtoType any] func(ctx *gin.Context, requestDto RequestDtoType)
