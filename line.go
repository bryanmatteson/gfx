package gfx

import "math"

type Lines []Line

type Line struct {
	Start, End Point
}

func MakeLine(x0, y0, x1, y1 float64) Line {
	return Line{Point{x0, y0}, Point{x1, y1}}
}

func (l Line) Length() float64 {
	dx := l.Start.X - l.End.X
	dy := l.Start.Y - l.End.Y
	return math.Sqrt(dx*dx + dy*dy)
}

func (l Line) Angle() float64      { return math.Atan2(l.End.Y-l.Start.Y, l.End.X-l.Start.X) }
func (l Line) IsAxisAligned() bool { return l.End.X == l.Start.X || l.End.Y == l.Start.Y }
func (l Line) IsVertical() bool    { return l.End.X == l.Start.X }
func (l Line) IsHorizontal() bool  { return l.End.Y == l.Start.Y }
