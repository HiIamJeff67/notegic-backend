package authe2etest

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	scookies "github.com/HiIamJeff67/notegic-backend/shared/cookies"
	spostgres "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres"

	testroutes "github.com/HiIamJeff67/notegic-backend/internal/clientgateway/transports/api/routes/testroutes"
	coreadapters "github.com/HiIamJeff67/notegic-backend/internal/clientgateway/transports/core/adapters"
)

const testAuthRouteNamespace = "/testRoute/auth"

func TestAuthE2E(t *testing.T) {
	databaseConfig := spostgres.Config{
		Host:     strings.TrimSpace(os.Getenv("CORE_DB_HOST")),
		User:     strings.TrimSpace(os.Getenv("CORE_DB_USER")),
		Password: os.Getenv("CORE_DB_PASSWORD"),
		Name:     strings.TrimSpace(os.Getenv("CORE_DB_NAME")),
		Port:     strings.TrimSpace(os.Getenv("CORE_DB_PORT")),
	}
	if databaseConfig.Host == "" || databaseConfig.User == "" || databaseConfig.Password == "" || databaseConfig.Name == "" || databaseConfig.Port == "" {
		t.Skipf("auth E2E test requires CORE_DB_HOST, CORE_DB_USER, CORE_DB_PASSWORD, CORE_DB_NAME, and CORE_DB_PORT")
	}
	db, err := spostgres.Connect(databaseConfig)
	if err != nil {
		t.Skipf("auth E2E test requires an available database: %v", err)
	}
	t.Cleanup(func() {
		_ = spostgres.Disconnect(db)
	})

	gin.SetMode(gin.TestMode)
	router := gin.New()
	testroutes.ConfigureTestAuthRoutes(
		router.Group(testAuthRouteNamespace),
		testroutes.AuthRouteDependencies{
			CoreAdapter: coreadapters.NewCoreAdapter("http://127.0.0.1:7778", 10*time.Second),
			AccessTokenCookieHandler: scookies.New(scookies.Config{
				Name:     scookies.ValidCookieName_AccessToken,
				Path:     "/",
				Duration: 30 * time.Minute,
				HTTPOnly: true,
				SameSite: http.SameSiteLaxMode,
			}),
			RefreshTokenCookieHandler: scookies.New(scookies.Config{
				Name:     scookies.ValidCookieName_RefreshToken,
				Path:     "/",
				Duration: 14 * 24 * time.Hour,
				HTTPOnly: true,
				SameSite: http.SameSiteStrictMode,
			}),
			AuthorizedRateLimiter: nil,
		},
	)
	server := httptest.NewServer(router)
	t.Cleanup(server.Close)
	client := server.Client()

	t.Run("register", func(t *testing.T) {
		registerE2ETester := NewRegisterE2ETester(server.URL, client)
		if registerE2ETester == nil {
			t.Fatal("NewRegisterE2ETester returned nil, router may be nil")
		}

		t.Run("valid_test_account", func(t *testing.T) {
			registerE2ETester.TestRegisterValidTestAccount(t)
		})
		t.Run("valid_user_account", func(t *testing.T) {
			registerE2ETester.TestRegisterValidUserAccount(t)
		})
		t.Run("no_name", func(t *testing.T) {
			registerE2ETester.TestRegisterNoName(t)
		})
		t.Run("name_without_number", func(t *testing.T) {
			registerE2ETester.TestRegisterNameWithoutNumber(t)
		})
		t.Run("short_name", func(t *testing.T) {
			registerE2ETester.TestRegisterShortName(t)
		})
		t.Run("invalid_email", func(t *testing.T) {
			registerE2ETester.TestRegisterInvalidEmail(t)
		})
		t.Run("short_password", func(t *testing.T) {
			registerE2ETester.TestRegisterShortPassword(t)
		})
		t.Run("password_without_lower_case_letter", func(t *testing.T) {
			registerE2ETester.TestRegisterPasswordWithoutLowerCaseLetter(t)
		})
		t.Run("password_without_upper_case_letter", func(t *testing.T) {
			registerE2ETester.TestRegisterPasswordWithoutUpperCaseLetter(t)
		})
		t.Run("password_without_number", func(t *testing.T) {
			registerE2ETester.TestRegisterPasswordWithoutNumber(t)
		})
		t.Run("password_without_sign", func(t *testing.T) {
			registerE2ETester.TestRegisterPasswordWithoutSign(t)
		})
	})

	t.Run("login", func(t *testing.T) {
		loginE2ETester := NewLoginE2ETester(server.URL, client)
		if loginE2ETester == nil {
			t.Fatal("NewLoginE2ETester returned nil, router may be nil")
		}

		t.Run("valid_test_account_by_name", func(t *testing.T) {
			loginE2ETester.TestLoginValidTestAccountByName(t)
		})
		t.Run("valid_test_account_by_email", func(t *testing.T) {
			loginE2ETester.TestLoginValidTestAccountByEmail(t)
		})
	})
}
