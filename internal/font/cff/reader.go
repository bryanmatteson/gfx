package cff

import (
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
)

type reader struct {
	b   []byte
	pos int
}

func newReader(b []byte) *reader {
	return &reader{b: b}
}

// func (r *reader) data() []byte { return r.b }
func (r *reader) slice(off, n int) ([]byte, error) {
	if off+n > len(r.b) {
		return nil, fmt.Errorf("read past end of buffer")
	}

	return r.b[off : off+n], nil
}

func (r *reader) eof() bool {
	return r.pos >= len(r.b)
}

func (r *reader) seek(off int) error {
	if off > len(r.b) {
		return fmt.Errorf("seek out of bounds")
	}
	r.pos = off
	return nil
}

func (r *reader) string(n int) (string, error) {
	b, err := r.bytes(n)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

func (r *reader) decodeString(n int, decoder *encoding.Decoder) (string, error) {
	s, err := r.string(n)
	if err != nil {
		return "", err
	}

	s, err = decoder.String(s)
	if err != nil {
		return "", err
	}

	return s, nil
}

func (r *reader) card8() (int, error) {
	b, err := r.byte()
	if err != nil {
		return 0, err
	}
	return int(b), nil
}

func (r *reader) card16() (int, error) {
	b, err := r.bytes(2)
	if err != nil {
		return 0, err
	}

	return int(binary.BigEndian.Uint16(b)), nil
}

func (r *reader) card32() (int, error) {
	b, err := r.bytes(4)
	if err != nil {
		return 0, err
	}

	return int(binary.BigEndian.Uint32(b)), nil
}

func (r *reader) sid() (int, error) {
	return r.card16()
}

func (r *reader) offsize() (int, error) {
	b, err := r.byte()
	if err != nil {
		return 0, err
	}
	return int(b), nil
}

func (r *reader) offset(n int) (val uint64, err error) {
	b, err := r.bytes(n)
	if err != nil {
		return 0, err
	}

	for _, v := range b {
		val = (val << 8) | uint64(v)
	}

	return
}

func (r *reader) bytes(n int) (b []byte, err error) {
	if r.pos+n > len(r.b) {
		return nil, fmt.Errorf("read past end of buffer")
	}

	b = r.b[r.pos : r.pos+n]
	r.pos += n
	return b, nil
}

func (r *reader) byte() (byte, error) {
	if r.eof() {
		return 0, fmt.Errorf("read past end of buffer")
	}
	b := r.b[r.pos]
	r.pos++
	return b, nil
}

// func (r *reader) peek() (byte, error) {
// 	if r.eof() {
// 		return 0, io.EOF
// 	}

// 	return r.b[r.pos], nil
// }

// func (r *reader) tag() (string, error) {
// 	return r.decodeString(4, charmap.ISO8859_1.NewDecoder())
// }

func (r *reader) strindex() (results []string, err error) {
	idx, err := r.index()
	if err != nil {
		return nil, err
	}

	if len(idx) == 0 {
		return nil, nil
	}

	count := len(idx) - 1

	results = make([]string, count)
	for i := 0; i < count; i++ {
		length := idx[i+1] - idx[i]
		if length < 0 {
			return nil, fmt.Errorf("negative object length %d at %d: pos=%d", length, i, r.pos)
		}

		s, err := r.decodeString(length, charmap.ISO8859_1.NewDecoder())
		if err != nil {
			return nil, err
		}
		results[i] = s
	}

	return
}

func (r *reader) index() (results []int, err error) {
	count, err := r.card16()
	if err != nil {
		return nil, err
	}

	if count == 0 {
		return nil, nil
	}

	offsetSize, err := r.offsize()
	if err != nil {
		return nil, err
	}

	results = make([]int, count+1)
	for i := 0; i < len(results); i++ {
		off, err := r.offset(offsetSize)
		if err != nil {
			return nil, err
		}
		results[i] = int(off)
	}

	return
}

func (r *reader) table() (results [][]byte, err error) {
	idx, err := r.index()
	if err != nil {
		return nil, err
	}

	if len(idx) == 0 {
		return nil, nil
	}

	count := len(idx) - 1

	results = make([][]byte, count)
	for i := 0; i < count; i++ {
		length := idx[i+1] - idx[i]
		if length < 0 {
			return nil, fmt.Errorf("negative object length %d at %d: pos=%d", length, i, r.pos)
		}

		b, err := r.bytes(length)
		if err != nil {
			return nil, err
		}
		results[i] = b
	}
	return
}

func (r *reader) real() (float64, error) {
	var builder strings.Builder
	var done, exponentMissing, hasExponent bool

	for !done {
		b, err := r.byte()
		if err != nil {
			return 0, err
		}
		var nibbles = []byte{b / 16, b % 16}

		for i := 0; i < 2; i++ {
			switch nibbles[i] {
			case 0, 1, 2, 3, 4, 5, 6, 7, 8, 9:
				builder.WriteString(fmt.Sprintf("%d", nibbles[i]))
			case 10:
				builder.WriteByte('.')
			case 11:
				if hasExponent {
					break
				}
				builder.WriteByte('E')
				exponentMissing = true
				hasExponent = true
			case 12:
				if hasExponent {
					break
				}
				builder.WriteString("E-")
				exponentMissing = true
				hasExponent = true
			case 13:
			case 14:
				builder.WriteByte('-')
			case 15:
				done = true
			default:
				return 0, fmt.Errorf("invalid nibble value %d", nibbles[i])
			}
		}
	}

	if exponentMissing {
		builder.WriteByte('0')
	}

	if builder.Len() == 0 {
		return 0, nil
	}

	return strconv.ParseFloat(builder.String(), 64)
}
