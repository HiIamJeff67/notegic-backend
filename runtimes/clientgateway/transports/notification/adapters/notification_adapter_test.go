package adapters

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	sharedcontexts "github.com/HiIamJeff67/notegic-backend/shared/lib/contexts"
)

type notificationRoundTripFunc func(*http.Request) (*http.Response, error)

func (f notificationRoundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

func TestCallSecurlyReturnsUpstreamStatusForNonJSONResponse(t *testing.T) {
	t.Setenv("CORE_DELEGATION_SECRET", "test-secret")
	t.Setenv("CORE_DELEGATION_AUDIENCE", "test-audience")
	t.Setenv("CORE_DELEGATION_ISSUER", "test-issuer")

	gin.SetMode(gin.TestMode)
	ginContext, _ := gin.CreateTestContext(httptest.NewRecorder())
	ginContext.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	ginContext.Set(sharedcontexts.ContextFieldName_User_PublicId.String(), uuid.New())

	client := &NotificationAdapter{
		baseURL: "http://notification",
		httpClient: &http.Client{
			Transport: notificationRoundTripFunc(func(*http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusNotFound,
					Body:       io.NopCloser(strings.NewReader("404 page not found\n")),
					Header:     make(http.Header),
				}, nil
			}),
		},
	}
	_, exception := CallSecurly[struct{}, struct{}](
		ginContext,
		client,
		&struct{}{},
		"SearchPrivateNotifications",
		"/runtimes/v1/notifications/search",
	)
	if exception == nil {
		t.Fatal("expected upstream failure")
	}
	if exception.Reason != "NotificationResponseFailed" {
		t.Fatalf("unexpected exception reason: %s", exception.Reason)
	}
	if exception.HTTPStatusCode() != http.StatusNotFound {
		t.Fatalf("unexpected HTTP status: %d", exception.HTTPStatusCode())
	}
	if exception.Origin() == nil || !strings.Contains(exception.Origin().Error(), "status 404") {
		t.Fatalf("expected upstream status in origin, got %v", exception.Origin())
	}
}
