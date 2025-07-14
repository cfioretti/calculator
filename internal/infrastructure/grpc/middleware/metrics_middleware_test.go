package middleware

import (
	"context"
	"errors"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	infraMetrics "github.com/cfioretti/calculator/internal/infrastructure/metrics"
)

// MockDomainMetrics for testing domain metrics interface
type MockDomainMetrics struct {
	calculationsTotal     map[string]int
	calculationDurations  map[string][]time.Duration
	activeCalculations    int
	calculationErrors     map[string]map[string]int
	doughAccuracies       []float64
	ingredientValidations map[string]map[bool]int
	doughWeights          []float64
	doughHydrations       []float64
	recipeTypes           map[string]int
}

// MockMetrics for testing
type MockGRPCMetrics struct {
	grpcRequests  map[string]map[string]int
	grpcDurations map[string][]time.Duration
}

func NewMockDomainMetrics() *MockDomainMetrics {
	return &MockDomainMetrics{
		calculationsTotal:     make(map[string]int),
		calculationDurations:  make(map[string][]time.Duration),
		calculationErrors:     make(map[string]map[string]int),
		ingredientValidations: make(map[string]map[bool]int),
		recipeTypes:           make(map[string]int),
	}
}

func (m *MockDomainMetrics) IncrementCalculationsTotal(calculationType string) {
	m.calculationsTotal[calculationType]++
}

func (m *MockDomainMetrics) RecordCalculationDuration(calculationType string, duration time.Duration) {
	m.calculationDurations[calculationType] = append(m.calculationDurations[calculationType], duration)
}

func (m *MockDomainMetrics) SetActiveCalculations(count int) {
	m.activeCalculations = count
}

func (m *MockDomainMetrics) IncrementCalculationErrors(calculationType string, errorType string) {
	if m.calculationErrors[calculationType] == nil {
		m.calculationErrors[calculationType] = make(map[string]int)
	}
	m.calculationErrors[calculationType][errorType]++
}

func (m *MockDomainMetrics) RecordDoughAccuracy(accuracy float64) {
	m.doughAccuracies = append(m.doughAccuracies, accuracy)
}

func (m *MockDomainMetrics) IncrementIngredientValidations(ingredient string, valid bool) {
	if m.ingredientValidations[ingredient] == nil {
		m.ingredientValidations[ingredient] = make(map[bool]int)
	}
	m.ingredientValidations[ingredient][valid]++
}

func (m *MockDomainMetrics) RecordDoughWeight(weight float64) {
	m.doughWeights = append(m.doughWeights, weight)
}

func (m *MockDomainMetrics) RecordDoughHydration(hydration float64) {
	m.doughHydrations = append(m.doughHydrations, hydration)
}

func (m *MockDomainMetrics) IncrementRecipeTypes(recipeType string) {
	m.recipeTypes[recipeType]++
}

// Getter methods for testing
func (m *MockDomainMetrics) GetCalculationsTotal(calculationType string) int {
	return m.calculationsTotal[calculationType]
}

func (m *MockDomainMetrics) GetCalculationDurations(calculationType string) []time.Duration {
	return m.calculationDurations[calculationType]
}

func (m *MockDomainMetrics) GetCalculationErrors(calculationType string, errorType string) int {
	if m.calculationErrors[calculationType] == nil {
		return 0
	}
	return m.calculationErrors[calculationType][errorType]
}

func NewMockGRPCMetrics() *MockGRPCMetrics {
	return &MockGRPCMetrics{
		grpcRequests:  make(map[string]map[string]int),
		grpcDurations: make(map[string][]time.Duration),
	}
}

func (m *MockGRPCMetrics) IncrementGRPCRequests(method, status string) {
	if m.grpcRequests[method] == nil {
		m.grpcRequests[method] = make(map[string]int)
	}
	m.grpcRequests[method][status]++
}

func (m *MockGRPCMetrics) RecordGRPCDuration(method string, duration time.Duration) {
	m.grpcDurations[method] = append(m.grpcDurations[method], duration)
}

