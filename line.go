package gfx

import "math"

type Lines []Line

func (l Lines) Bounds() Rect {
	var minx, miny, maxx, maxy float64
	minx, miny = math.Inf(1), math.Inf(1)
	maxx, maxy = math.Inf(-1), math.Inf(-1)

	for _, v := range l {
		minx = math.Min(minx, v.Start.X)
		minx = math.Min(minx, v.End.X)
		maxx = math.Max(maxx, v.Start.X)
		maxx = math.Max(maxx, v.End.X)
		miny = math.Min(miny, v.Start.Y)
		miny = math.Min(miny, v.End.Y)
		maxy = math.Max(maxy, v.Start.Y)
		maxy = math.Max(maxy, v.End.Y)
	}
	return MakeRect(minx, miny, maxx, maxy)
}

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
