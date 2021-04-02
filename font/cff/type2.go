package cff

import (
	"fmt"
	"log"

	"go.matteson.dev/gfx"
	"go.matteson.dev/gfx/font/cff/charsets"
)

type Type2Glyph struct {
	Width float64
	Path  *gfx.Path
}

type t2charstrings struct {
	selector    SubroutineSelector
	charset     charsets.Charset
	charstrings map[string]pscmdseq
}

type t2context struct {
	gsubs     []pscmdseq
	lsubs     []pscmdseq
	seenWidth bool
	dx, nx    float64
	stack     fstack
	width     float64
	path      gfx.Path
}

func (ctx *t2context) addHstemHints(hints [][2]float64) {}
func (ctx *t2context) addVstemHints(hints [][2]float64) {}

func (ctx *t2context) addRelHLine(dx float64) { ctx.addRelLine(dx, 0) }
func (ctx *t2context) addRelVLine(dy float64) { ctx.addRelLine(0, dy) }

func (ctx *t2context) addRelBezier(dx, dy, dcx1, dcy1, dcx2, dcy2 float64) {
	x, y := ctx.path.LastPoint()
	x, y = x+dx, y+dy

	var cx1, cy1 = x + dcx1, y + dcy1
	var cx2, cy2 = x + dcx2, y + dcy2

	ctx.path.CubicCurveTo(cx1, cy1, cx2, cy2, x, y)
}

func (ctx *t2context) addRelLine(dx, dy float64) {
	x, y := ctx.path.LastPoint()
	ctx.path.LineTo(x+dx, y+dy)
}

const (
	t2cmdHstem      = 1
	t2cmdVstem      = 3
	t2cmdVmoveto    = 4
	t2cmdRlineto    = 5
	t2cmdHlineto    = 6
	t2cmdVlineto    = 7
	t2cmdRrcurveto  = 8
	t2cmdCallsubr   = 10
	t2cmdReturn     = 11
	t2cmdEndchar    = 14
	t2cmdHstemhm    = 18
	t2cmdHintmask   = 19
	t2cmdCntrmask   = 20
	t2cmdRmoveto    = 21
	t2cmdHmoveto    = 22
	t2cmdVstemhm    = 23
	t2cmdRcurveline = 24
	t2cmdRlinecurve = 25
	t2cmdVvcurveto  = 26
	t2cmdHhcurveto  = 27
	t2cmdCallgsubr  = 29
	t2cmdVhcurveto  = 30
	t2cmdHvcurveto  = 31
	t2cmdHflex      = 1234
	t2cmdHflex1     = 1236
)

type fstack []float64

// func (s *fstack) popt() float64 {
// 	sz := len(*s)
// 	res := (*s)[sz-1]
// 	*s = (*s)[:sz-1]
// 	return res
// }

func (s *fstack) popb() float64 {
	res := (*s)[0]
	*s = (*s)[1:]
	return res
}

func (s *fstack) clear() {
	*s = (*s)[:0]
}

func (s *fstack) push(v ...float64) {
	*s = append(*s, v...)
}

func t2parse(data [][]byte, selector SubroutineSelector, charset charsets.Charset) (result *t2charstrings, err error) {
	commands := make(map[string]pscmdseq, len(data))
	for i := 0; i < len(data); i++ {
		cs := data[i]
		name := charset.GetNameByGlyphID(i)
		cmd, err := parseCommandSequence(cs, psType2Context)
		if err != nil {
			return nil, err
		}
		commands[name] = cmd
	}

	result = &t2charstrings{
		selector:    selector,
		charset:     charset,
		charstrings: commands,
	}
	return
}

func (cs *t2charstrings) Generate(c string, gid int, dx, nx float64) (*Type2Glyph, error) {
	seq, ok := cs.charstrings[c]
	if !ok {
		seq, ok = cs.charstrings[".notdef"]
		if !ok {
			return nil, fmt.Errorf("no sequence with name %s in this font", c)
		}
	}

	gsubs, lsubs := cs.selector.GetSubroutines(gid)
	ctx := t2context{
		lsubs: lsubs,
		gsubs: gsubs,
		dx:    dx,
		nx:    nx,
	}

	runseq(&ctx, seq)
	return &Type2Glyph{Width: ctx.width, Path: &ctx.path}, nil
}

func runseq(ctx *t2context, seq pscmdseq) {
	for _, cmd := range seq {
		ctx.stack.push(cmd.args...)
		t2ReadWidth(ctx, cmd.id)
		runcmd(ctx, cmd)
	}
}

