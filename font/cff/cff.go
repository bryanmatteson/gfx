package cff

import (
	"go.matteson.dev/gfx/font/cff/charsets"
	"go.matteson.dev/gfx/font/encoding"
	"go.matteson.dev/gfx"
)

type Header struct {
	// The major version of this font format. Starting at 1.
	MajorVersion int
	// The minor version of this font format. Starting at 0. Indicates extensions to the format which are undetectable by readers which do not support them.
	MinorVersion int
	// Indicates the size of this header in bytes so that future changes to the format may include extra data after the <see cref="OffsetSize"/> field.
	SizeInBytes int
	// Specifies the size of all offsets relative to the start of the data in the font.
	OffsetSize int
}

func readHeader(r *reader) (hdr *Header, err error) {
	hdr = &Header{}

	b, err := r.bytes(4)
	if err != nil {
		return hdr, err
	}

	hdr.MajorVersion = int(b[0])
	hdr.MinorVersion = int(b[1])
	hdr.SizeInBytes = int(b[2])
	hdr.OffsetSize = int(b[3])
	return
}

type Collection struct {
	Header    *Header
	Fonts     map[string]Font
	FirstFont Font
}

type Font interface {
	FontMatrix() gfx.Matrix
	Weight() string
	DefaultWidthX(c string) float64
	NominalWidthX(c string) float64
	GenerateGlyph(c string) (*Type2Glyph, error)
}

type font struct {
	tld         *TopLevelDictionary
	priv        *PrivateDictionary
	charset     charsets.Charset
	charstrings *t2charstrings
	encoding    encoding.Encoding
}

func (f *font) FontMatrix() gfx.Matrix         { return f.tld.FontMatrix }
func (f *font) Weight() string                 { return f.tld.Weight }
func (f *font) DefaultWidthX(c string) float64 { return f.priv.DefaultWidthX }
func (f *font) NominalWidthX(c string) float64 { return f.priv.NominalWidthX }

func (f *font) GenerateGlyph(c string) (*Type2Glyph, error) {
	gid := f.charset.GetGlyphIDByName(c)
	return f.charstrings.Generate(c, gid, f.DefaultWidthX(c), f.NominalWidthX(c))
}

type cidfont struct {
	*font
	selector *selector
}

func (f *cidfont) DefaultWidthX(c string) float64 {
	gid := f.charset.GetGlyphIDByName(c)
	fd := f.selector.GetFontDictionary(gid)
	if fd == nil {
		return 0
	}
	return fd.private.DefaultWidthX
}

func (f *cidfont) NominalWidthX(c string) float64 {
	gid := f.charset.GetGlyphIDByName(c)
	fd := f.selector.GetFontDictionary(gid)
	if fd == nil {
		return 0
	}
	return fd.private.NominalWidthX
}

func (f *cidfont) GenerateGlyph(c string) (*Type2Glyph, error) {
	gid := f.charset.GetGlyphIDByName(c)
	return f.charstrings.Generate(c, gid, f.DefaultWidthX(c), f.NominalWidthX(c))
}
