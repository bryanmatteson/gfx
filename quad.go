package gfx

import (
	"math"

	"github.com/ahmetb/go-linq"
)

type Quad struct {
	BottomLeft  Point
	TopLeft     Point
	TopRight    Point
	BottomRight Point
}

func MakeQuad(left, bottom, right, top float64) Quad {
	return Quad{Point{left, bottom}, Point{left, top}, Point{right, top}, Point{right, bottom}}
}

func (q Quad) Left() float64 {
	if q.TopLeft.X < q.TopRight.X {
		return q.TopLeft.X
	}
	return q.TopRight.X
}

func (q Quad) Right() float64 {
	if q.BottomRight.X < q.BottomLeft.X {
		return q.BottomLeft.X
	}
	return q.BottomRight.X
}

func (q Quad) Bottom() float64 {
	if q.BottomRight.Y < q.TopRight.Y {
		return q.BottomRight.Y
	}
	return q.TopRight.Y
}

func (q Quad) Top() float64 {
	if q.TopLeft.Y < q.BottomLeft.Y {
		return q.BottomLeft.Y
	}
	return q.TopLeft.Y
}

func (q Quad) T() float64 {
	if q.BottomRight == q.BottomLeft {
		return math.Atan2(q.TopLeft.Y-q.BottomLeft.Y, q.TopLeft.X-q.BottomLeft.X) - math.Pi/2
	}
	return math.Atan2(q.BottomRight.Y-q.BottomLeft.Y, q.BottomRight.X-q.BottomLeft.X)
}

func (q Quad) Rotation() float64 {
	return q.T() * 180 / math.Pi
}

func (q Quad) Width() float64 { return q.Right() - q.Left() }

func (q Quad) Height() float64 { return q.Top() - q.Bottom() }

func (q Quad) Orientation() Orientation {
	if EqualEpsilon(q.BottomLeft.Y, q.BottomRight.Y) {
		if q.BottomLeft.X > q.BottomRight.X {
			return PageDown
		}
		return PageUp
	}
	if EqualEpsilon(q.BottomLeft.X, q.BottomRight.X) {
		if q.BottomLeft.Y > q.BottomRight.Y {
			return PageRight
		}
		return PageLeft
	}
	return OtherOrientation
}

func (q Quad) Centroid() Point {
	cx := (q.BottomRight.X + q.TopRight.X + q.BottomLeft.X + q.TopLeft.X) / 4.0
	cy := (q.BottomRight.Y + q.TopRight.Y + q.BottomLeft.Y + q.TopLeft.Y) / 4.0
	return Point{cx, cy}
}

type Quads []Quad

func (q Quads) Normalize() (n Quad) {
	left, bottom, right, top := math.Inf(1), math.Inf(1), math.Inf(-1), math.Inf(-1)
	for _, quad := range q {
		left = math.Min(left, quad.Left())
		bottom = math.Min(bottom, quad.Bottom())
		right = math.Min(right, quad.Right())
		top = math.Min(top, quad.Top())
	}
	return MakeQuad(left, bottom, right, top)
}

func (q Quads) Orientation() (orientation Orientation) {
	if len(q) == 0 {
		return OtherOrientation
	}

	orientation = q[0].Orientation()
	if orientation == OtherOrientation {
		return
	}

	for _, quad := range q[1:] {
		if quad.Orientation() != orientation {
			return OtherOrientation
		}
	}
	return
}

