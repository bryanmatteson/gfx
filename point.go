package gfx

import (
	"fmt"
	"math"
)

type Points []Point

func (p Points) Bounds() Rect {
	var minx, miny, maxx, maxy float64
	minx, miny = math.Inf(1), math.Inf(1)
	maxx, maxy = math.Inf(-1), math.Inf(-1)

	for _, v := range p {
		minx = math.Min(minx, v.X)
		maxx = math.Max(maxx, v.X)
		miny = math.Min(miny, v.Y)
		maxy = math.Max(maxy, v.Y)
	}

	return MakeRect(minx, miny, maxx, maxy)
}

type Point struct {
	X, Y float64
}

func MakePoint(x, y float64) Point { return Point{X: x, Y: y} }

// Add returns the sum of p and op.
func (p Point) Add(op Point) Point { return Point{p.X + op.X, p.Y + op.Y} }

// Sub returns the difference of p and op.
func (p Point) Sub(op Point) Point { return Point{p.X - op.X, p.Y - op.Y} }

// Mul returns the scalar product of p and m.
func (p Point) Mul(m float64) Point { return Point{m * p.X, m * p.Y} }

// Ortho returns a counterclockwise orthogonal point with the same norm.
func (p Point) Ortho() Point { return Point{-p.Y, p.X} }

// Dot returns the dot product between p and op.
func (p Point) Dot(op Point) float64 { return p.X*op.X + p.Y*op.Y }

// Cross returns the cross product of p and op.
func (p Point) Cross(op Point) float64 { return p.X*op.Y - p.Y*op.X }

// Norm returns the vector's norm.
func (p Point) Norm() float64 { return math.Hypot(p.X, p.Y) }

// Normalize returns a unit point in the same direction as p.
func (p Point) Normalize() Point {
	if p.X == 0 && p.Y == 0 {
		return p
	}
	return p.Mul(1 / p.Norm())
}

// PerpDot returns the perp dot product between OP and OQ, ie. zero if aligned and |OP|*|OQ| if perpendicular.
func (p Point) PerpDot(q Point) float64 { return p.X*q.Y - p.Y*q.X }

func (p Point) String() string { return fmt.Sprintf("%f, %f", p.X, p.Y) }
