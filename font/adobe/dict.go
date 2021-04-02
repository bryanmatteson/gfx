package adobe

type PrivateDictionary struct {
	BlueValues               []int
	OtherBlues               []int
	FamilyBlues              []int
	FamilyOtherBlues         []int
	BlueScale                float64
	BlueShift                int
	BlueFuzz                 int
	StandardHorizontalWidth  float64
	StandardVerticalWidth    float64
	StemSnapHorizontalWidths []float64
	StemSnapVerticalWidths   []float64
	ForceBold                bool
	LanguageGroup            int
	ExpansionFactor          float64
}
