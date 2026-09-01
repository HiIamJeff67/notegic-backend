package routers

import (
	"github.com/gin-gonic/gin"

	cgateway "github.com/HiIamJeff67/notegic-backend/contracts/gateway/v1"

	slogs "github.com/HiIamJeff67/notegic-backend/shared/platform/observability/logs"

	middlewares "github.com/HiIamJeff67/notegic-backend/runtimes/core/transports/gateway/middlewares"
)

type RouterDependencies struct {
	Auth              AuthRouterDependencies
	APIKey            APIKeyRouterDependencies
	RootShelf         RootShelfRouterDependencies
	Station           StationRouterDependencies
	UserSetting       UserSettingRouterDependencies
	UserInfo          UserInfoRouterDependencies
	UserAccount       UserAccountRouterDependencies
	User              UserRouterDependencies
	Block             BlockRouterDependencies
	Realtime          RealtimeRouterDependencies
	RoutineTag        RoutineTagRouterDependencies
	RoutineRecord     RoutineRecordRouterDependencies
	RoutineTaskRecord RoutineTaskRecordRouterDependencies
	SubShelf          SubShelfRouterDependencies
	BlockPack         BlockPackRouterDependencies
	Material          MaterialRouterDependencies
	Routine           RoutineRouterDependencies
	RoutineTask       RoutineTaskRouterDependencies
	Theme             ThemeRouterDependencies
	Item              ItemRouterDependencies
	Badge             BadgeRouterDependencies
}

func NewRouter(deps RouterDependencies) *gin.Engine {
	router := slogs.WithGinLogger(gin.New())
	router.Use(middlewares.TokenResponseMiddleware())

	coreRouterGroup := router.Group("/core/" + cgateway.Version)
	anonymousCoreRouterGroup := coreRouterGroup.Group("")
	secureCoreRouterGroup := coreRouterGroup.Group("")

	configureAnonymousAuthRoutes(anonymousCoreRouterGroup, deps.Auth)
	configureAuthenticatedAuthRoutes(secureCoreRouterGroup, deps.Auth)
	configureAPIKeyRoutes(secureCoreRouterGroup, deps.APIKey)
	configureRootShelfRoutes(secureCoreRouterGroup, deps.RootShelf)
	configureStationRoutes(secureCoreRouterGroup, deps.Station)
	configureUserSettingRoutes(secureCoreRouterGroup, deps.UserSetting)
	configureUserInfoRoutes(secureCoreRouterGroup, deps.UserInfo)
	configureUserAccountRoutes(secureCoreRouterGroup, deps.UserAccount)
	configureUserRoutes(secureCoreRouterGroup, deps.User)
	configureBlockRoutes(secureCoreRouterGroup, deps.Block)
	configureRealtimeRoutes(secureCoreRouterGroup, deps.Realtime)
	configureRoutineTagRoutes(secureCoreRouterGroup, deps.RoutineTag)
	configureRoutineRecordRoutes(secureCoreRouterGroup, deps.RoutineRecord)
	configureRoutineTaskRecordRoutes(secureCoreRouterGroup, deps.RoutineTaskRecord)
	configureSubShelfRoutes(secureCoreRouterGroup, deps.SubShelf)
	configureBlockPackRoutes(secureCoreRouterGroup, deps.BlockPack)
	configureMaterialRoutes(secureCoreRouterGroup, deps.Material)
	configureRoutineRoutes(secureCoreRouterGroup, deps.Routine)
	configureRoutineTaskRoutes(secureCoreRouterGroup, deps.RoutineTask)
	configureThemeRoutes(anonymousCoreRouterGroup, deps.Theme)
	configureItemRoutes(secureCoreRouterGroup, deps.Item)
	configureBadgeRoutes(secureCoreRouterGroup, deps.Badge)

	return router
}
