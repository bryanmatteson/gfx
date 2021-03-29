package gfx

type Align int

const (
	AlignLeft Align = iota
	AlignCenter
	AlignRight
)

type FillRule int

const (
	FillRuleEvenOdd FillRule = iota
	FillRuleWinding
)

type LineJoin int

const (
	MiterJoin LineJoin = iota
	RoundJoin
	BevelJoin
)

type LineCap int

const (
	ButtCap LineCap = iota
	RoundCap
	SquareCap
	TriangleCap
)

type BlendMode int

const (
	BlendNormal BlendMode = iota
	BlendMultiply
	BlendScreen
	BlendOverlay
	BlendDarken
	BlendLighten
	BlendColorDodge
	BlendColorBurn
	BlendHardLight
	BlendSoftLight
	BlendDifference
	BlendExclusion
	BlendHue
	BlendSaturation
	BlendColor
	BlendLuminosity
)
