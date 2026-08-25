package testroutes

import (
	"github.com/gin-gonic/gin"

	scookies "github.com/HiIamJeff67/notegic-backend/shared/cookies"

	ratelimit "github.com/HiIamJeff67/notegic-backend/internal/clientgateway/ratelimit"
	binders "github.com/HiIamJeff67/notegic-backend/internal/clientgateway/transports/api/binders"
	controllers "github.com/HiIamJeff67/notegic-backend/internal/clientgateway/transports/api/controllers"
	middlewares "github.com/HiIamJeff67/notegic-backend/internal/clientgateway/transports/api/middlewares"
	coreadapters "github.com/HiIamJeff67/notegic-backend/internal/clientgateway/transports/core/adapters"
)

type AuthRouteDependencies struct {
	CoreAdapter               *coreadapters.CoreAdapter
	AccessTokenCookieHandler  *scookies.CookieHandler
	RefreshTokenCookieHandler *scookies.CookieHandler
	AuthorizedRateLimiter     *ratelimit.HybridRateLimiter
}

func testAuthorizedRateLimitMiddleware(rateLimiter *ratelimit.HybridRateLimiter) gin.HandlerFunc {
	if rateLimiter == nil {
		return func(ctx *gin.Context) { ctx.Next() }
	}
	return middlewares.AuthorizedRateLimitMiddleware(rateLimiter)
}

// the route structure is different here, since we use these routes to do the e2e test
// like it receive a database instance and a gin router group
// and its function name also start with the upper case letter
func ConfigureTestAuthRoutes(
	routerGroup *gin.RouterGroup,
	deps AuthRouteDependencies,
) {
	coreAdapter, accessTokenCookieHandler, refreshTokenCookieHandler, authorizedRateLimiter := deps.CoreAdapter, deps.AccessTokenCookieHandler, deps.RefreshTokenCookieHandler, deps.AuthorizedRateLimiter
	authBinder := binders.NewAuthBinder()
	authController := controllers.NewAuthController(
		coreAdapter,
		accessTokenCookieHandler,
		refreshTokenCookieHandler,
	)

	authRoutes := routerGroup.Group("/auth")
	{
		authRoutes.POST(
			"/register",
			authBinder.BindRegister(authController.Register),
		)
		authRoutes.POST(
			"/registerViaGoogle",
			authBinder.BindRegisterViaGoogle(authController.RegisterViaGoogle),
		)
		authRoutes.POST(
			"/login",
			authBinder.BindLogin(authController.Login),
		)
		authRoutes.POST(
			"/loginViaGoogle",
			authBinder.BindLoginViaGoogle(authController.LoginViaGoogle),
		)
		authRoutes.POST(
			"/logout",
			middlewares.JWTMiddleware(accessTokenCookieHandler, refreshTokenCookieHandler),
			testAuthorizedRateLimitMiddleware(authorizedRateLimiter),
			authBinder.BindLogout(authController.Logout),
		)
		authRoutes.POST(
			"/sendAuthCode",
			authBinder.BindSendAuthCode(authController.SendAuthCode),
		)
		authRoutes.PUT(
			"/validateEmail",
			middlewares.JWTMiddleware(accessTokenCookieHandler, refreshTokenCookieHandler),
			testAuthorizedRateLimitMiddleware(authorizedRateLimiter),
			authBinder.BindValidateEmail(authController.ValidateEmail),
		)
		authRoutes.PUT(
			"/resetEmail",
			middlewares.JWTMiddleware(accessTokenCookieHandler, refreshTokenCookieHandler),
			testAuthorizedRateLimitMiddleware(authorizedRateLimiter),
			authBinder.BindResetEmail(authController.ResetEmail),
		)
		authRoutes.PUT(
			"/forgetPassword",
			authBinder.BindForgetPassword(authController.ForgetPassword),
		)
		authRoutes.PUT(
			"/resetMe",
			middlewares.JWTMiddleware(accessTokenCookieHandler, refreshTokenCookieHandler),
			testAuthorizedRateLimitMiddleware(authorizedRateLimiter),
			authBinder.BindResetMe(authController.ResetMe),
		)
		authRoutes.DELETE(
			"/deleteMe",
			middlewares.JWTMiddleware(accessTokenCookieHandler, refreshTokenCookieHandler),
			testAuthorizedRateLimitMiddleware(authorizedRateLimiter),
			authBinder.BindDeleteMe(authController.DeleteMe),
		)
	}
}
