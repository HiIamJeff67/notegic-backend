package email

import (
	"context"
	"net"
	"net/http"
	"sync/atomic"
	"time"

	validator "github.com/go-playground/validator/v10"

	sobservability "github.com/HiIamJeff67/notegic-backend/shared/platform/observability"

	emailbuilders "github.com/HiIamJeff67/notegic-backend/runtimes/email/builders"
	renderers "github.com/HiIamJeff67/notegic-backend/runtimes/email/builders/renderers"
	emailconfig "github.com/HiIamJeff67/notegic-backend/runtimes/email/configs"
	coretransport "github.com/HiIamJeff67/notegic-backend/runtimes/email/transports/core"
	status "github.com/HiIamJeff67/notegic-backend/runtimes/email/transports/status"
)

type Application struct {
	healthy atomic.Bool
	ready   atomic.Bool
}

type ApplicationInterface interface {
	Start() func()
	IsHealthy() bool
	IsReady() bool
	loadConfig() emailconfig.Config
	initializeObservability() func()
	initializeWorkers(emailconfig.Config, func()) func()
	buildRouter() *http.ServeMux
	startHTTP(emailconfig.Config, *http.ServeMux, func(), func()) func()
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

func (a *Application) loadConfig() emailconfig.Config {
	config, err := emailconfig.LoadConfig()
	if err != nil {
		panic(err)
	}
	return config
}

func (a *Application) initializeObservability() func() {
	return sobservability.Initialize(
		context.Background(),
		sobservability.LoadConfig("notegic-email"),
	)
}

func (a *Application) initializeWorkers(
	config emailconfig.Config,
	shutdownObservability func(),
) func() {
	// Initialize renderers, the bounded delivery queue, and the Kafka consumer.
	deliverySender := NewEmailSender(config.SMTP)
	emailWorkerManager := NewEmailWorkerManager(16, deliverySender)
	welcomeRenderer, err := renderers.NewWelcomeEmailRenderer(config.Renderers.Welcome)
	if err != nil {
		emailWorkerManager.Shutdown()
		shutdownObservability()
		panic(err)
	}
	validationRenderer, err := renderers.NewValidationEmailRenderer(config.Renderers.Validation)
	if err != nil {
		emailWorkerManager.Shutdown()
		shutdownObservability()
		panic(err)
	}
	securityAlertRenderer, err := renderers.NewSecurityAlertEmailRenderer(config.Renderers.SecurityAlert)
	if err != nil {
		emailWorkerManager.Shutdown()
		shutdownObservability()
		panic(err)
	}
	welcomeBuilder := emailbuilders.NewWelcomeEmailBuilder(welcomeRenderer)
	validationBuilder := emailbuilders.NewValidationEmailBuilder(validationRenderer)
	securityAlertBuilder := emailbuilders.NewSecurityAlertEmailBuilder(securityAlertRenderer)
	emailRequestConsumer := coretransport.NewEmailRequestConsumer(
		welcomeBuilder,
		validationBuilder,
		securityAlertBuilder,
		emailWorkerManager,
		validator.New(),
		config.KafkaConsumer,
	)
	shutdownRequestConsumer := emailRequestConsumer.Start(context.Background())
	return func() {
		shutdownRequestConsumer()
		emailWorkerManager.Shutdown()
	}
}

func (a *Application) buildRouter() *http.ServeMux {
	mux := http.NewServeMux()
	status.ConfigureStartedRouter(mux, a.IsHealthy)
	status.ConfigureHealthRouter(mux, a.IsReady)
	return mux
}

func (a *Application) startHTTP(
	config emailconfig.Config,
	mux *http.ServeMux,
	shutdownWorkers func(),
	shutdownObservability func(),
) func() {
	// Bind the health server after the email consumer is running.
	listener, err := net.Listen("tcp", config.ListenAddress)
	if err != nil {
		shutdownWorkers()
		shutdownObservability()
		panic(err)
	}
	server := &http.Server{Handler: mux}
	a.healthy.Store(true)
	a.ready.Store(true)
	go func() {
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	return func() {
		// Drain email work after HTTP traffic has stopped, then release sobservability.
		a.ready.Store(false)
		a.healthy.Store(false)
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			panic(err)
		}
		shutdownWorkers()
		shutdownObservability()
	}
}

func (a *Application) Start() func() {
	shutdownObservability := a.initializeObservability()
	config := a.loadConfig()
	shutdownWorkers := a.initializeWorkers(config, shutdownObservability)
	router := a.buildRouter()
	return a.startHTTP(config, router, shutdownWorkers, shutdownObservability)
}

// make sure Application struct followed the ApplicationInterface implementations
var _ ApplicationInterface = (*Application)(nil)
