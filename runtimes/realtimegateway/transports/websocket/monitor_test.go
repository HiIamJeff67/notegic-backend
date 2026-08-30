package websocket

import (
	"os"
	"testing"

	"go.opentelemetry.io/otel"

	slogs "github.com/HiIamJeff67/notegic-backend/shared/platform/observability/logs"
	smetrics "github.com/HiIamJeff67/notegic-backend/shared/platform/observability/metrics"
	straces "github.com/HiIamJeff67/notegic-backend/shared/platform/observability/traces"
)

func TestMain(m *testing.M) {
	slogs.NotegicLogger = slogs.NewLogger(true)
	smetrics.NotegicMeter = smetrics.NewMeter(otel.Meter("realtime.test"))
	straces.NotegicTracer = straces.NewTracer(otel.Tracer("realtime.test"))

	os.Exit(m.Run())
}
