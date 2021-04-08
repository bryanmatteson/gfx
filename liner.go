package gfx

import (
	"math"

	"github.com/golang/freetype/raster"
	"golang.org/x/image/math/fixed"
)

// Liner receive segment definition
type Liner interface {
	// LineTo Draw a line from the current position to the point (x, y)
	LineTo(x, y float64)
}

// Flattener receive segment definition
type Flattener interface {
	// MoveTo Start a New line from the point (x, y)
	MoveTo(x, y float64)
	// LineTo Draw a line from the current position to the point (x, y)
	LineTo(x, y float64)
	// LineJoin use Round, Bevel or miter to join points
	LineJoin()
	// Close add the most recent starting point to close the path to create a polygon
	Close()
	// End mark the current line as finished so we can draw caps
	End()
}

// Flatten convert curves into straight segments keeping join segments info
func Flatten(path *Path, flattener Flattener, scale float64) {
	var startX, startY float64 = 0, 0

	for i, j := 0, 0; i < len(path.Components); i++ {
		cmd := path.Components[i]

		switch cmd {
		case MoveToComp:
			pt := path.Points[j]
			startX, startY = pt.X, pt.Y
			if j != 0 {
				flattener.End()
			}
			flattener.MoveTo(pt.X, pt.Y)

		case LineToComp:
			pt := path.Points[j]
			flattener.LineTo(pt.X, pt.Y)
			flattener.LineJoin()

		case QuadCurveToComp:
			TraceQuad(flattener, path.Points[j-1:], 0.5)
			pt := path.Points[j+1]
			flattener.LineTo(pt.X, pt.Y)

		case CubicCurveToComp:
			TraceCubic(flattener, path.Points[j-1:], 0.5)
			pt := path.Points[j+2]
			flattener.LineTo(pt.X, pt.Y)

		case ClosePathComp:
			flattener.LineTo(startX, startY)
			flattener.Close()
		}

		j += cmd.PointCount()
	}

	flattener.End()
}

// Transformer apply the Matrix transformation tr
type Transformer struct {
	Tr        Matrix
	Flattener Flattener
}

func (t Transformer) MoveTo(x, y float64) {
	t.Flattener.MoveTo(t.Tr.TransformXY(x, y))
}

func (t Transformer) LineTo(x, y float64) {
	t.Flattener.LineTo(t.Tr.TransformXY(x, y))
}

func (t Transformer) LineJoin() {
	t.Flattener.LineJoin()
}

func (t Transformer) Close() {
	t.Flattener.Close()
}

func (t Transformer) End() {
	t.Flattener.End()
}

type SegmentedPath struct {
	Points []float64
}

func (p *SegmentedPath) MoveTo(x, y float64) {
	p.Points = append(p.Points, x, y)
	// TODO need to mark this point as moveto
}

func (p *SegmentedPath) LineTo(x, y float64) {
	p.Points = append(p.Points, x, y)
}

func (p *SegmentedPath) LineJoin() {
	// TODO need to mark the current point as linejoin
}

func (p *SegmentedPath) Close() {
	// TODO Close
}

func (p *SegmentedPath) End() {
	// Nothing to do
}

type DemuxFlattener struct {
	Flatteners []Flattener
}

func (dc DemuxFlattener) MoveTo(x, y float64) {
	for _, flattener := range dc.Flatteners {
		flattener.MoveTo(x, y)
	}
}

func (dc DemuxFlattener) LineTo(x, y float64) {
	for _, flattener := range dc.Flatteners {
		flattener.LineTo(x, y)
	}
}

func (dc DemuxFlattener) LineJoin() {
	for _, flattener := range dc.Flatteners {
		flattener.LineJoin()
	}
}

func (dc DemuxFlattener) Close() {
	for _, flattener := range dc.Flatteners {
		flattener.Close()
	}
}

func (dc DemuxFlattener) End() {
	for _, flattener := range dc.Flatteners {
		flattener.End()
	}
}

type DashVertexConverter struct {
	next           Flattener
	x, y, distance float64
	dash           []float64
	currentDash    int
	dashOffset     float64
}

func NewDashConverter(dash []float64, dashOffset float64, flattener Flattener) *DashVertexConverter {
	var dasher DashVertexConverter
	dasher.dash = dash
	dasher.currentDash = 0
	dasher.dashOffset = dashOffset
	dasher.next = flattener
	return &dasher
}

func (dasher *DashVertexConverter) LineTo(x, y float64) {
	dasher.lineTo(x, y)
}

func (dasher *DashVertexConverter) MoveTo(x, y float64) {
	dasher.next.MoveTo(x, y)
	dasher.x, dasher.y = x, y
	dasher.distance = dasher.dashOffset
	dasher.currentDash = 0
}

func (dasher *DashVertexConverter) LineJoin() {
	dasher.next.LineJoin()
}

func (dasher *DashVertexConverter) Close() {
	dasher.next.Close()
}

func (dasher *DashVertexConverter) End() {
	dasher.next.End()
}

