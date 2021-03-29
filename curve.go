package gfx

import (
	"errors"
	"math"
)

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

const (
	CurveRecursionLimit = 32
)

// Cubic
//	x1, y1, cpx1, cpy1, cpx2, cpy2, x2, y2 float64

// SubdivideCubic a Bezier cubic curve in 2 equivalents Bezier cubic curves.
func SubdivideCubic(c, c1, c2 []float64) {
	c1[0], c1[1] = c[0], c[1]
	c2[6], c2[7] = c[6], c[7]

	c1[2] = (c[0] + c[2]) / 2
	c1[3] = (c[1] + c[3]) / 2

	midX := (c[2] + c[4]) / 2
	midY := (c[3] + c[5]) / 2

	c2[4] = (c[4] + c[6]) / 2
	c2[5] = (c[5] + c[7]) / 2

	c1[4] = (c1[2] + midX) / 2
	c1[5] = (c1[3] + midY) / 2

	c2[2] = (midX + c2[4]) / 2
	c2[3] = (midY + c2[5]) / 2

	c1[6] = (c1[4] + c2[2]) / 2
	c1[7] = (c1[5] + c2[3]) / 2

	c2[0], c2[1] = c1[6], c1[7]
}

// TraceCubic generate lines subdividing the cubic curve using a Liner
func TraceCubic(t Liner, cubic []Point, flatteningThreshold float64) error {
	if len(cubic) < 4 {
		return errors.New("cubic length must be >= 8")
	}

	cpoints := []float64{cubic[0].X, cubic[0].Y, cubic[1].X, cubic[1].Y, cubic[2].X, cubic[2].Y, cubic[3].X, cubic[3].Y}

	// Allocation curves
	var curves [CurveRecursionLimit * 8]float64
	copy(curves[0:8], cpoints[0:8])
	i := 0

	var c []float64

	var dx, dy, d2, d3 float64

	for i >= 0 {
		c = curves[i:]
		dx = c[6] - c[0]
		dy = c[7] - c[1]

		d2 = math.Abs((c[2]-c[6])*dy - (c[3]-c[7])*dx)
		d3 = math.Abs((c[4]-c[6])*dy - (c[5]-c[7])*dx)

		if (d2+d3)*(d2+d3) <= flatteningThreshold*(dx*dx+dy*dy) || i == len(curves)-8 {
			t.LineTo(c[6], c[7])
			i -= 8
		} else {
			SubdivideCubic(c, curves[i+8:], curves[i:])
			i += 8
		}
	}
	return nil
}

// Quad
// x1, y1, cpx1, cpy2, x2, y2 float64

// SubdivideQuad a Bezier quad curve in 2 equivalents Bezier quad curves.
func SubdivideQuad(c, c1, c2 []float64) {
	c1[0], c1[1] = c[0], c[1]
	c2[4], c2[5] = c[4], c[5]

	c1[2] = (c[0] + c[2]) / 2
	c1[3] = (c[1] + c[3]) / 2
	c2[2] = (c[2] + c[4]) / 2
	c2[3] = (c[3] + c[5]) / 2
	c1[4] = (c1[2] + c2[2]) / 2
	c1[5] = (c1[3] + c2[3]) / 2
	c2[0], c2[1] = c1[4], c1[5]
}

// TraceQuad generate lines subdividing the curve using a Liner
func TraceQuad(t Liner, quad []Point, flatteningThreshold float64) error {
	if len(quad) < 3 {
		return errors.New("quad length must be >= 6")
	}

	qpoints := []float64{quad[0].X, quad[0].Y, quad[1].X, quad[1].Y, quad[2].X, quad[2].Y}

	var curves [CurveRecursionLimit * 6]float64
	copy(curves[0:6], qpoints[0:6])
	i := 0
	var c []float64
	var dx, dy, d float64

	for i >= 0 {
		c = curves[i:]
		dx = c[4] - c[0]
		dy = c[5] - c[1]

		d = math.Abs(((c[2]-c[4])*dy - (c[3]-c[5])*dx))

		if (d*d) <= flatteningThreshold*(dx*dx+dy*dy) || i == len(curves)-6 {
			t.LineTo(c[4], c[5])
			i -= 6
		} else {
			SubdivideQuad(c, curves[i+6:], curves[i:])
			i += 6
		}
	}
	return nil
}

func TraceArc(t Liner, x, y, rx, ry, start, angle, scale float64) (lastX, lastY float64) {
	end := start + angle
	clockWise := true
	if angle < 0 {
		clockWise = false
	}
	ra := (math.Abs(rx) + math.Abs(ry)) / 2
	da := math.Acos(ra/(ra+0.125/scale)) * 2
	if !clockWise {
		da = -da
	}
	angle = start + da
	var curX, curY float64
	for {
		if (angle < end-da/4) != clockWise {
			curX = x + math.Cos(end)*rx
			curY = y + math.Sin(end)*ry
			return curX, curY
		}
		curX = x + math.Cos(angle)*rx
		curY = y + math.Sin(angle)*ry

		angle += da
		t.LineTo(curX, curY)
	}
}