func TestExtractMethodName(t *testing.T) {
	tests := []struct {
		name       string
		fullMethod string
		expected   string
	}{
		{
			name:       "Valid calculator method",
			fullMethod: "/calculator.CalculatorService/CalculateDough",
			expected:   "CalculateDough",
		},
		{
			name:       "Valid ingredient method",
			fullMethod: "/calculator.CalculatorService/CalculateIngredients",
			expected:   "CalculateIngredients",
		},
		{
			name:       "Invalid method format",
			fullMethod: "InvalidMethod",
			expected:   "unknown",
		},
		{
			name:       "Empty method",
			fullMethod: "",
			expected:   "unknown",
		},
		{
			name:       "Method without service",
			fullMethod: "/Method",
			expected:   "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractMethodName(tt.fullMethod)
			if result != tt.expected {
				t.Errorf("extractMethodName(%q) = %q, want %q",
					tt.fullMethod, result, tt.expected)
			}
		})
	}
}

func TestIsCalculationMethod(t *testing.T) {
	tests := []struct {
		name       string
		fullMethod string
		expected   bool
	}{
		{
			name:       "Dough calculation method",
			fullMethod: "/calculator.CalculatorService/CalculateDough",
			expected:   true,
		},
		{
			name:       "Ingredient calculation method",
			fullMethod: "/calculator.CalculatorService/CalculateIngredients",
			expected:   true,
		},
		{
			name:       "Recipe optimization method",
			fullMethod: "/calculator.CalculatorService/OptimizeRecipe",
			expected:   true,
		},
		{
			name:       "Non-calculation method",
			fullMethod: "/calculator.CalculatorService/GetHealth",
			expected:   false,
		},
		{
			name:       "Different service method",
			fullMethod: "/other.Service/CalculateDough",
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isCalculationMethod(tt.fullMethod)
			if result != tt.expected {
				t.Errorf("isCalculationMethod(%q) = %t, want %t",
					tt.fullMethod, result, tt.expected)
			}
		})
	}
}

func TestGetStatusCode(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "No error",
			err:      nil,
			expected: "success",
		},
		{
			name:     "Invalid argument error",
			err:      status.Error(codes.InvalidArgument, "invalid input"),
			expected: "invalid_argument",
		},
		{
			name:     "Not found error",
			err:      status.Error(codes.NotFound, "recipe not found"),
			expected: "not_found",
		},
		{
			name:     "Internal error",
			err:      status.Error(codes.Internal, "calculation failed"),
			expected: "internal_error",
		},
		{
			name:     "Unavailable error",
			err:      status.Error(codes.Unavailable, "service unavailable"),
			expected: "unavailable",
		},
		{
			name:     "Generic error",
			err:      errors.New("generic error"),
			expected: "error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getStatusCode(tt.err)
			if result != tt.expected {
				t.Errorf("getStatusCode(%v) = %q, want %q",
					tt.err, result, tt.expected)
			}
		})
	}
}

func TestGetCalculationType(t *testing.T) {
	tests := []struct {
		name       string
		fullMethod string
		expected   string
	}{
		{
			name:       "Dough calculation",
			fullMethod: "/calculator.CalculatorService/CalculateDough",
			expected:   "dough_calculation",
		},
		{
			name:       "Ingredient calculation",
			fullMethod: "/calculator.CalculatorService/CalculateIngredients",
			expected:   "ingredient_calculation",
		},
		{
			name:       "Recipe optimization",
			fullMethod: "/calculator.CalculatorService/OptimizeRecipe",
			expected:   "recipe_optimization",
		},
		{
			name:       "Unknown method",
			fullMethod: "/calculator.CalculatorService/UnknownMethod",
			expected:   "unknown_calculation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getCalculationType(tt.fullMethod)
			if result != tt.expected {
				t.Errorf("getCalculationType(%q) = %q, want %q",
					tt.fullMethod, result, tt.expected)
			}
		})
	}
}

func TestGetErrorType(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "Invalid argument",
			err:      status.Error(codes.InvalidArgument, "bad input"),
			expected: "invalid_input",
		},
		{
			name:     "Not found",
			err:      status.Error(codes.NotFound, "recipe missing"),
			expected: "recipe_not_found",
		},
		{
			name:     "Internal error",
			err:      status.Error(codes.Internal, "calculation failed"),
			expected: "calculation_error",
		},
		{
			name:     "Unavailable",
			err:      status.Error(codes.Unavailable, "service down"),
			expected: "service_unavailable",
		},
		{
			name:     "Deadline exceeded",
			err:      status.Error(codes.DeadlineExceeded, "timeout"),
			expected: "timeout",
		},
		{
			name:     "Unknown gRPC error",
			err:      status.Error(codes.Unknown, "unknown"),
			expected: "unknown_error",
		},
		{
			name:     "Generic error",
			err:      errors.New("generic"),
			expected: "unknown_error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getErrorType(tt.err)
			if result != tt.expected {
				t.Errorf("getErrorType(%v) = %q, want %q",
					tt.err, result, tt.expected)
			}
		})
	}
}

