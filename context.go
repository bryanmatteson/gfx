package gfx

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"math"

	"golang.org/x/image/draw"

	"github.com/golang/freetype/raster"
	"github.com/golang/freetype/truetype"
	"go.matteson.dev/gfx/font"
	"golang.org/x/image/math/fixed"
)

var DefaultFontData = font.FontData{Name: "Arial", Family: font.FontFamilySans, Style: font.FontStyleNormal}

// Painter implements the freetype raster.Painter and has a SetColor method like the RGBAPainter
type Painter interface {
	raster.Painter
	SetColor(color color.Color)
}

type ImageContext struct {
	*StackGraphicContext
	img              draw.Image
	painter          Painter
	fillRasterizer   *raster.Rasterizer
	strokeRasterizer *raster.Rasterizer
	FontCache        font.FontCache
	glyphCache       GlyphCache
	glyphBuf         *truetype.GlyphBuf
	DPI              int
	Filter           ImageFilter
}

// NewImageContext creates a new Graphic context from an image.
func NewImageContext(img draw.Image) *ImageContext {
	var painter Painter
	switch selectImage := img.(type) {
	case *image.RGBA:
		painter = raster.NewRGBAPainter(selectImage)
	default:
		img = ImageToRGBA(img)
		painter = raster.NewRGBAPainter(img.(*image.RGBA))
	}
	return NewImageContextWithPainter(img, painter)
}

// NewImageContextWithPainter creates a new Graphic context from an image and a Painter (see Freetype-go)
func NewImageContextWithPainter(img draw.Image, painter Painter) *ImageContext {
	width, height := img.Bounds().Dx(), img.Bounds().Dy()
	dpi := 92
	gc := &ImageContext{
		StackGraphicContext: NewStackGraphicContext(),
		img:                 img,
		painter:             painter,
		fillRasterizer:      raster.NewRasterizer(width, height),
		strokeRasterizer:    raster.NewRasterizer(width, height),
		FontCache:           font.GetGlobalFontCache(),
		glyphCache:          NewGlyphCache(),
		glyphBuf:            &truetype.GlyphBuf{},
		DPI:                 dpi,
		Filter:              BilinearFilter,
	}
	return gc
}

func (gc *ImageContext) GetDPI() int { return gc.DPI }

func (gc *ImageContext) Clear() {
	width, height := gc.img.Bounds().Dx(), gc.img.Bounds().Dy()
	gc.ClearRect(image.Rect(0, 0, width, height))
}

func (gc *ImageContext) ClearRect(rect image.Rectangle) {
	imageColor := image.NewUniform(gc.Current.FillColor)
	draw.Draw(gc.img, rect, imageColor, image.Point{}, draw.Src)
}

func (gc *ImageContext) DrawRect(rect Rect) {
	gc.MoveTo(rect.X.Min, rect.Y.Min)
	gc.LineTo(rect.X.Max, rect.Y.Min)
	gc.LineTo(rect.X.Max, rect.Y.Max)
	gc.LineTo(rect.X.Min, rect.Y.Max)
	gc.ClosePath()
}

func (gc *ImageContext) DrawQuad(quad Quad) {
	gc.MoveToPoint(quad.BottomLeft)
	gc.LineToPoint(quad.TopLeft)
	gc.LineToPoint(quad.TopRight)
	gc.LineToPoint(quad.BottomRight)
	gc.ClosePath()
}

func (gc *ImageContext) DrawLine(line Line) {
	gc.MoveToPoint(line.Pt1)
	gc.LineToPoint(line.Pt2)
}

func (gc *ImageContext) DrawPath(path *Path) {
	var j = 0
	for _, cmd := range path.Components {
		switch cmd {
		case MoveToComp:
			gc.MoveToPoint(path.Points[j])
		case LineToComp:
			gc.LineToPoint(path.Points[j])
		case QuadCurveToComp:
			gc.QuadCurveToPoints(path.Points[j], path.Points[j+1])
		case CubicCurveToComp:
			gc.CubicCurveToPoints(path.Points[j], path.Points[j+1], path.Points[j+2])
		case ClosePathComp:
			gc.ClosePath()
		}
		j += cmd.PointCount()
	}
}

