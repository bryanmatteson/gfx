package gfx

import (
	"strings"
)

type FontStyle byte

const FontStyleNormal FontStyle = 0
const (
	FontStyleBold FontStyle = 1 << iota
	FontStyleItalic
)

type FontFamily byte

// Families
const (
	FontFamilySans FontFamily = iota
	FontFamilySerif
	FontFamilyMono
)

type FontData struct {
	Name   string
	Style  FontStyle
	Family FontFamily
}

func (f FontData) IsBold() bool       { return f.Style&FontStyleBold == FontStyleBold }
func (f FontData) IsItalic() bool     { return f.Style&FontStyleItalic == FontStyleItalic }
func (f FontData) IsSerif() bool      { return f.Family == FontFamilySerif }
func (f FontData) IsSansSerif() bool  { return f.Family == FontFamilySans }
func (f FontData) IsMonospaced() bool { return f.Family == FontFamilyMono }

func (f FontData) String() string {
	var builder strings.Builder
	builder.WriteString(f.Name)
	builder.WriteByte('-')

	switch f.Family {
	case FontFamilySans:
		builder.WriteRune('s')
	case FontFamilySerif:
		builder.WriteRune('r')
	case FontFamilyMono:
		builder.WriteRune('m')
	}

	if f.IsBold() {
		builder.WriteRune('b')
	} else {
		builder.WriteRune('r')
	}

	if f.IsItalic() {
		builder.WriteRune('i')
	}

	return builder.String()
}

type Font interface {
	Name() string
	BoundingBox() Rect
	Info() FontData
	Glyph(chr rune, trm Matrix) *Glyph
	Advance(chr rune) float64
}

type FontCache interface {
	Load(FontData) (Font, error)
	Store(FontData, Font)
}

type Glyph struct {
	Path  *Path
	Width float64
}

func (g *Glyph) Copy() *Glyph {
	return &Glyph{
		Path:  g.Path.Copy(),
		Width: g.Width,
	}
}

func (g *Glyph) Fill(gc *ImageContext, x, y float64) float64 {
	gc.Save()
	gc.BeginPath()
	gc.Translate(x, y)
	gc.Fill(g.Path)
	gc.Restore()
	return g.Width
}

func (g *Glyph) Stroke(gc *ImageContext, x, y float64) float64 {
	gc.Save()
	gc.BeginPath()
	gc.Translate(x, y)
	gc.Stroke(g.Path)
	gc.Restore()
	return g.Width
}
