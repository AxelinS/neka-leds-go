package screen

import "math"

// ==================== SPATIAL HARMONIC MIXING ====================

// ApplySpatialHarmonic implements local neighbor mixing for smooth transitions
// Each LED blends with its left and right neighbors in a circular/wrap-around pattern
// This simulates light diffusion on the wall without NxN matrix inversion
//
// Formula per LED:
//   LED_i_final = weight_self * LED_i
//                 + weight_neighbor * LED_left
//                 + weight_neighbor * LED_right
//
// For corners/edges, only blend with existing neighbors
func ApplySpatialHarmonic(
	colors []LinearColor,
	weightSelf float64,
	weightNeighbor float64,
) []LinearColor {
	n := len(colors)
	if n == 0 {
		return colors
	}

	result := make([]LinearColor, n)

	for i := range n {
		// Start with self
		result[i] = colors[i]
		result[i].Scale(weightSelf)

		// Add left neighbor (circular)
		leftIdx := (i - 1 + n) % n
		result[i].AddScaled(&colors[leftIdx], weightNeighbor)

		// Add right neighbor (circular)
		rightIdx := (i + 1) % n
		result[i].AddScaled(&colors[rightIdx], weightNeighbor)

		result[i].Clamp()
	}

	return result
}

// ==================== TEMPORAL SMOOTHING ====================

// TemporalFilter maintains exponential smoothing state per LED
// output_t = prev * (1 - alpha) + current * alpha
// alpha: smoothing factor (0.15-0.25 recommended)
// Lower alpha = more smoothing, slower response
// Higher alpha = less smoothing, faster response
type TemporalFilter struct {
	Alpha    float64
	Previous []LinearColor
}

// NewTemporalFilter creates a new temporal filter
func NewTemporalFilter(alpha float64, ledCount int) *TemporalFilter {
	if alpha < 0 {
		alpha = 0
	}
	if alpha > 1 {
		alpha = 1
	}

	return &TemporalFilter{
		Alpha:    alpha,
		Previous: make([]LinearColor, ledCount),
	}
}

// Update applies exponential filtering to current colors
// Returns smoothed colors
func (tf *TemporalFilter) Update(current []LinearColor) []LinearColor {
	if len(current) != len(tf.Previous) {
		tf.Previous = make([]LinearColor, len(current))
		// First frame: return current
		copy(tf.Previous, current)
		return current
	}

	result := make([]LinearColor, len(current))
	oneMinusAlpha := 1.0 - tf.Alpha

	for i := range current {
		result[i].R = tf.Previous[i].R*oneMinusAlpha + current[i].R*tf.Alpha
		result[i].G = tf.Previous[i].G*oneMinusAlpha + current[i].G*tf.Alpha
		result[i].B = tf.Previous[i].B*oneMinusAlpha + current[i].B*tf.Alpha
		result[i].Clamp()
	}

	// Store for next frame
	copy(tf.Previous, result)
	return result
}

// SetAlpha updates the smoothing factor
func (tf *TemporalFilter) SetAlpha(alpha float64) {
	if alpha < 0 {
		alpha = 0
	}
	if alpha > 1 {
		alpha = 1
	}
	tf.Alpha = alpha
}

// Reset clears the filter state
func (tf *TemporalFilter) Reset() {
	for i := range tf.Previous {
		tf.Previous[i] = LinearColor{}
	}
}

// ==================== FULL PROCESSING PIPELINE ====================

// HarmonicProcessingConfig holds all parameters for the harmonic pipeline
type HarmonicProcessingConfig struct {
	// Sampling
	RegionSize             int // Size of region to sample per LED
	EnableWeightedSampling bool

	// Spatial mixing
	EnableSpatialHarmonic  bool
	HarmonicWeightSelf     float64 // typically 0.7
	HarmonicWeightNeighbor float64 // typically 0.15 each

	// Temporal smoothing
	EnableTemporalSmoothing bool
	TemporalAlpha           float64 // 0.15-0.25 recommended

	// Correction matrix
	CorrectionMatrix CorrectionMatrix

	// Power limiting
	EnablePowerLimit    bool
	PowerLimitThreshold float64 // maximum total luminance

	// Legacy compatibility
	Brightness       float64
	Saturation       float64
	Temperature      float64
	RCal, GCal, BCal float64
	Gamma            float64
}