func (gc *ImageContext) DrawImage(img image.Image) {
	DrawImage(img, gc.img, gc.Current.Trm, draw.Over, gc.Filter)
}

func (gc *ImageContext) FillString(text string) (width float64) {
	return gc.FillStringAt(text, 0, 0)
}

// FillStringAt draws the text at the specified point (x, y)
func (gc *ImageContext) FillStringAt(text string, x, y float64) (width float64) {
	f, err := gc.loadCurrentFont()
	if err != nil {
		log.Println(err)
		return 0.0
	}

	startx := x
	prev, hasPrev := truetype.Index(0), false
	fontName := gc.GetFontName()

	for _, r := range text {
		index := f.Index(r)
		if hasPrev {
			x += fUnitsToFloat64(f.Kern(fixed.Int26_6(gc.Current.Scale), prev, index))
		}
		glyph := gc.glyphCache.Fetch(gc, fontName, r)
		x += glyph.Fill(gc, x, y)
		prev, hasPrev = index, true
	}
	return x - startx
}

// StrokeString draws the contour of the text at point (0, 0)
func (gc *ImageContext) StrokeString(text string) (width float64) {
	return gc.StrokeStringAt(text, 0, 0)
}

// StrokeStringAt draws the contour of the text at point (x, y)
func (gc *ImageContext) StrokeStringAt(text string, x, y float64) (width float64) {
	f, err := gc.loadCurrentFont()
	if err != nil {
		log.Println(err)
		return 0.0
	}
	startx := x
	prev, hasPrev := truetype.Index(0), false
	fontName := gc.GetFontName()
	for _, r := range text {
		index := f.Index(r)
		if hasPrev {
			x += fUnitsToFloat64(f.Kern(fixed.Int26_6(gc.Current.Scale), prev, index))
		}
		glyph := gc.glyphCache.Fetch(gc, fontName, r)
		x += glyph.Stroke(gc, x, y)
		prev, hasPrev = index, true
	}
	return x - startx
}

func (gc *ImageContext) loadCurrentFont() (*truetype.Font, error) {
	font, err := gc.FontCache.Load(gc.Current.FontData)
	if err != nil {
		font, err = gc.FontCache.Load(DefaultFontData)
	}

	if font != nil {
		gc.SetFont(font)
		gc.SetFontSize(gc.Current.FontSize)
	}

	return font, err
}

// p is a truetype.Point measured in FUnits and positive Y going upwards.
// The returned value is the same thing measured in floating point and positive Y
// going downwards.

func (gc *ImageContext) drawGlyph(glyph truetype.Index, dx, dy float64) error {
	if err := gc.glyphBuf.Load(gc.Current.Font, fixed.Int26_6(gc.Current.Scale), glyph, font.HintingNone); err != nil {
		return err
	}
	e0 := 0
	for _, e1 := range gc.glyphBuf.Ends {
		DrawContour(gc, gc.glyphBuf.Points[e0:e1], dx, dy)
		e0 = e1
	}
	return nil
}

// CreateStringPath creates a path from the string s at x, y, and returns the string width.
// The text is placed so that the left edge of the em square of the first character of s
// and the baseline intersect at x, y. The majority of the affected pixels will be
// above and to the right of the point, but some may be below or to the left.
// For example, drawing a string that starts with a 'J' in an italic font may
// affect pixels below and left of the point.
func (gc *ImageContext) CreateStringPath(s string, x, y float64) float64 {
	f, err := gc.loadCurrentFont()
	if err != nil {
		log.Println(err)
		return 0.0
	}
	startx := x
	prev, hasPrev := truetype.Index(0), false
	for _, rune := range s {
		index := f.Index(rune)
		if hasPrev {
			x += fUnitsToFloat64(f.Kern(fixed.Int26_6(gc.Current.Scale), prev, index))
		}
		err := gc.drawGlyph(index, x, y)
		if err != nil {
			log.Println(err)
			return startx - x
		}
		x += fUnitsToFloat64(f.HMetric(fixed.Int26_6(gc.Current.Scale), index).AdvanceWidth)
		prev, hasPrev = index, true
	}
	return x - startx
}

