package screen

import (
	"math"
	"strconv"
	"strings"
)

// Reinicia todos los valores de los puntos, lineas y escalas.
func (l *LedsManager) Restart() {
	inner, outer := GetInnerOuterVals(l.Width, l.Height, l.S.LedsCount, l.S.Padding, l.S.LineLen, l.S.StartPoint)
	pixelLines := BuildPixelLinesBetweenPerimeters(
		outer,
		inner,
		l.Width,
		l.Height,
		l.S.LineThickness,
	)
	l.Points = outer
	l.PixelLines = pixelLines
	l.SampleLines = PixelLinesToSampleLines(pixelLines, l.Width)

	o_cine := ApplyCinemaPadding(outer, l.Width, l.Height, l.S.Padding, l.S.CinePaddingY)
	i_cine := ApplyCinemaPadding(inner, l.Width, l.Height, l.S.Padding+l.S.LineLen, l.S.CinePaddingY+(l.S.LineLen))
	pixelLinesCine := BuildPixelLinesBetweenPerimeters(
		o_cine,
		i_cine,
		l.Width,
		l.Height,
		l.S.LineThickness,
	)
	l.PointsCinema = o_cine
	l.PixelCinema = pixelLinesCine
	l.LinesCinema = PixelLinesToSampleLines(pixelLinesCine, l.Width)
}

func RasterizeThickLine(
	a, b Point,
	thickness int,
	w, h int,
) []Point {

	var pixels []Point

	dx := b.X - a.X
	dy := b.Y - a.Y
	steps := int(math.Max(math.Abs(float64(dx)), math.Abs(float64(dy))))

	if steps == 0 {
		return pixels
	}

	for i := 0; i <= steps; i++ {
		t := float64(i) / float64(steps)
		x := int(float64(a.X) + float64(dx)*t)
		y := int(float64(a.Y) + float64(dy)*t)

		// normal perpendicular
		nx := -dy
		ny := dx

		lenN := math.Hypot(float64(nx), float64(ny))
		if lenN == 0 {
			continue
		}

		nxF := float64(nx) / lenN
		nyF := float64(ny) / lenN

		for o := -thickness / 2; o <= thickness/2; o++ {
			xx := int(float64(x) + nxF*float64(o))
			yy := int(float64(y) + nyF*float64(o))

			if xx >= 0 && yy >= 0 && xx < w && yy < h {
				pixels = append(pixels, Point{xx, yy})
			}
		}
	}

	return pixels
}

func BuildPixelLinesBetweenPerimeters(
	outer []Point,
	inner []Point,
	w, h int,
	thickness int,
) []PixelLine {
	n := min(len(outer), len(inner))
	lines := make([]PixelLine, n)
	for i := range n {
		pixels := RasterizeThickLine(
			outer[i],
			inner[i],
			thickness,
			w,
			h,
		)
		lines[i] = PixelLine{Pixels: pixels}
	}
	return lines
}

func PixelLinesToSampleLines(
	lines []PixelLine,
	w int,
) []SampleLine {

	out := make([]SampleLine, len(lines))

	for i, l := range lines {
		offsets := make([]int, len(l.Pixels))

		for j, p := range l.Pixels {
			offsets[j] = (p.Y*w + p.X) * 4
		}

		out[i] = SampleLine{Offsets: offsets}
	}

	return out
}

func DetectSide(p Point, x0, y0, x1, y1 int) Side {
	switch {
	case p.Y == y0:
		return Top
	case p.X == x1:
		return Right
	case p.Y == y1:
		return Bottom
	default:
		return Left
	}
}

func GetValuesColor(r, g, b int, pts int) string {
	var sb strings.Builder
	sb.Grow(pts * 12)
	for range pts {
		sb.WriteString(strconv.Itoa(r))
		sb.WriteByte('-')
		sb.WriteString(strconv.Itoa(g))
		sb.WriteByte('-')
		sb.WriteString(strconv.Itoa(b))
		sb.WriteByte(' ')
	}
	return sb.String()
}

