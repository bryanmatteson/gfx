package afm

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"

	"go.matteson.dev/gfx"
)

const (
	Comment            = "Comment"
	StartFontMetrics   = "StartFontMetrics"
	EndFontMetrics     = "EndFontMetrics"
	FontName           = "FontName"
	FullName           = "FullName"
	FamilyName         = "FamilyName"
	Weight             = "Weight"
	FontBbox           = "FontBBox"
	Version            = "Version"
	Notice             = "Notice"
	EncodingScheme     = "EncodingScheme"
	MappingScheme      = "MappingScheme"
	EscChar            = "EscChar"
	CharacterSet       = "CharacterSet"
	Characters         = "Characters"
	IsBaseFont         = "IsBaseFont"
	VVector            = "VVector"
	IsFixedV           = "IsFixedV"
	CapHeight          = "CapHeight"
	XHeight            = "XHeight"
	Ascender           = "Ascender"
	Descender          = "Descender"
	UnderlinePosition  = "UnderlinePosition"
	UnderlineThickness = "UnderlineThickness"
	ItalicAngle        = "ItalicAngle"
	CharWidth          = "CharWidth"
	IsFixedPitch       = "IsFixedPitch"
	StartCharMetrics   = "StartCharMetrics"
	EndCharMetrics     = "EndCharMetrics"
	CharmetricsC       = "C"
	CharmetricsCh      = "CH"
	CharmetricsWx      = "WX"
	CharmetricsW0X     = "W0X"
	CharmetricsW1X     = "W1X"
	CharmetricsWy      = "WY"
	CharmetricsW0Y     = "W0Y"
	CharmetricsW1Y     = "W1Y"
	CharmetricsW       = "W"
	CharmetricsW0      = "W0"
	CharmetricsW1      = "W1"
	CharmetricsVv      = "VV"
	CharmetricsN       = "N"
	CharmetricsB       = "B"
	CharmetricsL       = "L"
	StdHw              = "StdHW"
	StdVw              = "StdVW"
	StartTrackKern     = "StartTrackKern"
	EndTrackKern       = "EndTrackKern"
	StartKernData      = "StartKernData"
	EndKernData        = "EndKernData"
	StartKernPairs     = "StartKernPairs"
	EndKernPairs       = "EndKernPairs"
	StartKernPairs0    = "StartKernPairs0"
	StartKernPairs1    = "StartKernPairs1"
	StartComposites    = "StartComposites"
	EndComposites      = "EndComposites"
	Cc                 = "CC"
	Pcc                = "PCC"
	KernPairKp         = "KP"
	KernPairKph        = "KPH"
	KernPairKpx        = "KPX"
	KernPairKpy        = "KPY"
)

func Parse(b []byte) (Metrics, error) { return ParseReader(bytes.NewReader(b)) }

