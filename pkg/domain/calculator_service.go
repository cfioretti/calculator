package domain

type DoughCalculatorService interface {
	TotalDoughWeightByPans(pans Pans) (*Pans, error)
}
