package cff

import (
	"encoding/binary"
	"errors"
	"fmt"
	"strconv"
)

var (
	errInvalidCFFTable               = errors.New("invalid cff table")
	errUnsupportedRealNumberEncoding = errors.New("unsupported real number encoding")
)

const (
	maxRealNumberStrLen = 64
	maxNibbleDefsLength = len("E-")
)

type pscmd struct {
	id   int
	args []float64
	op   *psop
}

type pscmdseq []*pscmd

type psContext uint32

const (
	psTLDContext psContext = iota
	psPrivContext
	psType2Context
)

func parseSubroutines(tab table, ctx psContext) (res []pscmdseq, err error) {
	res = make([]pscmdseq, len(tab))
	for i := range tab {
		seq, err := parseCommandSequence(tab[i], ctx)
		if err != nil {
			return nil, err
		}
		res[i] = seq
	}
	return
}

func parseCommandSequence(data []byte, ctx psContext) (pscmdseq, error) {
	bo := binary.BigEndian
	operands := make([]float64, 0, 8)
	commands := make(pscmdseq, 0)
	var realbuf [maxRealNumberStrLen]byte

	for len(data) > 0 {
		var val float64
		var handled bool
		b := data[0]
		data = data[1:]

		switch {
		case b == 28:
			if len(data) < 2 {
				return nil, errInvalidCFFTable
			}
			val, handled = float64(int16(bo.Uint16(data[:2]))), true
			data = data[2:]

		case b == 29 && ctx != psType2Context:
			if len(data) < 4 {
				return nil, errInvalidCFFTable
			}
			val, handled = float64(bo.Uint32(data[:4])), true
			data = data[4:]

		case b == 30 && ctx != psType2Context:
			s := realbuf[:0]

		loop:
			for {
				if len(data) == 0 {
					return nil, errInvalidCFFTable
				}

				b := data[0]
				data = data[1:]

				for i := 0; i < 2; i++ {
					nib := b >> 4
					b = b << 4

					if nib == 0x0f {
						f, err := strconv.ParseFloat(string(s), 64)
						if err != nil {
							return nil, errInvalidCFFTable
						}
						val, handled = f, true
						break loop
					}

					if nib == 0x0d {
						return nil, errInvalidCFFTable
					}

					if len(s)+maxNibbleDefsLength > len(realbuf) {
						return nil, errUnsupportedRealNumberEncoding
					}
					s = append(s, nibbleDefs[nib]...)
				}
			}

		case b < 32:

		case b < 247:
			val, handled = float64(b-139), true

		case b < 251:
			if len(data) == 0 {
				return nil, errInvalidCFFTable
			}

			b1 := data[0]
			data = data[1:]
			val, handled = float64(+int32(b-247)*256+int32(b1)+108), true

		case b < 255:
			if len(data) == 0 {
				return nil, errInvalidCFFTable
			}

			b1 := data[0]
			val, handled = float64(-int32(b-251)*256-int32(b1)-108), true
			data = data[1:]

		case b == 255 && ctx == psType2Context:
			if len(data) < 4 {
				return nil, errInvalidCFFTable
			}

			lead := float64(int16(bo.Uint16(data[:2])))
			fractional := float64(int16(bo.Uint16(data[:2])))
			val, handled = lead+(fractional/65535.0), true
			data = data[4:]
		}

		if handled {
			operands = append(operands, val)
			continue
		}

		id := int(b)
		ops, oplen := psoperators[ctx][0], 1
		if b == 12 {
			if len(data) == 0 {
				return nil, errInvalidCFFTable
			}
			b, data = data[0], data[1:]
			ops, oplen = psoperators[ctx][1], 2
			id = 1200 + int(b)
		}

		if int(b) >= len(ops) {
			return nil, fmt.Errorf("unrecognized CFF %d-byte operator: %d", oplen, b)
		}

		var args []float64
		if len(operands) > 0 {
			args = make([]float64, len(operands))
			copy(args, operands)
		}

		commands = append(commands, &pscmd{id: id, args: args, op: &ops[b]})
		operands = operands[:0]
	}

	return commands, nil
}

