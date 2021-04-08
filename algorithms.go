package gfx

import (
	"sort"
)

func Partition(starts, ends []float64) (ranges []Range) {
	all := append(starts, ends...)
	uniq := make(map[float64]struct{})

	for _, v := range all {
		uniq[v] = struct{}{}
	}

	events := make([]float64, 0, len(uniq))
	for y := range uniq {
		events = append(events, y)
	}
	sort.Float64s(events)

	contains := func(set []float64, e float64) bool {
		for _, s := range set {
			if s == e {
				return true
			}
		}
		return false
	}

	count, start := 0, 0.0
	for _, evt := range events {
		if contains(starts, evt) {
			if count == 0 {
				start = evt
			}
			count++
		}
		if contains(ends, evt) {
			count--
		}
		if count == 0 {
			ranges = append(ranges, Range{start, evt})
		}
	}

	return
}

func PartitionRectRows(rects Rects) []Range {
	starts, ends := make([]float64, len(rects)), make([]float64, len(rects))

	for i, r := range rects {
		starts[i], ends[i] = r.Y.Min, r.Y.Max
	}

	return Partition(starts, ends)
}

func PartitionRectColumns(rects Rects) []Range {
	starts, ends := make([]float64, len(rects)), make([]float64, len(rects))

	for i, r := range rects {
		starts[i], ends[i] = r.X.Min, r.X.Max
	}

	return Partition(starts, ends)
}

func PartitionLineRows(lines Lines) []Range {
	starts, ends := make([]float64, len(lines)), make([]float64, len(lines))

	for i, line := range lines {
		starts[i], ends[i] = line.Start.Y, line.End.Y
	}

	return Partition(starts, ends)
}

func PartitionLineColumns(lines Lines) []Range {
	starts, ends := make([]float64, len(lines)), make([]float64, len(lines))

	for i, line := range lines {
		starts[i], ends[i] = line.Start.X, line.End.X
	}

	return Partition(starts, ends)
}