func t2ReadWidth(ctx *t2context, cmdid int) {
	if ctx.seenWidth {
		return
	}

	ctx.seenWidth = true
	switch cmdid {
	case t2cmdHstem, t2cmdHstemhm, t2cmdVstemhm, t2cmdVstem:
		if len(ctx.stack)%2 != 0 {
			ctx.width = ctx.nx + ctx.stack.popb()
		}
	case t2cmdHmoveto, t2cmdVmoveto:
		if len(ctx.stack) > 1 {
			ctx.width = ctx.nx + ctx.stack.popb()
		}
	case t2cmdRmoveto:
		if len(ctx.stack) > 2 {
			ctx.width = ctx.nx + ctx.stack.popb()
		}
	case t2cmdCntrmask, t2cmdHintmask:
		if len(ctx.stack) > 0 {
			ctx.width = ctx.nx + ctx.stack.popb()
		}
	case t2cmdEndchar:
		if len(ctx.stack) == 0 {
			ctx.width = ctx.dx
		} else if len(ctx.stack) > 0 {
			ctx.width = ctx.nx + ctx.stack.popb()
		}
	default:
		ctx.seenWidth = true
	}
}

func runcmd(ctx *t2context, cmd *pscmd) {
	switch cmd.id {
	case t2cmdHstem:
		t2Hstem(ctx)
	case t2cmdVstem:
		t2Vstem(ctx)
	case t2cmdVmoveto:
		t2Vmoveto(ctx)
	case t2cmdRlineto:
		t2Rlineto(ctx)
	case t2cmdHlineto:
		t2Hlineto(ctx)
	case t2cmdVlineto:
		t2Vlineto(ctx)
	case t2cmdRrcurveto:
		t2Rrcurveto(ctx)
	case t2cmdCallsubr:
		t2Callsubr(ctx)
	case t2cmdReturn:
		t2Return(ctx)
	case t2cmdEndchar:
		t2Endchar(ctx)
	case t2cmdHstemhm:
		t2Hstemhm(ctx)
	case t2cmdHintmask:
		t2Hintmask(ctx)
	case t2cmdCntrmask:
		t2Cntrmask(ctx)
	case t2cmdRmoveto:
		t2Rmoveto(ctx)
	case t2cmdHmoveto:
		t2Hmoveto(ctx)
	case t2cmdVstemhm:
		t2Vstemhm(ctx)
	case t2cmdRcurveline:
		t2Rcurveline(ctx)
	case t2cmdRlinecurve:
		t2Rlinecurve(ctx)
	case t2cmdVvcurveto:
		t2Vvcurveto(ctx)
	case t2cmdHhcurveto:
		t2Hhcurveto(ctx)
	case t2cmdCallgsubr:
		t2Callgsubr(ctx)
	case t2cmdVhcurveto:
		t2Vhcurveto(ctx)
	case t2cmdHvcurveto:
		t2Hvcurveto(ctx)
	case t2cmdHflex:
		t2Hflex(ctx)
	case t2cmdHflex1:
		t2Hflex1(ctx)
	default:
		log.Printf("unknown command %d: '%s'", cmd.id, cmd.op.name)
	}
}

func t2Stem(ctx *t2context, dim int) {
	var numberOfEdgeHints = len(ctx.stack) / 2
	var hints = make([][2]float64, numberOfEdgeHints)

	var firstStart = ctx.stack.popb()
	var end = firstStart + ctx.stack.popb()

	hints[0] = [2]float64{firstStart, end}

	var current = end

	for i := 1; i < numberOfEdgeHints; i++ {
		var dstart = ctx.stack.popb()
		var dend = ctx.stack.popb()

		hints[i] = [2]float64{current + dstart, current + dstart + dend}
		current = current + dstart + dend
	}

	if dim == 0 {
		ctx.addHstemHints(hints)
	} else {
		ctx.addVstemHints(hints)
	}
	ctx.stack.clear()
}

func t2Hstem(ctx *t2context) { t2Stem(ctx, 0) }
func t2Vstem(ctx *t2context) { t2Stem(ctx, 1) }

func t2Vmoveto(ctx *t2context) {
	x, y := ctx.path.LastPoint()
	ctx.path.MoveTo(x, y+ctx.stack.popb())
	ctx.stack.clear()
}

func t2Rlineto(ctx *t2context) {
	var numberOfLines = len(ctx.stack) / 2

	for i := 0; i < numberOfLines; i++ {
		var dx = ctx.stack.popb()
		var dy = ctx.stack.popb()
		ctx.addRelLine(dx, dy)
	}

	ctx.stack.clear()
}

