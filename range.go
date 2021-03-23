package gfx

import "math"

type Range struct {
	Min, Max float64
}

// EmptyRange returns an empty range.
func EmptyRange() Range { return Range{1, 0} }

// IsEmpty reports whether the interval is empty.
func (r Range) IsEmpty() bool { return r.Min > r.Max }

func (r Range) Length() float64 { return r.Max - r.Min }

func (r Range) Contains(p float64) bool { return r.Min <= p && p <= r.Max }

func (r Range) InteriorContains(p float64) bool { return r.Min < p && p < r.Max }

// ContainsRange ..
func (r Range) ContainsRange(o Range) bool {
	if o.IsEmpty() {
		return true
	}
	return r.Min <= o.Min && o.Max <= r.Max
}

func (r Range) InteriorContainsRange(o Range) bool {
	if o.IsEmpty() {
		return true
	}

	return r.Min < o.Min && o.Max < r.Max
}

// Intersects ..
func (r Range) Intersects(o Range) bool {
	if r.Min <= o.Min {
		return o.Min <= r.Max && o.Min <= o.Max
	}
	return r.Min <= o.Max && r.Min <= r.Max
}

// InteriorIntersects returns true iff the interior of the interval contains any points in common with oi, including the latter's boundary.
func (r Range) InteriorIntersects(oi Range) bool {
	return oi.Min < r.Max && r.Min < oi.Max && r.Min < r.Max && oi.Min <= oi.Max
}

// Intersection returns the interval containing all points common to i and j.
func (r Range) Intersection(o Range) Range {
	return Range{
		Min: math.Max(r.Min, o.Min),
		Max: math.Min(r.Max, o.Max),
	}
}

// Clamp returns the closest point in the interval to the given point "p".
func (r Range) Clamp(v float64) float64 {
	return math.Max(r.Min, math.Min(r.Max, v))
}

// Expanded returns an interval that has been expanded on each side by margin.
// If margin is negative, then the function shrinks the interval on
// each side by margin instead. The resulting interval may be empty. Any
// expansion of an empty interval remains empty.
func (r Range) Expanded(margin float64) Range {
	if r.IsEmpty() {
		return r
	}
	return Range{r.Min - margin, r.Max + margin}
}

// Union returns the smallest interval that contains this interval and the given interval.
func (r Range) Union(other Range) Range {
	if r.IsEmpty() {
		return other
	}
	if other.IsEmpty() {
		return r
	}
	return Range{math.Min(r.Min, other.Min), math.Max(r.Max, other.Max)}
}
