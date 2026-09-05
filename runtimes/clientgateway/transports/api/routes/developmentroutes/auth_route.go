package developmentroutes

import (
	"time"

	"github.com/gin-gonic/gin"

	scookies "github.com/HiIamJeff67/notegic-backend/shared/cookies"

	ratelimit "github.com/HiIamJeff67/notegic-backend/runtimes/clientgateway/ratelimit"
	binders "github.com/HiIamJeff67/notegic-backend/runtimes/clientgateway/transports/api/binders"
	controllers "github.com/HiIamJeff67/notegic-backend/runtimes/clientgateway/transports/api/controllers"
	interceptors "github.com/HiIamJeff67/notegic-backend/runtimes/clientgateway/transports/api/interceptors"
	middlewares "github.com/HiIamJeff67/notegic-backend/runtimes/clientgateway/transports/api/middlewares"
	coreadapters "github.com/HiIamJeff67/notegic-backend/runtimes/clientgateway/transports/core/adapters"
)

type AuthRouteDependencies struct {
	CoreAdapter               *coreadapters.CoreAdapter
	AccessTokenCookieHandler  *scookies.CookieHandler
	RefreshTokenCookieHandler *scookies.CookieHandler
	UnauthorizedRateLimiter   *ratelimit.HybridRateLimiter
}

