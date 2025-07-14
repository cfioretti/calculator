package middleware

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	domainMetrics "github.com/cfioretti/calculator/internal/domain/metrics"
	infraMetrics "github.com/cfioretti/calculator/internal/infrastructure/metrics"
)

// MetricsMiddleware provides gRPC interceptors for metrics collection
type MetricsMiddleware struct {
	domainMetrics     domainMetrics.CalculatorMetrics
	prometheusMetrics *infraMetrics.PrometheusMetrics
}

// NewMetricsMiddleware creates a new metrics middleware
func NewMetricsMiddleware(
	domainMetrics domainMetrics.CalculatorMetrics,
	prometheusMetrics *infraMetrics.PrometheusMetrics,
) *MetricsMiddleware {
	return &MetricsMiddleware{
		domainMetrics:     domainMetrics,
		prometheusMetrics: prometheusMetrics,
	}
}

// UnaryServerInterceptor returns a gRPC unary server interceptor for metrics
func (m *MetricsMiddleware) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()

		// Increment active calculations for business methods
		if isCalculationMethod(info.FullMethod) {
			// Get current active count - in a real implementation,
			// this would be tracked in a service
			m.domainMetrics.SetActiveCalculations(getCurrentActiveCalculations() + 1)
		}

		// Call the actual handler
		resp, err := handler(ctx, req)

		duration := time.Since(start)

		// Record technical metrics
		statusCode := getStatusCode(err)
		m.prometheusMetrics.IncrementGRPCRequests(
			extractMethodName(info.FullMethod),
			statusCode,
		)
		m.prometheusMetrics.RecordGRPCDuration(
			extractMethodName(info.FullMethod),
			duration,
		)

		// Record business metrics for calculation methods
		if isCalculationMethod(info.FullMethod) {
			// Decrement active calculations
			m.domainMetrics.SetActiveCalculations(getCurrentActiveCalculations() - 1)

			calculationType := getCalculationType(info.FullMethod)

			if err != nil {
				// Record calculation error
				errorType := getErrorType(err)
				m.domainMetrics.IncrementCalculationErrors(calculationType, errorType)
			} else {
				// Record successful calculation
				m.domainMetrics.IncrementCalculationsTotal(calculationType)
				m.domainMetrics.RecordCalculationDuration(calculationType, duration)

				// Extract business metrics from response if available
				if businessMetrics := extractBusinessMetrics(resp); businessMetrics != nil {
					recordBusinessMetrics(m.domainMetrics, businessMetrics)
				}
			}
		}

		return resp, err
	}
}

// StreamServerInterceptor returns a gRPC stream server interceptor for metrics
func (m *MetricsMiddleware) StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		stream grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		start := time.Now()

		// Call the actual handler
		err := handler(srv, stream)

		duration := time.Since(start)

		// Record technical metrics for streaming
		statusCode := getStatusCode(err)
		m.prometheusMetrics.IncrementGRPCRequests(
			extractMethodName(info.FullMethod),
			statusCode,
		)
		m.prometheusMetrics.RecordGRPCDuration(
			extractMethodName(info.FullMethod),
			duration,
		)

		return err
	}
}

// Helper functions

func isCalculationMethod(fullMethod string) bool {
	calculationMethods := []string{
		"/calculator.CalculatorService/CalculateDough",
		"/calculator.CalculatorService/CalculateIngredients",
		"/calculator.CalculatorService/OptimizeRecipe",
	}

	for _, method := range calculationMethods {
		if fullMethod == method {
			return true
		}
	}
	return false
}

func extractMethodName(fullMethod string) string {
	// Extract method name from "/package.Service/Method" format
	if len(fullMethod) > 0 && fullMethod[0] == '/' {
		parts := splitMethod(fullMethod[1:]) // Remove leading slash
		if len(parts) == 2 {
			serviceParts := splitService(parts[0])
			if len(serviceParts) >= 2 {
				return parts[1] // Return just the method name
			}
		}
	}
	return "unknown"
}

func splitMethod(s string) []string {
	// Simple split on last slash
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == '/' {
			return []string{s[:i], s[i+1:]}
		}
	}
	return []string{s}
}

func splitService(s string) []string {
	// Simple split on dot
	parts := []string{}
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '.' {
			if i > start {
				parts = append(parts, s[start:i])
			}
			start = i + 1
		}
	}
	if start < len(s) {
		parts = append(parts, s[start:])
	}
	return parts
}

func getStatusCode(err error) string {
	if err == nil {
		return "success"
	}

	if st, ok := status.FromError(err); ok {
		switch st.Code() {
		case codes.OK:
			return "success"
		case codes.InvalidArgument:
			return "invalid_argument"
		case codes.NotFound:
			return "not_found"
		case codes.Internal:
			return "internal_error"
		case codes.Unavailable:
			return "unavailable"
		default:
			return "error"
		}
	}

	return "error"
}

func getCalculationType(fullMethod string) string {
	switch fullMethod {
	case "/calculator.CalculatorService/CalculateDough":
		return "dough_calculation"
	case "/calculator.CalculatorService/CalculateIngredients":
		return "ingredient_calculation"
	case "/calculator.CalculatorService/OptimizeRecipe":
		return "recipe_optimization"
	default:
		return "unknown_calculation"
	}
}

func getErrorType(err error) string {
	if st, ok := status.FromError(err); ok {
		switch st.Code() {
		case codes.InvalidArgument:
			return "invalid_input"
		case codes.NotFound:
			return "recipe_not_found"
		case codes.Internal:
			return "calculation_error"
		case codes.Unavailable:
			return "service_unavailable"
		case codes.DeadlineExceeded:
			return "timeout"
		default:
			return "unknown_error"
		}
	}
	return "unknown_error"
}

// getCurrentActiveCalculations would be implemented to track active calculations
// In a real implementation, this might be stored in a service or cache
func getCurrentActiveCalculations() int {
	// This is a placeholder - in reality you'd track this in your service
	return 0
}

// BusinessMetrics represents extracted business metrics from response
type BusinessMetrics struct {
	Weight      float64
	Hydration   float64
	Accuracy    float64
	Ingredients []string
	RecipeType  string
}

// extractBusinessMetrics extracts business metrics from gRPC response
func extractBusinessMetrics(resp interface{}) *BusinessMetrics {
	// This would be implemented based on your specific response types
	// For now, return nil - would be implemented when integrating with actual gRPC services
	return nil
}

// recordBusinessMetrics records the extracted business metrics
func recordBusinessMetrics(metrics domainMetrics.CalculatorMetrics, business *BusinessMetrics) {
	if business.Weight > 0 {
		metrics.RecordDoughWeight(business.Weight)
	}

	if business.Hydration > 0 {
		metrics.RecordDoughHydration(business.Hydration)
	}

	if business.Accuracy > 0 {
		metrics.RecordDoughAccuracy(business.Accuracy)
	}

	if business.RecipeType != "" {
		metrics.IncrementRecipeTypes(business.RecipeType)
	}

	for _, ingredient := range business.Ingredients {
		metrics.IncrementIngredientValidations(ingredient, true)
	}
}