func ParseReader(r io.Reader) (metrics Metrics, err error) {
	scanner := newScanner(r)
	token, err := scanner.word()
	if err != nil {
		return
	}

	if !strings.EqualFold(token, StartFontMetrics) {
		err = fmt.Errorf("invalid afm file, did not start with %s", StartFontMetrics)
		return
	}

	metrics.AfmVersion, err = scanner.float()
	if err != nil {
		return
	}

	for !scanner.eof() {
		token, err = scanner.word()
		if err != nil {
			return
		}

		switch token {
		case Comment:
			comment, err := scanner.line()
			if err != nil {
				return metrics, err
			}
			metrics.Comments = append(metrics.Comments, comment)
		case FontName:
			metrics.FontName, err = scanner.line()
			if err != nil {
				return metrics, err
			}
		case FullName:
			metrics.FullName, err = scanner.line()
			if err != nil {
				return metrics, err
			}
		case FamilyName:
			metrics.FamilyName, err = scanner.line()
			if err != nil {
				return metrics, err
			}
		case Weight:
			metrics.Weight, err = scanner.line()
			if err != nil {
				return metrics, err
			}
		case ItalicAngle:
			metrics.ItalicAngle, err = scanner.float()
			if err != nil {
				return metrics, err
			}
		case IsFixedPitch:
			metrics.IsFixedPitch, err = scanner.bool()
			if err != nil {
				return metrics, err
			}
		case FontBbox:
			x1, err := scanner.float()
			if err != nil {
				return metrics, err
			}
			y1, err := scanner.float()
			if err != nil {
				return metrics, err
			}
			x2, err := scanner.float()
			if err != nil {
				return metrics, err
			}
			y2, err := scanner.float()
			if err != nil {
				return metrics, err
			}
			metrics.BoundingBox = gfx.MakeQuad(x1, y1, x2, y2)
		case UnderlinePosition:
			metrics.UnderlinePosition, err = scanner.float()
			if err != nil {
				return metrics, err
			}
		case UnderlineThickness:
			metrics.UnderlineThickness, err = scanner.float()
			if err != nil {
				return metrics, err
			}
		case Version:
			metrics.Version, err = scanner.line()
			if err != nil {
				return metrics, err
			}
		case Notice:
			metrics.Notice, err = scanner.line()
			if err != nil {
				return metrics, err
			}
		case EncodingScheme:
			metrics.EncodingScheme, err = scanner.line()
			if err != nil {
				return metrics, err
			}
		case MappingScheme:
			metrics.MappingScheme, err = scanner.int()
			if err != nil {
				return metrics, err
			}
		case CharacterSet:
			metrics.CharacterSet, err = scanner.line()
			if err != nil {
				return metrics, err
			}
		case EscChar:
			metrics.EscapeCharacter, err = scanner.int()
			if err != nil {
				return metrics, err
			}
		case Characters:
			metrics.Characters, err = scanner.int()
			if err != nil {
				return metrics, err
			}
		case IsBaseFont:
			metrics.IsBaseFont, err = scanner.bool()
			if err != nil {
				return metrics, err
			}
		case CapHeight:
			metrics.CapHeight, err = scanner.float()
			if err != nil {
				return metrics, err
			}
		case XHeight:
			metrics.XHeight, err = scanner.float()
			if err != nil {
				return metrics, err
			}
		case Ascender:
			metrics.Ascender, err = scanner.float()
			if err != nil {
				return metrics, err
			}
		case Descender:
			metrics.Descender, err = scanner.float()
			if err != nil {
				return metrics, err
			}
		case StdHw:
			metrics.HorizontalStemWidth, err = scanner.float()
			if err != nil {
				return metrics, err
			}
		case StdVw:
			metrics.VerticalStemWidth, err = scanner.float()
			if err != nil {
				return metrics, err
			}
		case CharWidth:
			x1, err := scanner.float()
			if err != nil {
				return metrics, err
			}
			y1, err := scanner.float()
			if err != nil {
				return metrics, err
			}
			metrics.CharacterWidth = gfx.Point{X: x1, Y: y1}
		case VVector:
			x1, err := scanner.float()
			if err != nil {
				return metrics, err
			}
			y1, err := scanner.float()
			if err != nil {
				return metrics, err
			}
			metrics.VVector = gfx.Point{X: x1, Y: y1}
		case IsFixedV:
			metrics.IsFixedV, err = scanner.bool()
			if err != nil {
				return metrics, err
			}
		case StartCharMetrics:
			count, err := scanner.int()
			if err != nil {
				return metrics, err
			}
			if metrics.CharacterMetrics == nil {
				metrics.CharacterMetrics = make(map[string]IndividualCharacterMetric, count)
			}

			for i := 0; i < count; i++ {
				m, err := parseCharMetric(scanner)
				if err != nil {
					return metrics, err
				}
				metrics.CharacterMetrics[m.Name] = m
			}
			end, err := scanner.word()
			if err != nil {
				return metrics, err
			}
			if end != EndCharMetrics {
				return metrics, fmt.Errorf("character metrics section did not end with %s, instead it was %s", EndCharMetrics, end)
			}
		case EndFontMetrics:
		case StartKernData:
		default:
		}
	}

	return metrics, nil
}