func GetValues(r, g, b int, pts int) []byte {
	frame := make([]byte, pts*3)
	for i := range pts {
		frame[i*3] = byte(r)
		frame[i*3+1] = byte(g)
		frame[i*3+2] = byte(b)
	}
	return frame
}

func (l *LedsManager) GetLedValues(img []byte, w, h int, pts []Point) []byte {
	switch l.S.PixelMethod {
	case 0:
		return l.ComputeKernelLEDValues(img, w, h, pts)
	case 1:
		return l.ComputeLineLEDValues(img)
	default:
		return l.ComputeSimpleLEDValues(img, w, h, pts)
	}
}

func (l *LedsManager) ComputeSimpleLEDValues(img []byte, w, h int, pts []Point) []byte {
	frame := make([]byte, len(pts)*3)
	for i, p := range pts {
		idx := (p.Y*w + p.X) * 4
		b := int(img[idx])
		g := int(img[idx+1])
		r := int(img[idx+2])

		r, g, b = AdjustColorCalibration(r, g, b, l.S.RCal, l.S.GCal, l.S.BCal, l.S.Temperature, l.S.Brightness, l.S.Saturation, l.S.Gamma)

		frame[i*3] = byte(r)
		frame[i*3+1] = byte(g)
		frame[i*3+2] = byte(b)
	}
	return frame
}

func (l *LedsManager) ComputeKernelLEDValues(img []byte, w, h int, pts []Point) []byte {
	frame := make([]byte, len(pts)*3)
	for i, p := range pts {
		r, g, b := l.AvgColorKernel(img, w, h, p.X, p.Y)
		r, g, b = AdjustColorCalibration(r, g, b, l.S.RCal, l.S.GCal, l.S.BCal, l.S.Temperature, l.S.Brightness, l.S.Saturation, l.S.Gamma)

		frame[i*3] = byte(r)
		frame[i*3+1] = byte(g)
		frame[i*3+2] = byte(b)
	}
	return frame
}

func (l *LedsManager) computeLine(line SampleLine, img []byte) (int, int, int) {
	var rs, gs, bs int
	for _, off := range line.Offsets {
		bs += int(img[off])
		gs += int(img[off+1])
		rs += int(img[off+2])
	}
	n := len(line.Offsets)
	if n > 0 {
		rs /= n
		gs /= n
		bs /= n
	}
	rs, gs, bs = AdjustColorCalibration(rs, gs, bs, l.S.RCal, l.S.GCal, l.S.BCal, l.S.Temperature, l.S.Brightness, l.S.Saturation, l.S.Gamma)
	return rs, gs, bs
}

func (l *LedsManager) ComputeLineLEDValues(
	img []byte,
) []byte {
	frame := make([]byte, len(l.LinesCinema)*3)
	// computa las lineas del modo cine
	if l.S.Cinema {
		for i, line := range l.LinesCinema {
			rs, gs, bs := l.computeLine(line, img)
			frame[i*3] = byte(rs)
			frame[i*3+1] = byte(gs)
			frame[i*3+2] = byte(bs)
		}
		return frame
	}
	for i, line := range l.SampleLines {
		rs, gs, bs := l.computeLine(line, img)
		frame[i*3] = byte(rs)
		frame[i*3+1] = byte(gs)
		frame[i*3+2] = byte(bs)
	}
	return frame
}

func (l *LedsManager) AvgColorKernel(
	img []byte,
	w, h int,
	x, y int,
) (r, g, b int) {
	radius := l.S.KernelSize >> 1
	var rs, gs, bs, count int
	y0 := max(y-radius, 0)
	y1 := y + radius
	if y1 >= h {
		y1 = h - 1
	}
	x0 := max(x-radius, 0)
	x1 := x + radius
	if x1 >= w {
		x1 = w - 1
	}
	for yy := y0; yy <= y1; yy++ {
		row := yy * w * 4
		for xx := x0; xx <= x1; xx++ {
			i := row + xx*4
			bs += int(img[i])
			gs += int(img[i+1])
			rs += int(img[i+2])
			count++
		}
	}
	if count == 0 {
		return 0, 0, 0
	}
	return rs / count, gs / count, bs / count
}

