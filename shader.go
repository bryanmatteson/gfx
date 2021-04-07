package gfx

type ShaderKind int

// Shade types
const (
	FunctionShader ShaderKind = 1 + iota
	LinearShader
	RadialShader
	MeshShader
)

type Shader struct {
	Kind   ShaderKind
	Matrix Matrix
	Bounds Rect
}
