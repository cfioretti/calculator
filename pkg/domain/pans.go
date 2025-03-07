package domain

type Pans struct {
	Pans      []Pan
	TotalArea float64
}

type Pan struct {
	Shape    string
	Measures Measures
	Name     string
	Area     float64
}

type Measures struct {
	Diameter *int
	Edge     *int
	Width    *int
	Length   *int
}
