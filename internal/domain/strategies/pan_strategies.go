package strategies

import (
	"errors"
	"fmt"
	"math"

	"github.com/cfioretti/calculator/pkg/domain"
)

type PanStrategy interface {
	Calculate(measures domain.Measures) (domain.Pan, error)
}

type RoundPanStrategy struct{}
type RectangularPanStrategy struct{}
type SquarePanStrategy struct{}

func GetStrategy(shape string) (PanStrategy, error) {
	switch shape {
	case "round":
		return &RoundPanStrategy{}, nil
	case "square":
		return &SquarePanStrategy{}, nil
	case "rectangular":
		return &RectangularPanStrategy{}, nil
	default:
		return nil, fmt.Errorf("unsupported shape: %s", shape)
	}
}

func (s *RoundPanStrategy) Calculate(measures domain.Measures) (domain.Pan, error) {
	if measures.Diameter == nil {
		return domain.Pan{}, errors.New("diameter is required")
	}

	shape := "round"
	radius := float64(*measures.Diameter) / 2
	area := math.Round(math.Pi*radius*radius*100) / 100
	name := fmt.Sprintf("%s %d cm", shape, *measures.Diameter)

	return domain.Pan{
		Shape:    shape,
		Measures: measures,
		Area:     area,
		Name:     name,
	}, nil
}

func (s *SquarePanStrategy) Calculate(measures domain.Measures) (domain.Pan, error) {
	if measures.Edge == nil {
		return domain.Pan{}, errors.New("edge is required")
	}

	shape := "square"
	area := float64(*measures.Edge * *measures.Edge)
	name := fmt.Sprintf("%s %d cm", shape, *measures.Edge)

	return domain.Pan{
		Shape:    shape,
		Measures: measures,
		Area:     area,
		Name:     name,
	}, nil
}

func (s *RectangularPanStrategy) Calculate(measures domain.Measures) (domain.Pan, error) {
	if measures.Width == nil || measures.Length == nil {
		return domain.Pan{}, errors.New("width and length are required")
	}

	shape := "rectangular"
	area := float64(*measures.Width * *measures.Length)
	name := fmt.Sprintf("%s %d x %d cm", shape, *measures.Width, *measures.Length)

	return domain.Pan{
		Shape:    shape,
		Measures: measures,
		Area:     area,
		Name:     name,
	}, nil
}
