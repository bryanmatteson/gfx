package gfx

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