// GetStringBounds returns the approximate pixel bounds of the string s at x, y.
// The the left edge of the em square of the first character of s
// and the baseline intersect at 0, 0 in the returned coordinates.
// Therefore the top and left coordinates may well be negative.
func (gc *ImageContext) GetStringBounds(s string) (left, top, right, bottom float64) {
	f, err := gc.loadCurrentFont()
	if err != nil {
		log.Println(err)
		return 0, 0, 0, 0
	}
	top, left, bottom, right = 10e6, 10e6, -10e6, -10e6
	cursor := 0.0
	prev, hasPrev := truetype.Index(0), false
	for _, rune := range s {
		index := f.Index(rune)
		if hasPrev {
			cursor += fUnitsToFloat64(f.Kern(fixed.Int26_6(gc.Current.Scale), prev, index))
		}
		if err := gc.glyphBuf.Load(gc.Current.Font, fixed.Int26_6(gc.Current.Scale), index, font.HintingNone); err != nil {
			log.Println(err)
			return 0, 0, 0, 0
		}
		e0 := 0
		for _, e1 := range gc.glyphBuf.Ends {
			ps := gc.glyphBuf.Points[e0:e1]
			for _, p := range ps {
				x, y := pointToF64Point(p)
				top = math.Min(top, y)
				bottom = math.Max(bottom, y)
				left = math.Min(left, x+cursor)
				right = math.Max(right, x+cursor)
			}
		}
		cursor += fUnitsToFloat64(f.HMetric(fixed.Int26_6(gc.Current.Scale), index).AdvanceWidth)
		prev, hasPrev = index, true
	}
	return left, top, right, bottom
}

// recalc recalculates scale and bounds values from the font size, screen
// resolution and font metrics, and invalidates the glyph cache.
func (gc *ImageContext) recalc() {
	gc.Current.Scale = gc.Current.FontSize * float64(gc.DPI) * (64.0 / 72.0)
}

// SetFilter sets the ImageFilter to use for transformations
func (gc *ImageContext) SetFilter(filter ImageFilter) {
	gc.Filter = filter
}

// SetDPI sets the screen resolution in dots per inch.
func (gc *ImageContext) SetDPI(dpi int) {
	gc.DPI = dpi
	gc.recalc()
}

// SetFont sets the font used to draw text.
func (gc *ImageContext) SetFont(font *truetype.Font) {
	gc.Current.Font = font
}

// SetFontSize sets the font size in points (as in ``a 12 point font'').
func (gc *ImageContext) SetFontSize(fontSize float64) {
	gc.Current.FontSize = fontSize
	gc.recalc()
}

func (gc *ImageContext) paint(rasterizer *raster.Rasterizer, color color.Color) {
	gc.painter.SetColor(color)
	rasterizer.Rasterize(gc.painter)
	rasterizer.Clear()
	gc.Current.Path.Clear()
}

// Stroke strokes the paths with the color specified by SetStrokeColor
func (gc *ImageContext) Stroke(paths ...*Path) {
	paths = append(paths, gc.Current.Path)
	gc.strokeRasterizer.UseNonZeroWinding = true

	stroker := NewLineStroker(gc.Current.Cap, gc.Current.Join, Transformer{Tr: gc.Current.Trm, Flattener: ftLineBuilder{Adder: gc.strokeRasterizer}})
	stroker.HalfLineWidth = gc.Current.LineWidth / 2

	var liner Flattener = stroker
	if len(gc.Current.Dash) > 0 {
		liner = NewDashConverter(gc.Current.Dash, gc.Current.DashOffset, stroker)
	}

	for _, p := range paths {
		Flatten(p, liner, gc.Current.Trm.GetScale())
	}

	gc.paint(gc.strokeRasterizer, gc.Current.StrokeColor)
}

