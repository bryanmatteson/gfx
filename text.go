package gfx

// // Writing modes
// const (
// 	WModeHorizontal int = iota
// 	WModeVertical
// )

// type Letters []Letter
// type Letter struct {
// 	Rune     rune
// 	GlyphID  int
// 	Origin   Point
// 	FontData *FontData
// 	FontSize float64
// 	Color    color.Color
// }

// func (ch Letters) Orientation() (orientation Orientation) {
// 	if len(ch) == 0 {
// 		return OtherOrientation
// 	}

// 	orientation = ch[0].Orientation
// 	if orientation == OtherOrientation {
// 		return
// 	}

// 	for _, word := range ch[1:] {
// 		if word.Orientation != orientation {
// 			return OtherOrientation
// 		}
// 	}
// 	return
// }

// func (ch Letters) OrderByReadingOrder() (ret Letters) {
// 	if len(ch) <= 1 {
// 		return ch
// 	}

// 	switch ch.Orientation() {
// 	case PageUp:
// 		linq.From(ch).OrderBy(func(i interface{}) interface{} { return i.(Letter).Quad.BottomLeft.X }).ToSlice(&ret)
// 	case PageDown:
// 		linq.From(ch).OrderByDescending(func(i interface{}) interface{} { return i.(Letter).Quad.BottomLeft.X }).ToSlice(&ret)
// 	case PageLeft:
// 		linq.From(ch).OrderByDescending(func(i interface{}) interface{} { return i.(Letter).Quad.BottomLeft.Y }).ToSlice(&ret)
// 	case PageRight:
// 		linq.From(ch).OrderBy(func(i interface{}) interface{} { return i.(Letter).Quad.BottomLeft.Y }).ToSlice(&ret)
// 	default:
// 		avgAngle := linq.From(ch).Select(func(i interface{}) interface{} { return i.(Letter).Quad.Rotation() }).Average()
// 		switch {
// 		case 0 < avgAngle && avgAngle <= 90:
// 			linq.From(ch).OrderBy(func(i interface{}) interface{} { return i.(Letter).Quad.BottomLeft.X }).ThenBy(func(i interface{}) interface{} { return i.(Letter).Quad.BottomLeft.Y }).ToSlice(&ret)
// 		case 90 < avgAngle && avgAngle <= 180:
// 			linq.From(ch).OrderByDescending(func(i interface{}) interface{} { return i.(Letter).Quad.BottomLeft.X }).ThenBy(func(i interface{}) interface{} { return i.(Letter).Quad.BottomLeft.Y }).ToSlice(&ret)
// 		case -180 < avgAngle && avgAngle <= -90:
// 			linq.From(ch).OrderByDescending(func(i interface{}) interface{} { return i.(Letter).Quad.BottomLeft.X }).ThenByDescending(func(i interface{}) interface{} { return i.(Letter).Quad.BottomLeft.Y }).ToSlice(&ret)
// 		case -90 < avgAngle && avgAngle <= 0:
// 			linq.From(ch).OrderBy(func(i interface{}) interface{} { return i.(Letter).Quad.BottomLeft.X }).ThenByDescending(func(i interface{}) interface{} { return i.(Letter).Quad.BottomLeft.Y }).ToSlice(&ret)
// 		default:
// 			return ch
// 		}
// 	}

// 	return
// }

// func (ch Letters) IsWhitespace() bool {
// 	for _, letter := range ch {
// 		if !letter.IsWhitespace() {
// 			return false
// 		}
// 	}
// 	return true
// }

// func (l Letter) IsWhitespace() bool { return unicode.IsSpace(l.Rune) }

// type TextWords []TextWord

// func (w TextWords) Orientation() (orientation Orientation) {
// 	if len(w) == 0 {
// 		return OtherOrientation
// 	}

// 	orientation = w[0].Orientation
// 	if orientation == OtherOrientation {
// 		return
// 	}

// 	for _, word := range w[1:] {
// 		if word.Orientation != orientation {
// 			return OtherOrientation
// 		}
// 	}
// 	return
// }

// func (w TextWords) OrderByReadingOrder() (ret TextWords) {
// 	if len(w) <= 1 {
// 		return w
// 	}

