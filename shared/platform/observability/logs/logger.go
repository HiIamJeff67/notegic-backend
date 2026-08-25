package logs

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"go.opentelemetry.io/otel/attribute"
	otellog "go.opentelemetry.io/otel/log"
	logglobal "go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/trace"

	traces "github.com/HiIamJeff67/notegic-backend/shared/platform/observability/traces"
)

type LoggerInterface interface {
	Debug(ctx context.Context, message string, attributes ...attribute.KeyValue)
	Info(ctx context.Context, message string, attributes ...attribute.KeyValue)
	Warn(ctx context.Context, message string, attributes ...attribute.KeyValue)
	Error(ctx context.Context, err error, message string, attributes ...attribute.KeyValue)
	Alert(ctx context.Context, err error, message string, attributes ...attribute.KeyValue)
	JSON(ctx context.Context, level slog.Level, message string, payload any, attributes ...attribute.KeyValue) error
}

type Logger struct {
	logger   *slog.Logger
	console  *CommandLineInterfaceLogger
	emitOTel bool
}

func NewLogger(emitOtel bool) LoggerInterface {
	logger := &Logger{
		logger:   slog.New(slog.NewJSONHandler(os.Stdout, nil)),
		emitOTel: emitOtel,
	}
	if ConsoleLoggingEnabled() {
		logger.console = NewCommandLineInterfaceLogger().(*CommandLineInterfaceLogger)
	}
	return logger
}

func ConsoleLoggingEnabled() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("NOTEGIC_LOG_FORMAT"))) {
	case "console", "text":
		return true
	case "json":
		return false
	default:
		return strings.EqualFold(strings.TrimSpace(os.Getenv("OTEL_DEPLOYMENT_ENVIRONMENT")), "development")
	}
}

var (
	NotegicLogger LoggerInterface
)

func (l *Logger) Debug(ctx context.Context, message string, attributes ...attribute.KeyValue) {
	l.write(ctx, slog.LevelDebug, otellog.SeverityDebug, nil, message, attributes...)
}

func (l *Logger) Info(ctx context.Context, message string, attributes ...attribute.KeyValue) {
	l.write(ctx, slog.LevelInfo, otellog.SeverityInfo, nil, message, attributes...)
}

func (l *Logger) Warn(ctx context.Context, message string, attributes ...attribute.KeyValue) {
	l.write(ctx, slog.LevelWarn, otellog.SeverityWarn, nil, message, attributes...)
}

func (l *Logger) Error(ctx context.Context, err error, message string, attributes ...attribute.KeyValue) {
	l.write(ctx, slog.LevelError, otellog.SeverityError, err, message, attributes...)
}

func (l *Logger) Alert(ctx context.Context, err error, message string, attributes ...attribute.KeyValue) {
	l.write(ctx, slog.LevelError, otellog.SeverityFatal, err, message, attributes...)
}

func (l *Logger) JSON(
	ctx context.Context,
	level slog.Level,
	message string,
	payload any,
	attributes ...attribute.KeyValue,
) error {
	marshaledPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	attributes = append(attributes, attribute.String("payload", string(marshaledPayload)))
	if l.console != nil {
		prettyPayload, err := json.MarshalIndent(payload, "", "  ")
		if err != nil {
			return err
		}
		l.writeLocal(ctx, level, nil, message+"\n"+string(prettyPayload), attributes...)
	} else {
		l.writeLocal(ctx, level, nil, message, attributes...)
	}
	l.emitOTelRecord(ctx, level, severityFromLevel(level), message, attributes...)

	return nil
}

func severityFromLevel(level slog.Level) otellog.Severity {
	switch {
	case level <= slog.LevelDebug:
		return otellog.SeverityDebug
	case level <= slog.LevelInfo:
		return otellog.SeverityInfo
	case level <= slog.LevelWarn:
		return otellog.SeverityWarn
	default:
		return otellog.SeverityError
	}
}

func (l *Logger) write(
	ctx context.Context,
	level slog.Level,
	severity otellog.Severity,
	err error,
	message string,
	attributes ...attribute.KeyValue,
) {
	if ctx == nil {
		ctx = context.Background()
	}

	caller := traces.GetTrace(2)
	attributes = append(attributes,
		attribute.String("code.filepath", caller.File),
		attribute.Int("code.lineno", caller.Line),
		attribute.String("code.function", caller.Function),
	)
	if err != nil {
		attributes = append(attributes,
			attribute.String("error.type", typeName(err)),
			attribute.String("error.message", err.Error()),
		)
	}

	spanContext := trace.SpanContextFromContext(ctx)
	if spanContext.IsValid() {
		attributes = append(attributes,
			attribute.String("trace.id", spanContext.TraceID().String()),
			attribute.String("span.id", spanContext.SpanID().String()),
		)
	}

	l.writeLocal(ctx, level, err, message, attributes...)
	l.emitOTelRecord(ctx, level, severity, message, attributes...)
}

func (l *Logger) writeLocal(
	ctx context.Context,
	level slog.Level,
	err error,
	message string,
	attributes ...attribute.KeyValue,
) {
	if l.console != nil {
		l.console.write(level.String(), err, message, attributes...)
		return
	}

	slogAttributes := make([]slog.Attr, 0, len(attributes))
	for _, item := range attributes {
		slogAttributes = append(slogAttributes, slog.Any(string(item.Key), item.Value.AsInterface()))
	}
	l.logger.LogAttrs(ctx, level, message, slogAttributes...)
}

func (l *Logger) emitOTelRecord(
	ctx context.Context,
	level slog.Level,
	severity otellog.Severity,
	message string,
	attributes ...attribute.KeyValue,
) {
	if !l.emitOTel {
		return
	}
	if ctx == nil {
		ctx = context.Background()
	}

	otelAttributes := make([]otellog.KeyValue, 0, len(attributes))
	for _, item := range attributes {
		otelAttributes = append(otelAttributes, otellog.KeyValueFromAttribute(item))
	}

	var record otellog.Record
	record.SetTimestamp(time.Now())
	record.SetSeverity(severity)
	record.SetSeverityText(severity.String())
	record.SetBody(otellog.StringValue(message))
	record.AddAttributes(otelAttributes...)
	logglobal.Logger("notegic").Emit(ctx, record)
}

func typeName(err error) string {
	return fmt.Sprintf("%T", err)
}
