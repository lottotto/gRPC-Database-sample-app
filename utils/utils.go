package utils

import (
	"fmt"
	"os"

	_ "github.com/lib/pq"

	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"

	// "go.opentelemetry.io/otel/sdk/trace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

func GetEnv(name string, defalutValue string) string {
	var value string

	if value = os.Getenv(name); value == "" {
		value = defalutValue
	}
	return value
}

func GetPostgresConnectionInfo() string {
	return fmt.Sprintf("host=%v port=%v user=%v password=%v dbname=%v sslmode=disable",
		GetEnv("POSTGRES_HOST", "127.0.0.1"),
		GetEnv("POSTGRES_PORT", "5432"),
		GetEnv("POSTGRES_USER", "postgres"),
		GetEnv("POSTGRES_PASS", "password"),
		GetEnv("POSTGRES_DB", "postgres"),
	)
}

func GetPostgresConnection() (conn *sqlx.DB, err error) {
	dbinfo := GetPostgresConnectionInfo()
	db, err := sqlx.Open("postgres", dbinfo)
	return db, err

}

// TracerProvider作成のため、SpanExporterのInterfaceで戻す。
func GetTraceExporterStdOut() (exporter sdktrace.SpanExporter, err error) {
	// Todo: 何のオプションか
	exporter, err = stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		return nil, err
	}
	return exporter, nil
}

// TracerProvider作成のため、SpanExporterのInterfaceで戻す。
func GetTraceExporterJeager() (exporter sdktrace.SpanExporter, err error) {

	jaegerUrl := GetEnv("JAEGER_AGENT_ENDPOINT", "http://localhost:14268/api/traces")

	exporter, err = jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(jaegerUrl)))
	if err != nil {
		return nil, err
	}
	return exporter, nil
}

func InitTraceProvider(exporter sdktrace.SpanExporter, serviceNameKey string, version string) (tp *sdktrace.TracerProvider, err error) {

	tp = sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(
			resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceNameKey.String(serviceNameKey),
				semconv.ServiceVersionKey.String(version),
			),
		),
	)

	return tp, nil
}

// func GetZapLogger() *zap.Logger {

// 	logConfig := zap.Config{}

// 	logger, _ := zap.NewDevelopment()
// 	return logger
// }
