package cff

import (
	"fmt"

	"github.com/bryanmatteson/gfx/font/cff/charsets"
	"github.com/bryanmatteson/gfx/font/encoding"
)

type table [][]byte

func Parse(b []byte) (*Collection, error) {
	reader := newReader(b)

	header, err := readHeader(reader)
	if err != nil {
		return nil, err
	}

	fontNames, err := reader.strindex()
	if err != nil {
		return nil, err
	}

	tldTable, err := reader.table()
	if err != nil {
		return nil, err
	}

	stringTable, err := reader.strindex()
	if err != nil {
		return nil, err
	}

	gsubTable, err := reader.table()
	if err != nil {
		return nil, err
	}

	gsubs, err := parseSubroutines(gsubTable, psType2Context)
	if err != nil {
		return nil, err
	}

	parser := newfontparser(reader, stringTable, gsubs)

	fonts := make(map[string]Font, len(fontNames))
	var firstFont Font
	for i, name := range fontNames {
		f, err := parser.parse(name, tldTable[i])
		if err != nil {
			return nil, err
		}
		if len(fonts) == 0 {
			firstFont = f
		}
		fonts[name] = f
	}

	return &Collection{
		Header:    header,
		Fonts:     fonts,
		FirstFont: firstFont,
	}, nil
}

type fontparser struct {
	r          *reader
	strtab     strtable
	globalsubs []pscmdseq
}

func newfontparser(reader *reader, strtab strtable, globalsubs []pscmdseq) *fontparser {
	return &fontparser{
		r:          reader,
		strtab:     strtab,
		globalsubs: globalsubs,
	}
}

func (p *fontparser) parse(name string, data []byte) (Font, error) {
	tld, priv, err := parseDictionaries(p.r, data, p.strtab)
	if err != nil {
		return nil, err
	}

	var subroutines []pscmdseq
	if priv.LocalSubroutineOffset != 0 && tld.PrivateDictOffset != 0 {
		if err = p.r.seek(priv.LocalSubroutineOffset + tld.PrivateDictOffset); err != nil {
			return nil, err
		}
		subs, err := p.r.table()
		if err != nil {
			return nil, err
		}
		subroutines, err = parseSubroutines(subs, psType2Context)
		if err != nil {
			return nil, err
		}
	}

	if tld.CharStringsOffset < 0 {
		return nil, fmt.Errorf("no char strings offset")
	}

	if err = p.r.seek(tld.CharStringsOffset); err != nil {
		return nil, err
	}

	charStrings, err := p.r.table()
	if err != nil {
		return nil, err
	}

	charset, err := p.getCharset(tld.CharSetOffset, tld.IsCidFont, charStrings)
	if err != nil {
		return nil, err
	}

	if tld.IsCidFont {
		return p.parsecid(tld, priv, charStrings, charset, subroutines)
	}

	encoding, err := p.getEncoding(tld.EncodingOffset, charset)
	if err != nil {
		return nil, err
	}

	selector := newselector(&fontdict{tld: tld, private: priv, subroutines: subroutines})
	subselector := newsubselector(p.globalsubs, subroutines, selector)

	if err = p.r.seek(tld.CharStringsOffset); err != nil {
		return nil, err
	}

	cs, err := p.getCharStrings(tld.CharStringType, charStrings, subselector, charset)
	if err != nil {
		return nil, err
	}

	return &font{
		tld:         tld,
		priv:        priv,
		charset:     charset,
		charstrings: cs,
		encoding:    encoding,
	}, nil
}

