package gfx

import (
	"sort"
)

func Partition(starts, ends []float64) (ranges []Range) {
	uniqStarts := make(map[float64]struct{})
	uniqEnds := make(map[float64]struct{})

	for _, v := range starts {
		uniqStarts[v] = struct{}{}
	}

	for _, v := range ends {
		uniqEnds[v] = struct{}{}
	}

	starts = make([]float64, 0, len(uniqStarts))
	for v := range uniqStarts {
		starts = append(starts, v)
	}
	ends = make([]float64, 0, len(uniqEnds))
	for v := range uniqEnds {
		ends = append(ends, v)
	}

	events := append(starts, ends...)
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

func PartitionRectRows(rects Rects) (results Rects) {
	starts, ends := make([]float64, len(rects)), make([]float64, len(rects))

	for i, r := range rects {
		starts[i], ends[i] = r.Y.Min, r.Y.Max
	}

	ranges := Partition(starts, ends)
	for _, r := range ranges {
		grouped := make(Rects, 0)
		for _, rect := range rects {
			if r.ContainsRange(rect.Y) {
				grouped = append(grouped, rect)
			}
		}
		results = append(results, grouped.Union())
	}
	return
}

func PartitionRectColumns(rects Rects) (results Rects) {
	starts, ends := make([]float64, len(rects)), make([]float64, len(rects))

	for i, r := range rects {
		starts[i], ends[i] = r.X.Min, r.X.Max
	}
	ranges := Partition(starts, ends)
	for _, r := range ranges {
		grouped := make(Rects, 0)
		for _, rect := range rects {
			if r.ContainsRange(rect.X) {
				grouped = append(grouped, rect)
			}
		}
		results = append(results, grouped.Union())
	}
	return
}

func PartitionLineRows(lines Lines) (results Rects) {
	starts, ends := make([]float64, len(lines)), make([]float64, len(lines))

	for i, line := range lines {
		if line.IsHorizontal() {
			continue
		}
		starts[i], ends[i] = line.Start.Y, line.End.Y
	}

	ranges := Partition(starts, ends)
	for _, r := range ranges {
		grouped := make(Rects, 0)
		for _, line := range lines {
			if r.Contains(line.Start.Y) || r.Contains(line.End.Y) {
				grouped = append(grouped, MakeRect(line.Start.X, line.Start.Y, line.End.X, line.End.Y))
			}
		}
		results = append(results, grouped.Union())
	}
	return
}

func PartitionLineColumns(lines Lines) (results Rects) {
	starts, ends := make([]float64, len(lines)), make([]float64, len(lines))

	for i, line := range lines {
		if line.IsVertical() {
			continue
		}
		starts[i], ends[i] = line.Start.X, line.End.X
	}

	ranges := Partition(starts, ends)
	for _, r := range ranges {
		grouped := make(Rects, 0)
		for _, line := range lines {
			if r.Contains(line.Start.Y) || r.Contains(line.End.Y) {
				grouped = append(grouped, MakeRect(line.Start.X, line.Start.Y, line.End.X, line.End.Y))
			}
		}
		results = append(results, grouped.Union())
	}
	return
}
