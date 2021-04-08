package gfx

import (
	"image"
	"math"
	"sort"
)

type Rect struct {
	X, Y Range
}

func ImageRect(r image.Rectangle) Rect {
	return MakeRect(float64(r.Min.X), float64(r.Min.Y), float64(r.Max.X), float64(r.Max.Y))
}

var UnitRect = MakeRect(0, 0, 1, 1)
var InfiniteRect = Rect{InfiniteRange, InfiniteRange}

func MakeRectWH(x, y, w, h float64) Rect {
	return MakeRect(x, y, x+w, y+h)
}

func MakeRect(x0, y0, x1, y1 float64) Rect {
	if x0 > x1 {
		x0, x1 = x1, x0
	}

	if y0 > y1 {
		y0, y1 = y1, y0
	}

	return Rect{Range{x0, x1}, Range{y0, y1}}
}

func EmptyRect() Rect {
	return Rect{EmptyRange(), EmptyRange()}
}

func (r Rect) IsInfiniteRect() bool {
	return r.X.IsInfinite() && r.Y.IsInfinite()
}

func (r Rect) ImageRect() image.Rectangle {
	return image.Rect(int(r.X.Min), int(r.Y.Min), int(r.X.Max), int(r.Y.Max))
}

func (r Rect) Path() *Path {
	path := &Path{}
	path.MoveTo(r.X.Min, r.Y.Min)
	path.LineTo(r.X.Min, r.Y.Max)
	path.LineTo(r.X.Max, r.Y.Max)
	path.LineTo(r.X.Max, r.Y.Min)
	path.Close()
	return path
}

// IsValid reports whether the rectangle is valid.
// This requires the width to be empty iff the height is empty.
func (r Rect) IsValid() bool {
	return r.X.IsEmpty() == r.Y.IsEmpty()
}

// Size returns the width and height of this rectangle in (x,y)-space. Empty
// rectangles have a negative width and height.
func (r Rect) Size() Point {
	return Point{r.X.Length(), r.Y.Length()}
}

// ContainsPoint reports whether the rectangle contains the given point.
// Rectangles are closed regions, i.e. they contain their boundary.
func (r Rect) ContainsPoint(p Point) bool {
	return r.X.Contains(p.X) && r.Y.Contains(p.Y)
}

// InteriorContainsPoint returns true iff the given point is contained in the interior
// of the region (i.e. the region excluding its boundary).
func (r Rect) InteriorContainsPoint(p Point) bool {
	return r.X.InteriorContains(p.X) && r.Y.InteriorContains(p.Y)
}

// Contains reports whether the rectangle contains the given rectangle.
func (r Rect) Contains(other Rect) bool {
	return r.X.ContainsRange(other.X) && r.Y.ContainsRange(other.Y)
}

// InteriorContains reports whether the interior of this rectangle contains all of the
// points of the given other rectangle (including its boundary).
func (r Rect) InteriorContains(other Rect) bool {
	return r.X.InteriorContainsRange(other.X) && r.Y.InteriorContainsRange(other.Y)
}

// Intersects reports whether this rectangle and the other rectangle have any points in common.
func (r Rect) Intersects(other Rect) bool {
	return r.X.Intersects(other.X) && r.Y.Intersects(other.Y)
}

// InteriorIntersects reports whether the interior of this rectangle intersects
// any point (including the boundary) of the given other rectangle.
func (r Rect) InteriorIntersects(other Rect) bool {
	return r.X.InteriorIntersects(other.X) && r.Y.InteriorIntersects(other.Y)
}

// Expanded returns a rectangle that has been expanded in the x-direction
// by margin.X, and in y-direction by margin.Y. If either margin is empty,
// then shrink the interval on the corresponding sides instead. The resulting
// rectangle may be empty. Any expansion of an empty rectangle remains empty.
func (r Rect) Expanded(margin Point) Rect {
	xx := r.X.Expanded(margin.X)
	yy := r.Y.Expanded(margin.Y)
	if xx.IsEmpty() || yy.IsEmpty() {
		return EmptyRect()
	}
	return Rect{xx, yy}
}

