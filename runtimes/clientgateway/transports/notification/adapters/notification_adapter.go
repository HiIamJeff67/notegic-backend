package adapters

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	cgateway "github.com/HiIamJeff67/notegic-backend/contracts/gateway/v1"
	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	sharedcontexts "github.com/HiIamJeff67/notegic-backend/shared/lib/contexts"
	sharedtokens "github.com/HiIamJeff67/notegic-backend/shared/tokens"

	gatewaycontexts "github.com/HiIamJeff67/notegic-backend/runtimes/clientgateway/contexts"
)

type NotificationAdapter struct {
	baseURL    string
	httpClient *http.Client
}

func NewNotificationAdapter(baseURL string, timeout time.Duration) *NotificationAdapter {
	return &NotificationAdapter{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func CallSecurly[RequestDto any, ResponseDto any](
	ctx *gin.Context,
	client *NotificationAdapter,
	requestDto *RequestDto,
	operation string,
	path string,
) (*cgateway.Response[ResponseDto], *cexceptions.Exception) {
	if client == nil {
		return nil, cexceptions.New(
			"NotificationAdapterRequired",
			"Gateway",
			operation,
			"The Notification service adapter is required",
			http.StatusInternalServerError,
			true,
		)
	}
	if requestDto == nil {
		return nil, cexceptions.New(
			"InvalidRequest",
			"Gateway",
			operation,
			"The Notification service request DTO is required",
			http.StatusBadRequest,
		)
	}

	userSubject, exception := gatewaycontexts.GetAndConvertContextFieldToUUID(
		ctx,
		sharedcontexts.ContextFieldName_User_PublicId,
	)
	if exception != nil {
		return nil, exception
	}
	if userSubject == nil || *userSubject == uuid.Nil {
		return nil, cexceptions.New(
			"ContextFieldInvalid",
			"Gateway",
			operation,
			"A valid user subject is required for a secure Notification service call",
			http.StatusInternalServerError,
			true,
		)
	}

	requestId := ctx.GetHeader("X-Request-Id")
	if requestId == "" {
		requestId = uuid.NewString()
	}
	delegationToken, err := sharedtokens.GenerateDelegationToken(sharedtokens.DelegationTokenClaims{
		Actor:       "gateway",
		UserSubject: userSubject.String(),
		Operation:   operation,
		RequestId:   requestId,
	})
	if err != nil {
		return nil, cexceptions.New(
			"NotificationDelegationFailed",
			"Gateway",
			operation,
			"Failed to issue a Notification service delegation token",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}

	request := &cgateway.Request[RequestDto]{
		Operation: operation,
		Metadata: cgateway.RequestMetadata{
			RequestId:      requestId,
			TraceParent:    ctx.GetHeader("Traceparent"),
			IdempotencyKey: ctx.GetHeader("Idempotency-Key"),
		},
		Dto: *requestDto,
	}
	body, err := json.Marshal(request)
	if err != nil {
		return nil, cexceptions.New(
			"NotificationRequestEncodingFailed",
			"Gateway",
			operation,
			"Failed to encode the Notification service request",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}

	httpRequest, err := http.NewRequestWithContext(
		ctx.Request.Context(),
		http.MethodPost,
		client.baseURL+"/"+strings.TrimLeft(path, "/"),
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, cexceptions.New(
			"NotificationRequestCreationFailed",
			"Gateway",
			operation,
			"Failed to create the Notification service request",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}
	httpRequest.Header.Set("Authorization", "Bearer "+*delegationToken)
	httpRequest.Header.Set("Content-Type", "application/json")
	httpRequest.Header.Set("X-Request-Id", requestId)
	if request.Metadata.TraceParent != "" {
		httpRequest.Header.Set("Traceparent", request.Metadata.TraceParent)
	}
	if request.Metadata.IdempotencyKey != "" {
		httpRequest.Header.Set("Idempotency-Key", request.Metadata.IdempotencyKey)
	}
	for _, header := range []string{"User-Agent", "X-Real-IP", "X-Forwarded-For"} {
		if value := ctx.GetHeader(header); value != "" {
			httpRequest.Header.Set(header, value)
		}
	}

	httpResponse, err := client.httpClient.Do(httpRequest)
	if err != nil {
		return nil, cexceptions.New(
			"NotificationRequestFailed",
			"Gateway",
			operation,
			"Failed to communicate with the Notification service",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}
	defer httpResponse.Body.Close()
	responseBody, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		return nil, cexceptions.New(
			"NotificationResponseReadFailed",
			"Gateway",
			operation,
			"Failed to read the Notification service response",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}
	if httpResponse.StatusCode < http.StatusOK || httpResponse.StatusCode >= http.StatusMultipleChoices {
		response := &cgateway.Response[ResponseDto]{}
		if err := json.Unmarshal(responseBody, response); err == nil && response.Exception != nil {
			return nil, response.Exception.Clone(httpResponse.StatusCode)
		}
		return nil, cexceptions.New(
			"NotificationResponseFailed",
			"Gateway",
			operation,
			"The Notification service returned an unsuccessful response",
			httpResponse.StatusCode,
			true,
		).WithOrigin(fmt.Errorf("status %d: %s", httpResponse.StatusCode, strings.TrimSpace(string(responseBody))))
	}
	response := &cgateway.Response[ResponseDto]{}
	if err := json.Unmarshal(responseBody, response); err != nil {
		return nil, cexceptions.New(
			"NotificationResponseDecodingFailed",
			"Gateway",
			operation,
			"Failed to decode the Notification service response",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}
	if response.Version != cgateway.Version {
		return nil, cexceptions.New(
			"NotificationResponseVersionInvalid",
			"Gateway",
			operation,
			"The Notification service response uses an unsupported version",
			http.StatusInternalServerError,
			true,
		)
	}
	if response.Metadata.RequestId != requestId {
		return nil, cexceptions.New(
			"NotificationResponseRequestIdInvalid",
			"Gateway",
			operation,
			"The Notification service response does not match the request",
			http.StatusInternalServerError,
			true,
		)
	}
	return response, nil
}
