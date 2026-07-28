//go:build !wasm

package style

type Space uint8

const (
	Space0 Space = iota
	Space1
	Space2
	Space3
	Space4
	Space5
	Space6
	Space7
	Space8
	Space9
	Space10
	Space11
	Space12
)

type Radius uint8

const (
	RadiusNone Radius = iota
	RadiusSm
	RadiusMd
	RadiusLg
	RadiusFull
)

type TextSize uint8

const (
	TextXs TextSize = iota
	TextSm
	TextBase
	TextLg
	TextXl
	Text2xl
)

type Weight uint8

const (
	WeightRegular Weight = iota
	WeightMedium
	WeightBold
)

type Elevation uint8

const (
	Flat Elevation = iota
	Raised
	Floating
	Overlay
)

type Ratio uint8

const (
	RatioHalf Ratio = iota
	RatioTwoThirds
	RatioThreeQuarters
)

type Track uint8

const (
	TrackSm Track = iota
	TrackMd
	TrackLg
)

// Size es la ÚNICA medida de ancho. Relativa al contenedor, nunca al viewport —
// las unidades vw/vh no existen en esta API.
type Size uint8

const (
	Content Size = iota // se ajusta a su contenido
	Prose               // ancho legible (~65ch)
	Third
	Half
	TwoThirds
	Full // 100% del contenedor
)
