package cff

import (
	"fmt"

	"github.com/bryanmatteson/gfx/font/adobe"

	"github.com/bryanmatteson/gfx"
)

const (
	DefaultBlueScale       float64 = 0.039625
	DefaultExpansionFactor float64 = 0.06
	DefaultBlueFuzz        int     = 1
	DefaultBlueShift       int     = 7
	DefaultLanguageGroup   int     = 0
)

type PrivateDictionary struct {
	adobe.PrivateDictionary
	InitialRandomSeed     float64
	LocalSubroutineOffset int
	DefaultWidthX         float64
	NominalWidthX         float64
}

func newPrivateDictionary() *PrivateDictionary {
	priv := &PrivateDictionary{
		PrivateDictionary: adobe.PrivateDictionary{
			BlueScale:       DefaultBlueScale,
			ExpansionFactor: DefaultExpansionFactor,
			BlueFuzz:        DefaultBlueFuzz,
			BlueShift:       DefaultBlueShift,
			LanguageGroup:   DefaultLanguageGroup,
		},
		LocalSubroutineOffset: -1,
	}
	return priv
}

func (priv *PrivateDictionary) init(data []byte) error {
	cmds, err := parseCommandSequence(data, psPrivContext)
	if err != nil {
		return err
	}

	for _, cmd := range cmds {
		switch cmd.id {
		case 6:
			priv.BlueValues = deltaInts(cmd.args)
		case 7:
			priv.OtherBlues = deltaInts(cmd.args)
		case 8:
			priv.FamilyBlues = deltaInts(cmd.args)
		case 9:
			priv.FamilyOtherBlues = deltaInts(cmd.args)
		case 10:
			priv.StandardHorizontalWidth = cmd.args[0]
		case 11:
			priv.StandardVerticalWidth = cmd.args[0]
		case 19:
			priv.LocalSubroutineOffset = int(cmd.args[0])
		case 20:
			priv.DefaultWidthX = cmd.args[0]
		case 21:
			priv.NominalWidthX = cmd.args[0]
		case 1209:
			priv.BlueScale = cmd.args[0]
		case 1210:
			priv.BlueShift = int(cmd.args[0])
		case 1211:
			priv.BlueFuzz = int(cmd.args[0])
		case 1212:
			priv.StemSnapHorizontalWidths = deltaFloats(cmd.args)
		case 1213:
			priv.StemSnapVerticalWidths = deltaFloats(cmd.args)
		case 1214:
			priv.ForceBold = cmd.args[0] == 1
		case 1217:
			priv.LanguageGroup = int(cmd.args[0])
		case 1218:
			priv.ExpansionFactor = cmd.args[0]
		case 1219:
			priv.InitialRandomSeed = cmd.args[0]
		}
	}
	return nil
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

	PrivateDictSize   int
	PrivateDictOffset int

	CharStringsOffset      int
	SyntheticBaseFontIndex int
	PostScript             string
	BaseFontName           string

	BaseFontBlend    []float64
	IsCidFont        bool
	CidFontOperators CidFontOperators
}

func newTopLevelDictionary() *TopLevelDictionary {
	tld := &TopLevelDictionary{
		UnderlinePosition:  -100,
		UnderlineThickness: 50,
		FontMatrix:         gfx.NewScaleMatrix(0.001, 0.001),
		CharStringType:     Type2,
		FontBoundingBox:    gfx.MakeQuad(0, 0, 0, 0),
		CharSetOffset:      -1,
		EncodingOffset:     -1,
		CharStringsOffset:  -1,
	}
	tld.CidFontOperators.Count = 8720

	return tld
}

