package metrics

import (
	"context"
	"time"
)

type CalculatorMetrics interface {
	IncrementCalculationsTotal(calculationType string)
	RecordCalculationDuration(calculationType string, duration time.Duration)
	SetActiveCalculations(count int)
	IncrementCalculationErrors(calculationType string, errorType string)

	RecordDoughAccuracy(accuracy float64)
	IncrementIngredientValidations(ingredient string, valid bool)

	RecordDoughWeight(weight float64)
	RecordDoughHydration(hydration float64)
	IncrementRecipeTypes(recipeType string)
}

type CalculationResult struct {
	Type            string
	Duration        time.Duration
	Success         bool
	ErrorType       string
	Weight          float64
	Hydration       float64
	Accuracy        float64
	IngredientsUsed []string
}

type MetricsRecorder struct {
	metrics CalculatorMetrics
}

func NewMetricsRecorder(metrics CalculatorMetrics) *MetricsRecorder {
	return &MetricsRecorder{
		metrics: metrics,
	}
}

func (r *MetricsRecorder) RecordCalculation(ctx context.Context, result CalculationResult) {
	r.metrics.IncrementCalculationsTotal(result.Type)
	r.metrics.RecordCalculationDuration(result.Type, result.Duration)

	if !result.Success {
		r.metrics.IncrementCalculationErrors(result.Type, result.ErrorType)
		return
	}

	if result.Weight > 0 {
		r.metrics.RecordDoughWeight(result.Weight)
	}

	if result.Hydration > 0 {
		r.metrics.RecordDoughHydration(result.Hydration)
	}

	if result.Accuracy > 0 {
		r.metrics.RecordDoughAccuracy(result.Accuracy)
	}

	for _, ingredient := range result.IngredientsUsed {
		r.metrics.IncrementIngredientValidations(ingredient, true)
	}
}
