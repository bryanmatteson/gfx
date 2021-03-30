package enc

type Encoding interface {
	EncodingName() string
	HasName(name string) bool
	HasCode(code int) bool
	GetName(code int) string
}

type BaseEncoding struct {
	CodeToName map[int]string
	NameToCode map[string]int
}

func NewBase() *BaseEncoding {
	return &BaseEncoding{
		CodeToName: make(map[int]string),
		NameToCode: make(map[string]int),
	}
}

func (e *BaseEncoding) EncodingName() string { return "" }

func (e *BaseEncoding) HasName(name string) bool {
	_, ok := e.NameToCode[name]
	return ok
}

func (e *BaseEncoding) HasCode(code int) bool {
	_, ok := e.CodeToName[code]
	return ok
}

func (e *BaseEncoding) GetName(code int) string {
	name, ok := e.CodeToName[code]
	if !ok {
		return ".notdef"
	}
	return name
}

func (e *BaseEncoding) Add(code int, name string) {
	e.CodeToName[code] = name
	if _, ok := e.NameToCode[name]; !ok {
		e.NameToCode[name] = code
	}
}