// ExpandedByMargin returns a Rect that has been expanded by the amount on all sides.
func (r Rect) ExpandedByMargin(margin float64) Rect {
	return r.Expanded(Point{margin, margin})
}

// Union returns the smallest rectangle containing the union of this rectangle and
// the given rectangle.
func (r Rect) Union(other Rect) Rect {
	return Rect{r.X.Union(other.X), r.Y.Union(other.Y)}
}

// Intersection returns the smallest rectangle containing the intersection of this
// rectangle and the given rectangle.
func (r Rect) Intersection(other Rect) Rect {
	xx := r.X.Intersection(other.X)
	yy := r.Y.Intersection(other.Y)
	if xx.IsEmpty() || yy.IsEmpty() {
		return EmptyRect()
	}

	return Rect{xx, yy}
}

func (r Rect) ContainsAllPointsInPath(p *Path) bool {
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

		if !r.ContainsPoint(pt) {
			return false
		}

		j += cmd.PointCount()
	}
	return true
}

func (r Rect) Width() float64  { return r.X.Length() }
func (r Rect) Height() float64 { return r.Y.Length() }
func (r Rect) Quad() Quad      { return MakeQuad(r.X.Min, r.Y.Min, r.X.Max, r.Y.Max) }
func (r Rect) IsEmpty() bool   { return r.X.IsEmpty() || r.Y.IsEmpty() }

type Rects []Rect

func (r Rects) GroupRows() []Rects {
	if len(r) == 0 {
		return []Rects{}
	}

	events := make(map[float64]struct{})
	for _, rect := range r {
		events[rect.Y.Min], events[rect.Y.Max] = struct{}{}, struct{}{}
	}

	ys := make([]float64, 0, len(events))
	for y := range events {
		ys = append(ys, y)
	}
	sort.Float64s(ys)

	rows := make([]Rects, 0)
	row := make(Rects, 0)

	count := 0
	for _, y := range ys {
		for _, rect := range r {
			if !(EqualEpsilon(rect.Y.Min, y) || EqualEpsilon(rect.Y.Max, y)) {
				continue
			}

			if EqualEpsilon(rect.Y.Min, y) {
				count++
			}
			if EqualEpsilon(rect.Y.Max, y) {
				count--
			}

			row = append(row, rect)
		}

		if count == 0 {
			rows = append(rows, row)
			row = nil
		}
	}

	return rows
}

func (r Rects) Coalesce() Rects {
	if len(r) == 0 {
		return Rects{}
	}

	groups := make([]map[int]struct{}, 0)

	for idx, rect := range r {
		var group map[int]struct{}
		for _, grp := range groups {
			if _, hasIdx := grp[idx]; hasIdx {
				group = grp
				break
			}
		}

		if group == nil {
			grp := make(map[int]struct{})
			grp[idx] = struct{}{}
			groups = append(groups, grp)
			group = grp
		}

		for i := idx + 1; i < len(r); i++ {
			next := r[i]
			if rect.Intersects(next) {
				group[i] = struct{}{}
			}
		}
	}

	coalesced := make(Rects, len(groups))
	for i, grp := range groups {
		var minx, miny, maxx, maxy float64
		minx, miny = math.Inf(1), math.Inf(1)
		maxx, maxy = math.Inf(-1), math.Inf(-1)

		for idx := range grp {
			minx = math.Min(minx, r[idx].X.Min)
			miny = math.Min(miny, r[idx].Y.Min)
			maxx = math.Max(maxx, r[idx].X.Max)
			maxy = math.Max(maxy, r[idx].Y.Max)
		}

		coalesced[i] = MakeRect(minx, miny, maxx, maxy)
	}

	return coalesced
}

func (r Rects) Union() (u Rect) {
	var minx, miny, maxx, maxy float64
	minx, miny = math.Inf(1), math.Inf(1)
	maxx, maxy = math.Inf(-1), math.Inf(-1)

	for _, rect := range r {
		minx = math.Min(minx, rect.X.Min)
		miny = math.Min(miny, rect.Y.Min)
		maxx = math.Max(maxx, rect.X.Max)
		maxy = math.Max(maxy, rect.Y.Max)
	}

	return MakeRect(minx, miny, maxx, maxy)
}