func (gc *ImageContext) Fill(paths ...*Path) {
	paths = append(paths, gc.Current.Path)
	gc.fillRasterizer.UseNonZeroWinding = gc.Current.FillRule == FillRuleWinding

	flattener := Transformer{Tr: gc.Current.Trm, Flattener: ftLineBuilder{Adder: gc.fillRasterizer}}
	for _, p := range paths {
		Flatten(p, flattener, gc.Current.Trm.GetScale())
	}

	gc.paint(gc.fillRasterizer, gc.Current.FillColor)
}

func (gc *ImageContext) FillStroke(paths ...*Path) {
	paths = append(paths, gc.Current.Path)
	gc.fillRasterizer.UseNonZeroWinding = gc.Current.FillRule == FillRuleWinding
	gc.strokeRasterizer.UseNonZeroWinding = true

	flattener := Transformer{Tr: gc.Current.Trm, Flattener: ftLineBuilder{Adder: gc.fillRasterizer}}

	stroker := NewLineStroker(gc.Current.Cap, gc.Current.Join, Transformer{Tr: gc.Current.Trm, Flattener: ftLineBuilder{Adder: gc.strokeRasterizer}})
	stroker.HalfLineWidth = gc.Current.LineWidth / 2

	var liner Flattener
	if gc.Current.Dash != nil && len(gc.Current.Dash) > 0 {
		liner = NewDashConverter(gc.Current.Dash, gc.Current.DashOffset, stroker)
	} else {
		liner = stroker
	}

	demux := DemuxFlattener{Flatteners: []Flattener{flattener, liner}}
	for _, p := range paths {
		Flatten(p, demux, gc.Current.Trm.GetScale())
	}

	gc.paint(gc.fillRasterizer, gc.Current.FillColor)
	gc.paint(gc.strokeRasterizer, gc.Current.StrokeColor)
}

type ImageFilter int

const (
	LinearFilter ImageFilter = iota
	BilinearFilter
	BicubicFilter
)

type StackGraphicContext struct {
	Current *ContextStack
}

type ContextStack struct {
	Trm         Matrix
	Path        *Path
	LineWidth   float64
	Dash        []float64
	DashOffset  float64
	StrokeColor color.Color
	FillColor   color.Color
	FillRule    FillRule
	Cap         LineCap
	Join        LineJoin
	FontSize    float64
	FontData    font.FontData
	Font        *truetype.Font
	Scale       float64

	Previous *ContextStack
}

func (cs *ContextStack) GetFontName() string {
	fontData := cs.FontData
	return fmt.Sprintf("%s:%d:%d:%9.2f", fontData.Name, fontData.Family, fontData.Style, cs.FontSize)
}

func NewStackGraphicContext() *StackGraphicContext {
	gc := &StackGraphicContext{}
	gc.Current = new(ContextStack)
	gc.Current.Trm = IdentityMatrix
	gc.Current.Path = new(Path)
	gc.Current.LineWidth = 1.0
	gc.Current.StrokeColor = image.Black
	gc.Current.FillColor = image.White
	gc.Current.Cap = RoundCap
	gc.Current.FillRule = FillRuleEvenOdd
	gc.Current.Join = RoundJoin
	gc.Current.FontSize = 10
	gc.Current.FontData = DefaultFontData
	return gc
}

func (gc *StackGraphicContext) GetTransformationMatrix() Matrix {
	return gc.Current.Trm
}

func (gc *StackGraphicContext) SetTransformationMatrix(trm Matrix) {
	gc.Current.Trm = trm
}

func (gc *StackGraphicContext) Concat(trm Matrix) {
	gc.Current.Trm = gc.Current.Trm.Concat(trm)
}

func (gc *StackGraphicContext) Rotate(angle float64) {
	gc.Current.Trm = gc.Current.Trm.Rotated(angle)
}

func (gc *StackGraphicContext) Translate(tx, ty float64) {
	gc.Current.Trm = gc.Current.Trm.Translated(tx, ty)
}

func (gc *StackGraphicContext) Scale(sx, sy float64) {
	gc.Current.Trm = gc.Current.Trm.Scaled(sx, sy)
}

func (gc *StackGraphicContext) SetStrokeColor(c color.Color) {
	gc.Current.StrokeColor = c
}

func (gc *StackGraphicContext) SetFillColor(c color.Color) {
	gc.Current.FillColor = c
}

