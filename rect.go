package gfx

type Rect struct {
	X, Y Range
}

func MakeRectWH(x, y, w, h float64) Rect {
	return MakeRectCorners(x, y, x+w, y+h)
}

func MakeRectCorners(x0, y0, x1, y1 float64) Rect {
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

func (r Rect) Width() float64  { return r.X.Length() }
func (r Rect) Height() float64 { return r.Y.Length() }
func (r Rect) IsEmpty() bool   { return r.X.IsEmpty() || r.Y.IsEmpty() }
