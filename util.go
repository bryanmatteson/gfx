package gfx

import (
	"image"
	"image/color"
	"math"

	"github.com/golang/freetype/truetype"
	"golang.org/x/image/draw"
	"golang.org/x/image/math/fixed"
)

func Min4(a float64, b float64, c float64, d float64) float64 {
	return math.Min(math.Min(a, b), math.Min(c, d))
}

func Max4(a float64, b float64, c float64, d float64) float64 {
	return math.Max(math.Max(a, b), math.Max(c, d))
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

func ImageToRGBA(src image.Image) *image.RGBA {
	bounds := src.Bounds()
	dst := image.NewRGBA(bounds)
	draw.Draw(dst, bounds, src, bounds.Min, draw.Src)
	return dst
}

func pointToF64Point(p truetype.Point) (x, y float64) {
	return fUnitsToFloat64(p.X), -fUnitsToFloat64(p.Y)
}

func fUnitsToFloat64(x fixed.Int26_6) float64 {
	scaled := x << 2
	return float64(scaled/256) + float64(scaled%256)/256.0
}

func AdjustRectForStroke(r Rect, stroke *Stroke) Rect {
	if stroke == nil {
		return r
	}

	expand := stroke.LineWidth
	if expand == 0.0 {
		expand = 1.0
	}

	return r.Expanded(Point{expand, expand})
}

// DrawContour draws the given closed contour at the given sub-pixel offset.
func DrawContour(path PathWalker, ps []truetype.Point, dx, dy float64) {
	if len(ps) == 0 {
		return
	}
	startX, startY := pointToF64Point(ps[0])
	var others []truetype.Point
	if ps[0].Flags&0x01 != 0 {
		others = ps[1:]
	} else {
		lastX, lastY := pointToF64Point(ps[len(ps)-1])
		if ps[len(ps)-1].Flags&0x01 != 0 {
			startX, startY = lastX, lastY
			others = ps[:len(ps)-1]
		} else {
			startX = (startX + lastX) / 2
			startY = (startY + lastY) / 2
			others = ps
		}
	}
	path.MoveTo(startX+dx, startY+dy)
	q0X, q0Y, on0 := startX, startY, true
	for _, p := range others {
		qX, qY := pointToF64Point(p)
		on := p.Flags&0x01 != 0
		if on {
			if on0 {
				path.LineTo(qX+dx, qY+dy)
			} else {
				path.QuadCurveTo(q0X+dx, q0Y+dy, qX+dx, qY+dy)
			}
		} else {
			if on0 {
				// No-op.
			} else {
				midX := (q0X + qX) / 2
				midY := (q0Y + qY) / 2
				path.QuadCurveTo(q0X+dx, q0Y+dy, midX+dx, midY+dy)
			}
		}
		q0X, q0Y, on0 = qX, qY, on
	}
	// Close the curve.
	if on0 {
		path.LineTo(startX+dx, startY+dy)
	} else {
		path.QuadCurveTo(q0X+dx, q0Y+dy, startX+dx, startY+dy)
	}
}

func CropImage(img image.Image, color color.Color) Rect {
	var minx, miny, maxx, maxy float64
	minx, miny = math.Inf(1), math.Inf(1)
	maxx, maxy = math.Inf(-1), math.Inf(-1)

	cr, cg, cb, _ := color.RGBA()

	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			if (r == cr && g == cg && b == cb) || a == 0 {
				continue
			}

			minx = math.Min(minx, float64(x))
			miny = math.Min(miny, float64(y))
			maxx = math.Max(maxx, float64(x))
			maxy = math.Max(maxy, float64(y))
		}
	}

	return MakeRect(minx, miny, maxx, maxy)
}

func ConstrictImageRect(img image.Image)
