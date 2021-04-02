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
