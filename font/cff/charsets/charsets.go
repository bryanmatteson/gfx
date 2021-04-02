package charsets

type Charset interface {
	GetNameByGlyphID(gid int) string
	GetNameByStringID(sid int) string
	GetStringIDByGlyphID(gid int) int
	GetGlyphIDByName(name string) int
}

var Expert = NewCharset(expertCharmap)
var ExpertSubset = NewCharset(expertSubsetCharmap)
var ISOAdobe = NewCharset(isoAdobeCharmap)

type charset struct {
	charmap []Entry
}

func NewCharset(table []Entry) Charset {
	return &charset{table}
}

func (c *charset) GetNameByGlyphID(gid int) string {
	return c.charmap[gid].Name
}

func (c *charset) GetNameByStringID(sid int) string {
	for _, pair := range c.charmap {
		if pair.Code == sid {
			return pair.Name
		}
	}
	return ""
}

func (c *charset) GetStringIDByGlyphID(glyphID int) int {
	return c.charmap[glyphID].Code
}

func (c *charset) GetGlyphIDByName(characterName string) int {
	for gid, pair := range c.charmap {
		if pair.Name == characterName {
			return gid
		}
	}
	return 0
}

type Entry struct {
	Code int
	Name string
}
