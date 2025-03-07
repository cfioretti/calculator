package application

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"

	bdomain "github.com/cfioretti/calculator/pkg/domain"
)

func TestTotalDoughWeightByPans(t *testing.T) {
	tests := []struct {
		name     string
		input    bdomain.Pans
		wantArea float64
		wantErr  bool
	}{
		{
			name: "success with single pan",
			input: bdomain.Pans{
				Pans: []bdomain.Pan{
					{
						Shape: "rectangular",
						Measures: bdomain.Measures{
							Width:  intPtr(30),
							Length: intPtr(40),
						},
					},
				},
			},
			wantArea: 1200,
			wantErr:  false,
		},
		{
			name: "success with multiple pans",
			input: bdomain.Pans{
				Pans: []bdomain.Pan{
					{
						Shape: "round",
						Measures: bdomain.Measures{
							Diameter: intPtr(20),
						},
					},
					{
						Shape: "square",
						Measures: bdomain.Measures{
							Edge: intPtr(20),
						},
					},
				},
			},
			wantArea: math.Pi*100 + 400,
			wantErr:  false,
		},
		{
			name: "invalid shape",
			input: bdomain.Pans{
				Pans: []bdomain.Pan{
					{
						Shape: "triangle",
						Measures: bdomain.Measures{
							Width:  intPtr(20),
							Length: intPtr(30),
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid measures",
			input: bdomain.Pans{
				Pans: []bdomain.Pan{
					{
						Shape: "round",
						Measures: bdomain.Measures{
							Diameter: nil,
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "empty pans",
			input: bdomain.Pans{
				Pans: []bdomain.Pan{},
			},
			wantArea: 0,
			wantErr:  false,
		},
	}

	calculator := NewCalculatorService()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := calculator.TotalDoughWeightByPans(tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.InDelta(t, tt.wantArea, result.TotalArea, 0.001)
		})
	}
}

func intPtr(value int) *int {
	return &value
}
