package notification

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	validator "github.com/go-playground/validator/v10"
	"gorm.io/gorm"

	types "github.com/HiIamJeff67/notegic-backend/contracts/types"
	platformkafka "github.com/HiIamJeff67/notegic-backend/shared/platform/kafka"
	observability "github.com/HiIamJeff67/notegic-backend/shared/platform/observability"
	logs "github.com/HiIamJeff67/notegic-backend/shared/platform/observability/logs"
	platformpostgres "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres"
	sharedvalidations "github.com/HiIamJeff67/notegic-backend/shared/validations"

	configs "github.com/HiIamJeff67/notegic-backend/internal/notification/configs"
	postgres "github.com/HiIamJeff67/notegic-backend/internal/notification/data/postgres"
	repositories "github.com/HiIamJeff67/notegic-backend/internal/notification/data/postgres/repositories"
	services "github.com/HiIamJeff67/notegic-backend/internal/notification/services"
	notificationtransports "github.com/HiIamJeff67/notegic-backend/internal/notification/transports"
	consumers "github.com/HiIamJeff67/notegic-backend/internal/notification/transports/core/consumers"
	endpoints "github.com/HiIamJeff67/notegic-backend/internal/notification/transports/gateway/endpoints"
	routers "github.com/HiIamJeff67/notegic-backend/internal/notification/transports/gateway/routers"
	validations "github.com/HiIamJeff67/notegic-backend/internal/notification/validations"
	workers "github.com/HiIamJeff67/notegic-backend/internal/notification/workers"
)

type Application struct {
	healthy atomic.Bool
	ready   atomic.Bool
}

type ApplicationInterface interface {
	Start() func()
	IsHealthy() bool
	IsReady() bool
	loadConfig() configs.Config
	initializeObservability() func()
	initializeDatabase(platformpostgres.Config, func()) *gorm.DB
	initializeKafka(platformkafka.ConnectionConfig, *gorm.DB, func()) *platformkafka.Producer
	initializeService(*gorm.DB) services.NotificationServiceInterface
	initializeWorkers(configs.Config, services.NotificationServiceInterface, *gorm.DB, *platformkafka.Producer) func()
	buildRouter(services.NotificationServiceInterface) *gin.Engine
	startHTTP(configs.Config, *gin.Engine, func(), *gorm.DB, *platformkafka.Producer, func()) func()
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

func (a *Application) loadConfig() configs.Config {
	config, err := configs.LoadConfig()
	if err != nil {
		panic(err)
	}
	return config
}

func (a *Application) initializeObservability() func() {
	return observability.Initialize(
		context.Background(),
		observability.LoadConfig("notegic-notification"),
	)
}

func (a *Application) initializeDatabase(config platformpostgres.Config, shutdownObservability func()) *gorm.DB {
	db, err := postgres.Connect(config)
	if err != nil {
		shutdownObservability()
		panic(err)
	}
	if err := platformpostgres.Migrate(
		db,
		types.Runtime_Notification,
		postgres.DatabaseMigrationManifest,
	); err != nil {
		_ = postgres.Disconnect(db)
		shutdownObservability()
		panic(fmt.Errorf("failed to initialize Notification database schema: %w", err))
	}
	return db
}

func (a *Application) initializeKafka(
	config platformkafka.ConnectionConfig,
	db *gorm.DB,
	shutdownObservability func(),
) *platformkafka.Producer {
	producer, err := platformkafka.NewProducer(platformkafka.ClientConfig{
		ConnectionConfig: config,
		ClientId:         "notegic-notification-producer",
	})
	if err != nil {
		_ = postgres.Disconnect(db)
		shutdownObservability()
		panic(err)
	}
	return producer
}

func (a *Application) initializeService(db *gorm.DB) services.NotificationServiceInterface {
	repository := repositories.NewNotificationRepository(db)
	notificationValidator := validator.New()
	sharedvalidations.RegisterStringsValidation(notificationValidator)
	sharedvalidations.RegisterTimesValidation(notificationValidator)
	validations.RegisterNotificationValidation(notificationValidator)
	validations.RegisterNewsValidation(notificationValidator)
	validations.RegisterWarningValidation(notificationValidator)
	validations.RegisterImportantValidation(notificationValidator)
	return services.NewNotificationService(repository, notificationValidator)
}

func (a *Application) initializeWorkers(
	config configs.Config,
	service services.NotificationServiceInterface,
	db *gorm.DB,
	producer *platformkafka.Producer,
) func() {
	repository := repositories.NewNotificationRepository(db)
	consumer := consumers.NewNotificationRequestConsumer(service, config.Kafka.ConsumerConfig())
	relay := notificationtransports.NewOutboxRelay(
		repository,
		producer,
		config.OutboxPollInterval,
		config.OutboxClaimTimeout,
		config.OutboxInitialBackoff,
		config.OutboxMaximumBackoff,
		config.OutboxBatchSize,
		config.OutboxCleanupInterval,
		config.OutboxRetention,
	)
	cleanup := workers.NewCleanupWorker(
		service,
		config.OutboxCleanupInterval,
		config.NotificationRetention,
	)
	shutdownConsumer := consumer.Start(context.Background())
	shutdownRelay := relay.Start(context.Background())
	shutdownCleanup := cleanup.Start(context.Background())
	return func() {
		shutdownCleanup()
		shutdownRelay()
		shutdownConsumer()
	}
}

func (a *Application) buildRouter(service services.NotificationServiceInterface) *gin.Engine {
	router := logs.WithGinLogger(gin.New())
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
	routers.ConfigureNotificationRoutes(router.Group("/internal/v1"), endpoint)
	return router
}

func (a *Application) startHTTP(
	config configs.Config,
	router *gin.Engine,
	shutdownWorkers func(),
	db *gorm.DB,
	producer *platformkafka.Producer,
	shutdownObservability func(),
) func() {
	listener, err := net.Listen("tcp", config.ListenAddress)
	if err != nil {
		shutdownWorkers()
		producer.Close()
		_ = postgres.Disconnect(db)
		shutdownObservability()
		panic(err)
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
		// Stop background workers before closing the HTTP, Kafka, and database resources.
		a.ready.Store(false)
		a.healthy.Store(false)
		shutdownWorkers()
		shutdownContext, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownContext); err != nil {
			fmt.Printf("Failed to shutdown Notification server: %v\n", err)
		}
		producer.Close()
		if err := postgres.Disconnect(db); err != nil {
			fmt.Printf("Failed to disconnect Notification database: %v\n", err)
		}
		shutdownObservability()
	}
}

func (a *Application) Start() func() {
	shutdownObservability := a.initializeObservability()
	config := a.loadConfig()
	db := a.initializeDatabase(config.Postgres, shutdownObservability)
	producer := a.initializeKafka(config.Kafka.Connection, db, shutdownObservability)
	service := a.initializeService(db)
	shutdownWorkers := a.initializeWorkers(config, service, db, producer)
	router := a.buildRouter(service)
	return a.startHTTP(config, router, shutdownWorkers, db, producer, shutdownObservability)
}

// make sure Application struct followed the ApplicationInterface implementations
var _ ApplicationInterface = (*Application)(nil)
