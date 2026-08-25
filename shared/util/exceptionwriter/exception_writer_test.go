package exceptionwriter

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"
)

func TestSafelyAbortAndResponseWithJSONMasksInternalException(t *testing.T) {
	gin.SetMode(gin.TestMode)
	responseRecorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(responseRecorder)

	SafelyAbortAndResponseWithJSON(
		cexceptions.New(
			"DatabaseFailure",
			"API",
			"CreateRootShelf",
			"insert failed: duplicate key",
			http.StatusInternalServerError,
			true,
		).WithOrigin(errors.New("duplicate key value violates unique constraint")),
		ctx,
	)

	responseBody := responseRecorder.Body.String()
	if responseRecorder.Code != http.StatusInternalServerError {
		t.Fatalf("unexpected status: %d", responseRecorder.Code)
	}
	if strings.Contains(responseBody, "duplicate key") || strings.Contains(responseBody, "DatabaseFailure") {
		t.Fatalf("response leaked internal exception data: %s", responseBody)
	}
	if !strings.Contains(responseBody, "InternalServerError") {
		t.Fatalf("response did not contain the safe reason: %s", responseBody)
	}
}