func (tld *TopLevelDictionary) init(data []byte, strIndex strtable) error {
	cmds, err := parseCommandSequence(data, psTLDContext)
	if err != nil {
		return err
	}

	for _, cmd := range cmds {
		switch cmd.id {
		case 0:
			tld.Version = strIndex.GetName(int(cmd.args[0]))
		case 1:
			tld.Notice = strIndex.GetName(int(cmd.args[0]))
		case 2:
			tld.FullName = strIndex.GetName(int(cmd.args[0]))
		case 3:
			tld.FamilyName = strIndex.GetName(int(cmd.args[0]))
		case 4:
			tld.Weight = strIndex.GetName(int(cmd.args[0]))
		case 5:
			tld.FontBoundingBox = gfx.MakeQuad(cmd.args[0], cmd.args[1], cmd.args[2], cmd.args[3])
		case 13:
			tld.UniqueId = cmd.args[0]
		case 14:
			tld.Xuid = cmd.args
		case 15:
			tld.CharSetOffset = int(cmd.args[0])
		case 16:
			tld.EncodingOffset = int(cmd.args[0])
		case 17:
			tld.CharStringsOffset = int(cmd.args[0])
		case 18:
			tld.PrivateDictSize = int(cmd.args[0])
			tld.PrivateDictOffset = int(cmd.args[1])
		case 1200:
			tld.Copyright = strIndex.GetName(int(cmd.args[0]))
		case 1201:
			tld.IsFixedPitch = int(cmd.args[0]) == 1
		case 1202:
			tld.ItalicAngle = cmd.args[0]
		case 1203:
			tld.UnderlinePosition = cmd.args[0]
		case 1204:
			tld.UnderlineThickness = cmd.args[0]
		case 1205:
			tld.PaintType = cmd.args[0]
		case 1206:
			tld.CharStringType = CharStringType(int(cmd.args[0]))
		case 1207:
			switch {
			case len(cmd.args) == 4:
				tld.FontMatrix = gfx.NewMatrix(cmd.args[0], cmd.args[1], 0, cmd.args[2], cmd.args[3], 0)
			case len(cmd.args) == 6:
				tld.FontMatrix = gfx.NewMatrix(cmd.args[0], cmd.args[1], cmd.args[2], cmd.args[3], cmd.args[4], cmd.args[5])
			default:
				return fmt.Errorf("invalid number of values for font matrix, got %d", len(cmd.args))
			}
		case 1208:
			tld.StrokeWidth = cmd.args[0]
		case 1220:
			tld.SyntheticBaseFontIndex = int(cmd.args[0])
		case 1221:
			tld.PostScript = strIndex.GetName(int(cmd.args[0]))
		case 1222:
			tld.BaseFontName = strIndex.GetName(int(cmd.args[0]))
		case 1223:
			tld.BaseFontBlend = deltaFloats(cmd.args)

		case 1230:
			tld.IsCidFont = true
			tld.CidFontOperators.Ros = RegistryOrderingSupplement{
				Registry:   strIndex.GetName(int(cmd.args[0])),
				Ordering:   strIndex.GetName(int(cmd.args[1])),
				Supplement: cmd.args[2],
			}

		case 1231:
			tld.IsCidFont = true
			tld.CidFontOperators.Version = int(cmd.args[0])
		case 1232:
			tld.IsCidFont = true
			tld.CidFontOperators.Revision = int(cmd.args[0])
		case 1233:
			tld.IsCidFont = true
			tld.CidFontOperators.Type = int(cmd.args[0])
		case 1234:
			tld.IsCidFont = true
			tld.CidFontOperators.Count = int(cmd.args[0])
		case 1235:
			tld.IsCidFont = true
			tld.CidFontOperators.UIDBase = cmd.args[0]
		case 1236:
			tld.IsCidFont = true
			tld.CidFontOperators.FontDictionaryArray = int(cmd.args[0])
		case 1237:
			tld.IsCidFont = true
			tld.CidFontOperators.FontDictionarySelect = int(cmd.args[0])
		case 1238:
			tld.IsCidFont = true
			tld.CidFontOperators.FontName = strIndex.GetName(int(cmd.args[0]))
		}
	}
	return nil
}

func deltaFloats(input []float64) []float64 {
	deltas := make([]float64, len(input))
	deltas[0] = input[0]
	for i := 1; i < len(input); i++ {
		deltas[i] = deltas[i-1] + input[i]
	}
	return deltas
}

func deltaInts(input []float64) []int {
	deltas := make([]int, len(input))
	deltas[0] = int(input[0])
	for i := 1; i < len(input); i++ {
		deltas[i] = deltas[i-1] + int(input[i])
	}
	return deltas
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
