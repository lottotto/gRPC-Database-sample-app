package utils

import (
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"

	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

func Init() *sdktrace.TracerProvider {
	var jaegerUrl string

	if jaegerUrl = os.Getenv("JAEGER_AGENT_ENDPOINT"); jaegerUrl == "" {
		jaegerUrl = "http://localhost:14268/api/traces"

	}

	exporter, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(jaegerUrl)))

	if err != nil {
		log.Fatal(err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("mygrpc"),
		)),
	)
	return tp

}

func GetEnv(name string, defalutValue string) string {
	var value string
	if value = os.Getenv(name); value != "" {
		value = defalutValue
	}
	return value
}

func GetPostgresConnection() (conn *sqlx.DB, err error) {
	dbinfo := fmt.Sprintf("host=%v port=%v user=%v password=%v dbname=%v",
		GetEnv("POSTGRES_HOST", "127.0.0.1"),
		GetEnv("POSTGRES_PORT", "5432"),
		GetEnv("POSTGRES_USER", "postgres"),
		GetEnv("POSTGRES_PASS", "password"),
		GetEnv("POSTGRES_DB", "postgres"),
	)
	db, err := sqlx.Open("postgres", dbinfo)
	return db, err

}
