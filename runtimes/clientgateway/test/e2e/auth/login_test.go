package authe2etest

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"
)

/* ============================== Test Case Types ============================== */
type LoginRequestType = commonRequest[
	struct {
		Account  string `json:"account"`
		Password string `json:"password"`
	},
	commonCookies,
]
type LoginResponseType = commonResponse[
	struct {
		AccessToken string    `json:"accessToken"`
		UpdatedAt   time.Time `json:"updatedAt"`
	},
	commonCookies,
]
type LoginE2ETestCase = e2eTestCase[
	LoginRequestType,
	LoginResponseType,
]

/* ============================== Test Data Path & Some Constants ============================== */

const (
	loginTestdataPath = "testdata/login_testdata/"
	loginRoute        = testAuthRouteNamespace + "/login"
)

type LoginE2ETesterInterface interface {
	getLoginTestDataAndResponse(
		t *testing.T,
		method string,
		loginTestdataPath string,
	) (
		w *testHTTPResponse,
		testCase LoginE2ETestCase,
		res LoginResponseType,
		cookieMap map[string]string,
	)
	TestLoginValidTestAccountByName(t *testing.T)
	TestLoginValidTestAccountByEmail(t *testing.T)
}

type LoginE2ETester struct {
	baseURL string
	client  *http.Client
}

func NewLoginE2ETester(baseURL string, client *http.Client) LoginE2ETesterInterface {
	if baseURL == "" || client == nil {
		return nil
	}
	return &LoginE2ETester{
		baseURL: baseURL,
		client:  client,
	}
}

/* ============================== Auxiliary Functions ============================== */

func (et *LoginE2ETester) getLoginTestDataAndResponse(
	t *testing.T,
	method string,
	loginTestdataPath string,
) (
	w *testHTTPResponse,
	testCase LoginE2ETestCase,
	res LoginResponseType,
	cookieMap map[string]string,
) {
	if et == nil || et.client == nil {
		t.Fatalf("loginE2ETester or client is nil")
	}

	testCase = loadTestCase[LoginE2ETestCase](
		t, loginTestdataPath,
	)

	jsonBody, err := json.Marshal(testCase.Request.Body)
	if err != nil {
		t.Fatalf("failed to marshal request body: %v", err)
	}
	req, err := http.NewRequest(
		method,
		et.baseURL+loginRoute,
		bytes.NewReader(jsonBody),
	)
	if err != nil {
		t.Fatalf("failed to create HTTP request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if ua := testCase.Request.Header.UserAgent; ua != nil {
		req.Header.Set("User-Agent", *ua)
	}

	response, err := et.client.Do(req)
	if err != nil {
		t.Fatalf("failed to send HTTP request: %v", err)
	}
	w, err = newTestHTTPResponse(response)
	if err != nil {
		t.Fatalf("failed to read HTTP response: %v", err)
	}
	if err := json.Unmarshal(w.Body.Bytes(), &res.Body); err != nil {
		t.Errorf("failed to unmarshal response body: %v, body: %s", err, w.Body.String())
	}

	cookies := w.Result().Cookies()
	cookieMap = make(map[string]string)
	for _, c := range cookies {
		cookieMap[c.Name] = c.Value
	}
	return w, testCase, res, cookieMap
}

/* ============================== Test Cases ============================== */

func (et *LoginE2ETester) TestLoginValidTestAccountByName(t *testing.T) {
	if et.client == nil {
		return
	}

	w, testCase, res, cookieMap := et.getLoginTestDataAndResponse(
		t, "POST", loginTestdataPath+"valid_test_account_by_name_testdata.json",
	)

	// check status code
	if w.Code != testCase.Response.HTTPStatusCode {
		t.Errorf("expected http status code to be %d, got %d", testCase.Response.HTTPStatusCode, w.Code)
	}

	// check the body
	if err := json.Unmarshal(w.Body.Bytes(), &res.Body); err != nil {
		t.Errorf("failed to unmarshal response body: %v, body: %s", err, w.Body.String())
	}

	if !res.Body.Success {
		t.Errorf("expected body.success to be true, got false")
	}

	if res.Body.Data == nil {
		t.Errorf("expected response data to be not nil, got nil")
	}
	if len(strings.ReplaceAll(res.Body.Data.AccessToken, " ", "")) == 0 {
		t.Errorf("expected body.data.accessToken to be exist, got nil")
	}

	now := time.Now()
	if !isTimeWithin(res.Body.Data.UpdatedAt, now, 10*time.Second) {
		t.Errorf("expected body.data.createdAt to be %v (within tolerable time duration of %v), got %v", testCase.Response.Body.Data.UpdatedAt, 10*time.Second, now)
	}

	if res.Body.Exception != nil {
		t.Errorf("expected body.exception to be nil, got not %v", res.Body.Exception)
	}

	// check the accessToken in cookies
	if _, ok := cookieMap["accessToken"]; !ok {
		t.Errorf("expected cookie.accessToken to be set, got nil")
	}

	// check the refreshToken in cookies
	if _, ok := cookieMap["refreshToken"]; !ok {
		t.Errorf("expected cookie.refreshToken to be set, got nil")
	}
}

func (et *LoginE2ETester) TestLoginValidTestAccountByEmail(t *testing.T) {
	if et.client == nil {
		return
	}

	w, testCase, res, cookieMap := et.getLoginTestDataAndResponse(
		t, "POST", loginTestdataPath+"valid_test_account_by_email_testdata.json",
	)

	// check status code
	if w.Code != testCase.Response.HTTPStatusCode {
		t.Errorf("expected http status code to be %d, got %d", testCase.Response.HTTPStatusCode, w.Code)
	}

	// check the body
	if err := json.Unmarshal(w.Body.Bytes(), &res.Body); err != nil {
		t.Errorf("failed to unmarshal response body: %v, body: %s", err, w.Body.String())
	}

	if !res.Body.Success {
		t.Errorf("expected body.success to be true, got false")
	}

	if res.Body.Data == nil {
		t.Errorf("expected response data to be not nil, got nil")
	}
	if len(strings.ReplaceAll(res.Body.Data.AccessToken, " ", "")) == 0 {
		t.Errorf("expected body.data.accessToken to be exist, got nil")
	}

	now := time.Now()
	if !isTimeWithin(res.Body.Data.UpdatedAt, now, 10*time.Second) {
		t.Errorf("expected body.data.createdAt to be %v (within tolerable time duration of %v), got %v", testCase.Response.Body.Data.UpdatedAt, 10*time.Second, now)
	}

	if res.Body.Exception != nil {
		t.Errorf("expected body.exception to be nil, got not %v", res.Body.Exception)
	}

	// check the accessToken in cookies
	if _, ok := cookieMap["accessToken"]; !ok {
		t.Errorf("expected cookie.accessToken to be set, got nil")
	}

	// check the refreshToken in cookies
	if _, ok := cookieMap["refreshToken"]; !ok {
		t.Errorf("expected cookie.refreshToken to be set, got nil")
	}
}

// may test using access token or refresh token to login
