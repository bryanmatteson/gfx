package gfx

import (
	"image"
	"image/color"

	"golang.org/x/image/draw"
	"golang.org/x/image/math/f64"

	"github.com/golang/freetype/raster"
)

var defaultFontData = FontData{Name: "Arial", Family: FontFamilySans, Style: FontStyleNormal}

var (
	defaultFillStyle   = NewSolidPattern(color.White)
	defaultStrokeStyle = NewSolidPattern(color.Black)
)

type ImageContext struct {
	*StackGraphicContext
	img        *image.RGBA
	fontCache  FontCache
	rasterizer *raster.Rasterizer
	dpi        int
	filter     ImageFilter
}

func NewContext(width, height int) *ImageContext {
	return NewContextForImage(image.NewRGBA(image.Rect(0, 0, width, height)))
}

// NewImageContext creates a new Graphic context from an image.
func NewContextForImage(img draw.Image) *ImageContext {
	width, height := img.Bounds().Dx(), img.Bounds().Dy()
	var rgbaImage *image.RGBA

	switch selectImage := img.(type) {
	case *image.RGBA:
		rgbaImage = selectImage
	default:
		rgbaImage = ImageToRGBA(img)
	}

	gc := &ImageContext{
		StackGraphicContext: NewStackGraphicContext(),
		img:                 rgbaImage,
		rasterizer:          raster.NewRasterizer(width, height),
		dpi:                 72,
		filter:              BilinearFilter,
	}
	return gc
}

func (gc *ImageContext) SetFontCache(cache FontCache) { gc.fontCache = cache }

func (gc *ImageContext) GetDPI() int { return gc.dpi }

func (gc *ImageContext) Clear(color color.Color) {
	width, height := gc.img.Bounds().Dx(), gc.img.Bounds().Dy()
	gc.ClearRect(image.Rect(0, 0, width, height), color)
}

