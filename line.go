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

func (l Line) Angle() float64        { return math.Atan2(l.End.Y-l.Start.Y, l.End.X-l.Start.X) }
func (l Line) IsAxisAligned() bool   { return l.End.X == l.Start.X || l.End.Y == l.Start.Y }
func (l Line) IsVertical() bool      { return l.End.X == l.Start.X }
func (l Line) IsHorizontal() bool    { return l.End.Y == l.Start.Y }
func (l Line) Contains(p Point) bool { return l.Closest(p).Eq(p) }

func (l Line) Closest(v Point) Point {
	// between is a helper function which determines whether x is greater than min(a, b) and less than max(a, b)
	between := func(a, b, x float64) bool {
		min := math.Min(a, b)
		max := math.Max(a, b)
		return min < x && x < max
	}

	// Closest point will be on a line which perpendicular to this line.
	// If and only if the infinite perpendicular line intersects the segment.
	m, b := l.Formula()

	// Account for horizontal lines
	if m == 0 {
		x := v.X
		y := l.Start.Y

		// check if the X coordinate of v is on the line
		if between(l.Start.X, l.End.X, v.X) {
			return MakePoint(x, y)
		}

		// Otherwise get the closest endpoint
		if l.Start.Sub(v).VectorLength() < l.End.Sub(v).VectorLength() {
			return l.Start
		}
		return l.End
	}

	// Account for vertical lines
	if math.IsInf(math.Abs(m), 1) {
		x := l.Start.X
		y := v.Y

		// check if the Y coordinate of v is on the line
		if between(l.Start.Y, l.End.Y, v.Y) {
			return MakePoint(x, y)
		}

		// Otherwise get the closest endpoint
		if l.Start.Sub(v).VectorLength() < l.End.Sub(v).VectorLength() {
			return l.Start
		}
		return l.End
	}

	perpendicularM := -1 / m
	perpendicularB := v.Y - (perpendicularM * v.X)

	// Coordinates of intersect (of infinite lines)
	x := (perpendicularB - b) / (m - perpendicularM)
	y := m*x + b

	// Check if the point lies between the x and y bounds of the segment
	if !between(l.Start.X, l.End.X, x) && !between(l.Start.Y, l.End.Y, y) {
		// Not within bounding box
		toStart := v.Sub(l.Start)
		toEnd := v.Sub(l.End)

		if toStart.VectorLength() < toEnd.VectorLength() {
			return l.Start
		}
		return l.End
	}

	return MakePoint(x, y)
}

// Formula will return the values that represent the line in the formula: y = mx + b
// This function will return math.Inf+, math.Inf- for a vertical line.
func (l Line) Formula() (m, b float64) {
	// Account for horizontal lines
	if l.End.Y == l.Start.Y {
		return 0, l.Start.Y
	}

	m = (l.End.Y - l.Start.Y) / (l.End.X - l.Start.X)
	b = l.Start.Y - (m * l.Start.X)

	return m, b
}

func (l Line) Intersects(k Line) bool {
	_, ok := l.Intersection(k)
	return ok
}

func (l Line) Intersection(k Line) (Point, bool) {
	// Check if the lines are parallel
	lDir := l.Start.Sub(l.End)
	kDir := k.Start.Sub(k.End)

	if lDir.X == kDir.X && lDir.Y == kDir.Y {
		return Point{}, false
	}

	// The lines intersect - but potentially not within the line segments.
	// Get the intersection point for the lines if they were infinitely long, check if the point exists on both of the
	// segments
	lm, lb := l.Formula()
	km, kb := k.Formula()

	// Account for vertical lines
	if math.IsInf(math.Abs(lm), 1) && math.IsInf(math.Abs(km), 1) {
		// Both vertical, therefore parallel
		return Point{}, false
	}

	var x, y float64

	if math.IsInf(math.Abs(lm), 1) || math.IsInf(math.Abs(km), 1) {
		// One line is vertical
		intersectM := lm
		intersectB := lb
		verticalLine := k

		if math.IsInf(math.Abs(lm), 1) {
			intersectM = km
			intersectB = kb
			verticalLine = l
		}

		y = intersectM*verticalLine.Start.X + intersectB
		x = verticalLine.Start.X
	} else {
		// Coordinates of intersect
		x = (kb - lb) / (lm - km)
		y = lm*x + lb
	}

	if l.Contains(MakePoint(x, y)) && k.Contains(MakePoint(x, y)) {
		// The intersect point is on both line segments, they intersect.
		return MakePoint(x, y), true
	}

	return Point{}, false
}

type Lines []Line

func (l Lines) GroupIntersecting(minIntersections int) (results []Lines) {
	if len(l) == 0 {
		return
	}

	clusters := make([][]int, 0)
	status := make(map[int]int, len(l))

	regionQuery := func(id int) (neighbors []int) {
		obj := l[id]
		for i := 0; i < len(l); i++ {
			if id == i {
				continue
			}
			if obj.Intersects(l[i]) {
				neighbors = append(neighbors, i)
			}
		}
		return
	}

	var expandCluster func(id int, cluster int, neighbors []int)
	expandCluster = func(id int, cluster int, neighbors []int) {
		clusters[cluster-1] = append(clusters[cluster-1], id)
		status[id] = cluster

		for i := 0; i < len(neighbors); i++ {
			nid := neighbors[i]
			if _, ok := status[nid]; !ok {
				status[nid] = 0
				curneighbors := regionQuery(nid)
				if len(curneighbors) >= minIntersections {
					expandCluster(nid, cluster, curneighbors)
				}
			}

			if status[nid] < 1 {
				status[nid] = cluster
				clusters[cluster-1] = append(clusters[cluster-1], nid)
			}
		}
	}

	for i := 0; i < len(l); i++ {
		if _, ok := status[i]; ok {
			continue
		}

		status[i] = 0
		neighbors := regionQuery(i)
		if len(neighbors) >= minIntersections {
			clusters = append(clusters, make([]int, 0))
			clusterid := len(clusters)
			expandCluster(i, clusterid, neighbors)
		}
	}

	results = make([]Lines, len(clusters))
	for i, grp := range clusters {
		for _, idx := range grp {
			results[i] = append(results[i], l[idx])
		}
	}

	return
}

func (l Lines) Bounds() Rect {
	if len(l) == 0 {
		return EmptyRect()
	}

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
	if len(l) == 0 {
		return
	}

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

func (l *Lines) Smooth(maxLineGap float64, tolerance float64) {
	if len(*l) == 0 {
		return
	}

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
