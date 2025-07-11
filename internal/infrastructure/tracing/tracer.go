package tracing

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	oteltrace "go.opentelemetry.io/otel/trace"
)

// TracerConfig holds the configuration for the tracer
type TracerConfig struct {
	ServiceName        string
	ServiceVersion     string
	JaegerEndpoint     string
	Environment        string
	SamplingRatio      float64
	BatchTimeout       time.Duration
	MaxExportBatchSize int
}

// DefaultTracerConfig returns a default tracer configuration
func DefaultTracerConfig() *TracerConfig {
	return &TracerConfig{
		ServiceName:        getEnvOrDefault("OTEL_SERVICE_NAME", "calculator"),
		ServiceVersion:     getEnvOrDefault("OTEL_SERVICE_VERSION", "1.0.0"),
		JaegerEndpoint:     getEnvOrDefault("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4318"),
		Environment:        getEnvOrDefault("ENVIRONMENT", "development"),
		SamplingRatio:      1.0, // Sample all traces in development
		BatchTimeout:       5 * time.Second,
		MaxExportBatchSize: 512,
	}
}

// TracerProvider holds the tracer provider and cleanup function
type TracerProvider struct {
	provider *trace.TracerProvider
	cleanup  func(context.Context) error
}

// NewTracerProvider creates a new tracer provider with OTLP exporter
func NewTracerProvider(config *TracerConfig) (*TracerProvider, error) {
	// Create OTLP HTTP exporter
	exporter, err := otlptracehttp.New(
		context.Background(),
		otlptracehttp.WithEndpoint(config.JaegerEndpoint),
		otlptracehttp.WithInsecure(),            // Use HTTP instead of HTTPS for local development
		otlptracehttp.WithURLPath("/v1/traces"), // Explicit path
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
	}

	// Create resource with service information
	res, err := resource.New(
		context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(config.ServiceName),
			semconv.ServiceVersionKey.String(config.ServiceVersion),
			semconv.DeploymentEnvironmentKey.String(config.Environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create tracer provider
	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter,
			trace.WithBatchTimeout(config.BatchTimeout),
			trace.WithMaxExportBatchSize(config.MaxExportBatchSize),
		),
		trace.WithResource(res),
		trace.WithSampler(trace.TraceIDRatioBased(config.SamplingRatio)),
	)

	// Set global tracer provider
	otel.SetTracerProvider(tp)

	// Set global propagator for context propagation
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return &TracerProvider{
		provider: tp,
		cleanup: func(ctx context.Context) error {
			return tp.Shutdown(ctx)
		},
	}, nil
}

// GetTracer returns a tracer for the given name
func (tp *TracerProvider) GetTracer(name string) oteltrace.Tracer {
	return tp.provider.Tracer(name)
}

// Shutdown gracefully shuts down the tracer provider
func (tp *TracerProvider) Shutdown(ctx context.Context) error {
	return tp.cleanup(ctx)
}

// getEnvOrDefault returns the value of an environment variable or a default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Global tracer provider instance
var globalTracerProvider *TracerProvider

// InitTracing initializes the global tracer provider
func InitTracing(config *TracerConfig) error {
	if config == nil {
		config = DefaultTracerConfig()
	}

	tp, err := NewTracerProvider(config)
	if err != nil {
		return fmt.Errorf("failed to initialize tracing: %w", err)
	}

	globalTracerProvider = tp
	return nil
}

// GetGlobalTracer returns the global tracer for the given name
func GetGlobalTracer(name string) oteltrace.Tracer {
	if globalTracerProvider == nil {
		// Return a no-op tracer if tracing is not initialized
		return otel.Tracer(name)
	}
	return globalTracerProvider.GetTracer(name)
}

// ShutdownTracing gracefully shuts down the global tracer provider
func ShutdownTracing(ctx context.Context) error {
	if globalTracerProvider == nil {
		return nil
	}
	return globalTracerProvider.Shutdown(ctx)
}