func configureDevelopmentAuthRoutes(
	router *gin.RouterGroup,
	deps AuthRouteDependencies,
) {
	coreAdapter, accessTokenCookieHandler, refreshTokenCookieHandler, unauthorizedRateLimiter := deps.CoreAdapter, deps.AccessTokenCookieHandler, deps.RefreshTokenCookieHandler, deps.UnauthorizedRateLimiter

	authBinder := binders.NewAuthBinder()
	authController := controllers.NewAuthController(
		coreAdapter,
		accessTokenCookieHandler,
		refreshTokenCookieHandler,
	)

	authRoutes := router.Group("/auth")
	{
		authRoutes.POST(
			"/register",
			middlewares.ApplyTracerMiddleware("register"),
			middlewares.ApplyMeterMiddleware("server.requests.auth.register"),
			middlewares.UnauthorizedRateLimitMiddleware(unauthorizedRateLimiter),
			middlewares.TimeoutMiddleware(5*time.Second),
			authBinder.BindRegister(authController.Register),
		)
		authRoutes.POST(
			"/register-via-google",
			middlewares.ApplyTracerMiddleware("registerViaGoogle"),
			middlewares.ApplyMeterMiddleware("server.requests.auth.registerViaGoogle"),
			middlewares.UnauthorizedRateLimitMiddleware(unauthorizedRateLimiter),
			middlewares.TimeoutMiddleware(5*time.Second),
			authBinder.BindRegisterViaGoogle(authController.RegisterViaGoogle),
		)
		authRoutes.POST(
			"/login",
			middlewares.ApplyTracerMiddleware("login"),
			middlewares.ApplyMeterMiddleware("server.requests.auth.login"),
			middlewares.UnauthorizedRateLimitMiddleware(unauthorizedRateLimiter),
			middlewares.TimeoutMiddleware(3*time.Second),
			authBinder.BindLogin(authController.Login),
		)
		authRoutes.POST(
			"/login-via-google",
			middlewares.ApplyTracerMiddleware("loginViaGoogle"),
			middlewares.ApplyMeterMiddleware("server.requests.auth.loginViaGoogle"),
			middlewares.UnauthorizedRateLimitMiddleware(unauthorizedRateLimiter),
			middlewares.TimeoutMiddleware(3*time.Second),
			authBinder.BindLoginViaGoogle(authController.LoginViaGoogle),
		)
		authRoutes.POST(
			"/logout",
			middlewares.ApplyTracerMiddleware("logout"),
			middlewares.ApplyMeterMiddleware("server.requests.auth.logout"),
			middlewares.UnauthorizedRateLimitMiddleware(unauthorizedRateLimiter),
			middlewares.TimeoutMiddleware(3*time.Second),
			middlewares.GatewayAuthenticationMiddleware(accessTokenCookieHandler, refreshTokenCookieHandler),
			interceptors.ShareableResponseWriterInterceptor(
				interceptors.EmbeddedInterceptor,
			),
			authBinder.BindLogout(authController.Logout),
		)
		authRoutes.POST(
			"/send-auth-code",
			middlewares.ApplyTracerMiddleware("sendAuthCode"),
			middlewares.ApplyMeterMiddleware("server.requests.auth.sendAuthCode"),
			middlewares.UnauthorizedRateLimitMiddleware(unauthorizedRateLimiter),
			middlewares.TimeoutMiddleware(3*time.Second),
			authBinder.BindSendAuthCode(authController.SendAuthCode),
		)
		authRoutes.PUT(
			"/validate-email",
			middlewares.ApplyTracerMiddleware("validateEmail"),
			middlewares.ApplyMeterMiddleware("server.requests.auth.validateEmail"),
			middlewares.UnauthorizedRateLimitMiddleware(unauthorizedRateLimiter),
			middlewares.TimeoutMiddleware(3*time.Second),
			middlewares.GatewayAuthenticationMiddleware(accessTokenCookieHandler, refreshTokenCookieHandler),
			interceptors.ShareableResponseWriterInterceptor(
				interceptors.RefreshTokenInterceptor(accessTokenCookieHandler),
				interceptors.EmbeddedInterceptor,
			),
			authBinder.BindValidateEmail(authController.ValidateEmail),
		)
		authRoutes.PUT(
			"/reset-email",
			middlewares.ApplyTracerMiddleware("resetEmail"),
			middlewares.ApplyMeterMiddleware("server.requests.auth.resetEmail"),
			middlewares.UnauthorizedRateLimitMiddleware(unauthorizedRateLimiter),
			middlewares.TimeoutMiddleware(3*time.Second),
			middlewares.GatewayAuthenticationMiddleware(accessTokenCookieHandler, refreshTokenCookieHandler),
			interceptors.ShareableResponseWriterInterceptor(
				interceptors.RefreshTokenInterceptor(accessTokenCookieHandler),
				interceptors.EmbeddedInterceptor,
			),
			authBinder.BindResetEmail(authController.ResetEmail),
		)
		authRoutes.PUT(
			"/forget-password",
			middlewares.ApplyTracerMiddleware("forgetPassword"),
			middlewares.ApplyMeterMiddleware("server.requests.auth.forgetPassword"),
			middlewares.UnauthorizedRateLimitMiddleware(unauthorizedRateLimiter),
			middlewares.TimeoutMiddleware(3*time.Second),
			authBinder.BindForgetPassword(authController.ForgetPassword),
		)
		authRoutes.PUT(
			"/reset-me",
			middlewares.ApplyTracerMiddleware("resetMe"),
			middlewares.ApplyMeterMiddleware("server.requests.auth.resetMe"),
			middlewares.UnauthorizedRateLimitMiddleware(unauthorizedRateLimiter),
			middlewares.TimeoutMiddleware(3*time.Second),
			middlewares.GatewayAuthenticationMiddleware(accessTokenCookieHandler, refreshTokenCookieHandler),
			interceptors.ShareableResponseWriterInterceptor(
				interceptors.RefreshTokenInterceptor(accessTokenCookieHandler),
				interceptors.EmbeddedInterceptor,
			),
			authBinder.BindResetMe(authController.ResetMe),
		)
		authRoutes.DELETE(
			"/delete-me",
			middlewares.ApplyTracerMiddleware("deleteMe"),
			middlewares.ApplyMeterMiddleware("server.requests.auth.deleteMe"),
			middlewares.UnauthorizedRateLimitMiddleware(unauthorizedRateLimiter),
			middlewares.TimeoutMiddleware(5*time.Second),
			middlewares.GatewayAuthenticationMiddleware(accessTokenCookieHandler, refreshTokenCookieHandler),
			interceptors.ShareableResponseWriterInterceptor(
				interceptors.EmbeddedInterceptor,
			),
			authBinder.BindDeleteMe(authController.DeleteMe),
		)
	}
}
