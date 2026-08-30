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

type RegisterRequestType = commonRequest[
	struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	},
	commonCookies,
]
type RegisterResponseType = commonResponse[
	struct {
		AccessToken string    `json:"accessToken"`
		CreatedAt   time.Time `json:"createdAt"`
	},
	commonCookies,
]
type RegisterE2ETestCase = e2eTestCase[
	RegisterRequestType,
	RegisterResponseType,
]

/* ============================== Test Data Path & Some Constants ============================== */

const (
	registerTestdataPath = "testdata/register_testdata/"
	registerRoute        = testAuthRouteNamespace + "/register"
)

type RegisterE2ETesterInterface interface {
	getRegisterTestdataAndResponse(
		t *testing.T,
		method string,
		registerTestdataPath string,
	) (
		w *testHTTPResponse,
		testCase RegisterE2ETestCase,
		res RegisterResponseType,
		cookieMap map[string]string,
	)
	TestRegisterValidTestAccount(t *testing.T)
	TestRegisterValidUserAccount(t *testing.T)
	TestRegisterNoName(t *testing.T)
	TestRegisterNameWithoutNumber(t *testing.T)
	TestRegisterShortName(t *testing.T)
	TestRegisterInvalidEmail(t *testing.T)
	TestRegisterShortPassword(t *testing.T)
	TestRegisterPasswordWithoutLowerCaseLetter(t *testing.T)
	TestRegisterPasswordWithoutUpperCaseLetter(t *testing.T)
	TestRegisterPasswordWithoutNumber(t *testing.T)
	TestRegisterPasswordWithoutSign(t *testing.T)
}

type RegisterE2ETester struct {
	baseURL string
	client  *http.Client
}

func NewRegisterE2ETester(baseURL string, client *http.Client) RegisterE2ETesterInterface {
	if baseURL == "" || client == nil {
		return nil
	}
	return &RegisterE2ETester{
		baseURL: baseURL,
		client:  client,
	}
}

/* ============================== Auxiliary Functions ============================== */

