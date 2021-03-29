package gfx

type Stroke struct {
	StartCap LineCap
	DashCap  LineCap
	EndCap   LineCap

	LineJoin  LineJoin
	LineWidth float64

	MiterLimit float64
	DashPhase  float64
	Dashes     []float64
}
