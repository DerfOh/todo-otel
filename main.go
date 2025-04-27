package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time" // <-- Import time package

	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

// Declare global variables used across files
// Consider using a struct + dependency injection for better organization
var (
	store          *Store
	meterProvider  *sdkmetric.MeterProvider // Keep provider if needed globally
	meter          metric.Meter             // Keep meter if needed globally
	taskCounter    metric.Int64Counter
	handlerLatency metric.Float64Histogram
	errorCounter   metric.Int64Counter
	// logFile and logFileMutex are now managed within logger.go
)

func main() {
	// Initialize components by calling functions from other files
	initLogger()         // From logger.go
	defer closeLogFile() // From logger.go

	startLogRotation() // From logger.go

	shutdownTracer := initTracer() // From telemetry.go
	defer shutdownTracer()

	initMetrics() // From telemetry.go

	store = NewStore() // From store.go

	// Setup HTTP routes using handlers from handlers.go
	mux := http.NewServeMux()
	// Wrap handlers with OpenTelemetry instrumentation
	mux.Handle("/add", otelhttp.NewHandler(http.HandlerFunc(addHandler), "addHandler"))
	mux.Handle("/list", otelhttp.NewHandler(http.HandlerFunc(listHandler), "listHandler"))
	mux.Handle("/delete", otelhttp.NewHandler(http.HandlerFunc(deleteHandler), "deleteHandler"))
	mux.Handle("/update", otelhttp.NewHandler(http.HandlerFunc(updateHandler), "updateHandler"))
	mux.Handle("/get", otelhttp.NewHandler(http.HandlerFunc(getHandler), "getHandler"))
	mux.Handle("/complete", otelhttp.NewHandler(http.HandlerFunc(completeHandler), "completeHandler"))
	mux.Handle("/search", otelhttp.NewHandler(http.HandlerFunc(searchHandler), "searchHandler"))
	// Note: The /metrics endpoint is started within initMetrics()

	// Start HTTP server
	server := &http.Server{
		Addr:    ":8080", // Main application server port
		Handler: mux,
	}

	go func() {
		log.Info().Msg("Main HTTP server starting on :8080")
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Main server failed to start")
		}
	}()

	// Graceful shutdown handling
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Info().Msg("Shutting down server...")

	// Add context with timeout for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second) // Use a reasonable timeout e.g. 5*time.Second
	defer cancel()

	// Shutdown main HTTP server
	if err := server.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("Server shutdown failed")
	} else {
		log.Info().Msg("Server gracefully stopped")
	}

	// Shutdown meter provider (if applicable and needed)
	if meterProvider != nil {
		if err := meterProvider.Shutdown(context.Background()); err != nil {
			log.Error().Err(err).Msg("MeterProvider shutdown failed")
		}
	}

	// Tracer shutdown is handled by defer earlier
}