func t2Hlineto(ctx *t2context) {
	/*
	 * Appends a horizontal line of length dx1 to the current point.
	 * With an odd number of arguments, subsequent argument pairs are interpreted as alternating values of dy and dx.
	 * With an even number of arguments, the arguments are interpreted as alternating horizontal and vertical lines (dx and dy).
	 * The number of lines is determined from the number of arguments on the stack.
	 */
	var isOdd = len(ctx.stack)%2 != 0

	var numberOfAdditionalLines = len(ctx.stack)
	if isOdd {
		numberOfAdditionalLines--
	}

	if isOdd {
		ctx.addRelHLine(ctx.stack.popb())
		for i := 0; i < numberOfAdditionalLines; i += 2 {
			ctx.addRelVLine(ctx.stack.popb())
			ctx.addRelHLine(ctx.stack.popb())
		}
	} else {
		for i := 0; i < numberOfAdditionalLines; i += 2 {
			ctx.addRelHLine(ctx.stack.popb())
			ctx.addRelVLine(ctx.stack.popb())
		}
	}

	ctx.stack.clear()
}

func t2Vlineto(ctx *t2context) {
	var isOdd = len(ctx.stack)%2 != 0

	var numberOfAdditionalLines = len(ctx.stack)
	if isOdd {
		numberOfAdditionalLines--
	}

	if isOdd {
		ctx.addRelVLine(ctx.stack.popb())
		for i := 0; i < numberOfAdditionalLines; i += 2 {
			ctx.addRelHLine(ctx.stack.popb())
			ctx.addRelVLine(ctx.stack.popb())
		}
	} else {
		for i := 0; i < numberOfAdditionalLines; i += 2 {
			ctx.addRelVLine(ctx.stack.popb())
			ctx.addRelHLine(ctx.stack.popb())
		}
	}

	ctx.stack.clear()
}

func t2Rrcurveto(ctx *t2context) {
	var curveCount = len(ctx.stack) / 6
	for i := 0; i < curveCount; i++ {
		ctx.addRelBezier(
			ctx.stack.popb(), ctx.stack.popb(), ctx.stack.popb(),
			ctx.stack.popb(), ctx.stack.popb(), ctx.stack.popb())
	}

	ctx.stack.clear()
}

func t2Callsubr(ctx *t2context)  { t2Call(ctx, ctx.lsubs) }
func t2Callgsubr(ctx *t2context) { t2Call(ctx, ctx.gsubs) }
func t2Call(ctx *t2context, subroutines []pscmdseq) {
	idx := int(ctx.stack.popb()) + subrBias(len(subroutines))
	subr := subroutines[idx]
	runseq(ctx, subr)
}

func t2Return(ctx *t2context) {}

func t2Endchar(ctx *t2context) {
	ctx.path.ClosePath()
	ctx.stack.clear()
}

func t2Hstemhm(ctx *t2context) { t2Hstem(ctx) }
func t2Vstemhm(ctx *t2context) { t2Vstem(ctx) }

func t2Hintmask(ctx *t2context) { ctx.stack.clear() } // TODO
func t2Cntrmask(ctx *t2context) { ctx.stack.clear() } // TODO

func t2Rmoveto(ctx *t2context) {
	x, y := ctx.path.LastPoint()
	ctx.path.MoveTo(x+ctx.stack.popb(), y+ctx.stack.popb())
	ctx.stack.clear()
}

func t2Hmoveto(ctx *t2context) {
	x, y := ctx.path.LastPoint()
	ctx.path.MoveTo(x+ctx.stack.popb(), y)
	ctx.stack.clear()
}

func t2Rcurveline(ctx *t2context) {
	var numberOfCurves = (len(ctx.stack) - 2) / 6
	for i := 0; i < numberOfCurves; i++ {
		ctx.addRelBezier(
			ctx.stack.popb(), ctx.stack.popb(), ctx.stack.popb(),
			ctx.stack.popb(), ctx.stack.popb(), ctx.stack.popb())
	}

	ctx.addRelLine(ctx.stack.popb(), ctx.stack.popb())
	ctx.stack.clear()
}

func t2Rlinecurve(ctx *t2context) {
	var numberOfLines = (len(ctx.stack) - 6) / 2
	for i := 0; i < numberOfLines; i++ {
		ctx.addRelLine(ctx.stack.popb(), ctx.stack.popb())
	}

	ctx.addRelBezier(
		ctx.stack.popb(), ctx.stack.popb(), ctx.stack.popb(),
		ctx.stack.popb(), ctx.stack.popb(), ctx.stack.popb())

	ctx.stack.clear()
}

