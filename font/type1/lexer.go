package type1

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

type toktype int

const (
	None toktype = iota
	String
	Name
	Literal
	Real
	Integer
	StartArray
	EndArray
	StartProc
	EndProc
	StartDict
	EndDict
	Charstring

	Error toktype = 10000
)

type token struct {
	v   []byte
	typ toktype
}

type lexer struct {
	rd         *reader
	tok        *token
	comments   []string
	commentbuf bytes.Buffer
	strbuf     bytes.Buffer
	litbuf     bytes.Buffer
	openparens int
	err        error
}

func newlexer(b []byte) *lexer {
	lex := &lexer{rd: newreader(b)}
	lex.next()
	return lex
}

func (lex *lexer) next() *token {
	prev := lex.tok
	for lex.rd.advance() {
		c := lex.rd.current()
		switch c {
		case '%':
			lex.comments = append(lex.comments, lex.comment())
		case '(':
			str, err := lex.string()
			if err != nil {
				lex.err = err
				return nil
			}
			lex.tok = &token{[]byte(str), String}
			return lex.tok
		case ')':
			lex.err = fmt.Errorf("encountered an end of string ')' outside of a string")
			return nil
		case '[':
			lex.tok = &token{[]byte{c}, StartArray}
			return lex.tok
		case ']':
			lex.tok = &token{[]byte{c}, EndArray}
			return lex.tok
		case '{':
			lex.tok = &token{[]byte{c}, StartProc}
			return lex.tok
		case '}':
			lex.tok = &token{[]byte{c}, EndProc}
			return lex.tok
		case '/':
			lex.rd.advance()
			lex.tok = &token{[]byte(lex.literal()), Literal}
			return lex.tok
		case '<':
			next := lex.rd.peek()
			if next == '<' {
				lex.rd.advance()
				lex.tok = &token{[]byte("<<"), StartDict}
				return lex.tok
			}
			lex.tok = &token{[]byte{c}, Name}
			return lex.tok
		case '>':
			next := lex.rd.peek()
			if next == '>' {
				lex.rd.advance()
				lex.tok = &token{[]byte(">>"), EndDict}
				return lex.tok
			}
			lex.tok = &token{[]byte{c}, Name}
			return lex.tok
		default:
			if unicode.IsSpace(rune(c)) || c == 0 {
				continue
			}

			if tok, ok := lex.trynumber(c); ok {
				lex.tok = tok
				return tok
			}

			name := lex.literal()
			if name == "" {
				lex.err = fmt.Errorf("invalid data")
				return nil
			}

			if strings.EqualFold(name, RdProcedure) {
				if prev.typ != Integer {
					lex.err = fmt.Errorf("expected integer token before %s", name)
					return nil
				}
				n, err := strconv.ParseInt(string(prev.v), 10, 32)
				if err != nil {
					lex.err = err
					return nil
				}

				lex.tok = lex.charstring(int(n))
				return lex.tok
			}

			lex.tok = &token{[]byte(name), Name}
			return lex.tok
		}
	}
	return nil
}

func (lex *lexer) charstring(n int) *token {
	lex.rd.advance()
	data := make([]byte, n)
	for i := 0; i < n && !lex.rd.eof(); i++ {
		lex.rd.advance()
		data[i] = lex.rd.current()
	}

	return &token{data, Charstring}
}

