package middlewares

import "github.com/gin-gonic/gin"

// EitherMiddleware keeps route registration declarative while selecting a middleware
// branch per request. Each returned wrapper occupies one middleware slot, so
// selected handlers may safely call ctx.Next() just like ordinary Gin
// middleware. The condition must be derived from verified request context,
// never from a startup-time gateway flag.
func EitherMiddleware(
	passed []gin.HandlerFunc,
	failed []gin.HandlerFunc,
	condition func(*gin.Context) bool,
) []gin.HandlerFunc {
	length := len(passed)
	if len(failed) > length {
		length = len(failed)
	}
	if length == 0 {
		return nil
	}
	if condition == nil {
		condition = func(*gin.Context) bool { return false }
	}

	handlers := make([]gin.HandlerFunc, length)
	for index := range handlers {
		index := index
		handlers[index] = func(ctx *gin.Context) {
			selected := failed
			if condition(ctx) {
				selected = passed
			}
			if index >= len(selected) {
				ctx.Next()
				return
			}
			selected[index](ctx)
		}
	}
	return handlers
}