// 	switch w.Orientation() {
// 	case PageUp:
// 		linq.From(w).OrderBy(func(i interface{}) interface{} { return i.(TextWord).Quad.BottomLeft.X }).ToSlice(&ret)
// 	case PageDown:
// 		linq.From(w).OrderByDescending(func(i interface{}) interface{} { return i.(TextWord).Quad.BottomLeft.X }).ToSlice(&ret)
// 	case PageLeft:
// 		linq.From(w).OrderByDescending(func(i interface{}) interface{} { return i.(TextWord).Quad.BottomLeft.Y }).ToSlice(&ret)
// 	case PageRight:
// 		linq.From(w).OrderBy(func(i interface{}) interface{} { return i.(TextWord).Quad.BottomLeft.Y }).ToSlice(&ret)
// 	default:
// 		avgAngle := linq.From(w).Select(func(i interface{}) interface{} { return i.(TextWord).Quad.Rotation() }).Average()
// 		switch {
// 		case 0 < avgAngle && avgAngle <= 90:
// 			linq.From(w).OrderBy(func(i interface{}) interface{} { return i.(TextWord).Quad.BottomLeft.X }).ThenBy(func(i interface{}) interface{} { return i.(TextWord).Quad.BottomLeft.Y }).ToSlice(&ret)
// 		case 90 < avgAngle && avgAngle <= 180:
// 			linq.From(w).OrderByDescending(func(i interface{}) interface{} { return i.(TextWord).Quad.BottomLeft.X }).ThenBy(func(i interface{}) interface{} { return i.(TextWord).Quad.BottomLeft.Y }).ToSlice(&ret)
// 		case -180 < avgAngle && avgAngle <= -90:
// 			linq.From(w).OrderByDescending(func(i interface{}) interface{} { return i.(TextWord).Quad.BottomLeft.X }).ThenByDescending(func(i interface{}) interface{} { return i.(TextWord).Quad.BottomLeft.Y }).ToSlice(&ret)
// 		case -90 < avgAngle && avgAngle <= 0:
// 			linq.From(w).OrderBy(func(i interface{}) interface{} { return i.(TextWord).Quad.BottomLeft.X }).ThenByDescending(func(i interface{}) interface{} { return i.(TextWord).Quad.BottomLeft.Y }).ToSlice(&ret)
// 		default:
// 			return w
// 		}
// 	}

// 	return
// }

// type TextWord struct {
// 	Value         string
// 	Quad          Quad
// 	Confidence    float64
// 	Orientation   Orientation
// 	DeskewAngle   float64
// 	StartBaseline Point
// 	EndBaseline   Point
// }

// func MakeWord(letters Letters) (word TextWord) {
// 	if len(letters) == 0 {
// 		return
// 	}

// 	if len(letters) == 1 {
// 		word.Confidence = letters[0].Confidence
// 		word.DeskewAngle = letters[0].DeskewAngle
// 		word.Quad = letters[0].Quad
// 		word.EndBaseline = letters[0].EndBaseline
// 		word.StartBaseline = letters[0].StartBaseline
// 		word.EndBaseline = letters[0].EndBaseline
// 		word.Orientation = letters[0].Orientation
// 		word.Value = fmt.Sprintf("%c", letters[0].Rune)
// 		return
// 	}

// 	letters = letters.OrderByReadingOrder()

// 	quads := make(Quads, len(letters))
// 	for i, letter := range letters {
// 		quads[i] = letter.Quad
// 	}

// 	var builder strings.Builder
// 	for _, letter := range letters {
// 		builder.WriteRune(letter.Rune)
// 	}
// 	word.Value = builder.String()
// 	word.Confidence = letters.GetMeanConfidence()
// 	word.DeskewAngle = letters.GetMeanDeskewAngle()
// 	word.Orientation = letters.Orientation()
// 	word.StartBaseline = letters[0].StartBaseline
// 	word.EndBaseline = letters[len(letters)-1].EndBaseline
// 	word.Quad = quads.Union()

// 	return
// }

// func (w TextWord) IsWhitespace() bool {
// 	return strings.TrimSpace(w.Value) == ""
// }

// func (w TextWord) String() string {
// 	return w.Value
// }

// type TextLines []TextLine

// func (l TextLines) Orientation() (orientation Orientation) {
// 	if len(l) == 0 {
// 		return OtherOrientation
// 	}

// 	orientation = l[0].Orientation
// 	if orientation == OtherOrientation {
// 		return
// 	}

// 	for _, word := range l[1:] {
// 		if word.Orientation != orientation {
// 			return OtherOrientation
// 		}
// 	}
// 	return
// }

