package middlewares

import (
	"fmt"
	"slices"

	"github.com/gin-gonic/gin"

	enums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"

	contexts "github.com/HiIamJeff67/notegic-backend/internal/clientgateway/contexts"
)

var orderedAccessControlPermissions = []enums.AccessControlPermission{
	enums.AccessControlPermission_Read,
	enums.AccessControlPermission_Write,
	enums.AccessControlPermission_Admin,
	enums.AccessControlPermission_Owner,
}

func AllowedPermissionsAbove(permission enums.AccessControlPermission) gin.HandlerFunc {
	index := slices.Index(orderedAccessControlPermissions, permission)
	if index < 0 {
		panic(fmt.Sprintf("invalid access control permission: %s", permission))
	}

	return AllowedPermissionsWithin(orderedAccessControlPermissions[index:]...)
}

func AllowedPermissionsBelow(permission enums.AccessControlPermission) gin.HandlerFunc {
	index := slices.Index(orderedAccessControlPermissions, permission)
	if index < 0 {
		panic(fmt.Sprintf("invalid access control permission: %s", permission))
	}

	return AllowedPermissionsWithin(orderedAccessControlPermissions[:index+1]...)
}

func AllowedPermissionsWithin(allowedPermissions ...enums.AccessControlPermission) gin.HandlerFunc {
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