var nibbleDefs = [16]string{
	0x00: "0",
	0x01: "1",
	0x02: "2",
	0x03: "3",
	0x04: "4",
	0x05: "5",
	0x06: "6",
	0x07: "7",
	0x08: "8",
	0x09: "9",
	0x0a: ".",
	0x0b: "E",
	0x0c: "E-",
	0x0d: "",
	0x0e: "-",
	0x0f: "",
}

type psop struct {
	stackeffect int32
	name        string
}

var psoperators = [...][2][]psop{
	psTLDContext: {{
		// 1-byte operators.
		0:  {+1, "version"},
		1:  {+1, "Notice"},
		2:  {+1, "FullName"},
		3:  {+1, "FamilyName"},
		4:  {+1, "Weight"},
		5:  {-1, "FontBBox"},
		13: {+1, "UniqueID"},
		14: {-1, "XUID"},
		15: {+1, "charset"},
		16: {+1, "Encoding"},
		17: {+1, "CharStrings"},
		18: {+2, "Private"},
	}, {
		// 2-byte operators. The first byte is the escape byte.
		0:  {+1, "Copyright"},
		1:  {+1, "isFixedPitch"},
		2:  {+1, "ItalicAngle"},
		3:  {+1, "UnderlinePosition"},
		4:  {+1, "UnderlineThickness"},
		5:  {+1, "PaintType"},
		6:  {+1, "CharstringType"},
		7:  {-1, "FontMatrix"},
		8:  {+1, "StrokeWidth"},
		20: {+1, "SyntheticBase"},
		21: {+1, "PostScript"},
		22: {+1, "BaseFontName"},
		23: {-2, "BaseFontBlend"},
		30: {+3, "ROS"},
		31: {+1, "CIDFontVersion"},
		32: {+1, "CIDFontRevision"},
		33: {+1, "CIDFontType"},
		34: {+1, "CIDCount"},
		35: {+1, "UIDBase"},
		36: {+1, "FDArray"},
		37: {+1, "FDSelect"},
		38: {+1, "FontName"},
	}},
	psPrivContext: {{
		// 1-byte operators.
		6:  {-2, "BlueValues"},
		7:  {-2, "OtherBlues"},
		8:  {-2, "FamilyBlues"},
		9:  {-2, "FamilyOtherBlues"},
		10: {+1, "StdHW"},
		11: {+1, "StdVW"},
		19: {+1, "Subrs"},
		20: {+1, "defaultWidthX"},
		21: {+1, "nominalWidthX"},
	}, {
		// 2-byte operators. The first byte is the escape byte.
		9:  {+1, "BlueScale"},
		10: {+1, "BlueShift"},
		11: {+1, "BlueFuzz"},
		12: {-2, "StemSnapH"},
		13: {-2, "StemSnapV"},
		14: {+1, "ForceBold"},
		17: {+1, "LanguageGroup"},
		18: {+1, "ExpansionFactor"},
		19: {+1, "initialRandomSeed"},
	}},
	psType2Context: {{
		// 1-byte operators.
		0:  {}, // Reserved.
		1:  {-1, "hstem"},
		2:  {}, // Reserved.
		3:  {-1, "vstem"},
		4:  {-1, "vmoveto"},
		5:  {-1, "rlineto"},
		6:  {-1, "hlineto"},
		7:  {-1, "vlineto"},
		8:  {-1, "rrcurveto"},
		9:  {}, // Reserved.
		10: {+1, "callsubr"},
		11: {+0, "return"},
		12: {}, // escape.
		13: {}, // Reserved.
		14: {-1, "endchar"},
		15: {}, // Reserved.
		16: {}, // Reserved.
		17: {}, // Reserved.
		18: {-1, "hstemhm"},
		19: {-1, "hintmask"},
		20: {-1, "cntrmask"},
		21: {-1, "rmoveto"},
		22: {-1, "hmoveto"},
		23: {-1, "vstemhm"},
		24: {-1, "rcurveline"},
		25: {-1, "rlinecurve"},
		26: {-1, "vvcurveto"},
		27: {-1, "hhcurveto"},
		28: {}, // shortint.
		29: {+1, "callgsubr"},
		30: {-1, "vhcurveto"},
		31: {-1, "hvcurveto"},
	}, {
		// 2-byte operators. The first byte is the escape byte.
		34: {+7, "hflex"},
		36: {+9, "hflex1"},
	}},
}
