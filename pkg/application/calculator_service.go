package application

import (
	"context"
	"errors"

	"github.com/cfioretti/calculator/internal/domain/strategies"
	"github.com/cfioretti/calculator/pkg/domain"
)

type DoughCalculatorService struct{}

func NewCalculatorService() *DoughCalculatorService {
	return &DoughCalculatorService{}
}

type Input struct {
	Pans []PanInput `json:"pans"`
}

type PanInput struct {
	Shape    string                 `json:"shape"`
	Measures map[string]interface{} `json:"measures"`
}

func (dc DoughCalculatorService) TotalDoughWeightByPans(ctx context.Context, body domain.Pans) (*domain.Pans, error) {
	var result domain.Pans
	for _, item := range body.Pans {
		strategy, err := strategies.GetStrategy(item.Shape)
		if err != nil {
			return nil, errors.New("unsupported shape")
		}

		pan, err := strategy.Calculate(item.Measures)
		if err != nil {
			return nil, errors.New("error processing pan")
		}

		result.Pans = append(result.Pans, pan)
		result.TotalArea += pan.Area
	}
	return &result, nil
}
