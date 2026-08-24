package contexts

import (
	"context"
	"net/http"
	"slices"

	"github.com/google/uuid"

	exceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	sharedcontexts "github.com/HiIamJeff67/notegic-backend/shared/lib/contexts"
	sharedtokens "github.com/HiIamJeff67/notegic-backend/shared/tokens"

	enums "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/schemas/enums"
)

func WithGatewaySource(ctx context.Context, source string) context.Context {
	return sharedcontexts.WithValue(ctx, sharedcontexts.ContextFieldName_Gateway_Source, source)
}

func WithAuthMethod(ctx context.Context, method string) context.Context {
	return sharedcontexts.WithValue(ctx, sharedcontexts.ContextFieldName_Auth_Method, method)
}

func WithAPIKeyId(ctx context.Context, apiKeyId string) context.Context {
	return sharedcontexts.WithValue(ctx, sharedcontexts.ContextFieldName_API_Key_Id, apiKeyId)
}

func WithDelegationMetadata(ctx context.Context, claims *sharedtokens.DelegationTokenClaims) context.Context {
	if claims == nil {
		return ctx
	}
	source := claims.GatewaySource
	if source == "" {
		source = sharedtokens.GatewaySourceClient
	}
	method := claims.AuthMethod
	if method == "" {
		method = sharedtokens.AuthMethodJWT
	}
	ctx = WithGatewaySource(ctx, source)
	ctx = WithAuthMethod(ctx, method)
	if claims.ApiKeyId != "" {
		ctx = WithAPIKeyId(ctx, claims.ApiKeyId)
	}
	return ctx
}

func GetGatewaySource(ctx context.Context) (string, *exceptions.Exception) {
	source, err := sharedcontexts.GetValue[string](ctx, sharedcontexts.ContextFieldName_Gateway_Source)
	if err != nil || (source != sharedtokens.GatewaySourceClient && source != sharedtokens.GatewaySourceAPI) {
		return "", exceptions.New(
			"DelegationClaimsInvalid", "API", "ReadGatewaySource",
			"The verified delegation context does not contain a valid gateway source",
			http.StatusInternalServerError, true,
		).WithOrigin(err)
	}
	return source, nil
}

func IsClientGateway(ctx context.Context) bool {
	source, err := sharedcontexts.GetValue[string](ctx, sharedcontexts.ContextFieldName_Gateway_Source)
	return err == nil && source == sharedtokens.GatewaySourceClient
}

func GetAuthMethod(ctx context.Context) (string, *exceptions.Exception) {
	method, err := sharedcontexts.GetValue[string](ctx, sharedcontexts.ContextFieldName_Auth_Method)
	if err != nil || (method != sharedtokens.AuthMethodJWT && method != sharedtokens.AuthMethodAPIKey) {
		return "", exceptions.New(
			"DelegationClaimsInvalid", "API", "ReadAuthMethod",
			"The verified delegation context does not contain a valid authentication method",
			http.StatusInternalServerError, true,
		).WithOrigin(err)
	}
	return method, nil
}

func GetAPIKeyId(ctx context.Context) (string, *exceptions.Exception) {
	apiKeyId, err := sharedcontexts.GetValue[string](ctx, sharedcontexts.ContextFieldName_API_Key_Id)
	if err != nil || apiKeyId == "" {
		return "", exceptions.New(
			"DelegationClaimsInvalid", "API", "ReadAPIKeyId",
			"The verified delegation context does not contain a valid API key ID",
			http.StatusInternalServerError, true,
		).WithOrigin(err)
	}
	return apiKeyId, nil
}

func WithAllowedPermissions(
	ctx context.Context,
	allowedPermissions []enums.AccessControlPermission,
) context.Context {
	return sharedcontexts.WithValue(
		ctx,
		sharedcontexts.ContextFieldName_Allowed_Permissions,
		slices.Clone(allowedPermissions),
	)
}

func WithActorUserId(ctx context.Context, actorUserId uuid.UUID) context.Context {
	return sharedcontexts.WithValue(
		ctx,
		sharedcontexts.ContextFieldName_User_Id,
		actorUserId,
	)
}

func WithActorUserName(ctx context.Context, actorUserName string) context.Context {
	return sharedcontexts.WithValue(
		ctx,
		sharedcontexts.ContextFieldName_User_Name,
		actorUserName,
	)
}

func WithActorUserPublicId(ctx context.Context, actorUserPublicId uuid.UUID) context.Context {
	return sharedcontexts.WithValue(
		ctx,
		sharedcontexts.ContextFieldName_User_PublicId,
		actorUserPublicId,
	)
}

func GetAllowedPermissions(
	ctx context.Context,
) ([]enums.AccessControlPermission, *exceptions.Exception) {
	allowedPermissions, err := sharedcontexts.GetValue[[]enums.AccessControlPermission](
		ctx,
		sharedcontexts.ContextFieldName_Allowed_Permissions,
	)
	if err != nil {
		return nil, exceptions.New(
			"DelegationClaimsInvalid",
			"API",
			"ReadAllowedPermissions",
			"The verified delegation context does not contain valid allowed permissions",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}

	return slices.Clone(allowedPermissions), nil
}

func GetActorUserId(ctx context.Context) (uuid.UUID, *exceptions.Exception) {
	actorUserId, err := sharedcontexts.GetValue[uuid.UUID](
		ctx,
		sharedcontexts.ContextFieldName_User_Id,
	)
	if err != nil || actorUserId == uuid.Nil {
		return uuid.Nil, exceptions.New(
			"DelegationClaimsInvalid",
			"API",
			"ReadActorUserId",
			"The verified delegation context does not contain a valid actor user ID",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}

	return actorUserId, nil
}

func GetActorUserName(ctx context.Context) (string, *exceptions.Exception) {
	actorUserName, err := sharedcontexts.GetValue[string](
		ctx,
		sharedcontexts.ContextFieldName_User_Name,
	)
	if err != nil || actorUserName == "" {
		return "", exceptions.New(
			"DelegationClaimsInvalid",
			"API",
			"ReadActorUserName",
			"The verified delegation context does not contain a valid actor user name",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}

	return actorUserName, nil
}

func GetActorUserPublicId(ctx context.Context) (uuid.UUID, *exceptions.Exception) {
	actorUserPublicId, err := sharedcontexts.GetValue[uuid.UUID](
		ctx,
		sharedcontexts.ContextFieldName_User_PublicId,
	)
	if err != nil || actorUserPublicId == uuid.Nil {
		return uuid.Nil, exceptions.New(
			"DelegationClaimsInvalid",
			"API",
			"ReadActorUserPublicId",
			"The verified delegation context does not contain a valid actor public ID",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}

	return actorUserPublicId, nil
}
