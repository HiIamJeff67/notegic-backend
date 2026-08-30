package contexts

import (
	"context"
	"net/http"
	"slices"

	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	sharedcontexts "github.com/HiIamJeff67/notegic-backend/shared/lib/contexts"
)

func WithAllowedPermissions(
	ctx context.Context,
	allowedPermissions []cenums.AccessControlPermission,
) context.Context {
	return sharedcontexts.WithValue(
		ctx,
		sharedcontexts.ContextFieldName_Allowed_Permissions,
		slices.Clone(allowedPermissions),
	)
}

func GetAllowedPermissions(
	ctx context.Context,
) ([]cenums.AccessControlPermission, *cexceptions.Exception) {
	allowedPermissions, err := sharedcontexts.GetValue[[]cenums.AccessControlPermission](
		ctx,
		sharedcontexts.ContextFieldName_Allowed_Permissions,
	)
	if err != nil {
		return nil, cexceptions.New(
			"ContextFieldInvalid",
			"Gateway",
			"ReadAllowedPermissions",
			"The request context does not contain valid allowed permissions",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}

	return slices.Clone(allowedPermissions), nil
}

func GetOptionalAllowedPermissions(
	ctx context.Context,
) ([]cenums.AccessControlPermission, *cexceptions.Exception) {
	if ctx.Value(sharedcontexts.ContextFieldName_Allowed_Permissions) == nil {
		return nil, nil
	}

	return GetAllowedPermissions(ctx)
}
