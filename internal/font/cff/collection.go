package cff

type Header struct {
	// The major version of this font format. Starting at 1.
	MajorVersion int
	// The minor version of this font format. Starting at 0. Indicates extensions to the format which are undetectable by readers which do not support them.
	MinorVersion int
	// Indicates the size of this header in bytes so that future changes to the format may include extra data after the <see cref="OffsetSize"/> field.
	SizeInBytes int
	// Specifies the size of all offsets relative to the start of the data in the font.
	OffsetSize int
}

func readHeader(r *reader) (hdr *Header, err error) {
	hdr = &Header{}

	b, err := r.bytes(4)
	if err != nil {
		return hdr, err
	}

	hdr.MajorVersion = int(b[0])
	hdr.MinorVersion = int(b[1])
	hdr.SizeInBytes = int(b[2])
	hdr.OffsetSize = int(b[3])
	return
}

type Collection struct {
	Header    *Header
	Fonts     map[string]*Font
	FirstFont *Font
}
