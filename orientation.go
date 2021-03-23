package gfx

type Orientation int

// orientations
const (
	OtherOrientation Orientation = iota
	Horizontal
	Rotate180
	Rotate270
	Rotate90
)