func (p *fontparser) parsecid(tld *TopLevelDictionary, priv *PrivateDictionary, charStrings table, charset charsets.Charset, localSubs []pscmdseq) (Font, error) {
	glyphCount := len(charStrings)

	offset := tld.CidFontOperators.FontDictionaryArray
	if err := p.r.seek(offset); err != nil {
		return nil, err
	}

	fontDict, err := p.r.table()
	if err != nil {
		return nil, err
	}

	dictionaries := make([]*fontdict, len(fontDict))
	for i, index := range fontDict {
		fontTld, fontPriv, err := parseDictionaries(p.r, index, p.strtab)
		if err != nil {
			return nil, err
		}

		if fontPriv == nil {
			return nil, fmt.Errorf("no private dictionary")
		}

		var subroutines []pscmdseq
		if fontPriv.LocalSubroutineOffset >= 0 && fontTld.PrivateDictOffset >= 0 {
			if err = p.r.seek(fontPriv.LocalSubroutineOffset + fontTld.PrivateDictOffset); err != nil {
				return nil, err
			}
			subs, err := p.r.table()
			if err != nil {
				return nil, err
			}
			subroutines, err = parseSubroutines(subs, psType2Context)
			if err != nil {
				return nil, err
			}
		}

		dictionaries[i] = &fontdict{tld: fontTld, private: fontPriv, subroutines: subroutines}
	}

	if err = p.r.seek(tld.CidFontOperators.FontDictionarySelect); err != nil {
		return nil, err
	}

	fdsel, err := p.getFontDictionarySelect(glyphCount)
	if err != nil {
		return nil, err
	}

	selector := newcidselector(dictionaries, fdsel)
	subselector := newsubselector(p.globalsubs, localSubs, selector)

	if err = p.r.seek(tld.CharStringsOffset); err != nil {
		return nil, err
	}

	cs, err := p.getCharStrings(tld.CharStringType, charStrings, subselector, charset)
	if err != nil {
		return nil, err
	}

	return &cidfont{
		selector: selector,
		font: &font{
			tld:         tld,
			priv:        priv,
			charset:     charset,
			charstrings: cs,
			encoding:    nil,
		},
	}, nil
}

func (p *fontparser) getCharStrings(typ CharStringType, data table, selector SubroutineSelector, charset charsets.Charset) (cs *t2charstrings, err error) {
	switch typ {
	case Type1:
		return nil, fmt.Errorf("type1 charstrings are not supported")
	case Type2:
		return t2parse(data, selector, charset)
	default:
		return nil, fmt.Errorf("unexpected char strings type: %v", typ)
	}
}

func (p *fontparser) getFontDictionarySelect(glyphCount int) (sel FontDictionarySelect, err error) {
	format, err := p.r.card8()
	if err != nil {
		return nil, err
	}

	switch format {
	case 0:
		fds := make([]int, glyphCount)
		for i := 0; i < glyphCount; i++ {
			fds[i], err = p.r.card8()
			if err != nil {
				return nil, err
			}
		}

		sel = &format0FdSelect{fds: fds}
	case 3:
		rc, err := p.r.card16()
		if err != nil {
			return nil, err
		}

		if p.r.rem() < 3*rc+2 {
			return sel, errInvalidCFFTable
		}

		ranges := make([]range3, rc)
		for i := 0; i < rc; i++ {
			first, _ := p.r.card16()
			d, _ := p.r.card8()
			ranges[i] = range3{first, d}
		}
		sentinel, _ := p.r.card16()
		sel = &format1FdSelect{ranges: ranges, sentinel: sentinel}

	default:
		return nil, fmt.Errorf("invalid fd select format: %d", format)
	}
	return
}

func (p *fontparser) getCharset(offID int, cid bool, charStrings table) (charset charsets.Charset, err error) {
	if offID < 0 {
		if !cid {
			charset = charsets.ISOAdobe
		}
		return
	}

	switch {
	case !cid && offID == 0:
		charset = charsets.ISOAdobe
	case !cid && offID == 1:
		charset = charsets.Expert
	case !cid && offID == 2:
		charset = charsets.ExpertSubset
	default:
		if err = p.r.seek(offID); err != nil {
			return
		}
		charset, err = parseCharset(p.r, charStrings, p.strtab)
	}
	return
}

func (p *fontparser) getEncoding(offID int, charset charsets.Charset) (e encoding.Encoding, err error) {
	switch offID {
	case 0:
		e = StandardEncoding
	case 1:
		e = ExpertEncoding
	case -1:
	default:
		if err = p.r.seek(offID); err != nil {
			return nil, err
		}
		e, err = parseEncoding(p.r, charset, p.strtab)
	}
	return
}

