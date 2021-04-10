package gfx

import (
	"fmt"
	"math"
)

// PathComponent represents component of a path
type PathComponent int

func (c PathComponent) PointCount() int {
	switch c {
	case MoveToComp, LineToComp:
		return 1
	case QuadCurveToComp:
		return 2
	case CubicCurveToComp:
		return 3
	default:
		return 0
	}
}

// Command kinds
const (
	MoveToComp PathComponent = iota
	LineToComp
	CubicCurveToComp
	QuadCurveToComp
	ClosePathComp
)

type Path struct {
	Components []PathComponent
	Points     []Point
	x, y       float64
}

func (p *Path) appendToPath(cmd PathComponent, points ...Point) {
	p.Components = append(p.Components, cmd)
	p.Points = append(p.Points, points...)
}

// LastPoint returns the current point of the current path
func (p *Path) LastPoint() (x, y float64) {
	return p.x, p.y
}

// MoveTo starts a new path at (x, y) position
func (p *Path) MoveTo(x, y float64) {
	p.appendToPath(MoveToComp, Point{x, y})
	p.x = x
	p.y = y
}

// LineTo adds a line to the current path
func (p *Path) LineTo(x, y float64) {
	if len(p.Components) == 0 {
		p.MoveTo(x, y)
		return
	}

	p.appendToPath(LineToComp, Point{x, y})
	p.x = x
	p.y = y
}

// QuadCurveTo adds a quadratic bezier curve to the current path
func (p *Path) QuadCurveTo(cx, cy, x, y float64) {
	if len(p.Components) == 0 { //special case when no move has been done
		p.MoveTo(x, y)
		return
	}

	p.appendToPath(QuadCurveToComp, Point{cx, cy}, Point{x, y})
	p.x = x
	p.y = y
}

// CubicCurveTo adds a cubic bezier curve to the current path
func (p *Path) CubicCurveTo(cx1, cy1, cx2, cy2, x, y float64) {
	if len(p.Components) == 0 { //special case when no move has been done
		p.MoveTo(x, y)
		return
	}

	p.appendToPath(CubicCurveToComp, Point{cx1, cy1}, Point{cx2, cy2}, Point{x, y})
	p.x = x
	p.y = y
}

// Close closes the current path
func (p *Path) Close() {
	p.appendToPath(ClosePathComp)
}

// Copy make a clone of the current path and return it
func (p *Path) Copy() (dest *Path) {
	dest = new(Path)
	dest.Components = make([]PathComponent, len(p.Components))
	copy(dest.Components, p.Components)
	dest.Points = make([]Point, len(p.Points))
	copy(dest.Points, p.Points)
	dest.x, dest.y = p.x, p.y
	return dest
}

// Clear reset the path
func (p *Path) Clear() {
	p.Components = p.Components[0:0]
	p.Points = p.Points[0:0]
}

// IsEmpty returns true if the path is empty
func (p *Path) IsEmpty() bool {
	return len(p.Components) == 0
}

func (p *Path) IsRect() bool {
	moved, closed := false, false
	lineCount := 0
	for _, cmp := range p.Components {
		switch cmp {
		case LineToComp:
			lineCount++
		case QuadCurveToComp, CubicCurveToComp:
			return false
		case MoveToComp:
			if moved {
				return false
			}
			moved = true
		case ClosePathComp:
			if closed {
				return false
			}
			closed = true
		}
	}
	if lineCount < 3 || lineCount > 4 {
		return false
	}

	lines := make([]Line, 0, 4)
	p.Walk(NewLineWalker(&lines, IdentityMatrix, nil))

	if len(lines) != 4 {
		return false
	}

	vcount, hcount := 0, 0
	for _, line := range lines {
		if line.IsHorizontal() {
			hcount++
		} else if line.IsVertical() {
			vcount++
		}
	}

	return vcount == 2 && hcount == 2
}

// String returns a debug text view of the path
func (p *Path) String() string {
	var s string
	var j = 0
	for _, cmd := range p.Components {
		switch cmd {
		case MoveToComp:
			s += fmt.Sprintf("MoveTo: %v\n", p.Points[j])
		case LineToComp:
			s += fmt.Sprintf("LineTo: %v\n", p.Points[j])
		case QuadCurveToComp:
			s += fmt.Sprintf("QuadCurveTo: %v, %v\n", p.Points[j], p.Points[j+1])
		case CubicCurveToComp:
			s += fmt.Sprintf("CubicCurveTo: %v, %v, %v\n", p.Points[j], p.Points[j+1], p.Points[j+2])
		case ClosePathComp:
			s += "Close\n"
		}
		j += cmd.PointCount()
	}
	return s
}

