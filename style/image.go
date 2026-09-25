package style

// Wrapping styles for floating images.
const (
	WrappingInline  = "inline"
	WrappingSquare  = "square"
	WrappingTight   = "tight"
	WrappingBehind  = "behind"
	WrappingInFront = "infront"
)

// Image is the visual style of an inline or floating picture.
type Image struct {
	Width         float64 // pixels; 0 = intrinsic
	Height        float64
	WidthEMU      int64
	HeightEMU     int64
	Alignment     string
	WrappingStyle string
	MarginTop     float64
	MarginLeft    float64
	MarginRight   float64
	MarginBottom  float64
	OffsetX       float64
	OffsetY       float64
	IsWatermark   bool
	Name          string
	AltText       string
}
