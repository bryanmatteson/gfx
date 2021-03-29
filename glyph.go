package gfx

// GlyphCache manage a cache of glyphs
type GlyphCache interface {
	// Fetch fetches a glyph from the cache, storing with Render first if it doesn't already exist
	Fetch(gc *ImageContext, fontName string, chr rune) *Glyph
}

// GlyphCacheImp manage a map of glyphs without sync mecanism, not thread safe
type GlyphCacheImp struct {
	glyphs map[string]map[rune]*Glyph
}

// NewGlyphCache initializes a GlyphCache
func NewGlyphCache() *GlyphCacheImp {
	glyphs := make(map[string]map[rune]*Glyph)
	return &GlyphCacheImp{
		glyphs: glyphs,
	}
}

// Fetch fetches a glyph from the cache, calling renderGlyph first if it doesn't already exist
func (glyphCache *GlyphCacheImp) Fetch(gc *ImageContext, fontName string, chr rune) *Glyph {
	if glyphCache.glyphs[fontName] == nil {
		glyphCache.glyphs[fontName] = make(map[rune]*Glyph, 60)
	}
	if glyphCache.glyphs[fontName][chr] == nil {
		glyphCache.glyphs[fontName][chr] = renderGlyph(gc, fontName, chr)
	}
	return glyphCache.glyphs[fontName][chr].Copy()
}

// renderGlyph renders a glyph then caches and returns it
func renderGlyph(gc *ImageContext, fontName string, chr rune) *Glyph {
	gc.Save()
	defer gc.Restore()
	gc.BeginPath()
	width := gc.CreateStringPath(string(chr), 0, 0)
	path := gc.GetPath()
	return &Glyph{
		Path:  &path,
		Width: width,
	}
}

// Glyph represents a rune which has been converted to a Path and width
type Glyph struct {
	// path represents a glyph, it is always at (0, 0)
	Path  *Path
	Width float64
}

// Copy Returns a copy of a Glyph
func (g *Glyph) Copy() *Glyph {
	return &Glyph{
		Path:  g.Path.Copy(),
		Width: g.Width,
	}
}

// Fill copies a glyph from the cache, and fills it
func (g *Glyph) Fill(gc *ImageContext, x, y float64) float64 {
	gc.Save()
	gc.BeginPath()
	gc.Translate(x, y)
	gc.Fill(g.Path)
	gc.Restore()
	return g.Width
}

// Stroke fetches a glyph from the cache, and strokes it
func (g *Glyph) Stroke(gc *ImageContext, x, y float64) float64 {
	gc.Save()
	gc.BeginPath()
	gc.Translate(x, y)
	gc.Stroke(g.Path)
	gc.Restore()
	return g.Width
}
