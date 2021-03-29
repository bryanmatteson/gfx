package gfx

type ShaderKind int

// Shade types
const (
	LinearShaderKind ShaderKind = iota
	RadialShaderKind
	MeshShaderKind
	FunctionShaderKind
)

type Shader struct {
	Kind   ShaderKind
	Matrix Matrix
	Bounds Rect
}