func (et *RegisterE2ETester) getRegisterTestdataAndResponse(
	t *testing.T,
	method string,
	registerTestdataPath string,
) (
	w *testHTTPResponse,
	testCase RegisterE2ETestCase,
	res RegisterResponseType,
	cookieMap map[string]string,
) {
	if et == nil || et.client == nil {
		t.Fatal("registerE2ETester or client is nil")
	}

	testCase = loadTestCase[RegisterE2ETestCase](
		t, registerTestdataPath,
	)

	jsonBody, err := json.Marshal(testCase.Request.Body)
	if err != nil {
		t.Fatalf("failed to marshal request body: %v", err)
	}
	req, err := http.NewRequest(
		method,
		et.baseURL+registerRoute,
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

func (et *RegisterE2ETester) TestRegisterValidTestAccount(t *testing.T) {
	if et.client == nil {
		return
	}

	w, testCase, res, cookieMap := et.getRegisterTestdataAndResponse(
		t, "POST", registerTestdataPath+"valid_test_account_testdata.json",
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
	if !isTimeWithin(res.Body.Data.CreatedAt, now, 10*time.Second) {
		t.Errorf("expected body.data.createdAt to be %v (within tolerable time duration of %v), got %v", testCase.Response.Body.Data.CreatedAt, 10*time.Second, now)
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

func (et *RegisterE2ETester) TestRegisterValidUserAccount(t *testing.T) {
	if et.client == nil {
		return
	}

	w, testCase, res, cookieMap := et.getRegisterTestdataAndResponse(
		t, "POST", registerTestdataPath+"valid_user_account_testdata.json",
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
		t.Errorf("expected response data to be exist, got nil")
	}
	if len(strings.ReplaceAll(res.Body.Data.AccessToken, " ", "")) == 0 {
		t.Errorf("expected body.data.accessToken to be exist, got nil")
	}
	now := time.Now()
	if !isTimeWithin(res.Body.Data.CreatedAt, now, 10*time.Second) {
		t.Errorf("expected body.data.createdAt to be %v (within tolerable time duration of %v), got %v", testCase.Response.Body.Data.CreatedAt, 10*time.Second, now)
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

func (et *RegisterE2ETester) TestRegisterNoName(t *testing.T) {
	if et.client == nil {
		return
	}

	w, testCase, res, cookieMap := et.getRegisterTestdataAndResponse(
		t, "POST", registerTestdataPath+"no_name_testdata.json",
	)

	// check status code
	if w.Code != testCase.Response.HTTPStatusCode {
		t.Errorf("expected http status code to be %d, got %d", testCase.Response.HTTPStatusCode, w.Code)
	}

	// check the body
	if err := json.Unmarshal(w.Body.Bytes(), &res.Body); err != nil {
		t.Errorf("failed to unmarshal response body: %v, body: %s", err, w.Body.String())
	}

	if res.Body.Success {
		t.Errorf("expected body.success to be false, got true")
	}

	if res.Body.Data != nil {
		t.Errorf("expected response data to be nil, got %v", res.Body.Data)
	}

	if res.Body.Exception == nil {
		t.Errorf("expected body.exception to be exist, got nil")
	}

	// check the accessToken in cookies
	if val, ok := cookieMap["accessToken"]; ok {
		t.Errorf("expected cookie.accessToken to be not set, got %v", val)
	}

	// check the refreshToken in cookies
	if val, ok := cookieMap["refreshToken"]; ok {
		t.Errorf("expected cookie.refreshToken to be not set, got %v", val)
	}
}

func (et *RegisterE2ETester) TestRegisterNameWithoutNumber(t *testing.T) {
	if et.client == nil {
		return
	}

	w, testCase, res, cookieMap := et.getRegisterTestdataAndResponse(
		t, "POST", registerTestdataPath+"name_without_number_testdata.json",
	)

	// check status code
	if w.Code != testCase.Response.HTTPStatusCode {
		t.Errorf("expected http status code to be %d, got %d", testCase.Response.HTTPStatusCode, w.Code)
	}

	// check the body
	if err := json.Unmarshal(w.Body.Bytes(), &res.Body); err != nil {
		t.Errorf("failed to unmarshal response body: %v, body: %s", err, w.Body.String())
	}

	if res.Body.Success {
		t.Errorf("expected body.success to be false, got true")
	}

	if res.Body.Data != nil {
		t.Errorf("expected response data to be nil, got %v", res.Body.Data)
	}

	if res.Body.Exception == nil {
		t.Errorf("expected body.exception to be exist, got nil")
	}

	// check the accessToken in cookies
	if val, ok := cookieMap["accessToken"]; ok {
		t.Errorf("expected cookie.accessToken to be not set, got %v", val)
	}

	// check the refreshToken in cookies
	if val, ok := cookieMap["refreshToken"]; ok {
		t.Errorf("expected cookie.refreshToken to be not set, got %v", val)
	}
}

func (et *RegisterE2ETester) TestRegisterShortName(t *testing.T) {
	if et.client == nil {
		return
	}

	w, testCase, res, cookieMap := et.getRegisterTestdataAndResponse(
		t, "POST", registerTestdataPath+"short_name_testdata.json",
	)

	// check status code
	if w.Code != testCase.Response.HTTPStatusCode {
		t.Errorf("expected http status code to be %d, got %d", testCase.Response.HTTPStatusCode, w.Code)
	}

	// check the body
	if err := json.Unmarshal(w.Body.Bytes(), &res.Body); err != nil {
		t.Errorf("failed to unmarshal response body: %v, body: %s", err, w.Body.String())
	}

	if res.Body.Success {
		t.Errorf("expected body.success to be false, got true")
	}

	if res.Body.Data != nil {
		t.Errorf("expected response data to be nil, got %v", res.Body.Data)
	}

	if res.Body.Exception == nil {
		t.Errorf("expected body.exception to be exist, got nil")
	}

	// check the accessToken in cookies
	if val, ok := cookieMap["accessToken"]; ok {
		t.Errorf("expected cookie.accessToken to be not set, got %v", val)
	}

	// check the refreshToken in cookies
	if val, ok := cookieMap["refreshToken"]; ok {
		t.Errorf("expected cookie.refreshToken to be not set, got %v", val)
	}
}

func (et *RegisterE2ETester) TestRegisterInvalidEmail(t *testing.T) {
	if et.client == nil {
		return
	}

	w, testCase, res, cookieMap := et.getRegisterTestdataAndResponse(
		t, "POST", registerTestdataPath+"invalid_email_testdata.json",
	)

	// check status code
	if w.Code != testCase.Response.HTTPStatusCode {
		t.Errorf("expected http status code to be %d, got %d", testCase.Response.HTTPStatusCode, w.Code)
	}

	// check the body
	if err := json.Unmarshal(w.Body.Bytes(), &res.Body); err != nil {
		t.Errorf("failed to unmarshal response body: %v, body: %s", err, w.Body.String())
	}

	if res.Body.Success {
		t.Errorf("expected body.success to be false, got true")
	}

	if res.Body.Data != nil {
		t.Errorf("expected response data to be nil, got %v", res.Body.Data)
	}

	if res.Body.Exception == nil {
		t.Errorf("expected body.exception to be exist, got nil")
	}

	// check the accessToken in cookies
	if val, ok := cookieMap["accessToken"]; ok {
		t.Errorf("expected cookie.accessToken to be not set, got %v", val)
	}

	// check the refreshToken in cookies
	if val, ok := cookieMap["refreshToken"]; ok {
		t.Errorf("expected cookie.refreshToken to be not set, got %v", val)
	}
}

func (et *RegisterE2ETester) TestRegisterShortPassword(t *testing.T) {
	if et.client == nil {
		return
	}

	w, testCase, res, cookieMap := et.getRegisterTestdataAndResponse(
		t, "POST", registerTestdataPath+"short_password_testdata.json",
	)

	// check status code
	if w.Code != testCase.Response.HTTPStatusCode {
		t.Errorf("expected http status code to be %d, got %d", testCase.Response.HTTPStatusCode, w.Code)
	}

	// check the body
	if err := json.Unmarshal(w.Body.Bytes(), &res.Body); err != nil {
		t.Errorf("failed to unmarshal response body: %v, body: %s", err, w.Body.String())
	}

	if res.Body.Success {
		t.Errorf("expected body.success to be false, got true")
	}

	if res.Body.Data != nil {
		t.Errorf("expected response data to be nil, got %v", res.Body.Data)
	}

	if res.Body.Exception == nil {
		t.Errorf("expected body.exception to be exist, got nil")
	}

	// check the accessToken in cookies
	if val, ok := cookieMap["accessToken"]; ok {
		t.Errorf("expected cookie.accessToken to be not set, got %v", val)
	}

	// check the refreshToken in cookies
	if val, ok := cookieMap["refreshToken"]; ok {
		t.Errorf("expected cookie.refreshToken to be not set, got %v", val)
	}
}

func (et *RegisterE2ETester) TestRegisterPasswordWithoutLowerCaseLetter(t *testing.T) {
	if et.client == nil {
		return
	}

	w, testCase, res, cookieMap := et.getRegisterTestdataAndResponse(
		t, "POST", registerTestdataPath+"password_without_lower_case_letter_testdata.json",
	)

	// check status code
	if w.Code != testCase.Response.HTTPStatusCode {
		t.Errorf("expected http status code to be %d, got %d", testCase.Response.HTTPStatusCode, w.Code)
	}

	// check the body
	if err := json.Unmarshal(w.Body.Bytes(), &res.Body); err != nil {
		t.Errorf("failed to unmarshal response body: %v, body: %s", err, w.Body.String())
	}

	if res.Body.Success {
		t.Errorf("expected body.success to be false, got true")
	}

	if res.Body.Data != nil {
		t.Errorf("expected response data to be nil, got %v", res.Body.Data)
	}

	if res.Body.Exception == nil {
		t.Errorf("expected body.exception to be exist, got nil")
	}

	// check the accessToken in cookies
	if val, ok := cookieMap["accessToken"]; ok {
		t.Errorf("expected cookie.accessToken to be not set, got %v", val)
	}

	// check the refreshToken in cookies
	if val, ok := cookieMap["refreshToken"]; ok {
		t.Errorf("expected cookie.refreshToken to be not set, got %v", val)
	}
}

func (et *RegisterE2ETester) TestRegisterPasswordWithoutUpperCaseLetter(t *testing.T) {
	if et.client == nil {
		return
	}

	w, testCase, res, cookieMap := et.getRegisterTestdataAndResponse(
		t, "POST", registerTestdataPath+"password_without_upper_case_letter_testdata.json",
	)

	// check status code
	if w.Code != testCase.Response.HTTPStatusCode {
		t.Errorf("expected http status code to be %d, got %d", testCase.Response.HTTPStatusCode, w.Code)
	}

	// check the body
	if err := json.Unmarshal(w.Body.Bytes(), &res.Body); err != nil {
		t.Errorf("failed to unmarshal response body: %v, body: %s", err, w.Body.String())
	}

	if res.Body.Success {
		t.Errorf("expected body.success to be false, got true")
	}

	if res.Body.Data != nil {
		t.Errorf("expected response data to be nil, got %v", res.Body.Data)
	}

	if res.Body.Exception == nil {
		t.Errorf("expected body.exception to be exist, got nil")
	}

	// check the accessToken in cookies
	if val, ok := cookieMap["accessToken"]; ok {
		t.Errorf("expected cookie.accessToken to be not set, got %v", val)
	}

	// check the refreshToken in cookies
	if val, ok := cookieMap["refreshToken"]; ok {
		t.Errorf("expected cookie.refreshToken to be not set, got %v", val)
	}
}

func (et *RegisterE2ETester) TestRegisterPasswordWithoutNumber(t *testing.T) {
	if et.client == nil {
		return
	}

	w, testCase, res, cookieMap := et.getRegisterTestdataAndResponse(
		t, "POST", registerTestdataPath+"password_without_number_testdata.json",
	)

	// check status code
	if w.Code != testCase.Response.HTTPStatusCode {
		t.Errorf("expected http status code to be %d, got %d", testCase.Response.HTTPStatusCode, w.Code)
	}

	// check the body
	if err := json.Unmarshal(w.Body.Bytes(), &res.Body); err != nil {
		t.Errorf("failed to unmarshal response body: %v, body: %s", err, w.Body.String())
	}

	if res.Body.Success {
		t.Errorf("expected body.success to be false, got true")
	}

	if res.Body.Data != nil {
		t.Errorf("expected response data to be nil, got %v", res.Body.Data)
	}

	if res.Body.Exception == nil {
		t.Errorf("expected body.exception to be exist, got nil")
	}

	// check the accessToken in cookies
	if val, ok := cookieMap["accessToken"]; ok {
		t.Errorf("expected cookie.accessToken to be not set, got %v", val)
	}

	// check the refreshToken in cookies
	if val, ok := cookieMap["refreshToken"]; ok {
		t.Errorf("expected cookie.refreshToken to be not set, got %v", val)
	}
}

func (et *RegisterE2ETester) TestRegisterPasswordWithoutSign(t *testing.T) {
	if et.client == nil {
		return
	}

	w, testCase, res, cookieMap := et.getRegisterTestdataAndResponse(
		t, "POST", registerTestdataPath+"password_without_sign_testdata.json",
	)

	// check status code
	if w.Code != testCase.Response.HTTPStatusCode {
		t.Errorf("expected http status code to be %d, got %d", testCase.Response.HTTPStatusCode, w.Code)
	}

	// check the body
	if err := json.Unmarshal(w.Body.Bytes(), &res.Body); err != nil {
		t.Errorf("failed to unmarshal response body: %v, body: %s", err, w.Body.String())
	}

	if res.Body.Success {
		t.Errorf("expected body.success to be false, got true")
	}

	if res.Body.Data != nil {
		t.Errorf("expected response data to be nil, got %v", res.Body.Data)
	}

	if res.Body.Exception == nil {
		t.Errorf("expected body.exception to be exist, got nil")
	}

	// check the accessToken in cookies
	if val, ok := cookieMap["accessToken"]; ok {
		t.Errorf("expected cookie.accessToken to be not set, got %v", val)
	}

	// check the refreshToken in cookies
	if val, ok := cookieMap["refreshToken"]; ok {
		t.Errorf("expected cookie.refreshToken to be not set, got %v", val)
	}
}
