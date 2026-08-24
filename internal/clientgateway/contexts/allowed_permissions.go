package contexts

import (
	"context"
	"net/http"
	"slices"

	exceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	sharedcontexts "github.com/HiIamJeff67/notegic-backend/shared/lib/contexts"

	enumcontract "github.com/HiIamJeff67/notegic-backend/contracts/types/models/enums"
)

func WithAllowedPermissions(
	ctx context.Context,
	allowedPermissions []enumcontract.AccessControlPermission,
) context.Context {
	return sharedcontexts.WithValue(
		ctx,
		sharedcontexts.ContextFieldName_Allowed_Permissions,
		slices.Clone(allowedPermissions),
	)
}

func GetAllowedPermissions(
	ctx context.Context,
) ([]enumcontract.AccessControlPermission, *exceptions.Exception) {
	allowedPermissions, err := sharedcontexts.GetValue[[]enumcontract.AccessControlPermission](
		ctx,
		sharedcontexts.ContextFieldName_Allowed_Permissions,
	)
	if err != nil {
		return nil, exceptions.New(
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
) ([]enumcontract.AccessControlPermission, *exceptions.Exception) {
	if ctx.Value(sharedcontexts.ContextFieldName_Allowed_Permissions) == nil {
		return nil, nil
	}

	return GetAllowedPermissions(ctx)
}
