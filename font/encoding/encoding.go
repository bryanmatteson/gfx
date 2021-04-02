package encoding

type Encoding interface {
	EncodingName() string
	GetCharacterCode(name string) (int, bool)
	GetGlyphName(code int) (string, bool)
}

type Entry struct {
	Code int
	Name string
}

type encoding struct {
	name string
	ctn  map[int]string
	ntc  map[string]int
}

func NewEncoding(name string, entries []Entry) Encoding {
	ctn := make(map[int]string)
	ntc := make(map[string]int)

	for _, pair := range entries {
		ctn[pair.Code] = pair.Name
		ntc[pair.Name] = pair.Code
	}

	e := &encoding{
		name: name,
		ctn:  ctn,
		ntc:  ntc,
	}
	return e
}

func (e *encoding) EncodingName() string { return e.name }
func (e *encoding) GetCharacterCode(name string) (code int, ok bool) {
	code, ok = e.ntc[name]
	return
}
func (e *encoding) GetGlyphName(code int) (name string, ok bool) {
	name, ok = e.ctn[code]
	if !ok {
		name = ".notdef"
	}
	return
}

var MacExpert = NewEncoding("MacExpertEncoding", macExpertTable)
var MacRoman = NewEncoding("MacRomanEncoding", macRomanTable)
var MacOsRoman = NewEncoding("MacOsRomanEncoding", append(macRomanTable, macOsRomanTable...))
var Standard = NewEncoding("StandardEncoding", standardTable)
var Symbol = NewEncoding("SymbolEncoding", symbolTable)
var WinAnsi = NewEncoding("WinAnsiEncoding", winAnsiTable)
var ZapfDingbats = NewEncoding("ZapfDingbatsEncoding", zapfDingbatsTable)