func parseCharMetric(scanner *scanner) (metric IndividualCharacterMetric, err error) {
	line, err := scanner.line()
	if err != nil {
		return metric, err
	}

	split := strings.Split(line, ";")
	for _, s := range split {
		parts := strings.Split(strings.TrimSpace(s), " ")

		switch parts[0] {
		case CharmetricsC, CharmetricsCh:
			val, err := strconv.ParseInt(parts[1], 0, 32)
			if err != nil {
				return metric, err
			}
			metric.CharacterCode = int(val)
		case CharmetricsWx:
			val, err := strconv.ParseFloat(parts[1], 32)
			if err != nil {
				return metric, err
			}
			metric.Width.X = float64(val)
		case CharmetricsW0X:
			val, err := strconv.ParseFloat(parts[1], 32)
			if err != nil {
				return metric, err
			}
			metric.WidthDirection0.X = float64(val)
		case CharmetricsW1X:
			val, err := strconv.ParseFloat(parts[1], 32)
			if err != nil {
				return metric, err
			}
			metric.WidthDirection1.X = float64(val)
		case CharmetricsWy:
			val, err := strconv.ParseFloat(parts[1], 32)
			if err != nil {
				return metric, err
			}
			metric.Width.Y = float64(val)
		case CharmetricsW0Y:
			val, err := strconv.ParseFloat(parts[1], 32)
			if err != nil {
				return metric, err
			}
			metric.WidthDirection0.Y = float64(val)
		case CharmetricsW1Y:
			val, err := strconv.ParseFloat(parts[1], 32)
			if err != nil {
				return metric, err
			}
			metric.WidthDirection1.Y = float64(val)
		case CharmetricsW:
			x, err := strconv.ParseFloat(parts[1], 32)
			if err != nil {
				return metric, err
			}
			y, err := strconv.ParseFloat(parts[2], 32)
			if err != nil {
				return metric, err
			}
			metric.Width.X = float64(x)
			metric.Width.Y = float64(y)

		case CharmetricsW0:
			x, err := strconv.ParseFloat(parts[1], 32)
			if err != nil {
				return metric, err
			}
			y, err := strconv.ParseFloat(parts[2], 32)
			if err != nil {
				return metric, err
			}
			metric.WidthDirection0.X = float64(x)
			metric.WidthDirection0.Y = float64(y)
		case CharmetricsW1:
			x, err := strconv.ParseFloat(parts[1], 32)
			if err != nil {
				return metric, err
			}
			y, err := strconv.ParseFloat(parts[2], 32)
			if err != nil {
				return metric, err
			}
			metric.WidthDirection1.X = float64(x)
			metric.WidthDirection1.Y = float64(y)
		case CharmetricsVv:
			x, err := strconv.ParseFloat(parts[1], 32)
			if err != nil {
				return metric, err
			}
			y, err := strconv.ParseFloat(parts[2], 32)
			if err != nil {
				return metric, err
			}
			metric.VVector.X = float64(x)
			metric.VVector.Y = float64(y)
		case CharmetricsB:
			x1, err := strconv.ParseFloat(parts[1], 32)
			if err != nil {
				return metric, err
			}
			y1, err := strconv.ParseFloat(parts[2], 32)
			if err != nil {
				return metric, err
			}
			x2, err := strconv.ParseFloat(parts[3], 32)
			if err != nil {
				return metric, err
			}
			y2, err := strconv.ParseFloat(parts[4], 32)
			if err != nil {
				return metric, err
			}
			metric.BoundingBox = gfx.MakeQuad(x1, y1, x2, y2)
		case CharmetricsL:
			metric.Ligature.Successor = parts[1]
			metric.Ligature.Value = parts[2]
		case CharmetricsN:
			metric.Name = parts[1]
		}
	}

	return metric, nil
}
