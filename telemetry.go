package main

import (
	"context"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0" // Ensure correct semconv version
	oteltrace "go.opentelemetry.io/otel/trace"
)

// Note: Assumes global variables 'meterProvider', 'meter', 'taskCounter', 'handlerLatency', 'errorCounter' are declared in main.go or similar.

// initTracer initializes the OpenTelemetry tracer provider and returns a shutdown function.
func initTracer() func() {
	// Configure the OTLP exporter
	exp, err := otlptracehttp.New(
		context.Background(),
		otlptracehttp.WithEndpoint("otel-collector:4318"), // Ensure this matches your collector setup
		otlptracehttp.WithInsecure(),                      // Use WithInsecure for HTTP
	)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create OTLP trace exporter")
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("todo-app"), // Consistent service name
		)),
	)
	otel.SetTracerProvider(tp)

	log.Info().Msg("Tracer initialized and spans are being created")

	// Return the shutdown function
	return func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Error().Err(err).Msg("Failed to shutdown TracerProvider")
		}
	}
}

// initMetrics initializes the OpenTelemetry meter provider and metrics.
func initMetrics() {
	exp, err := prometheus.New()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Prometheus exporter")
	}
	// Assign to the global meterProvider
	meterProvider = sdkmetric.NewMeterProvider(sdkmetric.WithReader(exp))
	otel.SetMeterProvider(meterProvider)

	// Assign to the global meter
	meter = meterProvider.Meter("todo-service") // Use a consistent meter name

	// Assign to global metric variables
	taskCounter, err = meter.Int64Counter(
		"todo_tasks_added_total", // Use standard naming convention (_total for counters)
		metric.WithDescription("Total number of ToDo tasks added"),
		metric.WithUnit("{tasks}"),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create task counter metric")
	}

	handlerLatency, err = meter.Float64Histogram(
		"todo_handler_latency_milliseconds",
		metric.WithDescription("Latency of HTTP handlers in milliseconds"),
		metric.WithUnit("ms"),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create latency histogram metric")
	}

	errorCounter, err = meter.Int64Counter(
		"todo_handler_errors_total", // Use standard naming convention
		metric.WithDescription("Total number of errors encountered by HTTP handlers"),
		metric.WithUnit("{errors}"),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create error counter metric")
	}

	// Expose metrics via HTTP endpoint
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		log.Info().Msg("Prometheus metrics exposed at :2112/metrics")
		// Consider error handling for ListenAndServe
		if err := http.ListenAndServe(":2112", nil); err != nil {
			log.Error().Err(err).Msg("Metrics server failed")
		}
	}()
}

// logWithTrace adds trace and span IDs to log events.
func logWithTrace(ctx context.Context) *zerolog.Event {
	spanCtx := oteltrace.SpanContextFromContext(ctx)
	event := log.Info() // Start with Info level, can be changed as needed
	if spanCtx.HasTraceID() {
		event = event.Str("trace_id", spanCtx.TraceID().String())
	}
	if spanCtx.HasSpanID() {
		event = event.Str("span_id", spanCtx.SpanID().String())
	}
	return event
}

// handleError logs an error with trace context and writes an HTTP error response.
func handleError(ctx context.Context, w http.ResponseWriter, statusCode int, message string, err error) {
	logEntry := logWithTrace(ctx).Int("status_code", statusCode)
	if err != nil {
		logEntry = logEntry.Err(err) // Log the actual error if provided
	} else {
		logEntry = logEntry.Str("error", message) // Log the message as error string if no error object
	}
	logEntry.Msg(message) // Use the user-facing message as the log message

	http.Error(w, message, statusCode)
}
