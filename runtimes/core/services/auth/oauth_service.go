package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"golang.org/x/oauth2"

	cauthdto "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/auth"
	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"
)

type OAuthServiceInterface interface {
	GetGoogleUserInfo(
		ctx context.Context,
		request *cauthdto.GetGoogleUserInfoRequestDto,
	) (*cauthdto.GetGoogleUserInfoResponseDto, *cexceptions.Exception)
}

type OAuthService struct {
	oauthGoogleConfig *oauth2.Config
}

func NewOAuthService(oauthGoogleConfig *oauth2.Config) OAuthServiceInterface {
	return &OAuthService{
		oauthGoogleConfig: oauthGoogleConfig,
	}
}

func (s *OAuthService) GetGoogleUserInfo(
	ctx context.Context,
	request *cauthdto.GetGoogleUserInfoRequestDto,
) (*cauthdto.GetGoogleUserInfoResponseDto, *cexceptions.Exception) {
	token, err := s.oauthGoogleConfig.Exchange(ctx, request.AuthenticationCode)
	if err != nil {
		var retrieveError *oauth2.RetrieveError
		if errors.As(err, &retrieveError) && retrieveError.Response != nil &&
			retrieveError.Response.StatusCode >= http.StatusBadRequest &&
			retrieveError.Response.StatusCode < http.StatusInternalServerError {
			return nil, cexceptions.New(
				"InvalidAuthenticationCode",
				"OAuth",
				"GetGoogleUserInfo",
				"Authentication code is invalid or expired",
				http.StatusBadRequest,
			).WithOrigin(err)
		}

		return nil, cexceptions.New(
			"TokenExchangeFailed",
			"OAuth",
			"GetGoogleUserInfo",
			"Failed to exchange the OAuth token",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}

	client := s.oauthGoogleConfig.Client(ctx, token)
	response, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, cexceptions.New(
			"InvalidAuthenticationCode",
			"OAuth",
			"GetGoogleUserInfo",
			"Authentication code is invalid",
			http.StatusBadRequest,
		).WithOrigin(err)
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		if response.StatusCode >= http.StatusBadRequest && response.StatusCode < http.StatusInternalServerError {
			return nil, cexceptions.New(
				"InvalidAuthenticationCode",
				"OAuth",
				"GetGoogleUserInfo",
				"Authentication code is invalid or expired",
				http.StatusBadRequest,
			).WithOrigin(
				fmt.Errorf("Google userinfo returned HTTP status %d", response.StatusCode),
			)
		}

		return nil, cexceptions.New(
			"OAuthProviderUnavailable",
			"OAuth",
			"GetGoogleUserInfo",
			"The OAuth provider is unavailable",
			http.StatusBadGateway,
			true,
		).WithOrigin(
			fmt.Errorf("Google userinfo returned HTTP status %d", response.StatusCode),
		)
	}

	data, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, cexceptions.New(
			"ResponseReadFailed",
			"OAuth",
			"GetGoogleUserInfo",
			"Failed to read the OAuth provider response",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}

	var userInfo cauthdto.GetGoogleUserInfoResponseDto
	if err := json.Unmarshal(data, &userInfo); err != nil {
		return nil, cexceptions.New(
			"InvalidResponse",
			"OAuth",
			"GetGoogleUserInfo",
			"OAuth provider response is invalid",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}
	if userInfo.Id == "" || userInfo.Email == "" {
		return nil, cexceptions.New(
			"InvalidResponse",
			"OAuth",
			"GetGoogleUserInfo",
			"OAuth provider response does not contain a user ID and email",
			http.StatusBadGateway,
			true,
		)
	}

	return &userInfo, nil
}