func (dasher *DashVertexConverter) lineTo(x, y float64) {
	rest := dasher.dash[dasher.currentDash] - dasher.distance
	for rest < 0 {
		dasher.distance = dasher.distance - dasher.dash[dasher.currentDash]
		dasher.currentDash = (dasher.currentDash + 1) % len(dasher.dash)
		rest = dasher.dash[dasher.currentDash] - dasher.distance
	}
	d := distance(dasher.x, dasher.y, x, y)
	for d >= rest {
		k := rest / d
		lx := dasher.x + k*(x-dasher.x)
		ly := dasher.y + k*(y-dasher.y)
		if dasher.currentDash%2 == 0 {
			// line
			dasher.next.LineTo(lx, ly)
		} else {
			// gap
			dasher.next.End()
			dasher.next.MoveTo(lx, ly)
		}
		d = d - rest
		dasher.x, dasher.y = lx, ly
		dasher.currentDash = (dasher.currentDash + 1) % len(dasher.dash)
		rest = dasher.dash[dasher.currentDash]
	}
	dasher.distance = d
	if dasher.currentDash%2 == 0 {
		// line
		dasher.next.LineTo(x, y)
	} else {
		// gap
		dasher.next.End()
		dasher.next.MoveTo(x, y)
	}
	if dasher.distance >= dasher.dash[dasher.currentDash] {
		dasher.distance = dasher.distance - dasher.dash[dasher.currentDash]
		dasher.currentDash = (dasher.currentDash + 1) % len(dasher.dash)
	}
	dasher.x, dasher.y = x, y
}

func distance(x1, y1, x2, y2 float64) float64 {
	return vectorDistance(x2-x1, y2-y1)
}

type ftLineBuilder struct {
	Adder raster.Adder
}

func (liner ftLineBuilder) MoveTo(x, y float64) {
	liner.Adder.Start(fixed.Point26_6{X: fixed.Int26_6(x * 64), Y: fixed.Int26_6(y * 64)})
}

func (liner ftLineBuilder) LineTo(x, y float64) {
	liner.Adder.Add1(fixed.Point26_6{X: fixed.Int26_6(x * 64), Y: fixed.Int26_6(y * 64)})
}

func (liner ftLineBuilder) LineJoin() {}
func (liner ftLineBuilder) Close()    {}
func (liner ftLineBuilder) End()      {}

type LineStroker struct {
	Flattener     Flattener
	HalfLineWidth float64
	Cap           LineCap
	Join          LineJoin
	vertices      []float64
	rewind        []float64
	x, y, nx, ny  float64
}

func NewLineStroker(c LineCap, j LineJoin, flattener Flattener) *LineStroker {
	l := new(LineStroker)
	l.Flattener = flattener
	l.HalfLineWidth = 0.5
	l.Cap = c
	l.Join = j
	return l
}

func (l *LineStroker) MoveTo(x, y float64) {
	l.x, l.y = x, y
}

func (l *LineStroker) LineTo(x, y float64) {
	l.line(l.x, l.y, x, y)
}

func (l *LineStroker) LineJoin() {

}

func (l *LineStroker) line(x1, y1, x2, y2 float64) {
	dx := (x2 - x1)
	dy := (y2 - y1)
	d := vectorDistance(dx, dy)
	if d != 0 {
		nx := dy * l.HalfLineWidth / d
		ny := -(dx * l.HalfLineWidth / d)
		l.appendVertex(x1+nx, y1+ny, x2+nx, y2+ny, x1-nx, y1-ny, x2-nx, y2-ny)
		l.x, l.y, l.nx, l.ny = x2, y2, nx, ny
	}
}

func (l *LineStroker) Close() {
	if len(l.vertices) > 1 {
		l.appendVertex(l.vertices[0], l.vertices[1], l.rewind[0], l.rewind[1])
	}
}

func (l *LineStroker) End() {
	if len(l.vertices) > 1 {
		l.Flattener.MoveTo(l.vertices[0], l.vertices[1])
		for i, j := 2, 3; j < len(l.vertices); i, j = i+2, j+2 {
			l.Flattener.LineTo(l.vertices[i], l.vertices[j])
		}
	}
	for i, j := len(l.rewind)-2, len(l.rewind)-1; j > 0; i, j = i-2, j-2 {
		l.Flattener.LineTo(l.rewind[i], l.rewind[j])
	}
	if len(l.vertices) > 1 {
		l.Flattener.LineTo(l.vertices[0], l.vertices[1])
	}
	l.Flattener.End()

	l.vertices = l.vertices[0:0]
	l.rewind = l.rewind[0:0]
	l.x, l.y, l.nx, l.ny = 0, 0, 0, 0

}

func (l *LineStroker) appendVertex(vertices ...float64) {
	s := len(vertices) / 2
	l.vertices = append(l.vertices, vertices[:s]...)
	l.rewind = append(l.rewind, vertices[s:]...)
}

func vectorDistance(dx, dy float64) float64 {
	return float64(math.Sqrt(dx*dx + dy*dy))
}

type LineWalker struct {
	lines  *[]Line
	filter func(x0, y0, x1, y1 float64) bool
	trm    Matrix

	curx, cury float64
}

func NewLineWalker(lines *[]Line, trm Matrix, filter func(x0, y0, x1, y1 float64) bool) PathWalker {
	return &LineWalker{lines: lines, filter: filter, trm: trm}
}

func (f *LineWalker) MoveTo(x, y float64) { f.curx, f.cury = x, y }

func (f *LineWalker) LineTo(x, y float64) {
	curx, cury := f.trm.TransformXY(f.curx, f.cury)
	f.curx, f.cury = x, y
	x, y = f.trm.TransformXY(x, y)

	if f.filter != nil && !f.filter(curx, cury, x, y) {
		return
	}

	*f.lines = append(*f.lines, MakeLine(curx, cury, x, y))
}

func (f *LineWalker) QuadCurveTo(cx, cy, x, y float64) {
	f.curx, f.cury = x, y
}

func (f *LineWalker) CubicCurveTo(cx1, cy1, cx2, cy2, x, y float64) {
	f.curx, f.cury = x, y
}

func (f *LineWalker) Close() {}
