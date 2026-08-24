package middlewares

import (
	"fmt"
	"slices"

	"github.com/gin-gonic/gin"

	enumcontract "github.com/HiIamJeff67/notegic-backend/contracts/types/models/enums"

	contexts "github.com/HiIamJeff67/notegic-backend/internal/apigateway/contexts"
)

var orderedAccessControlPermissions = []enumcontract.AccessControlPermission{
	enumcontract.AccessControlPermission_Read,
	enumcontract.AccessControlPermission_Write,
	enumcontract.AccessControlPermission_Admin,
	enumcontract.AccessControlPermission_Owner,
}

func AllowedPermissionsAbove(permission enumcontract.AccessControlPermission) gin.HandlerFunc {
	index := slices.Index(orderedAccessControlPermissions, permission)
	if index < 0 {
		panic(fmt.Sprintf("invalid access control permission: %s", permission))
	}

	return AllowedPermissionsWithin(orderedAccessControlPermissions[index:]...)
}

func AllowedPermissionsBelow(permission enumcontract.AccessControlPermission) gin.HandlerFunc {
	index := slices.Index(orderedAccessControlPermissions, permission)
	if index < 0 {
		panic(fmt.Sprintf("invalid access control permission: %s", permission))
	}

	return AllowedPermissionsWithin(orderedAccessControlPermissions[:index+1]...)
}

func AllowedPermissionsWithin(allowedPermissions ...enumcontract.AccessControlPermission) gin.HandlerFunc {
	if len(allowedPermissions) == 0 {
		panic("allowed permissions are required")
	}
	for _, permission := range allowedPermissions {
		if !slices.Contains(orderedAccessControlPermissions, permission) {
			panic(fmt.Sprintf("invalid access control permission: %s", permission))
		}
	}

	return func(ctx *gin.Context) {
		ctx.Request = ctx.Request.WithContext(
			contexts.WithAllowedPermissions(ctx.Request.Context(), allowedPermissions),
		)

		ctx.Next()
	}
}