func TestMetricsMiddleware_UnaryServerInterceptor_Success(t *testing.T) {
	// Arrange
	mockDomainMetrics := NewMockDomainMetrics()
	prometheusMetrics := infraMetrics.NewPrometheusMetrics()

	middleware := NewMetricsMiddleware(mockDomainMetrics, prometheusMetrics)

	interceptor := middleware.UnaryServerInterceptor()

	// Mock handler that succeeds
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		time.Sleep(10 * time.Millisecond) // Simulate processing time
		return "success response", nil
	}

	info := &grpc.UnaryServerInfo{
		FullMethod: "/calculator.CalculatorService/CalculateDough",
	}

	// Act
	resp, err := interceptor(context.Background(), "test request", info, handler)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if resp != "success response" {
		t.Errorf("Expected 'success response', got %v", resp)
	}

	// Verify domain metrics were called (using the interface)
	if mockDomainMetrics.GetCalculationsTotal("dough_calculation") != 1 {
		t.Errorf("Expected 1 calculation total, got %d",
			mockDomainMetrics.GetCalculationsTotal("dough_calculation"))
	}

	if len(mockDomainMetrics.GetCalculationDurations("dough_calculation")) != 1 {
		t.Errorf("Expected 1 duration record, got %d",
			len(mockDomainMetrics.GetCalculationDurations("dough_calculation")))
	}
}

func TestMetricsMiddleware_UnaryServerInterceptor_Error(t *testing.T) {
	// Arrange
	mockDomainMetrics := NewMockDomainMetrics()
	prometheusMetrics := infraMetrics.NewPrometheusMetrics()

	middleware := NewMetricsMiddleware(mockDomainMetrics, prometheusMetrics)
	interceptor := middleware.UnaryServerInterceptor()

	// Mock handler that fails
	expectedError := status.Error(codes.InvalidArgument, "invalid input")
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, expectedError
	}

	info := &grpc.UnaryServerInfo{
		FullMethod: "/calculator.CalculatorService/CalculateDough",
	}

	// Act
	resp, err := interceptor(context.Background(), "test request", info, handler)

	// Assert
	if err != expectedError {
		t.Errorf("Expected error %v, got %v", expectedError, err)
	}

	if resp != nil {
		t.Errorf("Expected nil response, got %v", resp)
	}

	// Verify error was recorded
	if mockDomainMetrics.GetCalculationErrors("dough_calculation", "invalid_input") != 1 {
		t.Errorf("Expected 1 error record, got %d",
			mockDomainMetrics.GetCalculationErrors("dough_calculation", "invalid_input"))
	}

	// Should not record successful calculation
	if mockDomainMetrics.GetCalculationsTotal("dough_calculation") != 0 {
		t.Errorf("Expected 0 calculation totals for failed request, got %d",
			mockDomainMetrics.GetCalculationsTotal("dough_calculation"))
	}
}

func TestMetricsMiddleware_UnaryServerInterceptor_NonCalculationMethod(t *testing.T) {
	// Arrange
	mockDomainMetrics := NewMockDomainMetrics()
	prometheusMetrics := infraMetrics.NewPrometheusMetrics()

	middleware := NewMetricsMiddleware(mockDomainMetrics, prometheusMetrics)
	interceptor := middleware.UnaryServerInterceptor()

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "health ok", nil
	}

	info := &grpc.UnaryServerInfo{
		FullMethod: "/calculator.CalculatorService/GetHealth",
	}

	// Act
	resp, err := interceptor(context.Background(), "test request", info, handler)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if resp != "health ok" {
		t.Errorf("Expected 'health ok', got %v", resp)
	}

	// Should not record business metrics for non-calculation methods
	if mockDomainMetrics.GetCalculationsTotal("unknown_calculation") != 0 {
		t.Errorf("Expected 0 calculation totals for non-calculation method, got %d",
			mockDomainMetrics.GetCalculationsTotal("unknown_calculation"))
	}

	// Technical metrics should still be recorded (we can't easily test prometheus metrics here)
}
