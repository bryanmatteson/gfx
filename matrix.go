package gfx

import (
	"math"
)

type Matrix struct {
	A, B, C, D, E, F float64
}

var IdentityMatrix Matrix = Matrix{1, 0, 0, 1, 0, 0}

func NewMatrix(a, b, c, d, e, f float64) Matrix {
	return Matrix{a, b, c, d, e, f}
}

func NewScaleMatrix(sx, sy float64) Matrix {
	return NewMatrix(sx, 0, 0, sx, 0, 0)
}

func NewTranslationMatrix(tx, ty float64) Matrix {
	return NewMatrix(1, 0, 0, 1, tx, ty)
}

func NewRotationMatrixDeg(theta float64) Matrix {
	epsilon := math.Nextafter(1, 2) - 1
	var s, c float64

	for theta < 0 {
		theta += 360
	}

	for theta >= 360 {
		theta -= 360
	}

	if math.Abs(0-theta) < epsilon {
		s, c = 0, 1
	} else if math.Abs(90.0-theta) < epsilon {
		s, c = 1, 0
	} else if math.Abs(180.0-theta) < epsilon {
		s, c = 0, -1
	} else if math.Abs(270.0-theta) < epsilon {
		s, c = -1, 0
	} else {
		s, c = math.Sin(theta*math.Pi/180), math.Cos(theta*math.Pi/180)
	}
	return NewMatrix(c, s, -s, c, 0, 0)
}

// NewRotationMatrixRad creates a rotation transformation matrix. angle is in radian
func NewRotationMatrix(theta float64) Matrix {
	c := math.Cos(theta)
	s := math.Sin(theta)
	return Matrix{c, s, -s, c, 0, 0}
}

func (m Matrix) Translated(tx, ty float64) Matrix {
	return m.Concat(NewTranslationMatrix(tx, ty))
}

func (m Matrix) Rotated(theta float64) Matrix {
	return m.Concat(NewRotationMatrix(theta))
}

func (m Matrix) Scaled(sx, sy float64) Matrix {
	return m.Concat(NewScaleMatrix(sx, sy))
}

func (m Matrix) Determinant() float64 {
	return m.A*m.D - m.B*m.C
}

func (m Matrix) Compose(transforms ...Matrix) (result Matrix) {
	result = m
	for _, o := range transforms {
		result = result.Concat(o)
	}
	return
}

func (m Matrix) Concat(o Matrix) Matrix {
	return Matrix{
		(m.A * o.A) + (m.B * o.C),
		(m.A * o.B) + (m.B * o.D),
		(m.C * o.A) + (m.D * o.C),
		(m.C * o.B) + (m.D * o.D),
		(m.E * o.A) + (m.F * o.C) + o.E,
		(m.E * o.B) + (m.F * o.D) + o.F,
	}
}

// Inverted computes the inverse matrix
func (m Matrix) Inverted() (d Matrix) {
	det := m.Determinant()
	d.A = m.D / det
	d.B = -m.B / det
	d.C = -m.C / det
	d.D = m.A / det
	d.E = -m.E*d.A - m.F*d.C
	d.F = -m.E*d.B - m.F*d.D
	return
}

// PostScaled scales a matrix by postmultiplication.
func (m Matrix) PostScaled(sx, sy float64) Matrix {
	m.A *= sx
	m.B *= sy
	m.C *= sx
	m.D *= sy
	m.E *= sx
	m.F *= sy
	return m
}

// PreScaled scales a matrix by premultiplication.
func (m Matrix) PreScaled(sx, sy float64) Matrix {
	m.A *= sx
	m.B *= sx
	m.C *= sy
	m.D *= sy
	return m
}

func (m Matrix) Clone() Matrix { return NewMatrix(m.A, m.B, m.C, m.D, m.E, m.F) }

func (m Matrix) GetScaleFactor() (x, y float64) { return m.A, m.D }

// Expansion determines the average scaling factor of the matrix
func (m Matrix) Expansion() float64 { return math.Sqrt(math.Abs(m.Determinant())) }

// MaxExpansion finds the largest expansion performed by this matrix
func (m Matrix) MaxExpansion() float64 {
	return Max4(math.Abs(m.A), math.Abs(m.B), math.Abs(m.C), math.Abs(m.D))
}

func (m Matrix) GetTranslation() (x, y float64) { return m.E, m.F }

// GetScale computes a scale for the matrix
func (m Matrix) GetScale() float64 {
	x := 0.707106781*m.A + 0.707106781*m.B
	y := 0.707106781*m.C + 0.707106781*m.D
	return math.Sqrt(x*x + y*y)
}

func (m Matrix) TransformPoint(p Point) Point {
	return Point{
		X: (p.X * m.A) + (p.Y * m.C) + m.E,
		Y: (p.X * m.B) + (p.Y * m.D) + m.F,
	}
}

func (m Matrix) TransformVec(p Point) Point {
	return Point{
		X: (p.X * m.A) + (p.Y * m.C),
		Y: (p.X * m.B) + (p.Y * m.D),
	}
}

func (m Matrix) TransformXY(x, y float64) (xres, yres float64) {
	return (x * m.A) + (y * m.C) + m.E, (x * m.B) + (y * m.D) + m.F
}

func (m Matrix) TransformRect(r Rect) Rect {
	s := m.TransformPoint(Point{X: r.X.Min, Y: r.Y.Max})
	t := m.TransformPoint(Point{X: r.X.Min, Y: r.Y.Min})
	u := m.TransformPoint(Point{X: r.X.Max, Y: r.Y.Min})
	v := m.TransformPoint(Point{X: r.X.Max, Y: r.Y.Max})

	minX := Min4(s.X, t.X, u.X, v.X)
	minY := Min4(s.Y, t.Y, u.Y, v.Y)
	maxX := Max4(s.X, t.X, u.X, v.X)
	maxY := Max4(s.Y, t.Y, u.Y, v.Y)

	return MakeRectCorners(minX, minY, maxX, maxY)
}

func (m Matrix) TransformQuad(q Quad) (result Quad) {
	result.BottomLeft = m.TransformPoint(q.BottomLeft)
	result.BottomRight = m.TransformPoint(q.BottomRight)
	result.TopLeft = m.TransformPoint(q.TopLeft)
	result.TopRight = m.TransformPoint(q.TopRight)
	return
}

// Transform applies the transformation matrix to points. It modify the points passed in parameter.
func (m Matrix) Transform(pts ...Point) (points []Point) {
	points = make([]Point, len(pts))
	for i := 0; i < len(pts); i++ {
		points[i] = m.TransformPoint(pts[i])
	}
	return
}