func clamp(v int) int {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return v
}

// clampF restringe un valor float a un rango mínimo y máximo
func clampF(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

// AdjustColorCalibration aplica calibración RGB profesional en 3 pasos
//
// Paso 1: RGB Gains (ganancias independientes)
//
//	R_out = R_in * gain_R
//	G_out = G_in * gain_G
//	B_out = B_in * gain_B
//	Corrige: diferencias de brillo entre LEDs, dominantes de color, desbalance del panel
//
// Paso 2: Corrección de Temperatura de Color (CT)
//
//	Rango: -1.0 (frío/azul) a +1.0 (cálido/rojo)
//	Solo afecta R y B inversamente, manteniendo G estable
//	Simula cambios en Kelvin (ej: 3000K a 6500K)
//
// Paso 3: Brillo
//
//	Multiplicador final que aplica a todos los canales
//
// Método estándar en monitores profesionales, calibradores de cámara y postproducción
func AdjustColorCalibration(r, g, b int, rCal, gCal, bCal, temperature, brightness, saturation, gamma float64) (int, int, int) {
	rF := float64(r)
	gF := float64(g)
	bF := float64(b)

	// === PASO 1: Aplicar ganancias RGB independientes ===
	rF = rF * rCal
	gF = gF * gCal
	bF = bF * bCal

	// === PASO 2: Aplicar corrección de temperatura de color ===
	temp := clampF(temperature, -1.0, 1.0)
	if temp >= 0 {
		rF = rF * (1.0 + temp*0.5)
		bF = bF * (1.0 - temp*0.8)
	} else {
		rF = rF * (1.0 + temp*0.3)
		bF = bF * (1.0 - temp*0.5)
	}

	// === PASO 3: Aplicar brillo ===
	rF = rF * brightness
	gF = gF * brightness
	bF = bF * brightness

	// === PASO 4: Ajuste de saturación (en espacio RGB lineal aproximado) ===
	// Normalizar a [0,1]
	rn := clampFloat(rF/255.0, 0.0, 1.0)
	gn := clampFloat(gF/255.0, 0.0, 1.0)
	bn := clampFloat(bF/255.0, 0.0, 1.0)

	// Luminancia aproximada (Rec.709)
	lum := 0.2126*rn + 0.7152*gn + 0.0722*bn
	sat := clampF(saturation, 0.0, 4.0)
	rn = lum + (rn-lum)*sat
	gn = lum + (gn-lum)*sat
	bn = lum + (bn-lum)*sat

	// === PASO 5: Corrección gamma de salida ===
	gCorr := gamma
	if gCorr <= 0 {
		gCorr = 2.2
	}
	// Aplicar gamma (convertir lineal a espacio perceptual)
	rn = clampFloat(math.Pow(clampFloat(rn, 0.0, 1.0), 1.0/gCorr), 0.0, 1.0)
	gn = clampFloat(math.Pow(clampFloat(gn, 0.0, 1.0), 1.0/gCorr), 0.0, 1.0)
	bn = clampFloat(math.Pow(clampFloat(bn, 0.0, 1.0), 1.0/gCorr), 0.0, 1.0)

	return clamp(int(rn * 255.0)), clamp(int(gn * 255.0)), clamp(int(bn * 255.0))
}

// clampFloat limita un float en un rango
func clampFloat(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

// AdjustTemperature (deprecado) - mantener por compatibilidad
func AdjustTemperature(r, g, b int, t float64) (int, int, int) {
	// Usar calibración por defecto (1.0 para cada canal)
	return AdjustColorCalibration(r, g, b, 1.0, 1.0, 1.0, t, 1.0, 1.0, 2.2)
}

// AdjustBrightness (deprecado) - mantener por compatibilidad
func AdjustBrightness(r, g, b int, a float64) (int, int, int) {
	return int(float64(r) * a), int(float64(g) * a), int(float64(b) * a)
}
