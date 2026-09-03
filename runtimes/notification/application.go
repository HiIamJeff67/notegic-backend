package notification

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	validator "github.com/go-playground/validator/v10"
	"gorm.io/gorm"

	ctypes "github.com/HiIamJeff67/notegic-backend/contracts/types"

	sobservability "github.com/HiIamJeff67/notegic-backend/shared/platform/observability"
	slogs "github.com/HiIamJeff67/notegic-backend/shared/platform/observability/logs"
	spostgres "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres"
	srepositories "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories"
	svalidations "github.com/HiIamJeff67/notegic-backend/shared/validations"

	configs "github.com/HiIamJeff67/notegic-backend/runtimes/notification/configs"
	notificationexceptions "github.com/HiIamJeff67/notegic-backend/runtimes/notification/exceptions"
	services "github.com/HiIamJeff67/notegic-backend/runtimes/notification/services"
	consumers "github.com/HiIamJeff67/notegic-backend/runtimes/notification/transports/core/consumers"
	endpoints "github.com/HiIamJeff67/notegic-backend/runtimes/notification/transports/gateway/endpoints"
	routers "github.com/HiIamJeff67/notegic-backend/runtimes/notification/transports/gateway/routers"
	validations "github.com/HiIamJeff67/notegic-backend/runtimes/notification/validations"
	workers "github.com/HiIamJeff67/notegic-backend/runtimes/notification/workers"
)

type Application struct {
	healthy atomic.Bool
	ready   atomic.Bool
}

type ApplicationInterface interface {
	Start() func()
	IsHealthy() bool
	IsReady() bool
	initializeObservability() func()
	initializeDatabase(spostgres.Config) (*gorm.DB, error)
	initializeService(*gorm.DB) services.NotificationServiceInterface
	initializeWorkers(configs.Config, services.NotificationServiceInterface) func()
	buildRouter(services.NotificationServiceInterface) *gin.Engine
	startHTTP(configs.Config, *gin.Engine) (func(), error)
}

func NewApplication() *Application {
	return &Application{}
}

func (a *Application) IsHealthy() bool {
	return a.healthy.Load()
}

func (a *Application) IsReady() bool {
	return a.ready.Load()
}

func (a *Application) initializeObservability() func() {
	return sobservability.Initialize(
		context.Background(),
		sobservability.LoadConfig("notegic-notification"),
	)
}

func (a *Application) initializeDatabase(
	config spostgres.Config,
) (*gorm.DB, error) {
	if config.User != ctypes.Runtime_Notification.RoleName() {
		return nil, fmt.Errorf("Notification PostgreSQL user must be %q", ctypes.Runtime_Notification.RoleName())
	}
	db, err := spostgres.Connect(config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect Notification database: %w", err)
	}
	return db, nil
}

func (a *Application) initializeService(db *gorm.DB) services.NotificationServiceInterface {
	repository := srepositories.NewNotificationRepository(
		db,
		srepositories.NewUserProjectionRepository(db),
		srepositories.NewInboxEventRepository(),
	)
	notificationValidator := validator.New()
	svalidations.RegisterStringsValidation(notificationValidator)
	svalidations.RegisterTimesValidation(notificationValidator)
	validations.RegisterNotificationValidation(notificationValidator)
	validations.RegisterNewsValidation(notificationValidator)
	validations.RegisterWarningValidation(notificationValidator)
	validations.RegisterImportantValidation(notificationValidator)
	return services.NewNotificationService(
		repository,
		notificationValidator,
		notificationexceptions.NewEventException("Notification"),
		notificationexceptions.NewRequestException("Notification"),
		notificationexceptions.NewOperationException("Notification"),
		notificationexceptions.NewPayloadException("Notification"),
	)
}

func (a *Application) initializeWorkers(
	config configs.Config,
	service services.NotificationServiceInterface,
) func() {
	consumer := consumers.NewNotificationRequestConsumer(service, config.Kafka.ConsumerConfig())
	cleanup := workers.NewCleanupWorker(
		service,
		config.NotificationCleanupInterval,
		config.NotificationRetention,
	)
	shutdownConsumer := consumer.Start(context.Background())
	shutdownCleanup := cleanup.Start(context.Background())
	return func() {
		shutdownCleanup()
		shutdownConsumer()
	}
}

func (a *Application) buildRouter(service services.NotificationServiceInterface) *gin.Engine {
	router := slogs.WithGinLogger(gin.New())
	router.GET("/healthz", func(ctx *gin.Context) {
		if !a.IsReady() {
			ctx.Status(http.StatusServiceUnavailable)
			return
		}
		ctx.Status(http.StatusNoContent)
	})
	router.GET("/startedz", func(ctx *gin.Context) {
		if !a.IsHealthy() {
			ctx.Status(http.StatusServiceUnavailable)
			return
		}
		ctx.Status(http.StatusNoContent)
	})
	endpoint := endpoints.NewNotificationEndpoint(service)
	routers.ConfigureNotificationRoutes(router.Group("/runtimes/v1"), endpoint)
	return router
}

func (a *Application) startHTTP(
	config configs.Config,
	router *gin.Engine,
) (func(), error) {
	listener, err := net.Listen("tcp", config.ListenAddress)
	if err != nil {
		return nil, err
	}
	a.healthy.Store(true)
	a.ready.Store(true)
	server := &http.Server{Handler: router}
	go func() {
		if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
	}()

	return func() {
		a.ready.Store(false)
		a.healthy.Store(false)
		shutdownContext, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownContext); err != nil {
			fmt.Printf("Failed to shutdown Notification server: %v\n", err)
		}
	}, nil
}

func (a *Application) Start() func() {
	shutdownObservability := a.initializeObservability()
	var (
		db              *gorm.DB
		shutdownHTTP    func()
		shutdownWorkers func()
	)
	var shutdownOnce sync.Once
	shutdown := func() {
		shutdownOnce.Do(func() {
			if shutdownHTTP != nil {
				shutdownHTTP()
			}
			if shutdownWorkers != nil {
				shutdownWorkers()
			}
			if db != nil {
				if err := spostgres.Disconnect(db); err != nil {
					fmt.Printf("Failed to disconnect Notification database: %v\n", err)
				}
			}
			shutdownObservability()
		})
	}
	fail := func(err error) {
		shutdown()
		panic(err)
	}

	config, err := configs.LoadConfig()
	if err != nil {
		fail(err)
	}
	db, err = a.initializeDatabase(config.Postgres)
	if err != nil {
		fail(err)
	}
	service := a.initializeService(db)
	shutdownWorkers = a.initializeWorkers(config, service)
	router := a.buildRouter(service)
	shutdownHTTP, err = a.startHTTP(config, router)
	if err != nil {
		fail(err)
	}
	return shutdown
}

// make sure Application struct followed the ApplicationInterface implementations
var _ ApplicationInterface = (*Application)(nil)
