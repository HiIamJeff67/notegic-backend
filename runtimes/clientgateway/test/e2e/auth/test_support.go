package authe2etest

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"testing"
	"time"
)

type testHTTPResponse struct {
	Code     int
	Body     *bytes.Buffer
	response *http.Response
}

func newTestHTTPResponse(response *http.Response) (*testHTTPResponse, error) {
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return &testHTTPResponse{
		Code:     response.StatusCode,
		Body:     bytes.NewBuffer(body),
		response: response,
	}, nil
}

func (response *testHTTPResponse) Result() *http.Response {
	return response.response
}

type commonCookies struct {
	AccessToken  string
	RefreshToken string
}

type commonRequest[Body any, Cookies any] struct {
	Header struct {
		UserAgent *string
	}
	Body    Body
	Cookies *Cookies
}

type commonResponse[Data any, Cookies any] struct {
	HTTPStatusCode int
	Body           struct {
		Success   bool  `json:"success"`
		Data      *Data `json:"data"`
		Exception any   `json:"exception"`
	}
	Cookies *Cookies
}

type e2eTestCase[Request any, Response any] struct {
	Request  Request
	Response Response
}

func loadTestCase[Case any](t *testing.T, path string) Case {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read testdata: %v", err)
	}

	var testCase Case
	if err := json.Unmarshal(data, &testCase); err != nil {
		t.Fatalf("failed to unmarshal testdata: %v", err)
	}

	return testCase
}

func isTimeWithin(t1, t2 time.Time, delta time.Duration) bool {
	difference := t1.Sub(t2)
	if difference < 0 {
		difference = -difference
	}

	return difference <= delta
}
