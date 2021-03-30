package afm

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strconv"
	"unicode"
)

type scanner struct {
	data io.RuneReader
	buf  *bytes.Buffer
	ch   rune
}

func newScanner(r io.Reader) *scanner {
	p := &scanner{data: bufio.NewReader(r), buf: bytes.NewBuffer(make([]byte, 0, 1024))}
	p.advance()
	return p
}

func (s *scanner) advance() (result rune, err error) {
	r, _, e := s.data.ReadRune()
	result = s.ch

	if e != nil {
		if e != io.EOF {
			return 0, e
		}
		s.ch = -1
	} else {
		s.ch = r
	}

	return
}

func (s *scanner) eof() bool { return s.ch == -1 }

func (s *scanner) eatSpace() {
	for !s.eof() && unicode.IsSpace(s.ch) {
		s.advance()
	}
}

func (s *scanner) until(fn func(rune) bool) (string, error) {
	var r rune
	var err error

	s.buf.Reset()
	for !s.eof() && !fn(s.ch) {
		r, err = s.advance()
		if err != nil {
			break
		}
		s.buf.WriteRune(r)
	}

	return s.buf.String(), err
}

func (s *scanner) word() (string, error) {
	s.eatSpace()
	if s.eof() {
		return EndFontMetrics, nil
	}

	return s.until(unicode.IsSpace)
}

func (s *scanner) int() (int, error) {
	word, err := s.word()
	if err != nil {
		return 0, err
	}

	val, err := strconv.ParseInt(word, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("expected int, got %s", word)
	}

	return int(val), nil
}

func (s *scanner) float() (float64, error) {
	word, err := s.word()
	if err != nil {
		return 0, err
	}

	val, err := strconv.ParseFloat(word, 32)
	if err != nil {
		return 0, fmt.Errorf("expected float, got %s", word)
	}

	return float64(val), nil
}

func (s *scanner) bool() (bool, error) {
	word, err := s.word()
	if err != nil {
		return false, err
	}

	switch word {
	case "true":
		return true, nil
	case "false":
		return false, nil
	default:
		return false, fmt.Errorf("expected bool, got %s", word)
	}
}

func crlf(ch rune) bool { return ch == '\n' || ch == '\r' }

func (s *scanner) line() (string, error) {
	s.eatSpace()
	return s.until(crlf)
}
