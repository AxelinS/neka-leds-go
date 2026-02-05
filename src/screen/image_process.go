package screen

import (
	"go-neka-leds/src/esp32"
	"go-neka-leds/src/settings"
	"math"
	"strconv"
	"strings"
)

type MonitorSettings struct {
	Width, Height int
}

type LedsManager struct {
	Devs []esp32.ESP32
	MonitorSettings
	CountSide      SideCount
	Pause          bool
	WinCaptureMode int

	S settings.Settings

	Points       []Point
	ScaledPoints []Point

	SampleLines []SampleLine
	PixelLines  []PixelLine
	ScaledLines []PixelLine

	Cinema             bool
	PointsCinema       []Point
	ScaledPointsCinema []Point
	LinesCinema        []SampleLine
	PixelCinema        []PixelLine
	ScaledLinesCinema  []PixelLine
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

// Reinicia todos los valores de los puntos, lineas y escalas.
func (l *LedsManager) Restart() {
	inner, outer := GetInnerOuterVals(l.Width, l.Height, l.S.LedsCount, l.S.Padding, l.S.LineLen)
	pixelLines := BuildPixelLinesBetweenPerimeters(
		outer,
		inner,
		l.Width,
		l.Height,
		l.S.LineTickness,
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
		l.S.LineTickness,
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

func (l *LedsManager) GetLedValues(img []byte, w, h int, pts []Point) string {
	switch l.S.Mode {
	case 0:
		return l.ComputeKernelLEDValues(img, w, h, pts)
	case 1:
		return l.ComputeLineLEDValues(img)
	default:
		return l.ComputeSimpleLEDValues(img, w, h, pts)
	}
}

func (l *LedsManager) ComputeSimpleLEDValues(img []byte, w, h int, pts []Point) string {
	var sb strings.Builder
	sb.Grow(len(pts) * 12)

	for _, p := range pts {
		idx := (p.Y*w + p.X) * 4
		b := int(img[idx])
		g := int(img[idx+1])
		r := int(img[idx+2])

		r, g, b = AdjustTemperature(r, g, b, l.S.Temperature)
		r, g, b = AdjustBrightness(r, g, b, l.S.Brightness)

		sb.WriteString(strconv.Itoa(r))
		sb.WriteByte('-')
		sb.WriteString(strconv.Itoa(g))
		sb.WriteByte('-')
		sb.WriteString(strconv.Itoa(b))
		sb.WriteByte(' ')
	}
	return sb.String()
}

func (l *LedsManager) ComputeKernelLEDValues(img []byte, w, h int, pts []Point) string {
	var sb strings.Builder
	sb.Grow(len(pts) * 12)

	for _, p := range pts {
		r, g, b := l.AvgColorKernel(img, w, h, p.X, p.Y)
		r, g, b = AdjustTemperature(r, g, b, l.S.Temperature)
		r, g, b = AdjustBrightness(r, g, b, l.S.Brightness)

		sb.WriteString(strconv.Itoa(r))
		sb.WriteByte('-')
		sb.WriteString(strconv.Itoa(g))
		sb.WriteByte('-')
		sb.WriteString(strconv.Itoa(b))
		sb.WriteByte(' ')
	}
	return sb.String()
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
	rs, gs, bs = AdjustTemperature(rs, gs, bs, l.S.Temperature)
	rs, gs, bs = AdjustBrightness(rs, gs, bs, l.S.Brightness)
	return rs, gs, bs
}

func (l *LedsManager) ComputeLineLEDValues(
	img []byte,
) string {

	var sb strings.Builder
	sb.Grow(len(l.SampleLines) * 12)

	// computa las lineas del modo cine
	if l.Cinema {
		for _, line := range l.LinesCinema {
			rs, gs, bs := l.computeLine(line, img)
			sb.WriteString(strconv.Itoa(rs))
			sb.WriteByte('-')
			sb.WriteString(strconv.Itoa(gs))
			sb.WriteByte('-')
			sb.WriteString(strconv.Itoa(bs))
			sb.WriteByte(' ')
		}
		return sb.String()
	}

	for _, line := range l.SampleLines {
		rs, gs, bs := l.computeLine(line, img)
		sb.WriteString(strconv.Itoa(rs))
		sb.WriteByte('-')
		sb.WriteString(strconv.Itoa(gs))
		sb.WriteByte('-')
		sb.WriteString(strconv.Itoa(bs))
		sb.WriteByte(' ')
	}

	return sb.String()
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
