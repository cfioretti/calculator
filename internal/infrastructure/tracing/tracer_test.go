package tracing

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func TestDefaultTracerConfig(t *testing.T) {
	config := DefaultTracerConfig()

	assert.NotNil(t, config)
	assert.Equal(t, "calculator", config.ServiceName)
	assert.Equal(t, "1.0.0", config.ServiceVersion)
	assert.Equal(t, "development", config.Environment)
	assert.Equal(t, "localhost:4318", config.JaegerEndpoint)
	assert.Equal(t, 1.0, config.SamplingRatio)
}

func TestTracerConfigWithEnvironmentVariables(t *testing.T) {
	os.Setenv("OTEL_SERVICE_NAME", "test-calculator")
	os.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "test-endpoint:4318")
	os.Setenv("OTEL_RESOURCE_ATTRIBUTES", "service.version=2.0.0,deployment.environment=test")
	defer func() {
		os.Unsetenv("OTEL_SERVICE_NAME")
		os.Unsetenv("OTEL_EXPORTER_OTLP_ENDPOINT")
		os.Unsetenv("OTEL_RESOURCE_ATTRIBUTES")
	}()

	config := DefaultTracerConfig()

	assert.Equal(t, "test-calculator", config.ServiceName)
	assert.Equal(t, "test-endpoint:4318", config.JaegerEndpoint)
}

func TestNewTracerProvider(t *testing.T) {
	config := &TracerConfig{
		ServiceName:        "test-calculator",
		ServiceVersion:     "1.0.0",
		Environment:        "test",
		JaegerEndpoint:     "localhost:4318",
		SamplingRatio:      1.0,
		BatchTimeout:       5 * time.Second,
		MaxExportBatchSize: 512,
	}

	tp, err := NewTracerProvider(config)

	assert.NoError(t, err)
	assert.NotNil(t, tp)
	assert.NotNil(t, tp.provider)
	assert.NotNil(t, tp.cleanup)

	globalTP := otel.GetTracerProvider()
	assert.NotNil(t, globalTP)

	ctx := context.Background()
	err = tp.Shutdown(ctx)
	assert.NoError(t, err)
}

func TestTracerProviderGetTracer(t *testing.T) {
	config := DefaultTracerConfig()
	tp, err := NewTracerProvider(config)
	require.NoError(t, err)
	defer tp.Shutdown(context.Background())

	tracer := tp.GetTracer("test-tracer")
	assert.NotNil(t, tracer)

	ctx := context.Background()
	_, span := tracer.Start(ctx, "test-operation")
	assert.NotNil(t, span)

	spanCtx := span.SpanContext()
	assert.True(t, spanCtx.IsValid())
	assert.True(t, spanCtx.HasTraceID())
	assert.True(t, spanCtx.HasSpanID())

	span.End()
}

func TestGlobalTracerFunctions(t *testing.T) {
	tracer := GetGlobalTracer("test")
	assert.NotNil(t, tracer)

	config := DefaultTracerConfig()
	config.ServiceName = "test-global-calculator"

	err := InitTracing(config)
	assert.NoError(t, err)

	globalTracer := GetGlobalTracer("test-global")
	assert.NotNil(t, globalTracer)

	ctx := context.Background()
	err = ShutdownTracing(ctx)
	assert.NoError(t, err)
}

func TestInitTracingWithNilConfig(t *testing.T) {
	err := InitTracing(nil)
	assert.NoError(t, err)

	tracer := GetGlobalTracer("test")
	assert.NotNil(t, tracer)

	ctx := context.Background()
	err = ShutdownTracing(ctx)
	assert.NoError(t, err)
}

func TestSpanCreationAndAttributes(t *testing.T) {
	config := DefaultTracerConfig()
	tp, err := NewTracerProvider(config)
	require.NoError(t, err)
	defer tp.Shutdown(context.Background())

	tracer := tp.GetTracer("test-tracer")
	ctx := context.Background()

	ctx, span := tracer.Start(ctx, "test-operation")
	defer span.End()

	span.SetAttributes(
		attribute.String("test.key", "test.value"),
		attribute.Int("test.number", 42),
	)

	spanFromCtx := trace.SpanFromContext(ctx)
	assert.NotNil(t, spanFromCtx)
	assert.Equal(t, span.SpanContext().TraceID(), spanFromCtx.SpanContext().TraceID())
	assert.Equal(t, span.SpanContext().SpanID(), spanFromCtx.SpanContext().SpanID())
}

func TestContextPropagation(t *testing.T) {
	config := DefaultTracerConfig()
	tp, err := NewTracerProvider(config)
	require.NoError(t, err)
	defer tp.Shutdown(context.Background())

	tracer := tp.GetTracer("test-tracer")
	ctx := context.Background()

	ctx, parentSpan := tracer.Start(ctx, "parent-operation")
	defer parentSpan.End()

	parentTraceID := parentSpan.SpanContext().TraceID()

	ctx, childSpan := tracer.Start(ctx, "child-operation")
	defer childSpan.End()

	childTraceID := childSpan.SpanContext().TraceID()

	assert.Equal(t, parentTraceID, childTraceID)
	assert.NotEqual(t, parentSpan.SpanContext().SpanID(), childSpan.SpanContext().SpanID())
}

func TestGetEnvOrDefault(t *testing.T) {
	os.Setenv("TEST_VAR", "test_value")
	defer os.Unsetenv("TEST_VAR")

	result := getEnvOrDefault("TEST_VAR", "default_value")
	assert.Equal(t, "test_value", result)

	result = getEnvOrDefault("NON_EXISTING_VAR", "default_value")
	assert.Equal(t, "default_value", result)
}
