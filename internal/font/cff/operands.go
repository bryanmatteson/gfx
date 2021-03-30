package cff

import (
	"fmt"
	"reflect"

	"go.matteson.dev/gfx"
)

type operand struct{ value interface{} }

func (o operand) Int() (int, error) {
	i, ok := o.value.(int)
	if !ok {
		return 0, fmt.Errorf("expected int operand at index 0")
	}
	return i, nil
}

func (o operand) IntOrDef(def int) int {
	i, err := o.Int()
	if err != nil {
		return def
	}
	return i
}

func (o operand) Float() (float64, error) {
	switch v := o.value.(type) {
	case float64:
		return v, nil
	case int:
		return float64(v), nil
	default:
		return 0, fmt.Errorf("expected float operand, got %v", reflect.TypeOf(v))
	}
}

func (o operand) FloatOrDef(def float64) float64 {
	i, err := o.Float()
	if err != nil {
		return def
	}
	return i
}

func (o operand) GetString(strIndex strtable) (string, error) {
	i, err := o.Int()
	if err != nil {
		return "", err
	}

	return strIndex.GetName(i), nil
}

type operands []operand

func (o operands) GetString(strIndex strtable) (string, error) {
	if len(o) == 0 {
		return "", fmt.Errorf("empty operands")
	}
	return o[0].GetString(strIndex)
}

func (o operands) GetBoundingBox() (q gfx.Quad) {
	if len(o) != 4 {
		return
	}

	f1 := o[0].FloatOrDef(0)
	f2 := o[1].FloatOrDef(0)
	f3 := o[2].FloatOrDef(0)
	f4 := o[3].FloatOrDef(0)

	return gfx.MakeQuad(f1, f2, f3, f4)
}

func (o operands) GetFloats() (val []float64) {
	val = make([]float64, len(o))
	for i := range o {
		val[i] = o[i].FloatOrDef(0)
	}
	return
}

func (o operands) GetIntOrDef(idx int, def int) int {
	if idx >= len(o) {
		return def
	}
	return o[idx].IntOrDef(def)
}

func (o operands) GetFloatOrDef(idx int, def float64) float64 {
	if idx >= len(o) {
		return def
	}
	return o[idx].FloatOrDef(def)
}

func (o operands) GetDeltaInts() (val []int) {
	if len(o) == 0 {
		return
	}

	val = make([]int, len(o))
	val[0] = int(o[0].FloatOrDef(0))

	for i := 1; i < len(o); i++ {
		prev := val[i-1]
		cur := int(o[i].FloatOrDef(0))
		val[i] = prev + cur
	}
	return
}

func (o operands) GetDeltas() (val []float64) {
	if len(o) == 0 {
		return
	}
	val = make([]float64, len(o))
	val[0] = o[0].FloatOrDef(0)

	for i := 1; i < len(o); i++ {
		prev := val[i-1]
		cur := o[i].FloatOrDef(0)
		val[i] = prev + cur
	}
	return
}
