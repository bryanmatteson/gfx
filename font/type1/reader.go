package type1

type reader struct {
	data []byte
	b    byte
	pos  int
}

func newreader(b []byte) *reader {
	return &reader{data: b}
}

func (r *reader) eof() bool {
	return r.pos >= len(r.data)
}

func (r *reader) seek(o int) bool {
	if o < 0 || o >= len(r.data) {
		return false
	}

	r.pos = o
	return true
}

func (r *reader) advance() bool {
	if r.pos >= len(r.data) {
		return false
	}

	r.b = r.data[r.pos]
	r.pos++
	return true
}

func (r *reader) current() byte {
	return r.b
}

func (r *reader) peek() byte {
	return r.data[r.pos]
}
