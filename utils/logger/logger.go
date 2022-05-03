package logger

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func GetZapLogger() (*zap.Logger, error) {

	level := zap.NewAtomicLevel()
	level.SetLevel(zapcore.DebugLevel)
	cfg := zap.Config{
		Level:    level,
		Encoding: "json",
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "TimeStamp",
			LevelKey:       "SeverityText",
			NameKey:        "Name",
			CallerKey:      "Caller",
			MessageKey:     "Body",
			EncodeLevel:    zapcore.CapitalLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.StringDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
		OutputPaths: []string{"stdout", "/tmp/zap.log"},
	}
	logger, err := cfg.Build()
	if err != nil {
		return nil, err
	}
	return logger, nil
}

func GetOtelLogMetadataFields(ctx context.Context) []zap.Field {
	spanContext := trace.SpanContextFromContext(ctx)
	traceID := spanContext.TraceID().String()
	spanID := spanContext.SpanID().String()
	baggage := baggage.FromContext(ctx)
	// members := baggage.Members()
	fmt.Printf("%v \n", baggage.String())
	return []zap.Field{
		zap.String("TraceId", traceID),
		zap.String("SpanId", spanID),
		zap.Any("Resource", map[string]interface{}{
			"service.version": "1.0.0",
		}),
		zap.Any("Attribute", map[string]interface{}{
			"http.scheme": "http",
		}),
	}
}
