package screen

import (
	"go-neka-leds/src/esp32"
	"strconv"
	"strings"
)

type MonitorSettings struct {
	Width, Height int
}

type LedSettings struct {
	Devs []esp32.ESP32
	MonitorSettings
	CountSide SideCount
	Pause     bool
	LedsCount int

	Temperature float64
	Brightness  float64
	Mode        int

	KernelSize int
	Padding    int

	LineLen      int
	SampleLines  []SampleLine
	ScaledLines  []PixelLine
	Points       []Point
	ScaledPoints []Point
}

type Side int

const (
	Top Side = iota
	Right
	Bottom
	Left
)

type SampleLine struct {
	Offsets []int
}

func BuildPixelLines(
	pts []Point,
	w, h int,
	padding int,
	lineLen int,
) []PixelLine {

	x0, y0 := padding, padding
	x1, y1 := w-padding, h-padding

	lines := make([]PixelLine, len(pts))

	for i, p := range pts {
		side := DetectSide(p, x0, y0, x1, y1)

		var dx, dy int
		switch side {
		case Top:
			dx, dy = 0, 1
		case Bottom:
			dx, dy = 0, -1
		case Left:
			dx, dy = 1, 0
		case Right:
			dx, dy = -1, 0
		}

		pixels := make([]Point, 0, lineLen)

		for k := 0; k < lineLen; k++ {
			xx := p.X + dx*k
			yy := p.Y + dy*k

			if xx < 0 || yy < 0 || xx >= w || yy >= h {
				break
			}

			pixels = append(pixels, Point{xx, yy})
		}

		lines[i] = PixelLine{Pixels: pixels}
	}

	return lines
}

func BuildSampleLines(
	pts []Point,
	w, h int,
	padding int,
	lineLen int,
) []SampleLine {

	x0, y0 := padding, padding
	x1, y1 := w-padding, h-padding

	lines := make([]SampleLine, len(pts))

	for i, p := range pts {
		side := DetectSide(p, x0, y0, x1, y1)
		var dx, dy int
		switch side {
		case Top:
			dx, dy = 0, 1
		case Bottom:
			dx, dy = 0, -1
		case Left:
			dx, dy = 1, 0
		case Right:
			dx, dy = -1, 0
		}
		offsets := make([]int, 0, lineLen)
		for k := range lineLen {
			xx := p.X + dx*k
			yy := p.Y + dy*k
			if xx < 0 || yy < 0 || xx >= w || yy >= h {
				break
			}
			offsets = append(offsets, (yy*w+xx)*4)
		}
		lines[i] = SampleLine{Offsets: offsets}
	}
	return lines
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

func (l *LedSettings) GetLedValues(img []byte, w, h int, pts []Point) string {
	switch l.Mode {
	case 0:
		return l.ComputeKernelLEDValues(img, w, h, pts)
	case 1:
		return l.ComputeLineLEDValues(img)
	default:
		return l.ComputeSimpleLEDValues(img, w, h, pts)
	}
}

func (l *LedSettings) ComputeSimpleLEDValues(img []byte, w, h int, pts []Point) string {
	var sb strings.Builder
	sb.Grow(len(pts) * 12)

	for _, p := range pts {
		idx := (p.Y*w + p.X) * 4
		b := int(img[idx])
		g := int(img[idx+1])
		r := int(img[idx+2])

		r, g, b = AdjustTemperature(r, g, b, l.Temperature)
		r, g, b = AdjustBrightness(r, g, b, l.Brightness)

		sb.WriteString(strconv.Itoa(r))
		sb.WriteByte('-')
		sb.WriteString(strconv.Itoa(g))
		sb.WriteByte('-')
		sb.WriteString(strconv.Itoa(b))
		sb.WriteByte(' ')
	}
	return sb.String()
}

func (l *LedSettings) ComputeKernelLEDValues(img []byte, w, h int, pts []Point) string {
	var sb strings.Builder
	sb.Grow(len(pts) * 12)

	for _, p := range pts {
		r, g, b := l.AvgColorKernel(img, w, h, p.X, p.Y)
		r, g, b = AdjustTemperature(r, g, b, l.Temperature)
		r, g, b = AdjustBrightness(r, g, b, l.Brightness)

		sb.WriteString(strconv.Itoa(r))
		sb.WriteByte('-')
		sb.WriteString(strconv.Itoa(g))
		sb.WriteByte('-')
		sb.WriteString(strconv.Itoa(b))
		sb.WriteByte(' ')
	}
	return sb.String()
}

func (l *LedSettings) ComputeLineLEDValues(
	img []byte,
) string {

	var sb strings.Builder
	sb.Grow(len(l.SampleLines) * 12)

	for _, line := range l.SampleLines {
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

		rs, gs, bs = AdjustTemperature(rs, gs, bs, l.Temperature)
		rs, gs, bs = AdjustBrightness(rs, gs, bs, l.Brightness)

		sb.WriteString(strconv.Itoa(rs))
		sb.WriteByte('-')
		sb.WriteString(strconv.Itoa(gs))
		sb.WriteByte('-')
		sb.WriteString(strconv.Itoa(bs))
		sb.WriteByte(' ')
	}

	return sb.String()
}

func (l *LedSettings) AvgColorKernel(
	img []byte,
	w, h int,
	x, y int,
) (r, g, b int) {
	radius := l.KernelSize >> 1
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

func AdjustTemperature(r, g, b int, t float64) (int, int, int) {
	if t > 1 {
		t = 1
	} else if t < -1 {
		t = -1
	}
	var rGain, gGain, bGain float64
	if t >= 0 {
		rGain = 1.0 + 0.6*t
		gGain = 1.0 - 0.1*t
		bGain = 1.0 - 0.8*t
	} else {
		rGain = 1.0 + 0.5*t
		gGain = 1.0 - 0.1*t
		bGain = 1.0 - 0.2*t
	}
	return clamp(int(float64(r) * rGain)),
		clamp(int(float64(g) * gGain)),
		clamp(int(float64(b) * bGain))
}

func AdjustBrightness(r, g, b int, a float64) (int, int, int) {
	return int(float64(r) * a), int(float64(g) * a), int(float64(b) * a)
}