// ProcessLEDPipeline runs the complete harmonic processing pipeline
// img: RGBA image buffer
// w, h: image dimensions
// points: LED positions on screen
// config: processing configuration
// filter: temporal filter (can be nil to skip temporal smoothing)
func ProcessLEDPipeline(
	img []byte,
	w, h int,
	points []Point,
	config *HarmonicProcessingConfig,
	filter *TemporalFilter,
) []byte {
	if len(points) == 0 {
		return []byte{}
	}

	n := len(points)

	// =========== STEP 1: CAPTURE (sRGB) & CONVERT TO LINEAR ===========
	rawColors := make([]LinearColor, n)
	for i, p := range points {
		if config.EnableWeightedSampling && config.RegionSize > 1 {
			rawColors[i] = SampleRegionWeighted(img, w, h, p.X, p.Y, config.RegionSize)
		} else {
			// Point sampling
			idx := (p.Y*w + p.X) * 4
			rawColors[i].FromSRGB(img[idx+2], img[idx+1], img[idx])
		}
	}

	// =========== STEP 2: SPATIAL MIXING (LOCAL NEIGHBORS) ===========
	mixed := rawColors
	if config.EnableSpatialHarmonic {
		mixed = ApplySpatialHarmonic(
			rawColors,
			config.HarmonicWeightSelf,
			config.HarmonicWeightNeighbor,
		)
	}

	// =========== STEP 3: TEMPORAL SMOOTHING ===========
	smoothed := mixed
	if config.EnableTemporalSmoothing && filter != nil {
		smoothed = filter.Update(mixed)
	}

	// =========== STEP 4: CORRECTION MATRIX ===========
	corrected := make([]LinearColor, n)
	copy(corrected, smoothed)
	for i := range corrected {
		config.CorrectionMatrix.Apply(&corrected[i])
	}

	// =========== STEP 5: CALIBRATION IN LINEAR SPACE ===========
	// (Ganancias RGB, Temperatura, Brillo, Saturación - TODO en espacio lineal)
	calibrated := make([]LinearColor, n)
	copy(calibrated, corrected)

	// Apply RGB gains in linear space
	for i := range calibrated {
		calibrated[i].R *= config.RCal
		calibrated[i].G *= config.GCal
		calibrated[i].B *= config.BCal
	}

	// Apply color temperature in linear space
	// temp ∈ [-1.0, 1.0] affects R and B inversely, G stays stable
	if config.Temperature != 0 {
		temp := config.Temperature
		if temp < -1.0 {
			temp = -1.0
		}
		if temp > 1.0 {
			temp = 1.0
		}

		for i := range calibrated {
			if temp > 0 {
				// Warm: increase R, decrease B
				calibrated[i].R *= (1.0 + temp*0.4)
				calibrated[i].B *= (1.0 - temp*0.6)
			} else {
				// Cold: decrease R, increase B
				calibrated[i].R *= (1.0 + temp*0.3)
				calibrated[i].B *= (1.0 - temp*0.4)
			}
		}
	}

	// Apply global brightness in linear space
	if config.Brightness != 1.0 {
		for i := range calibrated {
			calibrated[i].R *= config.Brightness
			calibrated[i].G *= config.Brightness
			calibrated[i].B *= config.Brightness
		}
	}

	// Apply perceptual saturation in linear space
	// lum = 0.2126R + 0.7152G + 0.0722B (Rec.709)
	if config.Saturation != 1.0 {
		for i := range calibrated {
			lum := 0.2126*calibrated[i].R + 0.7152*calibrated[i].G + 0.0722*calibrated[i].B

			// Clamp saturation to valid range
			sat := config.Saturation
			if sat < 0 {
				sat = 0
			}
			if sat > 4.0 {
				sat = 4.0
			}

			calibrated[i].R = lum + (calibrated[i].R-lum)*sat
			calibrated[i].G = lum + (calibrated[i].G-lum)*sat
			calibrated[i].B = lum + (calibrated[i].B-lum)*sat
		}
	}

	// =========== STEP 6: GLOBAL POWER LIMITER ===========
	if config.EnablePowerLimit && config.PowerLimitThreshold > 0 {
		ApplyGlobalPowerLimit(calibrated, config.PowerLimitThreshold)
	}

	// =========== STEP 7: GAMMA CORRECTION (FINAL PERCEPTUAL TRANSFORM) ===========
	// Apply gamma ONLY at the end, before converting back to sRGB
	// gamma is applied as: value = pow(value, 1/gamma)
	gammaCorrection := 1.0 / config.Gamma
	if config.Gamma < 0.1 {
		gammaCorrection = 1.0 / 2.2 // Default to standard gamma if invalid
	}

	for i := range calibrated {
		// Clamp to [0, 1] before applying gamma
		r := calibrated[i].R
		if r < 0 {
			r = 0
		}
		if r > 1 {
			r = 1
		}
		g := calibrated[i].G
		if g < 0 {
			g = 0
		}
		if g > 1 {
			g = 1
		}
		b := calibrated[i].B
		if b < 0 {
			b = 0
		}
		if b > 1 {
			b = 1
		}

		calibrated[i].R = math.Pow(r, gammaCorrection)
		calibrated[i].G = math.Pow(g, gammaCorrection)
		calibrated[i].B = math.Pow(b, gammaCorrection)
	}

	// =========== STEP 8: CONVERT BACK TO sRGB 8-bit ===========
	framebuffer := make([]byte, n*3)
	for i, col := range calibrated {
		r, g, b := col.ToSRGB()
		framebuffer[i*3] = r
		framebuffer[i*3+1] = g
		framebuffer[i*3+2] = b
	}

	return framebuffer
}