func t2Vvcurveto(ctx *t2context) {
	var hasDeltaXFirstCurve = len(ctx.stack)%4 != 0

	var numberOfCurves = len(ctx.stack) / 4
	for i := 0; i < numberOfCurves; i++ {
		var dx1 = 0.0
		if i == 0 && hasDeltaXFirstCurve {
			dx1 = ctx.stack.popb()
		}

		var dy1 = ctx.stack.popb()
		var dx2 = ctx.stack.popb()
		var dy2 = ctx.stack.popb()
		var dy3 = ctx.stack.popb()

		ctx.addRelBezier(dx1, dy1, dx2, dy2, 0, dy3)
	}

	ctx.stack.clear()
}
func t2Hhcurveto(ctx *t2context) {
	var hasDeltaYFirstCurve = len(ctx.stack)%4 != 0

	if hasDeltaYFirstCurve {
		var dy1 = ctx.stack.popb()
		var dx1 = ctx.stack.popb()
		var dx2 = ctx.stack.popb()
		var dy2 = ctx.stack.popb()
		var dx3 = ctx.stack.popb()

		ctx.addRelBezier(dx1, dy1, dx2, dy2, dx3, 0)
	}

	var numberOfCurves = len(ctx.stack) / 4
	for i := 0; i < numberOfCurves; i++ {
		var dx1 = ctx.stack.popb()
		var dx2 = ctx.stack.popb()
		var dy2 = ctx.stack.popb()
		var dx3 = ctx.stack.popb()

		ctx.addRelBezier(dx1, 0, dx2, dy2, dx3, 0)
	}

	ctx.stack.clear()
}
func t2Vhcurveto(ctx *t2context) {
	var remainder = len(ctx.stack) % 8

	if remainder <= 1 {
		// {dya dxb dyb dxc dxd dxe dye dyf}+ dxf?
		// 2 curves, 1st starts vertical ends horizontal, second starts horizontal ends vertical

		var numberOfCurves = (len(ctx.stack) - remainder) / 8
		for i := 0; i < numberOfCurves; i++ {
			// First curve
			{
				var dy1 = ctx.stack.popb()
				var dx2 = ctx.stack.popb()
				var dy2 = ctx.stack.popb()
				var dx3 = ctx.stack.popb()
				ctx.addRelBezier(0, dy1, dx2, dy2, dx3, 0)
			}
			// Second curve
			{
				var dx1 = ctx.stack.popb()
				var dx2 = ctx.stack.popb()
				var dy2 = ctx.stack.popb()
				var dy3 = ctx.stack.popb()
				var dx3 = 0.0

				if i == numberOfCurves-1 && remainder == 1 {
					dx3 = ctx.stack.popb()
				}

				ctx.addRelBezier(dx1, 0, dx2, dy2, dx3, dy3)
			}
		}
	} else if remainder == 4 || remainder == 5 {
		// dy1 dx2 dy2 dx3 {dxa dxb dyb dyc dyd dxe dye dxf}* dyf?
		var numberOfCurves = (len(ctx.stack) - remainder) / 8

		{
			var dy1 = ctx.stack.popb()
			var dx2 = ctx.stack.popb()
			var dy2 = ctx.stack.popb()
			var dx3 = ctx.stack.popb()
			var dy3 = 0.0
			if len(ctx.stack) == 1 {
				dy3 = ctx.stack.popb()
			}
			ctx.addRelBezier(0, dy1, dx2, dy2, dx3, dy3)
		}

		for i := 0; i < numberOfCurves; i++ {
			// First curve
			{
				var dx1 = ctx.stack.popb()
				var dx2 = ctx.stack.popb()
				var dy2 = ctx.stack.popb()
				var dy3 = ctx.stack.popb()
				ctx.addRelBezier(dx1, 0, dx2, dy2, 0, dy3)
			}
			// Second curve
			{
				var dy1 = ctx.stack.popb()
				var dx2 = ctx.stack.popb()
				var dy2 = ctx.stack.popb()
				var dx3 = ctx.stack.popb()
				var dy3 = 0.0

				if i == numberOfCurves-1 && remainder == 5 {
					dy3 = ctx.stack.popb()
				}

				ctx.addRelBezier(0, dy1, dx2, dy2, dx3, dy3)
			}
		}
	}

	ctx.stack.clear()
}
func t2Hvcurveto(ctx *t2context) {
	var remainder = len(ctx.stack) % 8

	if remainder <= 1 {
		// {dxa dxb dyb dyc dyd dxe dye dxf}+ dyf?
		// 2 curves, 1st starts horizontal ends vertical, second starts vertical ends horizontal

		var numberOfCurves = (len(ctx.stack) - remainder) / 8
		for i := 0; i < numberOfCurves; i++ {
			// First curve
			{
				var dx1 = ctx.stack.popb()
				var dx2 = ctx.stack.popb()
				var dy2 = ctx.stack.popb()
				var dy3 = ctx.stack.popb()
				ctx.addRelBezier(dx1, 0, dx2, dy2, 0, dy3)
			}

			// Second curve
			{
				var dy1 = ctx.stack.popb()
				var dx2 = ctx.stack.popb()
				var dy2 = ctx.stack.popb()
				var dx3 = ctx.stack.popb()
				var dy3 = 0.0

				if i == numberOfCurves-1 && remainder == 1 {
					dy3 = ctx.stack.popb()
				}

				ctx.addRelBezier(0, dy1, dx2, dy2, dx3, dy3)
			}
		}
	} else if remainder == 4 || remainder == 5 {
		// dx1 dx2 dy2 dy3 {dya dxb dyb dxc dxd dxe dye dyf}* dxf?
		var numberOfCurves = (len(ctx.stack) - remainder) / 8

		{
			var dx1 = ctx.stack.popb()
			var dx2 = ctx.stack.popb()
			var dy2 = ctx.stack.popb()
			var dy3 = ctx.stack.popb()
			var dx3 = 0.0
			if len(ctx.stack) == 1 {
				dx3 = ctx.stack.popb()
			}
			ctx.addRelBezier(dx1, 0, dx2, dy2, dx3, dy3)
		}

		for i := 0; i < numberOfCurves; i++ {
			// First curve
			{
				var dy1 = ctx.stack.popb()
				var dx2 = ctx.stack.popb()
				var dy2 = ctx.stack.popb()
				var dx3 = ctx.stack.popb()
				ctx.addRelBezier(0, dy1, dx2, dy2, dx3, 0)
			}
			// Second curve
			{
				var dx1 = ctx.stack.popb()
				var dx2 = ctx.stack.popb()
				var dy2 = ctx.stack.popb()
				var dy3 = ctx.stack.popb()
				var dx3 = 0.0

				if i == numberOfCurves-1 && remainder == 5 {
					dx3 = ctx.stack.popb()
				}

				ctx.addRelBezier(dx1, 0, dx2, dy2, dx3, dy3)
			}
		}
	}

	ctx.stack.clear()
}

