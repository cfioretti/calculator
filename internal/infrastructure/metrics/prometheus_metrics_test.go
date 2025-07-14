package metrics

import (
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestPrometheusMetrics_IncrementCalculationsTotal(t *testing.T) {
	metrics := NewPrometheusMetrics()

	metrics.IncrementCalculationsTotal("dough_calculation")
	metrics.IncrementCalculationsTotal("dough_calculation")
	metrics.IncrementCalculationsTotal("ingredient_calculation")

	expected := `
		# HELP calculator_calculations_total Total number of calculations performed
		# TYPE calculator_calculations_total counter
		calculator_calculations_total{type="dough_calculation"} 2
		calculator_calculations_total{type="ingredient_calculation"} 1
	`

	if err := testutil.GatherAndCompare(
		prometheus.DefaultGatherer,
		strings.NewReader(expected),
		"calculator_calculations_total",
	); err != nil {
		t.Errorf("Unexpected metric value: %v", err)
	}
}

func TestPrometheusMetrics_RecordCalculationDuration(t *testing.T) {
	metrics := NewPrometheusMetrics()

	metrics.RecordCalculationDuration("dough_calculation", 100*time.Millisecond)
	metrics.RecordCalculationDuration("dough_calculation", 200*time.Millisecond)

	metricFamily, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		t.Fatalf("Failed to gather metrics: %v", err)
	}

	var found bool
	for _, mf := range metricFamily {
		if mf.GetName() == "calculator_calculation_duration_seconds" {
			found = true
			for _, metric := range mf.GetMetric() {
				if metric.GetHistogram() != nil {
					// Check that we have 2 observations
					if metric.GetHistogram().GetSampleCount() != 2 {
						t.Errorf("Expected 2 observations, got %d",
							metric.GetHistogram().GetSampleCount())
					}
					// Check approximate sum (0.1 + 0.2 = 0.3 seconds)
					expectedSum := 0.3
					actualSum := metric.GetHistogram().GetSampleSum()
					if actualSum < expectedSum-0.001 || actualSum > expectedSum+0.001 {
						t.Errorf("Expected sum ~%f, got %f", expectedSum, actualSum)
					}
				}
			}
		}
	}

	if !found {
		t.Error("Expected to find calculator_calculation_duration_seconds metric")
	}
}

func TestPrometheusMetrics_SetActiveCalculations(t *testing.T) {
	metrics := NewPrometheusMetrics()

	metrics.SetActiveCalculations(5)

	expected := `
		# HELP calculator_active_calculations Number of calculations currently in progress
		# TYPE calculator_active_calculations gauge
		calculator_active_calculations 5
	`

	if err := testutil.GatherAndCompare(
		prometheus.DefaultGatherer,
		strings.NewReader(expected),
		"calculator_active_calculations",
	); err != nil {
		t.Errorf("Unexpected metric value: %v", err)
	}

	metrics.SetActiveCalculations(3)

	expected = `
		# HELP calculator_active_calculations Number of calculations currently in progress
		# TYPE calculator_active_calculations gauge
		calculator_active_calculations 3
	`

	if err := testutil.GatherAndCompare(
		prometheus.DefaultGatherer,
		strings.NewReader(expected),
		"calculator_active_calculations",
	); err != nil {
		t.Errorf("Unexpected metric value after update: %v", err)
	}
}

func TestPrometheusMetrics_IncrementCalculationErrors(t *testing.T) {
	metrics := NewPrometheusMetrics()

	metrics.IncrementCalculationErrors("dough_calculation", "invalid_input")
	metrics.IncrementCalculationErrors("dough_calculation", "invalid_input")
	metrics.IncrementCalculationErrors("dough_calculation", "timeout")
	metrics.IncrementCalculationErrors("ingredient_calculation", "missing_data")

	expected := `
		# HELP calculator_calculation_errors_total Total number of calculation errors
		# TYPE calculator_calculation_errors_total counter
		calculator_calculation_errors_total{error_type="invalid_input",type="dough_calculation"} 2
		calculator_calculation_errors_total{error_type="timeout",type="dough_calculation"} 1
		calculator_calculation_errors_total{error_type="missing_data",type="ingredient_calculation"} 1
	`

	if err := testutil.GatherAndCompare(
		prometheus.DefaultGatherer,
		strings.NewReader(expected),
		"calculator_calculation_errors_total",
	); err != nil {
		t.Errorf("Unexpected metric value: %v", err)
	}
}

func TestPrometheusMetrics_RecordDoughAccuracy(t *testing.T) {
	metrics := NewPrometheusMetrics()

	metrics.RecordDoughAccuracy(95.5)
	metrics.RecordDoughAccuracy(98.2)
	metrics.RecordDoughAccuracy(92.1)

	metricFamily, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		t.Fatalf("Failed to gather metrics: %v", err)
	}

	var found bool
	for _, mf := range metricFamily {
		if mf.GetName() == "calculator_dough_accuracy_percentage" {
			found = true
			for _, metric := range mf.GetMetric() {
				if metric.GetHistogram() != nil {
					if metric.GetHistogram().GetSampleCount() != 3 {
						t.Errorf("Expected 3 observations, got %d",
							metric.GetHistogram().GetSampleCount())
					}
					expectedSum := 285.8
					actualSum := metric.GetHistogram().GetSampleSum()
					if actualSum < expectedSum-0.1 || actualSum > expectedSum+0.1 {
						t.Errorf("Expected sum ~%f, got %f", expectedSum, actualSum)
					}
				}
			}
		}
	}

	if !found {
		t.Error("Expected to find calculator_dough_accuracy_percentage metric")
	}
}