// ProcessLEDPipelineFromLines is variant that uses SampleLine for sampling
// (for cinema mode or special sampling patterns)
func ProcessLEDPipelineFromLines(
	img []byte,
	lines []SampleLine,
	config *HarmonicProcessingConfig,
	filter *TemporalFilter,
) []byte {
	if len(lines) == 0 {
		return []byte{}
	}

	n := len(lines)

	// =========== STEP 1: CAPTURE (sRGB) & CONVERT TO LINEAR ===========
	rawColors := make([]LinearColor, n)
	for i, line := range lines {
		rawColors[i] = SampleLineLinear(line, img)
	}

	// =========== STEP 2: SPATIAL MIXING ===========
	mixed := rawColors
	if config.EnableSpatialHarmonic {
		mixed = ApplySpatialHarmonic(
			rawColors,
			config.HarmonicWeightSelf,
			config.HarmonicWeightNeighbor,
		)
	}

	// =========== STEP 3: TEMPORAL SMOOTHING ===========
	smoothed := mixed
	if config.EnableTemporalSmoothing && filter != nil {
		smoothed = filter.Update(mixed)
	}

	// =========== STEP 4: CORRECTION MATRIX ===========
	corrected := make([]LinearColor, n)
	copy(corrected, smoothed)
	for i := range corrected {
		config.CorrectionMatrix.Apply(&corrected[i])
	}

	// =========== STEP 5: CALIBRATION IN LINEAR SPACE ===========
	// (Ganancias RGB, Temperatura, Brillo, Saturación - TODO en espacio lineal)
	calibrated := make([]LinearColor, n)
	copy(calibrated, corrected)

	// Apply RGB gains in linear space
	for i := range calibrated {
		calibrated[i].R *= config.RCal
		calibrated[i].G *= config.GCal
		calibrated[i].B *= config.BCal
	}

	// Apply color temperature in linear space
	// temp ∈ [-1.0, 1.0] affects R and B inversely, G stays stable
	if config.Temperature != 0 {
		temp := config.Temperature
		if temp < -1.0 {
			temp = -1.0
		}
		if temp > 1.0 {
			temp = 1.0
		}

		for i := range calibrated {
			if temp > 0 {
				// Warm: increase R, decrease B
				calibrated[i].R *= (1.0 + temp*0.4)
				calibrated[i].B *= (1.0 - temp*0.6)
			} else {
				// Cold: decrease R, increase B
				calibrated[i].R *= (1.0 + temp*0.3)
				calibrated[i].B *= (1.0 - temp*0.4)
			}
		}
	}

	// Apply global brightness in linear space
	if config.Brightness != 1.0 {
		for i := range calibrated {
			calibrated[i].R *= config.Brightness
			calibrated[i].G *= config.Brightness
			calibrated[i].B *= config.Brightness
		}
	}

	// Apply perceptual saturation in linear space
	// lum = 0.2126R + 0.7152G + 0.0722B (Rec.709)
	if config.Saturation != 1.0 {
		for i := range calibrated {
			lum := 0.2126*calibrated[i].R + 0.7152*calibrated[i].G + 0.0722*calibrated[i].B

			// Clamp saturation to valid range
			sat := config.Saturation
			if sat < 0 {
				sat = 0
			}
			if sat > 4.0 {
				sat = 4.0
			}

			calibrated[i].R = lum + (calibrated[i].R-lum)*sat
			calibrated[i].G = lum + (calibrated[i].G-lum)*sat
			calibrated[i].B = lum + (calibrated[i].B-lum)*sat
		}
	}

	// =========== STEP 6: GLOBAL POWER LIMITER ===========
	if config.EnablePowerLimit && config.PowerLimitThreshold > 0 {
		ApplyGlobalPowerLimit(calibrated, config.PowerLimitThreshold)
	}

	// =========== STEP 7: GAMMA CORRECTION (FINAL PERCEPTUAL TRANSFORM) ===========
	// Apply gamma ONLY at the end, before converting back to sRGB
	// gamma is applied as: value = pow(value, 1/gamma)
	gammaCorrection := 1.0 / config.Gamma
	if config.Gamma < 0.1 {
		gammaCorrection = 1.0 / 2.2 // Default to standard gamma if invalid
	}

	for i := range calibrated {
		// Clamp to [0, 1] before applying gamma
		r := calibrated[i].R
		if r < 0 {
			r = 0
		}
		if r > 1 {
			r = 1
		}
		g := calibrated[i].G
		if g < 0 {
			g = 0
		}
		if g > 1 {
			g = 1
		}
		b := calibrated[i].B
		if b < 0 {
			b = 0
		}
		if b > 1 {
			b = 1
		}

		calibrated[i].R = math.Pow(r, gammaCorrection)
		calibrated[i].G = math.Pow(g, gammaCorrection)
		calibrated[i].B = math.Pow(b, gammaCorrection)
	}

	// =========== STEP 8: CONVERT BACK TO sRGB 8-bit ===========
	framebuffer := make([]byte, n*3)
	for i, col := range calibrated {
		r, g, b := col.ToSRGB()
		framebuffer[i*3] = r
		framebuffer[i*3+1] = g
		framebuffer[i*3+2] = b
	}

	return framebuffer
}
