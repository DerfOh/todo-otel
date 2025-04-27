package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

// Global variables required by handlers and other components
var (
	store          *Store                   // In-memory task store
	meterProvider  *sdkmetric.MeterProvider // OTel meter provider for metrics
	meter          metric.Meter             // OTel meter for creating metrics
	taskCounter    metric.Int64Counter      // Counter for tracking task operations
	handlerLatency metric.Float64Histogram  // Histogram for tracking handler latencies
	errorCounter   metric.Int64Counter      // Counter for tracking errors
)

func main() {
	// Initialize structured logging
	initLogger()
	defer closeLogFile()
	startLogRotation()

	// Initialize OpenTelemetry tracing
	shutdownTracer := initTracer()
	defer shutdownTracer()

	// Initialize OpenTelemetry metrics
	initMetrics()

	// Initialize task store
	store = NewStore()

	// Configure and start HTTP server
	mux := setupRoutes()
	server := createServer(mux)
	startServerAsync(server)

	// Wait for shutdown signal and perform cleanup
	handleGracefulShutdown(server)
}

// setupRoutes configures all HTTP routes with OTel instrumentation
func setupRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	// Register API endpoints with OpenTelemetry instrumentation wrappers
	mux.Handle("/add", otelhttp.NewHandler(http.HandlerFunc(addHandler), "addHandler"))
	mux.Handle("/list", otelhttp.NewHandler(http.HandlerFunc(listHandler), "listHandler"))
	mux.Handle("/delete", otelhttp.NewHandler(http.HandlerFunc(deleteHandler), "deleteHandler"))
	mux.Handle("/update", otelhttp.NewHandler(http.HandlerFunc(updateHandler), "updateHandler"))
	mux.Handle("/get", otelhttp.NewHandler(http.HandlerFunc(getHandler), "getHandler"))
	mux.Handle("/complete", otelhttp.NewHandler(http.HandlerFunc(completeHandler), "completeHandler"))
	mux.Handle("/search", otelhttp.NewHandler(http.HandlerFunc(searchHandler), "searchHandler"))

	return mux
}

// createServer initializes the HTTP server with configuration
func createServer(handler http.Handler) *http.Server {
	return &http.Server{
		Addr:         ":8080",
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
}

// startServerAsync starts the HTTP server in a goroutine
func startServerAsync(server *http.Server) {
	go func() {
		log.Info().Msg("Main HTTP server starting on :8080")
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Main server failed to start")
		}
	}()
}

// handleGracefulShutdown waits for termination signals and shuts down cleanly
func handleGracefulShutdown(server *http.Server) {
	// Wait for interrupt signal
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Info().Msg("Shutting down server...")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("Server shutdown failed")
	} else {
		log.Info().Msg("Server gracefully stopped")
	}

	// Clean up OTel meter provider
	if meterProvider != nil {
		if err := meterProvider.Shutdown(context.Background()); err != nil {
			log.Error().Err(err).Msg("MeterProvider shutdown failed")
		}
	}
}