func TestPrometheusMetrics_IncrementIngredientValidations(t *testing.T) {
	metrics := NewPrometheusMetrics()

	metrics.IncrementIngredientValidations("flour", true)
	metrics.IncrementIngredientValidations("flour", true)
	metrics.IncrementIngredientValidations("flour", false)
	metrics.IncrementIngredientValidations("water", true)
	metrics.IncrementIngredientValidations("salt", false)

	expected := `
		# HELP calculator_ingredient_validations_total Total number of ingredient validations
		# TYPE calculator_ingredient_validations_total counter
		calculator_ingredient_validations_total{ingredient="flour",valid="true"} 2
		calculator_ingredient_validations_total{ingredient="flour",valid="false"} 1
		calculator_ingredient_validations_total{ingredient="water",valid="true"} 1
		calculator_ingredient_validations_total{ingredient="salt",valid="false"} 1
	`

	if err := testutil.GatherAndCompare(
		prometheus.DefaultGatherer,
		strings.NewReader(expected),
		"calculator_ingredient_validations_total",
	); err != nil {
		t.Errorf("Unexpected metric value: %v", err)
	}
}

func TestPrometheusMetrics_RecordDoughWeight(t *testing.T) {
	metrics := NewPrometheusMetrics()

	metrics.RecordDoughWeight(500.0)
	metrics.RecordDoughWeight(750.0)
	metrics.RecordDoughWeight(1000.0)

	metricFamily, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		t.Fatalf("Failed to gather metrics: %v", err)
	}

	var found bool
	for _, mf := range metricFamily {
		if mf.GetName() == "calculator_dough_weight_grams" {
			found = true
			for _, metric := range mf.GetMetric() {
				if metric.GetHistogram() != nil {
					if metric.GetHistogram().GetSampleCount() != 3 {
						t.Errorf("Expected 3 observations, got %d",
							metric.GetHistogram().GetSampleCount())
					}
					expectedSum := 2250.0
					actualSum := metric.GetHistogram().GetSampleSum()
					if actualSum < expectedSum-1.0 || actualSum > expectedSum+1.0 {
						t.Errorf("Expected sum ~%f, got %f", expectedSum, actualSum)
					}
				}
			}
		}
	}

	if !found {
		t.Error("Expected to find calculator_dough_weight_grams metric")
	}
}

func TestPrometheusMetrics_IncrementRecipeTypes(t *testing.T) {
	metrics := NewPrometheusMetrics()

	metrics.IncrementRecipeTypes("sourdough")
	metrics.IncrementRecipeTypes("sourdough")
	metrics.IncrementRecipeTypes("pizza")
	metrics.IncrementRecipeTypes("bread")
	metrics.IncrementRecipeTypes("pizza")

	expected := `
		# HELP calculator_recipe_types_total Total number of calculations by recipe type
		# TYPE calculator_recipe_types_total counter
		calculator_recipe_types_total{recipe_type="sourdough"} 2
		calculator_recipe_types_total{recipe_type="pizza"} 2
		calculator_recipe_types_total{recipe_type="bread"} 1
	`

	if err := testutil.GatherAndCompare(
		prometheus.DefaultGatherer,
		strings.NewReader(expected),
		"calculator_recipe_types_total",
	); err != nil {
		t.Errorf("Unexpected metric value: %v", err)
	}
}

func TestPrometheusMetrics_GRPCMetrics(t *testing.T) {
	metrics := NewPrometheusMetrics()

	metrics.IncrementGRPCRequests("CalculateDough", "success")
	metrics.IncrementGRPCRequests("CalculateDough", "success")
	metrics.IncrementGRPCRequests("CalculateDough", "error")
	metrics.IncrementGRPCRequests("ValidateIngredients", "success")

	metrics.RecordGRPCDuration("CalculateDough", 50*time.Millisecond)
	metrics.RecordGRPCDuration("ValidateIngredients", 25*time.Millisecond)

	expected := `
		# HELP calculator_grpc_requests_total Total number of gRPC requests
		# TYPE calculator_grpc_requests_total counter
		calculator_grpc_requests_total{method="CalculateDough",status="success"} 2
		calculator_grpc_requests_total{method="CalculateDough",status="error"} 1
		calculator_grpc_requests_total{method="ValidateIngredients",status="success"} 1
	`

	if err := testutil.GatherAndCompare(
		prometheus.DefaultGatherer,
		strings.NewReader(expected),
		"calculator_grpc_requests_total",
	); err != nil {
		t.Errorf("Unexpected gRPC requests metric: %v", err)
	}

	metricFamily, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		t.Fatalf("Failed to gather metrics: %v", err)
	}

	var found bool
	for _, mf := range metricFamily {
		if mf.GetName() == "calculator_grpc_request_duration_seconds" {
			found = true
			// We expect at least 2 observations total across all methods
			totalSamples := uint64(0)
			for _, metric := range mf.GetMetric() {
				if metric.GetHistogram() != nil {
					totalSamples += metric.GetHistogram().GetSampleCount()
				}
			}
			if totalSamples < 2 {
				t.Errorf("Expected at least 2 total observations, got %d", totalSamples)
			}
		}
	}

	if !found {
		t.Error("Expected to find calculator_grpc_request_duration_seconds metric")
	}
}
