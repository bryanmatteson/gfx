package gfx

type Stroke struct {
	LineCap   LineCap
	LineJoin  LineJoin
	LineWidth float64

	MiterLimit float64
	DashPhase  float64
	Dashes     []float64
}
