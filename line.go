package gfx

import (
	"math"
	"sort"
)

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

func (l Lines) NormalizeDirection() {
	for i := 0; i < len(l); i++ {
		line := l[i]
		switch {
		case line.IsHorizontal():
			if line.Start.X > line.End.X {
				line.Start.X, line.End.X = line.End.X, line.Start.X
			}
		case line.IsVertical():
			if line.Start.Y > line.End.Y {
				line.Start.Y, line.End.Y = line.End.Y, line.Start.Y
			}
		default:
			continue
		}
		l[i] = line
	}
}

func (l *Lines) Coalesce(maxLineGap float64, tolerance float64) {
	horizontalLines := map[float64]Lines{}
	verticalLines := map[float64]Lines{}

	var results Lines

	for _, line := range *l {
		switch {
		case line.IsHorizontal():
			if line.Start.X > line.End.X {
				line.Start.X, line.End.X = line.End.X, line.Start.X
			}
			handled := false
			for key := range horizontalLines {
				if ApproxEqual(key, line.Start.Y, tolerance) {
					horizontalLines[key] = append(horizontalLines[key], line)
					handled = true
					break
				}
			}

			if !handled {
				horizontalLines[line.Start.Y] = append(horizontalLines[line.Start.Y], line)
			}
		case line.IsVertical():
			if line.Start.Y > line.End.Y {
				line.Start.Y, line.End.Y = line.End.Y, line.Start.Y
			}
			handled := false
			for key := range verticalLines {
				if ApproxEqual(key, line.Start.X, tolerance) {
					verticalLines[key] = append(verticalLines[key], line)
					handled = true
					break
				}
			}

			if !handled {
				verticalLines[line.Start.X] = append(verticalLines[line.Start.X], line)
			}
		default:
			results = append(results, line)
		}
	}

	for y, lines := range horizontalLines {
		if len(lines) == 1 {
			results = append(results, lines...)
			continue
		}

		lines.SortAscendingX()
		var last Line

		for i, cur := range lines {
			if i == 0 {
				last = cur
				continue
			}

			// check intersection or within gap distance
			if last.End.X >= cur.Start.X && last.Start.X <= cur.End.X || math.Abs(cur.Start.X-last.End.X) < maxLineGap {
				last = MakeLine(math.Min(last.Start.X, cur.Start.X), y, math.Max(cur.End.X, last.End.X), y)
				continue
			}

			results = append(results, last)
			last = cur
		}
		results = append(results, last)
	}

	for x, lines := range verticalLines {
		if len(lines) == 1 {
			results = append(results, lines...)
			continue
		}

		lines.SortAscendingY()
		var last Line

		for i, cur := range lines {
			if i == 0 {
				last = cur
				continue
			}

			// check intersection or within gap distance
			if last.End.Y >= cur.Start.Y && last.Start.Y <= cur.End.Y || math.Abs(cur.Start.Y-last.End.Y) < maxLineGap {
				last = MakeLine(x, math.Min(last.Start.Y, cur.Start.Y), x, math.Max(cur.End.Y, last.End.Y))
				continue
			}

			results = append(results, last)
			last = cur
		}
		results = append(results, last)
	}

	*l = results
}

func (l Lines) SortAscendingX() {
	l.Sort(func(lhs, rhs *Line) bool { return lhs.Start.X < rhs.Start.X })
}

func (l Lines) SortAscendingY() {
	l.Sort(func(lhs, rhs *Line) bool { return lhs.Start.Y < rhs.Start.Y })
}

func (l Lines) SortDescendingX() {
	l.Sort(func(lhs, rhs *Line) bool { return lhs.Start.X > rhs.Start.X })
}

func (l Lines) SortDescendingY() {
	l.Sort(func(lhs, rhs *Line) bool { return lhs.Start.Y > rhs.Start.Y })
}

func (l Lines) SortAscendingLength() {
	l.Sort(func(lhs, rhs *Line) bool { return lhs.Length() < rhs.Length() })
}

func (l Lines) Sort(less func(lhs, rhs *Line) bool) { sort.Sort(&sortLines{l, less}) }

type sortLines struct {
	lines Lines
	less  func(lhs, rhs *Line) bool
}

func (s *sortLines) Len() int           { return len(s.lines) }
func (s *sortLines) Less(i, j int) bool { return s.less(&s.lines[i], &s.lines[j]) }
func (s *sortLines) Swap(i, j int)      { s.lines[i], s.lines[j] = s.lines[j], s.lines[i] }