func (gc *StackGraphicContext) SetFillRule(f FillRule) {
	gc.Current.FillRule = f
}

func (gc *StackGraphicContext) SetLineWidth(lineWidth float64) {
	gc.Current.LineWidth = lineWidth
}

func (gc *StackGraphicContext) SetLineCap(cap LineCap) {
	gc.Current.Cap = cap
}

func (gc *StackGraphicContext) SetLineJoin(join LineJoin) {
	gc.Current.Join = join
}

func (gc *StackGraphicContext) SetLineDash(dash []float64, dashOffset float64) {
	gc.Current.Dash = dash
	gc.Current.DashOffset = dashOffset
}

func (gc *StackGraphicContext) SetFontSize(fontSize float64) {
	gc.Current.FontSize = fontSize
}

func (gc *StackGraphicContext) GetFontSize() float64 {
	return gc.Current.FontSize
}

func (gc *StackGraphicContext) SetFontData(fontData font.FontData) {
	gc.Current.FontData = fontData
}

func (gc *StackGraphicContext) GetFontData() font.FontData {
	return gc.Current.FontData
}

func (gc *StackGraphicContext) BeginPath() {
	gc.Current.Path.Clear()
}

func (gc *StackGraphicContext) GetPath() Path {
	return *gc.Current.Path.Copy()
}

func (gc *StackGraphicContext) IsEmpty() bool {
	return gc.Current.Path.IsEmpty()
}

func (gc *StackGraphicContext) LastPoint() (float64, float64) {
	return gc.Current.Path.LastPoint()
}

func (gc *StackGraphicContext) MoveTo(x, y float64) {
	gc.Current.Path.MoveTo(x, y)
}

func (gc *StackGraphicContext) MoveToPoint(pt Point) {
	gc.MoveTo(pt.X, pt.Y)
}

func (gc *StackGraphicContext) LineTo(x, y float64) {
	gc.Current.Path.LineTo(x, y)
}

func (gc *StackGraphicContext) LineToPoint(pt Point) {
	gc.LineTo(pt.X, pt.Y)
}

func (gc *StackGraphicContext) QuadCurveTo(cx, cy, x, y float64) {
	gc.Current.Path.QuadCurveTo(cx, cy, x, y)
}

func (gc *StackGraphicContext) QuadCurveToPoints(cp, pt Point) {
	gc.Current.Path.QuadCurveTo(cp.X, cp.Y, pt.X, pt.Y)
}

func (gc *StackGraphicContext) CubicCurveTo(cx1, cy1, cx2, cy2, x, y float64) {
	gc.Current.Path.CubicCurveTo(cx1, cy1, cx2, cy2, x, y)
}

func (gc *StackGraphicContext) CubicCurveToPoints(cp1, cp2, pt Point) {
	gc.Current.Path.CubicCurveTo(cp1.X, cp1.Y, cp2.X, cp2.Y, pt.X, pt.Y)
}

func (gc *StackGraphicContext) ClosePath() {
	gc.Current.Path.ClosePath()
}

func (gc *StackGraphicContext) Save() {
	context := new(ContextStack)
	context.FontSize = gc.Current.FontSize
	context.FontData = gc.Current.FontData
	context.LineWidth = gc.Current.LineWidth
	context.StrokeColor = gc.Current.StrokeColor
	context.FillColor = gc.Current.FillColor
	context.FillRule = gc.Current.FillRule
	context.Dash = gc.Current.Dash
	context.DashOffset = gc.Current.DashOffset
	context.Cap = gc.Current.Cap
	context.Join = gc.Current.Join
	context.Path = gc.Current.Path.Copy()
	context.Font = gc.Current.Font
	context.Scale = gc.Current.Scale
	context.Trm = gc.Current.Trm
	context.Previous = gc.Current
	gc.Current = context
}

func (gc *StackGraphicContext) Restore() {
	if gc.Current.Previous != nil {
		oldContext := gc.Current
		gc.Current = gc.Current.Previous
		oldContext.Previous = nil
	}
}

func (gc *StackGraphicContext) GetFontName() string {
	return gc.Current.GetFontName()
}