// func t2And(ctx *t2context) {
// 	lhs := ctx.stack.popt()
// 	rhs := ctx.stack.popt()
// 	res := 0.0
// 	if lhs != 0 && rhs != 0 {
// 		res = 1
// 	}
// 	ctx.stack.push(res)
// }

// func t2Or(ctx *t2context) {
// 	lhs := ctx.stack.popt()
// 	rhs := ctx.stack.popt()
// 	res := 0.0
// 	if lhs != 0 || rhs != 0 {
// 		res = 1
// 	}
// 	ctx.stack.push(res)
// }

// func t2Not(ctx *t2context)    {}
// func t2Abs(ctx *t2context)    {}
// func t2Add(ctx *t2context)    {}
// func t2Sub(ctx *t2context)    {}
// func t2Div(ctx *t2context)    {}
// func t2Neg(ctx *t2context)    {}
// func t2Eq(ctx *t2context)     {}
// func t2Drop(ctx *t2context)   {}
// func t2Put(ctx *t2context)    {}
// func t2Get(ctx *t2context)    {}
// func t2IfElse(ctx *t2context) {}
// func t2Rand(ctx *t2context)   {}
// func t2Mul(ctx *t2context)    {}
// func t2Sqrt(ctx *t2context)   {}
// func t2Dup(ctx *t2context)    {}
// func t2Exch(ctx *t2context)   {}
// func t2Index(ctx *t2context)  {}
// func t2Roll(ctx *t2context)   {}

// func t2Flex(ctx *t2context)  {}
// func t2Flex1(ctx *t2context) {}

func t2Hflex(ctx *t2context)  { ctx.stack.clear() } // TODO
func t2Hflex1(ctx *t2context) { ctx.stack.clear() } // TODO

func subrBias(numSubroutines int) int {
	if numSubroutines < 1240 {
		return 107
	}
	if numSubroutines < 33900 {
		return 1131
	}
	return 32768
}
