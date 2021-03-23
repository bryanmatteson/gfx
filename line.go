package gfx

import "math"

type Lines []Line

type Line struct {
	Pt1, Pt2 Point
}

func (l Line) Length() float64 {
	dx := l.Pt1.X - l.Pt2.X
	dy := l.Pt1.Y - l.Pt2.Y
	return math.Sqrt(dx*dx + dy*dy)
}
