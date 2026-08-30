package middlewares

import (
	"fmt"
	"slices"

	"github.com/gin-gonic/gin"

	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"

	contexts "github.com/HiIamJeff67/notegic-backend/runtimes/apigateway/contexts"
)

var orderedAccessControlPermissions = []cenums.AccessControlPermission{
	cenums.AccessControlPermission_Read,
	cenums.AccessControlPermission_Write,
	cenums.AccessControlPermission_Admin,
	cenums.AccessControlPermission_Owner,
}

func AllowedPermissionsAbove(permission cenums.AccessControlPermission) gin.HandlerFunc {
	index := slices.Index(orderedAccessControlPermissions, permission)
	if index < 0 {
		panic(fmt.Sprintf("invalid access control permission: %s", permission))
	}

	return AllowedPermissionsWithin(orderedAccessControlPermissions[index:]...)
}

func AllowedPermissionsBelow(permission cenums.AccessControlPermission) gin.HandlerFunc {
	index := slices.Index(orderedAccessControlPermissions, permission)
	if index < 0 {
		panic(fmt.Sprintf("invalid access control permission: %s", permission))
	}

	return AllowedPermissionsWithin(orderedAccessControlPermissions[:index+1]...)
}

func AllowedPermissionsWithin(allowedPermissions ...cenums.AccessControlPermission) gin.HandlerFunc {
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
