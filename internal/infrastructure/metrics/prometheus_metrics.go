package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	domainMetrics "github.com/cfioretti/calculator/internal/domain/metrics"
)

type PrometheusMetrics struct {
	// Business Operations Metrics
	calculationsTotal   *prometheus.CounterVec
	calculationDuration *prometheus.HistogramVec
	activeCalculations  prometheus.Gauge
	calculationErrors   *prometheus.CounterVec

	// Quality Metrics
	doughAccuracy         prometheus.Histogram
	ingredientValidations *prometheus.CounterVec

	// Domain-specific metrics
	doughWeight    prometheus.Histogram
	doughHydration prometheus.Histogram
	recipeTypes    *prometheus.CounterVec

	// Technical metrics
	grpcRequestsTotal   *prometheus.CounterVec
	grpcRequestDuration *prometheus.HistogramVec
}

func NewPrometheusMetrics() *PrometheusMetrics {
	return &PrometheusMetrics{
		// Business Operations
		calculationsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "calculator_calculations_total",
				Help: "Total number of calculations performed",
			},
			[]string{"type"},
		),
		calculationDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "calculator_calculation_duration_seconds",
				Help:    "Duration of calculations in seconds",
				Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0},
			},
			[]string{"type"},
		),
		activeCalculations: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "calculator_active_calculations",
				Help: "Number of calculations currently in progress",
			},
		),
		calculationErrors: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "calculator_calculation_errors_total",
				Help: "Total number of calculation errors",
			},
			[]string{"type", "error_type"},
		),

		// Quality Metrics
		doughAccuracy: promauto.NewHistogram(
			prometheus.HistogramOpts{
				Name:    "calculator_dough_accuracy_percentage",
				Help:    "Accuracy of dough calculations as percentage",
				Buckets: []float64{70, 75, 80, 85, 90, 95, 97, 99, 99.5, 100},
			},
		),
		ingredientValidations: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "calculator_ingredient_validations_total",
				Help: "Total number of ingredient validations",
			},
			[]string{"ingredient", "valid"},
		),

		// Domain-specific metrics
		doughWeight: promauto.NewHistogram(
			prometheus.HistogramOpts{
				Name:    "calculator_dough_weight_grams",
				Help:    "Weight of calculated dough in grams",
				Buckets: []float64{100, 250, 500, 750, 1000, 1500, 2000, 3000, 5000, 10000},
			},
		),
		doughHydration: promauto.NewHistogram(
			prometheus.HistogramOpts{
				Name:    "calculator_dough_hydration_percentage",
				Help:    "Hydration percentage of calculated dough",
				Buckets: []float64{50, 55, 60, 65, 70, 75, 80, 85, 90, 95, 100},
			},
		),
		recipeTypes: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "calculator_recipe_types_total",
				Help: "Total number of calculations by recipe type",
			},
			[]string{"recipe_type"},
		),

		// Technical metrics
		grpcRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "calculator_grpc_requests_total",
				Help: "Total number of gRPC requests",
			},
			[]string{"method", "status"},
		),
		grpcRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "calculator_grpc_request_duration_seconds",
				Help:    "Duration of gRPC requests in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method"},
		),
	}
}

var _ domainMetrics.CalculatorMetrics = (*PrometheusMetrics)(nil)

func (m *PrometheusMetrics) IncrementCalculationsTotal(calculationType string) {
	m.calculationsTotal.WithLabelValues(calculationType).Inc()
}

func (m *PrometheusMetrics) RecordCalculationDuration(calculationType string, duration time.Duration) {
	m.calculationDuration.WithLabelValues(calculationType).Observe(duration.Seconds())
}

func (m *PrometheusMetrics) SetActiveCalculations(count int) {
	m.activeCalculations.Set(float64(count))
}

func (m *PrometheusMetrics) IncrementCalculationErrors(calculationType string, errorType string) {
	m.calculationErrors.WithLabelValues(calculationType, errorType).Inc()
}

func (m *PrometheusMetrics) RecordDoughAccuracy(accuracy float64) {
	m.doughAccuracy.Observe(accuracy)
}

func (m *PrometheusMetrics) IncrementIngredientValidations(ingredient string, valid bool) {
	validStr := "false"
	if valid {
		validStr = "true"
	}
	m.ingredientValidations.WithLabelValues(ingredient, validStr).Inc()
}

func (m *PrometheusMetrics) RecordDoughWeight(weight float64) {
	m.doughWeight.Observe(weight)
}

func (m *PrometheusMetrics) RecordDoughHydration(hydration float64) {
	m.doughHydration.Observe(hydration)
}

func (m *PrometheusMetrics) IncrementRecipeTypes(recipeType string) {
	m.recipeTypes.WithLabelValues(recipeType).Inc()
}

func (m *PrometheusMetrics) IncrementGRPCRequests(method string, status string) {
	m.grpcRequestsTotal.WithLabelValues(method, status).Inc()
}

func (m *PrometheusMetrics) RecordGRPCDuration(method string, duration time.Duration) {
	m.grpcRequestDuration.WithLabelValues(method).Observe(duration.Seconds())
}
