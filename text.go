package gfx

import (
	"strings"
	"unicode"

	"github.com/ahmetb/go-linq"
)

type Chars []Char

func (l Chars) IsWhitespace() bool {
	for _, letter := range l {
		if !letter.IsWhitespace() {
			return false
		}
	}
	return true
}

func (l Chars) GetMeanConfidence() float64 {
	sum := 0.0
	for _, c := range l {
		sum += c.Confidence
	}
	return sum / float64(len(l))
}

type Char struct {
	Rune          rune
	Quad          Quad
	Confidence    float64
	Orientation   Orientation
	DeskewAngle   float64
	StartBaseline Point
	EndBaseline   Point
}

func (l Char) IsWhitespace() bool { return unicode.IsSpace(l.Rune) }

type TextWords []TextWord

func (w TextWords) Orientation() (orientation Orientation) {
	if len(w) == 0 {
		return OtherOrientation
	}

	orientation = w[0].Orientation
	if orientation == OtherOrientation {
		return
	}

	for _, word := range w[1:] {
		if word.Orientation != orientation {
			return OtherOrientation
		}
	}
	return
}

func (w TextWords) OrderByReadingOrder() (ret TextWords) {
	if len(w) <= 1 {
		return w
	}

	switch w.Orientation() {
	case PageUp:
		linq.From(w).OrderBy(func(i interface{}) interface{} { return i.(*TextWord).Quad.BottomLeft.X }).ToSlice(&ret)
	case PageDown:
		linq.From(w).OrderByDescending(func(i interface{}) interface{} { return i.(*TextWord).Quad.BottomLeft.X }).ToSlice(&ret)
	case PageLeft:
		linq.From(w).OrderByDescending(func(i interface{}) interface{} { return i.(*TextWord).Quad.BottomLeft.Y }).ToSlice(&ret)
	case PageRight:
		linq.From(w).OrderBy(func(i interface{}) interface{} { return i.(*TextWord).Quad.BottomLeft.Y }).ToSlice(&ret)
	default:
		avgAngle := linq.From(w).Select(func(i interface{}) interface{} { return i.(*TextWord).Quad.Rotation() }).Average()
		switch {
		case 0 < avgAngle && avgAngle <= 90:
			linq.From(w).OrderBy(func(i interface{}) interface{} { return i.(*TextWord).Quad.BottomLeft.X }).ThenBy(func(i interface{}) interface{} { return i.(*TextWord).Quad.BottomLeft.Y }).ToSlice(&ret)
		case 90 < avgAngle && avgAngle <= 180:
			linq.From(w).OrderByDescending(func(i interface{}) interface{} { return i.(*TextWord).Quad.BottomLeft.X }).ThenBy(func(i interface{}) interface{} { return i.(*TextWord).Quad.BottomLeft.Y }).ToSlice(&ret)
		case -180 < avgAngle && avgAngle <= -90:
			linq.From(w).OrderByDescending(func(i interface{}) interface{} { return i.(*TextWord).Quad.BottomLeft.X }).ThenByDescending(func(i interface{}) interface{} { return i.(*TextWord).Quad.BottomLeft.Y }).ToSlice(&ret)
		case -90 < avgAngle && avgAngle <= 0:
			linq.From(w).OrderBy(func(i interface{}) interface{} { return i.(*TextWord).Quad.BottomLeft.X }).ThenByDescending(func(i interface{}) interface{} { return i.(*TextWord).Quad.BottomLeft.Y }).ToSlice(&ret)
		default:
			return w
		}
	}

	return
}

type TextWord struct {
	Value         string
	Quad          Quad
	Confidence    float64
	Orientation   Orientation
	DeskewAngle   float64
	StartBaseline Point
	EndBaseline   Point
}

func (w TextWord) IsWhitespace() bool {
	return strings.TrimSpace(w.Value) == ""
}

func (w TextWord) String() string {
	return w.Value
}

type TextLines []TextLine

