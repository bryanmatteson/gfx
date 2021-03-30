package cff

import (
	"fmt"

	"go.matteson.dev/gfx"
)

const (
	DefaultBlueScale       float64 = 0.039625
	DefaultExpansionFactor float64 = 0.06
	DefaultBlueFuzz        int     = 1
	DefaultBlueShift       int     = 7
	DefaultLanguageGroup   int     = 0
)

type dictionary interface {
	ApplyOperands(operands operands, key []byte, strIndex strtable) error
}

type AdobePrivateDictionary struct {
	BlueValues               []int
	OtherBlues               []int
	FamilyBlues              []int
	FamilyOtherBlues         []int
	BlueScale                float64
	BlueShift                int
	BlueFuzz                 int
	StandardHorizontalWidth  float64
	StandardVerticalWidth    float64
	StemSnapHorizontalWidths []float64
	StemSnapVerticalWidths   []float64
	ForceBold                bool
	LanguageGroup            int
	ExpansionFactor          float64
}

type PrivateDictionary struct {
	AdobePrivateDictionary
	InitialRandomSeed     float64
	LocalSubroutineOffset int
	DefaultWidthX         float64
	NominalWidthX         float64
}

func newPrivateDictionary() *PrivateDictionary {
	return &PrivateDictionary{
		LocalSubroutineOffset: -1,
	}
}

func (d *PrivateDictionary) ApplyOperands(operands operands, key []byte, strIndex strtable) (err error) {
	switch key[0] {
	case 6:
		d.BlueValues = operands.GetDeltaInts()
	case 7:
		d.OtherBlues = operands.GetDeltaInts()
	case 8:
		d.FamilyBlues = operands.GetDeltaInts()
	case 9:
		d.FamilyOtherBlues = operands.GetDeltaInts()
	case 10:
		d.StandardHorizontalWidth = operands.GetFloatOrDef(0, 0)
	case 11:
		d.StandardVerticalWidth = operands.GetFloatOrDef(0, 0)
	case 12:
		if len(key) < 2 {
			return fmt.Errorf("invalid key: 12 requires second byte")
		}
		switch key[1] {
		case 9:
			d.BlueScale = operands.GetFloatOrDef(0, 0)
		case 10:
			d.BlueShift = operands.GetIntOrDef(0, 0)
		case 11:
			d.BlueFuzz = operands.GetIntOrDef(0, 0)
		case 12:
			d.StemSnapHorizontalWidths = operands.GetDeltas()
		case 13:
			d.StemSnapVerticalWidths = operands.GetDeltas()
		case 14:
			d.ForceBold = operands.GetFloatOrDef(0, 0) == 1
		case 17:
			d.LanguageGroup = operands.GetIntOrDef(0, 0)
		case 18:
			d.ExpansionFactor = operands.GetFloatOrDef(0, 0)
		case 19:
			d.InitialRandomSeed = operands.GetFloatOrDef(0, 0)
		}
	case 19:
		d.LocalSubroutineOffset = operands.GetIntOrDef(0, -1)
	case 20:
		d.DefaultWidthX = operands.GetFloatOrDef(0, 0)
	case 21:
		d.NominalWidthX = operands.GetFloatOrDef(0, 0)
	}
	return
}

type TopLevelDictionary struct {
	Version            string
	Notice             string
	Copyright          string
	FullName           string
	FamilyName         string
	Weight             string
	IsFixedPitch       bool
	ItalicAngle        float64
	UnderlinePosition  float64
	UnderlineThickness float64
	PaintType          float64
	CharStringType     CharStringType
	FontMatrix         gfx.Matrix
	StrokeWidth        float64
	UniqueId           float64
	FontBoundingBox    gfx.Quad
	Xuid               []float64
	CharSetOffset      int
	EncodingOffset     int

	privateDictSize   int
	privateDictOffset int

	CharStringsOffset      int
	SyntheticBaseFontIndex int
	PostScript             string
	BaseFontName           string

	BaseFontBlend    []float64
	IsCidFont        bool
	CidFontOperators CidFontOperators
}