func (gc *ImageContext) ClearRect(rect image.Rectangle, color color.Color) {
	imageColor := image.NewUniform(color)
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

// func DrawImage(src image.Image, dest draw.Image, tr Matrix, op draw.Op, filter ImageFilter) {
func (gc *ImageContext) DrawImage(img image.Image) {
	var transformer draw.Transformer
	switch gc.filter {
	case LinearFilter:
		transformer = draw.NearestNeighbor
	case BilinearFilter:
		transformer = draw.BiLinear
	case BicubicFilter:
		transformer = draw.CatmullRom
	}

	trm := gc.Current.Trm
	transformer.Transform(gc.img, f64.Aff3{trm.A, trm.B, trm.E, trm.C, trm.D, trm.F}, img, img.Bounds(), draw.Over, nil)
}

// recalc recalculates scale and bounds values from the font size, screen
// resolution and font metrics, and invalidates the glyph cache.
func (gc *ImageContext) recalc() {
	gc.Current.Scale = gc.Current.FontSize * float64(gc.dpi) * (64.0 / 72.0)
}

// SetFilter sets the ImageFilter to use for transformations
func (gc *ImageContext) SetFilter(filter ImageFilter) {
	gc.filter = filter
}

// SetDPI sets the screen resolution in dots per inch.
func (gc *ImageContext) SetDPI(dpi int) {
	gc.dpi = dpi
	gc.recalc()
}

// SetFontSize sets the font size in points (as in ``a 12 point font'').
func (gc *ImageContext) SetFontSize(fontSize float64) {
	gc.Current.FontSize = fontSize
	gc.recalc()
}

func (gc *ImageContext) Stroke(paths ...*Path) {
	if len(paths) == 0 && gc.Current.Path.IsEmpty() {
		return
	}

	paths = append(paths, gc.Current.Path)
	gc.rasterizer.UseNonZeroWinding = true

	stroker := NewLineStroker(gc.Current.Cap, gc.Current.Join, Transformer{Tr: gc.Current.Trm, Flattener: ftLineBuilder{Adder: gc.rasterizer}})
	stroker.HalfLineWidth = gc.Current.LineWidth / 2

	var liner Flattener = stroker
	if len(gc.Current.Dash) > 0 {
		liner = NewDashConverter(gc.Current.Dash, gc.Current.DashOffset, stroker)
	}

	for _, p := range paths {
		Flatten(p, liner, gc.Current.Trm.GetScale())
	}

	var painter raster.Painter
	if gc.Current.Mask == nil {
		if pattern, ok := gc.Current.StrokePattern.(*solidPattern); ok {
			p := raster.NewRGBAPainter(gc.img)
			p.SetColor(pattern.color)
			painter = p
		}
	}
	if painter == nil {
		painter = newPatternPainter(gc.img, gc.Current.Mask, gc.Current.StrokePattern)
	}

	gc.rasterizer.Rasterize(painter)
}

func (gc *ImageContext) Fill(paths ...*Path) {
	if len(paths) == 0 && gc.Current.Path.IsEmpty() {
		return
	}

	gc.rasterizer.Clear()
	gc.rasterizer.UseNonZeroWinding = gc.Current.FillRule == FillRuleWinding

	paths = append(paths, gc.Current.Path)
	flattener := Transformer{Tr: gc.Current.Trm, Flattener: ftLineBuilder{Adder: gc.rasterizer}}
	for _, p := range paths {
		Flatten(p, flattener, gc.Current.Trm.GetScale())
	}

	var painter raster.Painter
	if gc.Current.Mask == nil {
		if pattern, ok := gc.Current.FillPattern.(*solidPattern); ok {
			p := raster.NewRGBAPainter(gc.img)
			p.SetColor(pattern.color)
			painter = p
		}
	}
	if painter == nil {
		painter = newPatternPainter(gc.img, gc.Current.Mask, gc.Current.FillPattern)
	}

	gc.rasterizer.Rasterize(painter)
}

func (gc *ImageContext) Clip(paths ...*Path) {
	if len(paths) == 0 && gc.Current.Path.IsEmpty() {
		return
	}

	paths = append(paths, gc.Current.Path)
	gc.rasterizer.UseNonZeroWinding = gc.Current.FillRule == FillRuleWinding

	flattener := Transformer{Tr: gc.Current.Trm, Flattener: ftLineBuilder{Adder: gc.rasterizer}}
	for _, p := range paths {
		Flatten(p, flattener, gc.Current.Trm.GetScale())
	}

	clip := image.NewAlpha(gc.img.Bounds())
	painter := raster.NewAlphaOverPainter(clip)
	gc.rasterizer.Rasterize(painter)
	gc.rasterizer.Clear()
	gc.Current.Path.Clear()

	if gc.Current.Mask == nil {
		gc.Current.Mask = clip
	} else {
		mask := image.NewAlpha(gc.img.Bounds())
		draw.DrawMask(mask, mask.Bounds(), clip, image.Point{}, gc.Current.Mask, image.Point{}, draw.Over)
		gc.Current.Mask = mask
	}
}

func (gc *ImageContext) ClipImage(m image.Image) {
	var transformer draw.Transformer
	switch gc.filter {
	case LinearFilter:
		transformer = draw.NearestNeighbor
	case BilinearFilter:
		transformer = draw.BiLinear
	case BicubicFilter:
		transformer = draw.CatmullRom
	}
	mask := image.NewAlpha(gc.img.Bounds())

	trm := f64.Aff3{gc.Current.Trm.A, gc.Current.Trm.B, gc.Current.Trm.E, gc.Current.Trm.C, gc.Current.Trm.D, gc.Current.Trm.F}
	transformer.Transform(mask, trm, m, m.Bounds(), draw.Over, nil)
	gc.Current.Mask = mask
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
	Trm           Matrix
	Path          *Path
	LineWidth     float64
	Dash          []float64
	DashOffset    float64
	FillRule      FillRule
	Cap           LineCap
	Join          LineJoin
	FontSize      float64
	FontData      FontData
	Font          Font
	Scale         float64
	FillPattern   Pattern
	StrokePattern Pattern
	Mask          *image.Alpha

	Previous *ContextStack
}

func NewStackGraphicContext() *StackGraphicContext {
	gc := &StackGraphicContext{
		Current: &ContextStack{
			Trm:           IdentityMatrix,
			Path:          new(Path),
			LineWidth:     1.0,
			Cap:           RoundCap,
			FillRule:      FillRuleEvenOdd,
			Join:          RoundJoin,
			FontSize:      10,
			FontData:      defaultFontData,
			FillPattern:   defaultFillStyle,
			StrokePattern: defaultStrokeStyle,
			Scale:         1.0,
		},
	}
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

func (gc *StackGraphicContext) SetStroke(stroke *Stroke) {
	gc.Current.Dash = stroke.Dashes
	gc.Current.DashOffset = stroke.DashPhase
	gc.Current.Cap = stroke.LineCap
	gc.Current.Join = stroke.LineJoin
	gc.Current.LineWidth = stroke.LineWidth
}

func (gc *StackGraphicContext) SetStrokeColor(c color.Color) {
	gc.Current.StrokePattern = NewSolidPattern(c)
}

func (gc *StackGraphicContext) SetFillColor(c color.Color) {
	gc.Current.FillPattern = NewSolidPattern(c)
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

func (gc *StackGraphicContext) SetFontData(fontData FontData) {
	gc.Current.FontData = fontData
}

func (gc *StackGraphicContext) GetFontData() FontData {
	return gc.Current.FontData
}

func (gc *StackGraphicContext) SetFont(font Font) {
	gc.Current.Font = font
}

func (gc *StackGraphicContext) GetFont() Font {
	return gc.Current.Font
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
	context.Font = gc.Current.Font
	context.LineWidth = gc.Current.LineWidth
	context.StrokePattern = gc.Current.StrokePattern
	context.FillPattern = gc.Current.FillPattern
	context.FillRule = gc.Current.FillRule
	context.Dash = gc.Current.Dash
	context.DashOffset = gc.Current.DashOffset
	context.Cap = gc.Current.Cap
	context.Join = gc.Current.Join
	context.Path = gc.Current.Path.Copy()
	context.Scale = gc.Current.Scale
	context.Trm = gc.Current.Trm
	context.Mask = gc.Current.Mask
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