func (q Quads) Union() (u Quad) {
	var left, bottom, right, top float64
	switch q.Orientation() {
	case PageUp:
		left, bottom, right, top = math.Inf(1), math.Inf(1), math.Inf(-1), math.Inf(-1)
		for _, quad := range q {
			left = math.Min(left, quad.Left())
			bottom = math.Min(bottom, quad.Bottom())
			right = math.Max(right, quad.Right())
			top = math.Max(top, quad.Top())
		}
	case PageDown:
		left, bottom, right, top = math.Inf(-1), math.Inf(-1), math.Inf(1), math.Inf(1)
		for _, quad := range q {
			right = math.Min(right, quad.Left())
			top = math.Min(top, quad.Bottom())
			left = math.Max(left, quad.Right())
			bottom = math.Max(bottom, quad.Top())
		}
	case PageLeft:
		left, bottom, right, top = math.Inf(1), math.Inf(-1), math.Inf(-1), math.Inf(1)
		for _, quad := range q {
			top = math.Min(top, quad.Left())
			left = math.Min(left, quad.Bottom())
			bottom = math.Max(bottom, quad.Right())
			right = math.Max(right, quad.Top())
		}
	case PageRight:
		left, bottom, right, top = math.Inf(-1), math.Inf(1), math.Inf(1), math.Inf(-1)
		for _, quad := range q {
			bottom = math.Min(bottom, quad.Left())
			right = math.Min(right, quad.Bottom())
			top = math.Max(top, quad.Right())
			left = math.Max(left, quad.Top())
		}
	default:
		baselines := make([]Point, 0, len(q)*2)
		var xAvg, yAvg float64
		for _, quad := range q {
			baselines = append(baselines, quad.BottomLeft, quad.BottomRight)
			xAvg += quad.BottomLeft.X + quad.BottomRight.X
			yAvg += quad.BottomLeft.Y + quad.BottomRight.Y
		}

		xAvg /= float64(len(q) * 2.0)
		yAvg /= float64(len(q) * 2.0)

		sumProduct := 0.0
		sumDiffSquaredX := 0.0

		for i := 0; i < len(baselines); i++ {
			pt := baselines[i]
			xdiff, ydiff := pt.X-xAvg, pt.Y-yAvg
			sumProduct += xdiff * ydiff
			sumDiffSquaredX += xdiff * xdiff
		}

		var cos, sin float64 = 0, 1
		if sumDiffSquaredX > 1e-3 {
			// not a vertical line
			angle := math.Atan(sumProduct / sumDiffSquaredX) // -π/2 ≤ θ ≤ π/2
			cos = math.Cos(angle)
			sin = math.Sin(angle)
		}

		trm := NewMatrix(cos, -sin, sin, cos, 0, 0)
		pts := make([]Point, 0, len(q)*4)
		linq.From(q).SelectMany(func(i interface{}) linq.Query {
			return linq.From([]Point{i.(Quad).BottomLeft, i.(Quad).BottomRight, i.(Quad).TopLeft, i.(Quad).TopRight})
		}).Distinct().Select(func(i interface{}) interface{} { return trm.TransformPoint(i.(Point)) }).ToSlice(&pts)

		left, bottom, right, top := math.Inf(1), math.Inf(1), math.Inf(-1), math.Inf(-1)
		for _, pt := range pts {
			left = math.Min(left, pt.X)
			right = math.Min(right, pt.X)
			bottom = math.Min(bottom, pt.Y)
			top = math.Min(top, pt.Y)
		}

		aabb := MakeQuad(left, bottom, right, top)

		rotateBack := NewMatrix(cos, sin, -sin, cos, 0, 0)
		obb := Quad{
			TopLeft:     rotateBack.TransformPoint(aabb.TopLeft),
			TopRight:    rotateBack.TransformPoint(aabb.TopRight),
			BottomLeft:  rotateBack.TransformPoint(aabb.BottomLeft),
			BottomRight: rotateBack.TransformPoint(aabb.BottomRight),
		}

		obb1 := Quad{TopLeft: obb.BottomLeft, TopRight: obb.TopLeft, BottomLeft: obb.BottomRight, BottomRight: obb.TopRight}
		obb2 := Quad{TopLeft: obb.BottomRight, TopRight: obb.BottomLeft, BottomLeft: obb.TopRight, BottomRight: obb.TopLeft}
		obb3 := Quad{TopLeft: obb.TopRight, TopRight: obb.BottomRight, BottomLeft: obb.TopLeft, BottomRight: obb.BottomLeft}

		firstq := q[0]
		lastq := q[len(q)-1]

		baselineAngle := math.Atan2(lastq.BottomRight.Y-firstq.BottomLeft.Y, lastq.BottomRight.X-firstq.BottomLeft.X) * 180 / math.Pi

		deltaAngle := math.Abs(BoundAngle180(obb.Rotation() - baselineAngle))
		deltaAngle1 := math.Abs(BoundAngle180(obb1.Rotation() - baselineAngle))

		if deltaAngle1 < deltaAngle {
			deltaAngle = deltaAngle1
			obb = obb1
		}

		deltaAngle2 := math.Abs(BoundAngle180(obb2.Rotation() - baselineAngle))
		if deltaAngle2 < deltaAngle {
			deltaAngle = deltaAngle2
			obb = obb2
		}

		deltaAngle3 := math.Abs(BoundAngle180(obb3.Rotation() - baselineAngle))
		if deltaAngle3 < deltaAngle {
			obb = obb3
		}
		return obb
	}
	return MakeQuad(left, bottom, right, top)
}
