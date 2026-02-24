package screen

import (
	"math"
)

// ==================== HELPERS ====================

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// ==================== COLOR SPACE CONVERSIONS ====================

// sRGBToLinear converts sRGB (gamma-corrected) value to linear color space
// sRGB is what we receive from screen capture
// Formula: if C_srgb <= 0.04045: C_linear = C_srgb / 12.92
//
//	else: C_linear = ((C_srgb + 0.055) / 1.055) ^ 2.4
func sRGBToLinear(c uint8) float64 {
	cf := float64(c) / 255.0
	if cf <= 0.04045 {
		return cf / 12.92
	}
	return math.Pow((cf+0.055)/1.055, 2.4)
}

// LinearToSRGB converts linear color space back to sRGB (gamma-corrected)
// Formula: if C_linear <= 0.0031308: C_srgb = C_linear * 12.92
//
//	else: C_srgb = 1.055 * C_linear^(1/2.4) - 0.055
func LinearToSRGB(c float64) uint8 {
	var cf float64
	if c <= 0.0031308 {
		cf = c * 12.92
	} else {
		cf = 1.055*math.Pow(c, 1.0/2.4) - 0.055
	}
	// Clamp to [0, 1]
	if cf < 0 {
		cf = 0
	} else if cf > 1 {
		cf = 1
	}
	return uint8(cf * 255.0)
}

// LinearColor represents color in linear RGB space (0.0-1.0)
type LinearColor struct {
	R, G, B float64
}

// FromSRGB creates a LinearColor from sRGB bytes
func (lc *LinearColor) FromSRGB(r, g, b uint8) {
	lc.R = sRGBToLinear(r)
	lc.G = sRGBToLinear(g)
	lc.B = sRGBToLinear(b)
}

// ToSRGB converts to sRGB bytes
func (lc *LinearColor) ToSRGB() (uint8, uint8, uint8) {
	return LinearToSRGB(lc.R), LinearToSRGB(lc.G), LinearToSRGB(lc.B)
}

// Add combines two linear colors
func (lc *LinearColor) Add(other *LinearColor) {
	lc.R += other.R
	lc.G += other.G
	lc.B += other.B
}

// AddScaled adds a scaled version of another color
func (lc *LinearColor) AddScaled(other *LinearColor, scale float64) {
	lc.R += other.R * scale
	lc.G += other.G * scale
	lc.B += other.B * scale
}

// Scale multiplies all channels by a scalar
func (lc *LinearColor) Scale(factor float64) {
	lc.R *= factor
	lc.G *= factor
	lc.B *= factor
}

// Clamp ensures all values are in [0, 1]
func (lc *LinearColor) Clamp() {
	if lc.R < 0 {
		lc.R = 0
	} else if lc.R > 1 {
		lc.R = 1
	}
	if lc.G < 0 {
		lc.G = 0
	} else if lc.G > 1 {
		lc.G = 1
	}
	if lc.B < 0 {
		lc.B = 0
	} else if lc.B > 1 {
		lc.B = 1
	}
}

// Luminance calculates perceptual luminance (Rec.709)
func (lc *LinearColor) Luminance() float64 {
	return 0.2126*lc.R + 0.7152*lc.G + 0.0722*lc.B
}

// ==================== REGION SAMPLING WITH WEIGHTS ====================

// SampleRegionWeighted samples a rectangular region with edge-weighted averaging
// Edges get higher weight: w = 1.0 + 0.5 * (normalized_distance_to_edge)
// img: RGBA pixel buffer (4 bytes per pixel)
// w, h: image dimensions
// cx, cy: center point
// regionSize: size of region (will be centered)
func SampleRegionWeighted(img []byte, w, h, cx, cy, regionSize int) LinearColor {
	var result LinearColor
	totalWeight := 0.0

	// Calculate region bounds
	halfSize := regionSize / 2
	x0 := cx - halfSize
	x1 := cx + halfSize
	y0 := cy - halfSize
	y1 := cy + halfSize

	// Clamp to image bounds
	if x0 < 0 {
		x0 = 0
	}
	if x1 >= w {
		x1 = w - 1
	}
	if y0 < 0 {
		y0 = 0
	}
	if y1 >= h {
		y1 = h - 1
	}

	// Sample each pixel in region with edge weighting
	for yy := y0; yy <= y1; yy++ {
		for xx := x0; xx <= x1; xx++ {
			// Calculate normalized distance to edge (0 at center, 1 at edge)
			normDistX := float64(abs(xx-cx)) / float64(halfSize+1)
			normDistY := float64(abs(yy-cy)) / float64(halfSize+1)
			normDistToEdge := math.Max(normDistX, normDistY)

			// Weight increases towards edges: 1.0 + 0.5 * distance
			weight := 1.0 + 0.5*normDistToEdge

			// Sample pixel (RGBA format)
			idx := (yy*w + xx) * 4
			var pixel LinearColor
			pixel.FromSRGB(img[idx+2], img[idx+1], img[idx])

			// Add weighted contribution
			result.AddScaled(&pixel, weight)
			totalWeight += weight
		}
	}

	if totalWeight > 0 {
		result.Scale(1.0 / totalWeight)
	}

	result.Clamp()
	return result
}

// SampleLine samples a SampleLine with linear color averaging
func SampleLineLinear(line SampleLine, img []byte) LinearColor {
	var result LinearColor
	count := 0

	for _, offset := range line.Offsets {
		// offset points to B in RGBA format
		// B at offset, G at offset+1, R at offset+2
		var pixel LinearColor
		pixel.FromSRGB(img[offset+2], img[offset+1], img[offset])
		result.Add(&pixel)
		count++
	}

	if count > 0 {
		result.Scale(1.0 / float64(count))
	}

	result.Clamp()
	return result
}

// ==================== CORRECTION MATRIX ====================

// CorrectionMatrix is a 3x3 color correction matrix
type CorrectionMatrix [3][3]float64

// IdentityMatrix creates a 3x3 identity matrix
func IdentityMatrix() CorrectionMatrix {
	return CorrectionMatrix{
		{1, 0, 0},
		{0, 1, 0},
		{0, 0, 1},
	}
}

// DiagonalMatrix creates a diagonal correction matrix (RGB gains)
func DiagonalMatrix(r, g, b float64) CorrectionMatrix {
	return CorrectionMatrix{
		{r, 0, 0},
		{0, g, 0},
		{0, 0, b},
	}
}

// Apply applies the color correction
func (m *CorrectionMatrix) Apply(c *LinearColor) {
	r := m[0][0]*c.R + m[0][1]*c.G + m[0][2]*c.B
	g := m[1][0]*c.R + m[1][1]*c.G + m[1][2]*c.B
	b := m[2][0]*c.R + m[2][1]*c.G + m[2][2]*c.B

	c.R = r
	c.G = g
	c.B = b
	c.Clamp()
}

// ==================== POWER LIMITING ====================

// ApplyGlobalPowerLimit scales all LEDs to stay within power budget
// threshold: maximum allowed luminance sum (0.0-255.0 scale)
func ApplyGlobalPowerLimit(ledColors []LinearColor, powerLimit float64) {
	if powerLimit <= 0 {
		return
	}

	// Calculate total luminance
	totalLum := 0.0
	for _, col := range ledColors {
		totalLum += col.Luminance()
	}

	// If within limit, nothing to do
	if totalLum <= powerLimit {
		return
	}

	// Scale all LEDs down proportionally
	scale := powerLimit / totalLum
	for i := range ledColors {
		ledColors[i].Scale(scale)
	}
}
