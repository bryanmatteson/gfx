package gfx

type ColorspaceKind int

// color spaces
const (
	ColorspaceNone ColorspaceKind = iota
	ColorspaceGray
	ColorspaceRGB
	ColorspaceBGR
	ColorspaceCMYK
	ColorspaceLab
	ColorspaceIndexed
	ColorspaceSeparation
)

// flags
const (
	ColorspaceIsDevice uint32 = 1 << iota
	ColorspaceIsICC
	ColorspaceHasCMYK
	ColorspaceHasSpots
	ColorspaceHasCMYKAndSpots
)

type Colorspace struct {
	Kind          ColorspaceKind
	Name          string
	ColorantCount int
	Flags         uint32
}

func (c Colorspace) IsSubtractive() bool {
	return c.Kind == ColorspaceCMYK || c.Kind == ColorspaceSeparation
}

func (c Colorspace) DeviceNHasOnlyCMYK() bool {
	return (c.Flags&ColorspaceHasCMYK != 0) && (c.Flags&ColorspaceHasCMYKAndSpots == 0)
}

func (c Colorspace) DeviceNHasCMYK() bool { return c.Flags&ColorspaceHasCMYK != 0 }

func (c Colorspace) IsGray() bool { return c.Kind == ColorspaceGray }

func (c Colorspace) IsRGB() bool { return c.Kind == ColorspaceBGR }

func (c Colorspace) IsCMYK() bool { return c.Kind == ColorspaceCMYK }

func (c Colorspace) IsLab() bool { return c.Kind == ColorspaceLab }

func (c Colorspace) IsIndexed() bool { return c.Kind == ColorspaceIndexed }

func (c Colorspace) IsDeviceN() bool { return c.Kind == ColorspaceSeparation }

func (c Colorspace) IsDevice() bool { return c.Flags&ColorspaceIsDevice != 0 }

func (c Colorspace) IsDeviceGray() bool { return c.IsDevice() && c.IsGray() }

func (c Colorspace) IsDeviceCMYK() bool { return c.IsDevice() && c.IsCMYK() }

func (c Colorspace) IsLabICC() bool { return c.IsLab() && c.Flags&ColorspaceIsICC != 0 }
