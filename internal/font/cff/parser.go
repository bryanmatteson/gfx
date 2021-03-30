package cff

import (
	"fmt"

	"go.matteson.dev/gfx/internal/font/enc"
)

type table [][]byte

func Parse(b []byte) (*Collection, error) {
	reader := newReader(b)

	// tag, err := reader.tag()
	// if err != nil {
	// 	return
	// }
	// _ = tag

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

	globalSubroutines, err := reader.table()
	if err != nil {
		return nil, err
	}

	fonts := make(map[string]*Font)
	var firstFont *Font
	for i, name := range fontNames {
		f, err := parseFont(reader, name, tldTable[i], stringTable, globalSubroutines)
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

func parseFont(reader *reader, name string, dictionaryData []byte, stringTable strtable, globalSubroutines table) (font *Font, err error) {
	tld := newTopLevelDictionary()
	if err = readDictionary(newReader(dictionaryData), tld, stringTable); err != nil {
		return nil, err
	}

	if tld.CharStringsOffset < 0 {
		return nil, fmt.Errorf("no char strings offset")
	}

	privateDict := newPrivateDictionary()
	if tld.privateDictSize > 0 {
		pdata, err := reader.slice(tld.privateDictOffset, tld.privateDictSize)
		if err != nil {
			return nil, err
		}

		if err = readDictionary(newReader(pdata), privateDict, stringTable); err != nil {
			return nil, err
		}
	}

	var localSubroutines table
	if privateDict.LocalSubroutineOffset >= 0 && tld.privateDictOffset >= 0 {
		if err = reader.seek(privateDict.LocalSubroutineOffset + tld.privateDictOffset); err != nil {
			return nil, err
		}
		localSubroutines, err = reader.table()
		if err != nil {
			return nil, err
		}
	}

	if err = reader.seek(tld.CharStringsOffset); err != nil {
		return nil, err
	}

	charStrings, err := reader.table()
	if err != nil {
		return nil, err
	}

	var charset Charset
	if tld.CharSetOffset >= 0 {
		charsetID := tld.CharSetOffset
		switch {
		case !tld.IsCidFont && charsetID == 0:
			charset = ISOAdobeCharset()
		case !tld.IsCidFont && charsetID == 1:
			charset = ExpertCharset()
		case !tld.IsCidFont && charsetID == 2:
			charset = ExpertSubsetCharset()
		default:
			charset, err = parseCharset(reader, tld, charStrings, stringTable)
			if err != nil {
				return nil, err
			}
		}
	} else if !tld.IsCidFont {
		charset = ISOAdobeCharset()
	}

	if tld.IsCidFont {
		return parseCIDFont(reader, tld, charStrings, stringTable, privateDict, charset, globalSubroutines, localSubroutines)
	}

	var fontEncoding Encoding
	switch tld.EncodingOffset {
	case 0:
		fontEncoding = NewStandardEncoding()
	case 1:
		fontEncoding = NewExpertEncoding()
	case -1:
	default:
		if err = reader.seek(tld.EncodingOffset); err != nil {
			return nil, err
		}

		fontEncoding, err = parseEncoding(reader, charset, stringTable)
		if err != nil {
			return nil, err
		}
	}
	_ = fontEncoding

	selector := &subroutineselector{global: globalSubroutines, local: localSubroutines, cid: false}
	if err := reader.seek(tld.CharStringsOffset); err != nil {
		return nil, err
	}

	switch tld.CharStringType {
	case Type1:
		return nil, fmt.Errorf("type1 charstrings are not supported")
	case Type2:
	default:
		return nil, fmt.Errorf("unexpected char strings type: %v", tld.CharStringType)
	}
	_ = selector

	return
}

func parseCIDFont(reader *reader, tld *TopLevelDictionary, charStrings table, strIndex strtable, priv *PrivateDictionary, charset Charset, globalSubroutines table, localSubroutines table) (font *Font, err error) {
	glyphCount := len(charStrings)

	offset := tld.CidFontOperators.FontDictionaryArray
	if err := reader.seek(offset); err != nil {
		return nil, err
	}

	fontDict, err := reader.table()
	if err != nil {
		return nil, err
	}

	var privateDictionaries []*PrivateDictionary
	var fontDictionaries []*TopLevelDictionary
	var fontLocalSubroutines []table

	for _, index := range fontDict {
		tldcid := newTopLevelDictionary()
		if err = readDictionary(newReader(index), tldcid, strIndex); err != nil {
			return nil, err
		}

		if tldcid.privateDictSize <= 0 {
			return nil, fmt.Errorf("no private dict")
		}

		pdbytes, err := reader.slice(tldcid.privateDictOffset, tldcid.privateDictSize)
		if err != nil {
			return nil, err
		}

		pdcid := newPrivateDictionary()
		if err = readDictionary(newReader(pdbytes), pdcid, strIndex); err != nil {
			return nil, err
		}

		var subroutines table
		if pdcid.LocalSubroutineOffset > 0 {
			if err = reader.seek(tldcid.privateDictOffset + pdcid.LocalSubroutineOffset); err != nil {
				return nil, err
			}
			subroutines, err = reader.table()
			if err != nil {
				return nil, err
			}
		}

		fontLocalSubroutines = append(fontLocalSubroutines, subroutines)
		fontDictionaries = append(fontDictionaries, tldcid)
		privateDictionaries = append(privateDictionaries, pdcid)
	}

	if err = reader.seek(tld.CidFontOperators.FontDictionarySelect); err != nil {
		return nil, err
	}
	format, err := reader.card8()
	if err != nil {
		return nil, err
	}

	var fdselect FDSelect
	switch format {
	case 0:
		fds := make([]int, glyphCount)
		for i := 0; i < glyphCount; i++ {
			fds[i], err = reader.card8()
			if err != nil {
				return nil, err
			}
		}
		fdselect = &fmtzerofdselect{fds: fds, ros: tld.CidFontOperators.Ros}
	case 3:
		rc, err := reader.card16()
		if err != nil {
			return nil, err
		}
		ranges := make([]range3, rc)
		for i := 0; i < rc; i++ {
			first, err := reader.card16()
			if err != nil {
				return nil, err
			}
			d, err := reader.card8()
			if err != nil {
				return nil, err
			}
			ranges[i] = range3{first, d}
		}
		sentinel, err := reader.card16()
		if err != nil {
			return nil, err
		}
		fdselect = &fmtonefdselect{ranges: ranges, ros: tld.CidFontOperators.Ros, sentinel: sentinel}
	default:
		return nil, fmt.Errorf("invalid fd select format: %d", format)
	}

	selector := &subroutineselector{globalSubroutines, localSubroutines, true, fdselect, fontLocalSubroutines}

	if err = reader.seek(tld.CharStringsOffset); err != nil {
		return nil, err
	}
	switch tld.CharStringType {
	case Type1:
		return nil, fmt.Errorf("type1 charstrings are not supported")
	case Type2:
	default:
		return nil, fmt.Errorf("unexpected char strings type: %v", tld.CharStringType)
	}
	_ = selector

	return
}

func parseEncoding(reader *reader, charset Charset, strIndex strtable) (Encoding, error) {
	format, err := reader.byte()
	if err != nil {
		return nil, err
	}

	baseFormat := format & 0x7f
	hasSupplements := (format & 0x80) != 0
	encoding := &baseencoding{BaseEncoding: enc.NewBase()}

	n, err := reader.card8()
	if err != nil {
		return nil, err
	}

	switch baseFormat {
	case 0:
		for i := 1; i < int(n); i++ {
			code, err := reader.card8()
			if err != nil {
				return nil, err
			}
			sid := charset.GetStringIdByGlyphId(i)
			name := strIndex.GetName(sid)
			encoding.Add(int(code), sid, name)
		}
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
				sid := charset.GetStringIdByGlyphId(gid)
				code := int(rf) + j
				name := strIndex.GetName(sid)
				encoding.Add(code, sid, name)
				gid++
			}
		}
	}

	if hasSupplements {
		encoding.supplements, err = parseSupplements(reader, strIndex)
		if err != nil {
			return nil, err
		}
	}

	return encoding, nil
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

func parseCharset(reader *reader, tld *TopLevelDictionary, charStringIndex table, strIndex strtable) (Charset, error) {
	if err := reader.seek(tld.CharSetOffset); err != nil {
		return nil, err
	}
	format, err := reader.byte()
	if err != nil {
		return nil, err
	}

	charmap := make([]charmapentry, 0, len(charStringIndex))
	charmap = append(charmap, charmapentry{0, ".notdef"})

	switch format {
	case 0:
		for gid := 1; gid < len(charStringIndex); gid++ {
			sid, err := reader.sid()
			if err != nil {
				return nil, err
			}
			charmap = append(charmap, charmapentry{sid, strIndex.GetName(sid)})
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
			charmap = append(charmap, charmapentry{fsid, strIndex.GetName(fsid)})
			for i := 0; i < int(nir); i++ {
				gid++
				sid := fsid + i + 1
				charmap = append(charmap, charmapentry{sid, strIndex.GetName(sid)})
			}
		}
	}

	return &charset{false, charmap}, err
}