func (l TextLines) Orientation() (orientation Orientation) {
	if len(l) == 0 {
		return OtherOrientation
	}

	orientation = l[0].Orientation
	if orientation == OtherOrientation {
		return
	}

	for _, word := range l[1:] {
		if word.Orientation != orientation {
			return OtherOrientation
		}
	}
	return
}

func (l TextLines) OrderByReadingOrder() (ret TextLines) {
	if len(l) <= 1 {
		return l
	}

	switch l.Orientation() {
	case PageUp:
		linq.From(l).OrderByDescending(func(i interface{}) interface{} { return i.(*TextLine).Quad.BottomLeft.Y }).ToSlice(&ret)
	case PageDown:
		linq.From(l).OrderBy(func(i interface{}) interface{} { return i.(*TextLine).Quad.BottomLeft.Y }).ToSlice(&ret)
	case PageLeft:
		linq.From(l).OrderByDescending(func(i interface{}) interface{} { return i.(*TextLine).Quad.BottomLeft.X }).ToSlice(&ret)
	case PageRight:
		linq.From(l).OrderBy(func(i interface{}) interface{} { return i.(*TextLine).Quad.BottomLeft.X }).ToSlice(&ret)
	default:
		avgAngle := linq.From(l).Select(func(i interface{}) interface{} { return i.(*TextLine).Quad.Rotation() }).Average()
		switch {
		case 0 < avgAngle && avgAngle <= 90:
			linq.From(l).OrderByDescending(func(i interface{}) interface{} { return i.(*TextLine).Quad.BottomLeft.Y }).ThenBy(func(i interface{}) interface{} { return i.(*TextLine).Quad.BottomLeft.X }).ToSlice(&ret)
		case 90 < avgAngle && avgAngle <= 180:
			linq.From(l).OrderBy(func(i interface{}) interface{} { return i.(*TextLine).Quad.BottomLeft.Y }).ThenBy(func(i interface{}) interface{} { return i.(*TextLine).Quad.BottomLeft.Y }).ToSlice(&ret)
		case -180 < avgAngle && avgAngle <= -90:
			linq.From(l).OrderBy(func(i interface{}) interface{} { return i.(*TextLine).Quad.BottomLeft.Y }).ThenByDescending(func(i interface{}) interface{} { return i.(*TextLine).Quad.BottomLeft.X }).ToSlice(&ret)
		case -90 < avgAngle && avgAngle <= 0:
			linq.From(l).OrderByDescending(func(i interface{}) interface{} { return i.(*TextLine).Quad.BottomLeft.X }).ThenByDescending(func(i interface{}) interface{} { return i.(*TextLine).Quad.BottomLeft.Y }).ToSlice(&ret)
		default:
			return l
		}
	}

	return
}

type TextLine struct {
	TextWords
	Quad          Quad
	Orientation   Orientation
	WordSeparator string
}

func MakeTextLine(words TextWords, sep string) TextLine {
	var quads Quads = make(Quads, len(words))
	for i, l := range words {
		quads[i] = l.Quad
	}

	return TextLine{
		TextWords:     words,
		Orientation:   quads.Orientation(),
		Quad:          quads.Union(),
		WordSeparator: sep,
	}
}

func (l TextLine) String() string {
	words := make([]string, 0, len(l.TextWords))
	for _, w := range l.TextWords {
		if strings.TrimSpace(w.Value) == "" {
			continue
		}
		words = append(words, w.String())
	}
	return strings.Join(words, l.WordSeparator)
}

type TextBlocks []TextBlock

type TextBlock struct {
	TextLines
	Quad          Quad
	Orientation   Orientation
	LineSeparator string
}

func MakeTextBlock(lines TextLines, sep string) TextBlock {
	var quads Quads = make(Quads, len(lines))
	for i, l := range lines {
		quads[i] = l.Quad
	}

	return TextBlock{
		TextLines:     lines,
		Orientation:   quads.Orientation(),
		Quad:          quads.Union(),
		LineSeparator: sep,
	}
}

func (b TextBlock) String() string {
	lines := make([]string, len(b.TextLines))
	for i, l := range b.TextLines {
		lines[i] = l.String()
	}
	return strings.Join(lines, b.LineSeparator)
}
