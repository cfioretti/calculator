package strategies

import (
	"github.com/cfioretti/calculator/pkg/domain"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetStrategy(t *testing.T) {
	tests := []struct {
		name    string
		shape   string
		wantErr bool
	}{
		{"round shape", "round", false},
		{"square shape", "square", false},
		{"rectangular shape", "rectangular", false},
		{"invalid shape", "triangle", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strategy, err := GetStrategy(tt.shape)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, strategy)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, strategy)
			}
		})
	}
}

func TestPanStrategies(t *testing.T) {
	tests := []struct {
		name     string
		strategy PanStrategy
		measures domain.Measures
		wantArea float64
		wantErr  bool
	}{
		{
			name:     "round 20 cm",
			strategy: &RoundPanStrategy{},
			measures: domain.Measures{Diameter: intPtr(20)},
			wantArea: 314.16,
			wantErr:  false,
		},
		{
			name:     "square 20 cm",
			strategy: &SquarePanStrategy{},
			measures: domain.Measures{Edge: intPtr(20)},
			wantArea: 400,
			wantErr:  false,
		},
		{
			name:     "rectangular 20 x 30 cm",
			strategy: &RectangularPanStrategy{},
			measures: domain.Measures{Width: intPtr(20), Length: intPtr(30)},
			wantArea: 600,
			wantErr:  false,
		},
		{
			name:     "invalid measures",
			strategy: &RoundPanStrategy{},
			measures: domain.Measures{},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pan, err := tt.strategy.Calculate(tt.measures)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.name, pan.Name)
			assert.Equal(t, tt.wantArea, pan.Area)
		})
	}
}

func intPtr(i int) *int {
	return &i
}
