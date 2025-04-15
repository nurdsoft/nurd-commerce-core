// Package main is the main app
package main

import (
	"context"
	"embed"
	"log"

	"github.com/nurdsoft/nurd-commerce-core/cmd"
	"github.com/nurdsoft/nurd-commerce-core/internal/swagger/static"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

//go:embed docs/swagger
var staticFiles embed.FS

func execute() error {
	return cmd.Execute()
}

func initTracer() (func(context.Context) error, error) {
	ctx := context.Background()

	exporter, err := otlptracehttp.New(ctx)
	if err != nil {
		return nil, err
	}

	bsp := trace.NewBatchSpanProcessor(exporter)
	tp := trace.NewTracerProvider(
		trace.WithSpanProcessor(bsp),
		trace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
		)),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	return tp.Shutdown, nil
}

//go:generate swagger generate spec -o ./docs/swagger/swagger.yml --scan-models
func main() {

	shutdown, err := initTracer()
	if err != nil {
		log.Fatalf("failed to initialize tracer: %v", err)
	}
	defer func() {
		if err := shutdown(context.Background()); err != nil {
			log.Fatalf("failed to shutdown tracer: %v", err)
		}
	}()

	static.StaticFiles = staticFiles

	if err := execute(); err != nil {
		log.Fatal(err)
	}
}