func (tld *TopLevelDictionary) ApplyOperands(operands operands, key []byte, strIndex strtable) (err error) {
	switch key[0] {
	case 0:
		tld.Version, err = operands.GetString(strIndex)
	case 1:
		tld.Notice, err = operands.GetString(strIndex)
	case 2:
		tld.FullName, err = operands.GetString(strIndex)
	case 3:
		tld.FamilyName, err = operands.GetString(strIndex)
	case 4:
		tld.Weight, err = operands.GetString(strIndex)
	case 5:
		tld.FontBoundingBox = operands.GetBoundingBox()
	case 12:
		if len(key) < 2 {
			return fmt.Errorf("invalid key: 12 requires second byte")
		}
		switch key[1] {
		case 0:
			tld.Copyright, err = operands.GetString(strIndex)
		case 1:
			tld.IsFixedPitch = int(operands.GetFloatOrDef(0, 0)) == 1
		case 2:
			tld.ItalicAngle = operands.GetFloatOrDef(0, 0)
		case 3:
			tld.UnderlinePosition = operands.GetFloatOrDef(0, 0)
		case 4:
			tld.UnderlineThickness = operands.GetFloatOrDef(0, 0)
		case 5:
			tld.PaintType = operands[0].FloatOrDef(0)
		case 6:
			tld.CharStringType = CharStringType(operands[0].IntOrDef(2))
		case 7:
			vals := operands.GetFloats()
			switch {
			case len(vals) == 4:
				tld.FontMatrix = gfx.NewMatrix(vals[0], vals[1], 0, vals[2], vals[3], 0)
			case len(vals) == 6:
				tld.FontMatrix = gfx.NewMatrix(vals[0], vals[1], vals[2], vals[3], vals[4], vals[5])
			default:
				return fmt.Errorf("invalid number of values for font matrix, got %d", len(vals))
			}
		case 8:
			tld.StrokeWidth = operands.GetFloatOrDef(0, 0)
		case 20:
			tld.SyntheticBaseFontIndex = operands.GetIntOrDef(0, 0)
		case 21:
			tld.PostScript, err = operands.GetString(strIndex)
		case 22:
			tld.BaseFontName, err = operands.GetString(strIndex)
		case 23:
			tld.BaseFontBlend = operands.GetDeltas()
		case 30:
			registry, err := operands.GetString(strIndex)
			if err != nil {
				return err
			}
			ordering, err := operands[1].GetString(strIndex)
			if err != nil {
				return err
			}

			supplement := operands.GetFloatOrDef(2, 0)
			tld.CidFontOperators.Ros = RegistryOrderingSupplement{
				Registry:   registry,
				Ordering:   ordering,
				Supplement: supplement,
			}
		case 31:
			tld.CidFontOperators.Version = operands.GetIntOrDef(0, 0)
		case 32:
			tld.CidFontOperators.Revision = operands.GetIntOrDef(0, 0)
		case 33:
			tld.CidFontOperators.Type = operands.GetIntOrDef(0, 0)
		case 34:
			tld.CidFontOperators.Count = operands.GetIntOrDef(0, 0)
		case 35:
			tld.CidFontOperators.UIDBase = operands.GetFloatOrDef(0, 0)
		case 36:
			tld.CidFontOperators.FontDictionaryArray = operands.GetIntOrDef(0, 0)
		case 37:
			tld.CidFontOperators.FontDictionarySelect = operands.GetIntOrDef(0, 0)
		case 38:
			tld.CidFontOperators.FontName, err = operands.GetString(strIndex)
		}
	case 13:
		tld.UniqueId = operands.GetFloatOrDef(0, 0)
	case 14:
		tld.Xuid = operands.GetFloats()
	case 15:
		tld.CharSetOffset = operands.GetIntOrDef(0, 0)
	case 16:
		tld.EncodingOffset = operands.GetIntOrDef(0, -1)
	case 17:
		tld.CharStringsOffset = operands.GetIntOrDef(0, -1)
	case 18:
		tld.privateDictSize = operands.GetIntOrDef(0, -1)
		tld.privateDictOffset = operands.GetIntOrDef(1, -1)
	}
	return
}

func newTopLevelDictionary() *TopLevelDictionary {
	dict := &TopLevelDictionary{
		UnderlinePosition:  -100,
		UnderlineThickness: 50,
		FontMatrix:         gfx.NewScaleMatrix(0.001, 0.001),
		CharStringType:     Type2,
		FontBoundingBox:    gfx.MakeQuad(0, 0, 0, 0),
		CharSetOffset:      -1,
		EncodingOffset:     -1,
		CharStringsOffset:  -1,
	}
	dict.CidFontOperators.Count = 8720
	return dict
}

type CidFontOperators struct {
	Ros                  RegistryOrderingSupplement
	Version              int
	Revision             int
	Type                 int
	Count                int
	UIDBase              float64
	FontDictionaryArray  int
	FontDictionarySelect int
	FontName             string
}

type RegistryOrderingSupplement struct {
	Registry   string
	Ordering   string
	Supplement float64
}

type CharStringType int

const (
	Type1 CharStringType = 1
	Type2 CharStringType = 2
)

func readDictionary(reader *reader, dict dictionary, strIndex strtable) (err error) {
	operands := make(operands, 0)

	for !reader.eof() {
		operands = operands[0:0]
		for {
			b, err := reader.byte()
			if err != nil {
				return err
			}

			if b <= 21 {
				key := []byte{b}
				if b == 12 {
					b2, err := reader.byte()
					if err != nil {
						return err
					}
					key = append(key, b2)
				}
				if err := dict.ApplyOperands(operands, key, strIndex); err != nil {
					return err
				}
				break
			}

			switch {
			case b == 28:
				val, err := reader.card16()
				if err != nil {
					return err
				}
				operands = append(operands, operand{int(val)})
			case b == 29:
				val, err := reader.card32()
				if err != nil {
					return err
				}
				operands = append(operands, operand{val})
			case b == 30:
				val, err := reader.real()
				if err != nil {
					return err
				}
				operands = append(operands, operand{val})
			case b >= 32 && b <= 246:
				operands = append(operands, operand{int(b - 139)})
			case b >= 247 && b <= 250:
				b1, err := reader.card8()
				if err != nil {
					return err
				}
				val := int(+int32(b-247)*256 + int32(b1) + 108)
				operands = append(operands, operand{val})
			case b >= 251 && b <= 254:
				b1, err := reader.card8()
				if err != nil {
					return err
				}
				val := int(-int32(b-251)*256 - int32(b1) - 108)
				operands = append(operands, operand{val})
			default:
				return fmt.Errorf("out of range in top level dict")
			}
		}
	}
	return
}