// func (l TextLines) OrderByReadingOrder() (ret TextLines) {
// 	if len(l) <= 1 {
// 		return l
// 	}

// 	switch l.Orientation() {
// 	case PageUp:
// 		linq.From(l).OrderByDescending(func(i interface{}) interface{} { return i.(TextLine).Quad.BottomLeft.Y }).ToSlice(&ret)
// 	case PageDown:
// 		linq.From(l).OrderBy(func(i interface{}) interface{} { return i.(TextLine).Quad.BottomLeft.Y }).ToSlice(&ret)
// 	case PageLeft:
// 		linq.From(l).OrderByDescending(func(i interface{}) interface{} { return i.(TextLine).Quad.BottomLeft.X }).ToSlice(&ret)
// 	case PageRight:
// 		linq.From(l).OrderBy(func(i interface{}) interface{} { return i.(TextLine).Quad.BottomLeft.X }).ToSlice(&ret)
// 	default:
// 		avgAngle := linq.From(l).Select(func(i interface{}) interface{} { return i.(TextLine).Quad.Rotation() }).Average()
// 		switch {
// 		case 0 < avgAngle && avgAngle <= 90:
// 			linq.From(l).OrderByDescending(func(i interface{}) interface{} { return i.(TextLine).Quad.BottomLeft.Y }).ThenBy(func(i interface{}) interface{} { return i.(TextLine).Quad.BottomLeft.X }).ToSlice(&ret)
// 		case 90 < avgAngle && avgAngle <= 180:
// 			linq.From(l).OrderBy(func(i interface{}) interface{} { return i.(TextLine).Quad.BottomLeft.Y }).ThenBy(func(i interface{}) interface{} { return i.(TextLine).Quad.BottomLeft.Y }).ToSlice(&ret)
// 		case -180 < avgAngle && avgAngle <= -90:
// 			linq.From(l).OrderBy(func(i interface{}) interface{} { return i.(TextLine).Quad.BottomLeft.Y }).ThenByDescending(func(i interface{}) interface{} { return i.(TextLine).Quad.BottomLeft.X }).ToSlice(&ret)
// 		case -90 < avgAngle && avgAngle <= 0:
// 			linq.From(l).OrderByDescending(func(i interface{}) interface{} { return i.(TextLine).Quad.BottomLeft.X }).ThenByDescending(func(i interface{}) interface{} { return i.(TextLine).Quad.BottomLeft.Y }).ToSlice(&ret)
// 		default:
// 			return l
// 		}
// 	}

// 	return
// }

// type TextLine struct {
// 	TextWords
// 	Quad          Quad
// 	Orientation   Orientation
// 	WordSeparator string
// }

// func MakeTextLine(words TextWords, sep string) TextLine {
// 	var quads = make(Quads, len(words))
// 	for i, l := range words {
// 		quads[i] = l.Quad
// 	}

// 	return TextLine{
// 		TextWords:     words,
// 		Orientation:   quads.Orientation(),
// 		Quad:          quads.Union(),
// 		WordSeparator: sep,
// 	}
// }

// func (l TextLine) String() string {
// 	words := make([]string, 0, len(l.TextWords))
// 	for _, w := range l.TextWords {
// 		if strings.TrimSpace(w.Value) == "" {
// 			continue
// 		}
// 		words = append(words, w.String())
// 	}
// 	return strings.Join(words, l.WordSeparator)
// }

// type TextBlocks []TextBlock

// func (tb TextBlocks) String() string {
// 	var builder strings.Builder
// 	for _, block := range tb {
// 		builder.WriteString(block.String())
// 		builder.WriteRune('\n')
// 	}
// 	return builder.String()
// }

// type TextBlock struct {
// 	TextLines
// 	Quad          Quad
// 	Orientation   Orientation
// 	LineSeparator string
// }

// func MakeTextBlock(lines TextLines, sep string) TextBlock {
// 	var quads = make(Quads, len(lines))
// 	for i, l := range lines {
// 		quads[i] = l.Quad
// 	}

// 	return TextBlock{
// 		TextLines:     lines,
// 		Orientation:   quads.Orientation(),
// 		Quad:          quads.Union(),
// 		LineSeparator: sep,
// 	}
// }

// func (b TextBlock) String() string {
// 	lines := make([]string, len(b.TextLines))
// 	for i, l := range b.TextLines {
// 		lines[i] = l.String()
// 	}
// 	return strings.Join(lines, b.LineSeparator)
// }