func parseDictionaries(reader *reader, data []byte, strtab strtable) (tld *TopLevelDictionary, priv *PrivateDictionary, err error) {
	tld = newTopLevelDictionary()
	if err = tld.init(data, strtab); err != nil {
		return nil, nil, err
	}

	priv = newPrivateDictionary()
	if tld.PrivateDictSize > 0 {
		pdata, err := reader.slice(tld.PrivateDictOffset, tld.PrivateDictSize)
		if err != nil {
			return nil, nil, err
		}

		if err = priv.init(pdata); err != nil {
			return nil, nil, err
		}
	}
	return tld, priv, nil
}

func parseEncoding(reader *reader, charset charsets.Charset, strIndex strtable) (encoding.Encoding, error) {
	format, err := reader.byte()
	if err != nil {
		return nil, err
	}

	baseFormat := format & 0x7f
	hasSupplements := (format & 0x80) != 0
	entries := make([]encoding.Entry, 0)

	n, err := reader.card8()
	if err != nil {
		return nil, err
	}

	ename := ""

	switch baseFormat {
	case 0:
		for i := 1; i < int(n); i++ {
			code, err := reader.card8()
			if err != nil {
				return nil, err
			}
			sid := charset.GetStringIDByGlyphID(i)
			name := strIndex.GetName(sid)
			entries = append(entries, encoding.Entry{Code: code, Name: name})
		}
		ename = "Format0"
	case 1:
		gid := 1
		for i := 1; i < int(n); i++ {
			rf, err := reader.card8()
			if err != nil {
				return nil, err
			}
			rl, err := reader.card8()
			if err != nil {
				return nil, err
			}

			for j := 0; j < 1+int(rl); j++ {
				sid := charset.GetStringIDByGlyphID(gid)
				code := int(rf) + j
				name := strIndex.GetName(sid)
				entries = append(entries, encoding.Entry{Code: code, Name: name})
				gid++
			}
		}
		ename = "Format1"
	default:
		return nil, fmt.Errorf("invalid encoding format %d", format)
	}

	if hasSupplements {
		supplements, err := parseSupplements(reader, strIndex)
		if err != nil {
			return nil, err
		}

		for _, s := range supplements {
			entries = append(entries, encoding.Entry{Code: s.code, Name: s.name})
		}
	}

	return encoding.NewEncoding(ename, entries), nil
}

func parseSupplements(reader *reader, strIndex strtable) (s []supplement, err error) {
	n, err := reader.card8()
	if err != nil {
		return nil, err
	}
	s = make([]supplement, n)
	for i := 0; i < int(n); i++ {
		code, err := reader.card8()
		if err != nil {
			return nil, err
		}
		sid, err := reader.sid()
		if err != nil {
			return nil, err
		}
		name := strIndex.GetName(sid)
		s[i] = supplement{int(code), sid, name}
	}
	return
}

func parseCharset(reader *reader, charStringIndex table, strIndex strtable) (charsets.Charset, error) {
	format, err := reader.byte()
	if err != nil {
		return nil, err
	}

	charmap := make([]charsets.Entry, 0, len(charStringIndex))
	charmap = append(charmap, charsets.Entry{Code: 0, Name: ".notdef"})

	switch format {
	case 0:
		for gid := 1; gid < len(charStringIndex); gid++ {
			sid, err := reader.sid()
			if err != nil {
				return nil, err
			}
			charmap = append(charmap, charsets.Entry{Code: sid, Name: strIndex.GetName(sid)})
		}
	case 1, 2:
		for gid := 1; gid < len(charStringIndex); gid++ {
			fsid, err := reader.sid()
			if err != nil {
				return nil, err
			}
			nir := 0
			if format == 1 {
				n, err := reader.card8()
				if err != nil {
					return nil, err
				}
				nir = int(n)
			} else {
				n, err := reader.card16()
				if err != nil {
					return nil, err
				}
				nir = int(n)
			}

			if err != nil {
				return nil, err
			}
			charmap = append(charmap, charsets.Entry{Code: fsid, Name: strIndex.GetName(fsid)})
			for i := 0; i < int(nir); i++ {
				gid++
				sid := fsid + i + 1
				charmap = append(charmap, charsets.Entry{Code: sid, Name: strIndex.GetName(sid)})
			}
		}
	}

	return charsets.NewCharset(charmap), err
}
