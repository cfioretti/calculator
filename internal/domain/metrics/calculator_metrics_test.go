package metrics

import (
	"context"
	"testing"
	"time"
)

type MockCalculatorMetrics struct {
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

func NewMockCalculatorMetrics() *MockCalculatorMetrics {
	return &MockCalculatorMetrics{
		calculationsTotal:     make(map[string]int),
		calculationDurations:  make(map[string][]time.Duration),
		calculationErrors:     make(map[string]map[string]int),
		ingredientValidations: make(map[string]map[bool]int),
		recipeTypes:           make(map[string]int),
	}
}

func (m *MockCalculatorMetrics) IncrementCalculationsTotal(calculationType string) {
	m.calculationsTotal[calculationType]++
}

func (m *MockCalculatorMetrics) RecordCalculationDuration(calculationType string, duration time.Duration) {
	m.calculationDurations[calculationType] = append(m.calculationDurations[calculationType], duration)
}

func (m *MockCalculatorMetrics) SetActiveCalculations(count int) {
	m.activeCalculations = count
}

func (m *MockCalculatorMetrics) IncrementCalculationErrors(calculationType string, errorType string) {
	if m.calculationErrors[calculationType] == nil {
		m.calculationErrors[calculationType] = make(map[string]int)
	}
	m.calculationErrors[calculationType][errorType]++
}

func (m *MockCalculatorMetrics) RecordDoughAccuracy(accuracy float64) {
	m.doughAccuracies = append(m.doughAccuracies, accuracy)
}

func (m *MockCalculatorMetrics) IncrementIngredientValidations(ingredient string, valid bool) {
	if m.ingredientValidations[ingredient] == nil {
		m.ingredientValidations[ingredient] = make(map[bool]int)
	}
	m.ingredientValidations[ingredient][valid]++
}

func (m *MockCalculatorMetrics) RecordDoughWeight(weight float64) {
	m.doughWeights = append(m.doughWeights, weight)
}

func (m *MockCalculatorMetrics) RecordDoughHydration(hydration float64) {
	m.doughHydrations = append(m.doughHydrations, hydration)
}

func (m *MockCalculatorMetrics) IncrementRecipeTypes(recipeType string) {
	m.recipeTypes[recipeType]++
}

func (m *MockCalculatorMetrics) GetCalculationsTotal(calculationType string) int {
	return m.calculationsTotal[calculationType]
}

func (m *MockCalculatorMetrics) GetCalculationDurations(calculationType string) []time.Duration {
	return m.calculationDurations[calculationType]
}

func (m *MockCalculatorMetrics) GetCalculationErrors(calculationType string, errorType string) int {
	if m.calculationErrors[calculationType] == nil {
		return 0
	}
	return m.calculationErrors[calculationType][errorType]
}

func TestMetricsRecorder_RecordCalculation_Success(t *testing.T) {
	mockMetrics := NewMockCalculatorMetrics()
	recorder := NewMetricsRecorder(mockMetrics)

	result := CalculationResult{
		Type:            "dough_calculation",
		Duration:        100 * time.Millisecond,
		Success:         true,
		Weight:          500.0,
		Hydration:       70.0,
		Accuracy:        95.5,
		IngredientsUsed: []string{"flour", "water", "salt"},
	}

	recorder.RecordCalculation(context.Background(), result)

	if mockMetrics.calculationsTotal["dough_calculation"] != 1 {
		t.Errorf("Expected calculations total to be 1, got %d",
			mockMetrics.calculationsTotal["dough_calculation"])
	}

	if len(mockMetrics.calculationDurations["dough_calculation"]) != 1 {
		t.Errorf("Expected 1 duration record, got %d",
			len(mockMetrics.calculationDurations["dough_calculation"]))
	}
	if mockMetrics.calculationDurations["dough_calculation"][0] != 100*time.Millisecond {
		t.Errorf("Expected duration 100ms, got %v",
			mockMetrics.calculationDurations["dough_calculation"][0])
	}

	if len(mockMetrics.doughWeights) != 1 || mockMetrics.doughWeights[0] != 500.0 {
		t.Errorf("Expected dough weight 500.0, got %v", mockMetrics.doughWeights)
	}

	if len(mockMetrics.doughHydrations) != 1 || mockMetrics.doughHydrations[0] != 70.0 {
		t.Errorf("Expected dough hydration 70.0, got %v", mockMetrics.doughHydrations)
	}

	if len(mockMetrics.doughAccuracies) != 1 || mockMetrics.doughAccuracies[0] != 95.5 {
		t.Errorf("Expected dough accuracy 95.5, got %v", mockMetrics.doughAccuracies)
	}

	expectedIngredients := []string{"flour", "water", "salt"}
	for _, ingredient := range expectedIngredients {
		if mockMetrics.ingredientValidations[ingredient][true] != 1 {
			t.Errorf("Expected ingredient %s to have 1 valid validation, got %d",
				ingredient, mockMetrics.ingredientValidations[ingredient][true])
		}
	}
}

func TestMetricsRecorder_RecordCalculation_Error(t *testing.T) {
	mockMetrics := NewMockCalculatorMetrics()
	recorder := NewMetricsRecorder(mockMetrics)

	result := CalculationResult{
		Type:      "dough_calculation",
		Duration:  50 * time.Millisecond,
		Success:   false,
		ErrorType: "invalid_ingredients",
		Weight:    0, // Should not be recorded for failed calculations
		Hydration: 0, // Should not be recorded for failed calculations
		Accuracy:  0, // Should not be recorded for failed calculations
	}

	recorder.RecordCalculation(context.Background(), result)

	if mockMetrics.calculationsTotal["dough_calculation"] != 1 {
		t.Errorf("Expected calculations total to be 1, got %d",
			mockMetrics.calculationsTotal["dough_calculation"])
	}

	if len(mockMetrics.calculationDurations["dough_calculation"]) != 1 {
		t.Errorf("Expected 1 duration record, got %d",
			len(mockMetrics.calculationDurations["dough_calculation"]))
	}

	if mockMetrics.calculationErrors["dough_calculation"]["invalid_ingredients"] != 1 {
		t.Errorf("Expected 1 error record, got %d",
			mockMetrics.calculationErrors["dough_calculation"]["invalid_ingredients"])
	}

	if len(mockMetrics.doughWeights) != 0 {
		t.Errorf("Expected no dough weights for failed calculation, got %v", mockMetrics.doughWeights)
	}

	if len(mockMetrics.doughHydrations) != 0 {
		t.Errorf("Expected no dough hydrations for failed calculation, got %v", mockMetrics.doughHydrations)
	}

	if len(mockMetrics.doughAccuracies) != 0 {
		t.Errorf("Expected no dough accuracies for failed calculation, got %v", mockMetrics.doughAccuracies)
	}
}

func TestMetricsRecorder_RecordCalculation_PartialData(t *testing.T) {
	mockMetrics := NewMockCalculatorMetrics()
	recorder := NewMetricsRecorder(mockMetrics)

	result := CalculationResult{
		Type:            "quick_calculation",
		Duration:        25 * time.Millisecond,
		Success:         true,
		Weight:          0,                 // No weight provided
		Hydration:       65.0,              // Only hydration provided
		Accuracy:        0,                 // No accuracy provided
		IngredientsUsed: []string{"flour"}, // Only one ingredient
	}

	recorder.RecordCalculation(context.Background(), result)

	if len(mockMetrics.doughWeights) != 0 {
		t.Errorf("Expected no dough weights when weight is 0, got %v", mockMetrics.doughWeights)
	}

	if len(mockMetrics.doughHydrations) != 1 || mockMetrics.doughHydrations[0] != 65.0 {
		t.Errorf("Expected dough hydration 65.0, got %v", mockMetrics.doughHydrations)
	}

	if len(mockMetrics.doughAccuracies) != 0 {
		t.Errorf("Expected no dough accuracies when accuracy is 0, got %v", mockMetrics.doughAccuracies)
	}

	if mockMetrics.ingredientValidations["flour"][true] != 1 {
		t.Errorf("Expected flour to have 1 valid validation, got %d",
			mockMetrics.ingredientValidations["flour"][true])
	}
}

func TestMetricsRecorder_MultipleCalculations_AggregatesCorrectly(t *testing.T) {
	mockMetrics := NewMockCalculatorMetrics()
	recorder := NewMetricsRecorder(mockMetrics)

	results := []CalculationResult{
		{
			Type:     "dough_calculation",
			Duration: 100 * time.Millisecond,
			Success:  true,
			Weight:   500.0,
		},
		{
			Type:     "dough_calculation",
			Duration: 150 * time.Millisecond,
			Success:  true,
			Weight:   750.0,
		},
		{
			Type:      "dough_calculation",
			Duration:  80 * time.Millisecond,
			Success:   false,
			ErrorType: "invalid_input",
		},
	}

	for _, result := range results {
		recorder.RecordCalculation(context.Background(), result)
	}

	if mockMetrics.calculationsTotal["dough_calculation"] != 3 {
		t.Errorf("Expected 3 total calculations, got %d",
			mockMetrics.calculationsTotal["dough_calculation"])
	}

	if len(mockMetrics.calculationDurations["dough_calculation"]) != 3 {
		t.Errorf("Expected 3 duration records, got %d",
			len(mockMetrics.calculationDurations["dough_calculation"]))
	}

	expectedWeights := []float64{500.0, 750.0}
	if len(mockMetrics.doughWeights) != 2 {
		t.Errorf("Expected 2 weight records, got %d", len(mockMetrics.doughWeights))
	}
	for i, weight := range expectedWeights {
		if mockMetrics.doughWeights[i] != weight {
			t.Errorf("Expected weight %f at index %d, got %f",
				weight, i, mockMetrics.doughWeights[i])
		}
	}

	if mockMetrics.calculationErrors["dough_calculation"]["invalid_input"] != 1 {
		t.Errorf("Expected 1 error record, got %d",
			mockMetrics.calculationErrors["dough_calculation"]["invalid_input"])
	}
}
