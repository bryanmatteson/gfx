package afm

import "go.matteson.dev/gfx"

type WritingDirections byte

const (
	/// Writing direction 0 only.
	Direction0Only WritingDirections = 0
	/// Writing direction 1 only.
	Direction1Only WritingDirections = 1
	/// Writing direction 0 and 1.
	Direction0And1 WritingDirections = 2
)

type Metrics struct {
	// Version of the Adobe Font Metrics specification used to generate this file.
	AfmVersion float64

	// Any comments in the file.
	Comments []string

	// The writing directions described by these metrics.
	MetricSets WritingDirections

	// Font name.
	FontName string

	// Font full name.
	FullName string

	// Font family name.
	FamilyName string

	// Font weight.
	Weight string

	// Minimum bounding box for all characters in the font.
	BoundingBox gfx.Quad

	// Font program version identifier.
	Version string

	// Font name trademark or copyright notice.
	Notice string

	// String indicating the default encoding vector for this font program.
	// Common ones are AdobeStandardEncoding and JIS12-88-CFEncoding.
	// Special font programs might state FontSpecific.
	EncodingScheme string

	// Describes the mapping scheme.
	MappingScheme int

	// The bytes value of the escape-character used if this font is escape-mapped.
	EscapeCharacter int

	// Describes the character set of this font.
	CharacterSet string

	// The number of characters in this font.
	Characters int

	// Whether this is a base font.
	IsBaseFont bool

	// A vector from the origin of writing direction 0 to direction 1.
	VVector gfx.Point

	// Whether VVector is the same for every character in this font.
	IsFixedV bool

	// Whether font is monospace or not
	IsFixedPitch bool

	// Usually the y-value of the top of capital 'H'.
	CapHeight float64

	// Usually the y-value of the top of lowercase 'x'.
	XHeight float64

	// Usually the y-value of the top of lowercase 'd'.
	Ascender float64

	// Usually the y-value of the bottom of lowercase 'p'.
	Descender float64

	// Distance from the baseline for underlining.
	UnderlinePosition float64

	// Width of the line for underlining.
	UnderlineThickness float64

	// Angle in degrees counter-clockwise from the vertical of the vertical linea.
	// Zero for non-italic fonts.
	ItalicAngle float64

	// If present all characters have this width and height.
	CharacterWidth gfx.Point

	// Horizontal stem width.
	HorizontalStemWidth float64

	// Vertical stem width.
	VerticalStemWidth float64

	// Metrics for the individual characters.
	CharacterMetrics map[string]IndividualCharacterMetric
}

type IndividualCharacterMetric struct {
	// Character code.
	CharacterCode int

	// PostScript language character name.
	Name string

	// Width.
	Width gfx.Point

	// Width for writing direction 0.
	WidthDirection0 gfx.Point

	// Width for writing direction 1.
	WidthDirection1 gfx.Point

	// Vector from origin of writing direction 1 to origin of writing direction 0.
	VVector gfx.Point

	// Character bounding box.
	BoundingBox gfx.Quad

	// Ligature information.
	Ligature Ligature
}

type Ligature struct {
	// The character to join with to form a ligature.
	Successor string

	// The current character.
	Value string
}