func (lex *lexer) trynumber(c byte) (*token, bool) {
	pos := lex.rd.pos

	var builder = &bytes.Buffer{}
	var radix *bytes.Buffer

	advance := func() {
		lex.rd.advance()
		c = lex.rd.current()
	}

	var hasDigit bool

	if c == '+' || c == '-' {
		builder.WriteByte(c)
		advance()
	}

	for unicode.IsDigit(rune(c)) {
		builder.WriteByte(c)
		advance()
		hasDigit = true
	}

	if c == '.' {
		builder.WriteByte(c)
		advance()
	} else if c == '#' {
		radix = builder
		builder = &bytes.Buffer{}
		advance()
	} else if builder.Len() == 0 || !hasDigit {
		lex.rd.seek(pos)
		return nil, false
	} else {
		lex.rd.seek(lex.rd.pos - 1)
		return &token{builder.Bytes(), Integer}, true
	}

	if unicode.IsDigit(rune(c)) {
		builder.WriteByte(c)
		advance()
	} else {
		lex.rd.seek(pos)
		return nil, false
	}

	for unicode.IsDigit(rune(c)) {
		builder.WriteByte(c)
		advance()
	}

	if c == 'E' || c == 'e' {
		builder.WriteByte(c)
		advance()

		if c == '-' {
			builder.WriteByte(c)
			advance()
		}

		if unicode.IsDigit(rune(c)) {
			builder.WriteByte(c)
			advance()
		} else {
			lex.rd.seek(pos)
			return nil, false
		}

		for unicode.IsDigit(rune(c)) {
			builder.WriteByte(c)
			advance()
		}
	}

	lex.rd.seek(lex.rd.pos - 1)
	if radix != nil {
		rdx, err := strconv.ParseInt(radix.String(), 10, 32)
		if err != nil {
			lex.err = err
			lex.rd.seek(pos)
			return nil, false
		}
		number, err := strconv.ParseInt(builder.String(), int(rdx), 32)
		if err != nil {
			lex.err = err
			lex.rd.seek(pos)
			return nil, false
		}
		return &token{[]byte(strconv.FormatInt(number, 10)), Integer}, true
	}

	return &token{builder.Bytes(), Real}, true
}

func (lex *lexer) literal() string {
	lex.litbuf.Reset()

	for !lex.rd.eof() {
		c := rune(lex.rd.current())
		if unicode.IsSpace(c) || c == '(' || c == ')' || c == '<' || c == '>' || c == '[' || c == ']' || c == '{' || c == '}' || c == '/' || c == '%' {
			break
		}
		lex.litbuf.WriteRune(c)
		lex.rd.advance()
	}

	return lex.litbuf.String()
}

func (lex *lexer) comment() string {
	lex.commentbuf.Reset()

	for lex.rd.advance() {
		c := lex.rd.current()
		if c == '\n' || c == '\r' {
			break
		}
		lex.commentbuf.WriteByte(c)
	}

	return lex.commentbuf.String()
}

func (lex *lexer) string() (string, error) {
	lex.strbuf.Reset()

	for lex.rd.advance() {
		c := lex.rd.current()
		switch c {
		case '(':
			lex.openparens++
			lex.strbuf.WriteByte(c)
		case ')':
			if lex.openparens == 0 {
				return lex.strbuf.String(), nil
			}
			lex.strbuf.WriteByte(c)
			lex.openparens--
		case '\\':
			lex.rd.advance()
			c1 := lex.rd.current()
			switch c1 {
			case 'n', 'r':
				lex.strbuf.WriteByte('\n')
			case 't':
				lex.strbuf.WriteByte('\t')
			case 'b':
				lex.strbuf.WriteByte('\b')
			case 'f':
				lex.strbuf.WriteByte('\f')
			case '\\':
				lex.strbuf.WriteByte('\\')
			case '(':
				lex.strbuf.WriteByte('(')
			case ')':
				lex.strbuf.WriteByte(')')
			}
			if unicode.IsDigit(rune(c1)) {
				lex.rd.advance()
				c2 := lex.rd.current()
				lex.rd.advance()
				c3 := lex.rd.current()
				code, err := strconv.ParseInt(string([]byte{c1, c2, c3}), 8, 8)
				if err != nil {
					return "", fmt.Errorf("invalid octal")
				}
				lex.strbuf.WriteByte(byte(code))
			}
		case '\r', '\n':
			lex.strbuf.WriteByte('\n')
		default:
			lex.strbuf.WriteByte(c)
		}
	}
	return "", nil
}
