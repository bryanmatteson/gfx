package gfx

import (
	"sort"
)

func Partition(starts, ends []float64) (ranges []Range) {
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

func PartitionLineRows(lines Lines) (results []Lines) {
	starts, ends := make([]float64, len(lines)), make([]float64, len(lines))

	for i, line := range lines {
		if line.IsHorizontal() {
			continue
		}
		starty, endy := line.Start.Y, line.End.Y
		if starty > endy {
			starty, endy = endy, starty
		}
		starts[i], ends[i] = starty, endy
	}

	ranges := Partition(starts, ends)
	for _, r := range ranges {
		grouped := make(Lines, 0)
		for _, line := range lines {
			if r.Contains(line.Start.Y) || r.Contains(line.End.Y) {
				grouped = append(grouped, line)
			}
		}
		results = append(results, grouped)
	}
	return
}

func PartitionLineColumns(lines Lines) (results []Lines) {
	starts, ends := make([]float64, len(lines)), make([]float64, len(lines))

	for i, line := range lines {
		if line.IsVertical() {
			continue
		}
		startx, endx := line.Start.X, line.End.X
		if startx > endx {
			startx, endx = endx, startx
		}
		starts[i], ends[i] = startx, endx
	}

	ranges := Partition(starts, ends)
	for _, r := range ranges {
		grouped := make(Lines, 0)
		for _, line := range lines {
			if r.Contains(line.Start.Y) || r.Contains(line.End.Y) {
				grouped = append(grouped, line)
			}
		}
		results = append(results, grouped)
	}
	return
}
