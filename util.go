package gfx

import (
	"math"
)

func Min4(a float64, b float64, c float64, d float64) float64 {
	return math.Min(math.Min(a, b), math.Min(c, d))
}

func Max4(a float64, b float64, c float64, d float64) float64 {
	return math.Max(math.Max(a, b), math.Max(c, d))
}

func QuadToRect(q Quad) (r Rect) {
	r.X.Min = Min4(q.BottomLeft.X, q.BottomRight.X, q.TopLeft.X, q.TopRight.X)
	r.Y.Min = Min4(q.BottomLeft.Y, q.BottomRight.Y, q.TopLeft.Y, q.TopRight.Y)
	r.X.Max = Max4(q.BottomLeft.X, q.BottomRight.X, q.TopLeft.X, q.TopRight.X)
	r.Y.Max = Max4(q.BottomLeft.Y, q.BottomRight.Y, q.TopLeft.Y, q.TopRight.Y)
	return
}

func RectToQuad(r Rect) Quad {
	return MakeQuad(r.X.Min, r.Y.Min, r.X.Max, r.Y.Max)
}

func bezierBounds(x0, y0, x1, y1, x2, y2, x3, y3 float64) Rect {
	var tvalues, xvalues, yvalues []float64
	var a, b, c, t, t1, t2, b2ac, sqrtb2ac float64

	for i := 0; i < 2; i++ {
		if i == 0 {
			b = 6*x0 - 12*x1 + 6*x2
			a = -3*x0 + 9*x1 - 9*x2 + 3*x3
			c = 3*x1 - 3*x0
		} else {
			b = 6*y0 - 12*y1 + 6*y2
			a = -3*y0 + 9*y1 - 9*y2 + 3*y3
			c = 3*y1 - 3*y0
		}

		if math.Abs(a) < 1e-12 {
			if math.Abs(b) < 1e-12 {
				continue
			}

			t = -c / b
			if 0 < t && t < 1 {
				tvalues = append(tvalues, t)
			}
			continue
		}
		b2ac = b*b - 4*c*a
		if b2ac < 0 {
			if math.Abs(b2ac) < 1e-12 {
				t = -b / (2 * a)
				if 0 < t && t < 1 {
					tvalues = append(tvalues, t)
				}
			}
			continue
		}

		sqrtb2ac = math.Sqrt(b2ac)
		t1 = (-b + sqrtb2ac) / (2 * a)
		if 0 < t1 && t1 < 1 {
			tvalues = append(tvalues, t1)
		}

		t2 = (-b - sqrtb2ac) / (2 * a)
		if 0 < t2 && t2 < 1 {
			tvalues = append(tvalues, t2)
		}
	}

	xvalues = make([]float64, len(tvalues))
	yvalues = make([]float64, len(tvalues))

	for j := len(tvalues) - 1; j >= 0; j-- {
		t = tvalues[j]
		mt := 1 - t
		xvalues[j] = (mt * mt * mt * x0) + (3 * mt * mt * t * x1) + (3 * mt * t * t * x2) + (t * t * t * x3)
		yvalues[j] = (mt * mt * mt * y0) + (3 * mt * mt * t * y1) + (3 * mt * t * t * y2) + (t * t * t * y3)
	}

	xvalues = append(xvalues, x0, x3)
	yvalues = append(yvalues, y0, y3)

	minx, miny := math.Inf(1), math.Inf(1)
	maxx, maxy := math.Inf(-1), math.Inf(-1)

	for _, x := range xvalues {
		minx = math.Min(minx, x)
		maxx = math.Max(maxx, x)
	}

	for _, y := range yvalues {
		miny = math.Min(miny, y)
		maxy = math.Max(maxy, y)
	}

	return MakeRectCorners(minx, miny, maxx, maxy)
}

// Epsilon is the smallest number below which we assume to be zero
var Epsilon = math.Nextafter(1.0, 2.0) - 1

func EqualEpsilon(a, b float64) bool {
	return ApproxEqual(a, b, Epsilon)
}

func ApproxEqual(a, b, epsilon float64) bool {
	return math.Abs(a-b) < epsilon
}

func ApproxZero(a, epsilon float64) bool {
	return ApproxEqual(a, 0, epsilon)
}

func ZeroEpsilon(a float64) bool {
	return ApproxEqual(a, 0, Epsilon)
}

func BoundAngle180(angle float64) float64 {
	angle = math.Mod(angle+180, 360)
	if angle < 0 {
		angle += 360
	}
	return angle - 180
}

func LineAngle(p1, p2 Point) float64 {
	return math.Atan2(p2.Y-p1.Y, p2.X-p1.X) * 180 / math.Pi
}
