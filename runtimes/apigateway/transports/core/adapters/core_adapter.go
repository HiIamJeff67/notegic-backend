package adapters

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	cgateway "github.com/HiIamJeff67/notegic-backend/contracts/gateway/v1"
	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	sharedtokens "github.com/HiIamJeff67/notegic-backend/shared/tokens"
)

type CoreAdapter struct {
	baseURL    string
	httpClient *http.Client
}

func NewCoreAdapter(baseURL string, timeout time.Duration) *CoreAdapter {
	return &CoreAdapter{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

/* ============================== Delegation Methods ============================== */

func IssueDelegationTokenFromSource(
	actor string,
	userSubject string,
	allowedPermissions []string,
	operation string,
	requestId string,
	gatewaySource string,
	authMethod string,
	apiKeyId string,
) (string, error) {
	token, err := sharedtokens.GenerateDelegationToken(sharedtokens.DelegationTokenClaims{
		Actor:              actor,
		GatewaySource:      gatewaySource,
		AuthMethod:         authMethod,
		ApiKeyId:           apiKeyId,
		UserSubject:        userSubject,
		AllowedPermissions: allowedPermissions,
		Operation:          operation,
		RequestId:          requestId,
	})
	if err != nil {
		return "", err
	}

	return *token, nil
}

/* ============================== Internal HTTP Methods ============================== */

func call[RequestDto any, ResponseDto any](
	client *CoreAdapter,
	ctx context.Context,
	method string,
	path string,
	delegationToken string,
	forwardedHeaders http.Header,
	request *cgateway.Request[RequestDto],
) (*cgateway.Response[ResponseDto], *cexceptions.Exception) {
	if client == nil {
		return nil, cexceptions.New(
			"CoreAdapterRequired",
			"Gateway",
			"CallCore",
			"The Core service client is required",
			http.StatusInternalServerError,
			true,
		)
	}
	if request == nil {
		return nil, cexceptions.New(
			"InvalidRequest",
			"Gateway",
			"CallCore",
			"The Core service request is required",
			http.StatusInternalServerError,
			true,
		)
	}
	if request.Version == "" {
		request.Version = cgateway.Version
	}

	body, err := json.Marshal(request)
	if err != nil {
		return nil, cexceptions.New(
			"CoreRequestEncodingFailed",
			"Gateway",
			"CallCore",
			"Failed to encode the Core service request",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}
	httpRequest, err := http.NewRequestWithContext(
		ctx,
		method,
		client.baseURL+"/"+strings.TrimLeft(path, "/"),
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, cexceptions.New(
			"CoreRequestCreationFailed",
			"Gateway",
			"CallCore",
			"Failed to create the Core service request",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}
	httpRequest.Header.Set("Authorization", "Bearer "+delegationToken)
	httpRequest.Header.Set("Content-Type", "application/json")
	for key, values := range forwardedHeaders {
		if strings.EqualFold(key, "Cookie") {
			continue
		}
		for _, value := range values {
			httpRequest.Header.Add(key, value)
		}
	}
	httpRequest.Header.Set("X-Request-Id", request.Metadata.RequestId)
	if request.Metadata.TraceParent != "" {
		httpRequest.Header.Set("Traceparent", request.Metadata.TraceParent)
	}
	if request.Metadata.IdempotencyKey != "" {
		httpRequest.Header.Set("Idempotency-Key", request.Metadata.IdempotencyKey)
	}

	httpResponse, err := client.httpClient.Do(httpRequest)
	if err != nil {
		return nil, cexceptions.New(
			"CoreRequestFailed",
			"Gateway",
			"CallCore",
			"Failed to communicate with the Core service",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}
	defer httpResponse.Body.Close()
	responseBody, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		return nil, cexceptions.New(
			"CoreResponseReadFailed",
			"Gateway",
			"CallCore",
			"Failed to read the Core service response",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}
	response := &cgateway.Response[ResponseDto]{}
	if err := json.Unmarshal(responseBody, response); err != nil {
		return nil, cexceptions.New(
			"CoreResponseDecodingFailed",
			"Gateway",
			"CallCore",
			"Failed to decode the Core service response",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}
	if response.Version != cgateway.Version {
		return nil, cexceptions.New(
			"CoreResponseVersionInvalid",
			"Gateway",
			"CallCore",
			"The Core service response uses an unsupported version",
			http.StatusInternalServerError,
			true,
		)
	}
	if response.Metadata.RequestId != request.Metadata.RequestId {
		return nil, cexceptions.New(
			"CoreResponseRequestIdInvalid",
			"Gateway",
			"CallCore",
			"The Core service response does not match the request",
			http.StatusInternalServerError,
			true,
		)
	}
	if httpResponse.StatusCode < http.StatusOK || httpResponse.StatusCode >= http.StatusMultipleChoices {
		if response.Exception != nil {
			return nil, response.Exception.Clone(httpResponse.StatusCode)
		}
		return nil, cexceptions.New(
			"CoreResponseFailed",
			"Gateway",
			"CallCore",
			"The Core service returned an unsuccessful response",
			http.StatusInternalServerError,
			true,
		)
	}

	return response, nil
}

// CallAsAPIKey is the Core adapter path for APIGateway requests. The edge
// middleware has already checked the header shape; Core remains responsible
// for hashing, cache/DB lookup, revocation, and actor-context population.
func CallAsAPIKey[RequestDto any, ResponseDto any](
	ctx *gin.Context,
	client *CoreAdapter,
	requestDto *RequestDto,
	operation string,
	path string,
) (*cgateway.Response[ResponseDto], *cexceptions.Exception) {
	if requestDto == nil {
		return nil, cexceptions.New(
			"InvalidRequest",
			"Gateway",
			operation,
			"The Core service request DTO is required",
			http.StatusBadRequest,
		)
	}
	requestId := ctx.GetHeader("X-Request-Id")
	if requestId == "" {
		requestId = uuid.NewString()
	}
	delegationToken, err := IssueDelegationTokenFromSource(
		"gateway",
		"",
		nil,
		operation,
		requestId,
		sharedtokens.GatewaySourceAPI,
		sharedtokens.AuthMethodAPIKey,
		"",
	)
	if err != nil {
		return nil, cexceptions.New(
			"CoreDelegationFailed",
			"Gateway",
			operation,
			"Failed to communicate with the Core service",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}
	forwardedHeaders := http.Header{}
	for _, header := range []string{"User-Agent", "X-Real-IP", "X-Forwarded-For", "X-API-Key"} {
		if value := ctx.GetHeader(header); value != "" {
			forwardedHeaders.Set(header, value)
		}
	}
	return call[RequestDto, ResponseDto](
		client,
		ctx.Request.Context(),
		http.MethodPost,
		path,
		delegationToken,
		forwardedHeaders,
		&cgateway.Request[RequestDto]{
			Operation: operation,
			Metadata: cgateway.RequestMetadata{
				RequestId:      requestId,
				TraceParent:    ctx.GetHeader("Traceparent"),
				IdempotencyKey: ctx.GetHeader("Idempotency-Key"),
			},
			Dto: *requestDto,
		},
	)
}

func CallSecurly[RequestDto any, ResponseDto any](
	ctx *gin.Context,
	client *CoreAdapter,
	requestDto *RequestDto,
	operation string,
	path string,
) (*cgateway.Response[ResponseDto], *cexceptions.Exception) {
	// APIGateway never falls back to the ClientGateway JWT delegation path.
	// Core authenticates the forwarded API key and hydrates the actor context.
	return CallAsAPIKey[RequestDto, ResponseDto](ctx, client, requestDto, operation, path)
}
