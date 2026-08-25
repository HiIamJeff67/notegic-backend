package logs

import (
	"context"
	"encoding/json"
	"fmt"
	stdlog "log"
	"log/slog"
	"os"
	"strings"

	colog "github.com/comail/colog"
	"go.opentelemetry.io/otel/attribute"
)

type CommandLineInterfaceLogger struct {
	logger *stdlog.Logger
}

func NewCommandLineInterfaceLogger() LoggerInterface {
	coloredLogger := colog.NewCoLog(os.Stderr, "", stdlog.Ltime)
	coloredLogger.SetFormatter(&colog.StdFormatter{
		Flag:   stdlog.Ltime,
		Colors: true,
	})

	return &CommandLineInterfaceLogger{
		logger: stdlog.New(coloredLogger, "", 0),
	}
}

func (l *CommandLineInterfaceLogger) Debug(ctx context.Context, message string, attributes ...attribute.KeyValue) {
	l.write("debug", nil, message, attributes...)
}

func (l *CommandLineInterfaceLogger) Info(ctx context.Context, message string, attributes ...attribute.KeyValue) {
	l.write("info", nil, message, attributes...)
}

func (l *CommandLineInterfaceLogger) Warn(ctx context.Context, message string, attributes ...attribute.KeyValue) {
	l.write("warn", nil, message, attributes...)
}

func (l *CommandLineInterfaceLogger) Error(ctx context.Context, err error, message string, attributes ...attribute.KeyValue) {
	l.write("error", err, message, attributes...)
}

func (l *CommandLineInterfaceLogger) Alert(ctx context.Context, err error, message string, attributes ...attribute.KeyValue) {
	l.write("alert", err, message, attributes...)
}

func (l *CommandLineInterfaceLogger) JSON(
	ctx context.Context,
	level slog.Level,
	message string,
	payload any,
	attributes ...attribute.KeyValue,
) error {
	marshaledPayload, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}

	l.write(level.String(), nil, message+"\n"+string(marshaledPayload), attributes...)

	return nil
}

func (l *CommandLineInterfaceLogger) write(
	level string,
	err error,
	message string,
	attributes ...attribute.KeyValue,
) {
	if err != nil {
		message += ": " + err.Error()
	}
	if len(attributes) > 0 {
		items := make([]string, 0, len(attributes))
		for _, item := range attributes {
			items = append(items, fmt.Sprintf("%s=%v", item.Key, item.Value.AsInterface()))
		}
		message += " " + strings.Join(items, " ")
	}

	l.logger.Printf("%s: %s", level, message)
}
