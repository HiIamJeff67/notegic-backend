package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	cgateway "github.com/HiIamJeff67/notegic-backend/contracts/gateway/v1"
)

func writeClientResponse[D any](ctx *gin.Context, data D) {
	ctx.JSON(http.StatusOK, cgateway.ClientResponse[D]{
		Success: true,
		Data:    data,
	})
}

func writeCreatedClientResponse[D any](ctx *gin.Context, data D) {
	ctx.JSON(http.StatusCreated, cgateway.ClientResponse[D]{
		Success: true,
		Data:    data,
	})
}