func (p *Path) Walk(builder PathWalker) {
	for i, j := 0, 0; i < len(p.Components); i++ {
		cmd := p.Components[i]
		switch cmd {
		case MoveToComp:
			builder.MoveTo(p.Points[j].X, p.Points[j].Y)
		case LineToComp:
			builder.LineTo(p.Points[j].X, p.Points[j].Y)
		case QuadCurveToComp:
			builder.QuadCurveTo(p.Points[j].X, p.Points[j].Y, p.Points[j+1].X, p.Points[j+1].Y)
		case CubicCurveToComp:
			builder.CubicCurveTo(p.Points[j].X, p.Points[j].Y, p.Points[j+1].X, p.Points[j+1].Y, p.Points[j+2].X, p.Points[j+2].Y)
		case ClosePathComp:
			builder.Close()
		}

		j += cmd.PointCount()
	}
}

// PathWalker describes the interface for path drawing.
type PathWalker interface {
	MoveTo(x, y float64)
	LineTo(x, y float64)
	QuadCurveTo(cx, cy, x, y float64)
	CubicCurveTo(cx1, cy1, cx2, cy2, x, y float64)
	Close()
}

func (p *Path) ApproxBounds() Rect {
	var minx, maxx, miny, maxy float64
	minx, miny = math.Inf(1), math.Inf(1)
	maxx, maxy = math.Inf(-1), math.Inf(-1)

	for i, j := 0, 0; i < len(p.Components); i++ {
		cmd := p.Components[i]

		var pt Point
		switch cmd {
		case MoveToComp, LineToComp:
			pt = p.Points[j]
		case QuadCurveToComp:
			pt = p.Points[j+1]
		case CubicCurveToComp:
			pt = p.Points[j+2]
		}

		minx = math.Min(minx, pt.X)
		maxx = math.Max(maxx, pt.X)
		miny = math.Min(miny, pt.Y)
		maxy = math.Max(maxy, pt.Y)

		j += cmd.PointCount()
	}
	return MakeRect(minx, miny, maxx, maxy)
}

func (p *Path) Bounds() Rect {
	var minx, maxx, miny, maxy float64
	minx, miny = math.Inf(1), math.Inf(1)
	maxx, maxy = math.Inf(-1), math.Inf(-1)

	var x0, y0, x1, y1, cx0, cy0, cx1, cy1, ax, ay, bx, by float64

	x0, ax = math.Inf(1), math.Inf(1)
	y0, ay = math.Inf(-1), math.Inf(-1)
	for i, j := 0, 0; i < len(p.Components); i++ {
		cmd := p.Components[i]
		switch cmd {
		case MoveToComp, LineToComp:
			x1, y1 = p.Points[j].X, p.Points[j].Y
			if math.IsInf(x0, 1) {
				x0 = x1
			}
			if math.IsInf(y0, -1) {
				y0 = y1
			}
			ax = math.Min(x0, x1)
			bx = math.Max(x0, x1)
			ay = math.Min(y0, y1)
			by = math.Min(y0, y1)
		case CubicCurveToComp:
			x1, y1 = p.Points[j+2].X, p.Points[j+2].Y
			if math.IsInf(x0, 1) {
				x0 = x1
			}
			if math.IsInf(y0, -1) {
				y0 = y1
			}

			cx0, cy0 = p.Points[j].X, p.Points[j].Y
			cx1, cy1 = p.Points[j+1].X, p.Points[j+1].Y

			bounds := bezierBounds(x0, y0, cx0, cy0, cx1, cy1, x1, y1)
			ax = bounds.X.Min
			ay = bounds.Y.Min
			bx = bounds.X.Max
			by = bounds.Y.Max
		}

		x0 = x1
		y0 = y1

		minx = math.Min(minx, ax)
		maxx = math.Max(maxx, bx)
		miny = math.Min(miny, ay)
		maxy = math.Max(maxy, by)
		j += cmd.PointCount()
	}

	return MakeRect(minx, miny, maxx, maxy)
}
