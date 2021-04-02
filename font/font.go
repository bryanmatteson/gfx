package font

import (
	"errors"
	"sync"

	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
)

type Hinting = font.Hinting

const (
	HintingNone     = font.HintingNone
	HintingVertical = font.HintingVertical
	HintingFull     = font.HintingFull
)

type FontStyle byte

// Styles
const (
	FontStyleNormal FontStyle = iota
	FontStyleBold   FontStyle = 1 << iota
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

func (f FontData) IsBold() bool   { return f.Style&FontStyleBold == FontStyleBold }
func (f FontData) IsItalic() bool { return f.Style&FontStyleItalic == FontStyleItalic }

type FontCache interface {
	Load(FontData) (*truetype.Font, error)
	Store(FontData, []byte) error
	Has(FontData) bool
}

type defaultFontCache struct {
	sync.RWMutex
	fonts map[string]*truetype.Font
}

func newDefaultFontCache() *defaultFontCache {
	return &defaultFontCache{
		fonts: make(map[string]*truetype.Font),
	}
}

func fontKeyName(fontData FontData) string {
	fontFileName := fontData.Name
	switch fontData.Family {
	case FontFamilySans:
		fontFileName += "s"
	case FontFamilySerif:
		fontFileName += "r"
	case FontFamilyMono:
		fontFileName += "m"
	}
	if fontData.Style&FontStyleBold != 0 {
		fontFileName += "b"
	} else {
		fontFileName += "r"
	}

	if fontData.Style&FontStyleItalic != 0 {
		fontFileName += "i"
	}
	fontFileName += ".ttf"
	return fontFileName
}

// Load a font from cache if exists otherwise it will load the font from file
func (cache *defaultFontCache) Load(fontData FontData) (*truetype.Font, error) {
	cache.RLock()
	font := cache.fonts[fontKeyName(fontData)]
	cache.RUnlock()

	if font != nil {
		return font, nil
	}

	return nil, errors.New("no font found")
}

// Store a font to this cache
func (cache *defaultFontCache) Store(fontData FontData, data []byte) error {
	font, err := truetype.Parse(data)
	if err != nil {
		return err
	}

	cache.Lock()
	cache.fonts[fontKeyName(fontData)] = font
	cache.Unlock()
	return nil
}

// Store a font to this cache
func (cache *defaultFontCache) Has(fontData FontData) bool {
	cache.Lock()
	font, ok := cache.fonts[fontKeyName(fontData)]
	cache.Unlock()
	return ok && font != nil
}

func GetGlobalFontCache() FontCache {
	return fontCache
}

var (
	defaultFonts           = newDefaultFontCache()
	fontCache    FontCache = defaultFonts
)
