package sdl_utils

import (
	"fmt"
	"go-neka-leds/src/screen"
	"go-neka-leds/src/utils"
	"strconv"
	"time"

	"github.com/Zyko0/go-sdl3/sdl"
)

func ComputeScale(
	srcW, srcH int,
	dstW, dstH int,
) (float64, float64) {
	return float64(dstW) / float64(srcW),
		float64(dstH) / float64(srcH)
}
func ScalePixelLines(
	lines []screen.PixelLine,
	srcW, srcH int,
	dstW, dstH int,
	padding int,
) []screen.PixelLine {
	scaleX := float64(dstW-2*padding) / float64(srcW)
	scaleY := float64(dstH-2*padding) / float64(srcH)
	out := make([]screen.PixelLine, len(lines))
	for i, l := range lines {
		sp := make([]screen.Point, len(l.Pixels))

		for j, p := range l.Pixels {
			sp[j] = screen.Point{
				X: padding + int(float64(p.X)*scaleX),
				Y: padding + int(float64(p.Y)*scaleY),
			}
		}

		out[i] = screen.PixelLine{Pixels: sp}
	}
	return out
}

func ScalePoints(
	points []screen.Point,
	srcW, srcH int, // resolución real
	dstW, dstH int, // ventana SDL
	padding int,
) []screen.Point {
	scaleX := float64(dstW-2*padding) / float64(srcW)
	scaleY := float64(dstH-2*padding) / float64(srcH)

	out := make([]screen.Point, len(points))
	for i, p := range points {
		out[i] = screen.Point{
			X: padding + int(float64(p.X)*scaleX),
			Y: padding + int(float64(p.Y)*scaleY),
		}
	}
	return out
}

func (m *MenuSystem) ModeTestPoints() {
	m.Led_s.Pause = !m.Led_s.Pause

	values := screen.GetValuesColor(0, 0, 180, m.Led_s.CountSide.Top)
	values += screen.GetValuesColor(180, 0, 0, m.Led_s.CountSide.Right)
	values += screen.GetValuesColor(180, 180, 180, m.Led_s.CountSide.Bottom)
	values += screen.GetValuesColor(0, 180, 0, m.Led_s.CountSide.Left)
	// Pintamos de diferente color los leds de cada lado
	time.Sleep(500 * time.Millisecond)
	for _, dev := range m.Led_s.Devs {
		if dev.Connected {
			dev.SafeWrite("RGB " + values + "\n")
		} else {
			if dev.Reconnect() {
				fmt.Println("[RECONNECTED]", dev.Id)
			}
		}
	}
}

func (m *MenuSystem) renderPoints(renderer *sdl.Renderer) {
	spp := 200 // start of preview page
	pW := m.Led_s.Width / 4
	pH := m.Led_s.Height / 4

	top_padding := float32(10)
	left_padding := float32(10)

	renderer.SetDrawColor(utils.ColorGris.R, utils.ColorGris.G, utils.ColorGris.B, utils.ColorGris.A)
	if m.Led_s.Pause {
		renderer.SetDrawColor(0, 0, 255, 255)
	}
	renderer.RenderLine(float32(spp)+left_padding, top_padding, float32(pW+spp)+left_padding, float32(1)+top_padding) // line top
	if m.Led_s.Pause {
		renderer.SetDrawColor(255, 0, 0, 255)
	}
	renderer.RenderLine(float32(pW+spp)+left_padding, top_padding+2, float32(pW+spp-1)+left_padding, float32(pH-2)+top_padding) // line right
	if m.Led_s.Pause {
		renderer.SetDrawColor(255, 255, 255, 255)
	}
	renderer.RenderLine(float32(spp)+left_padding, float32(pH)+top_padding, float32(pW+spp)+left_padding, float32(pH-1)+top_padding) // line bottom
	if m.Led_s.Pause {
		renderer.SetDrawColor(0, 255, 0, 255)
	}
	renderer.RenderLine(float32(spp)+left_padding, top_padding+2, float32(spp+1)+left_padding, float32(pH-2)+top_padding) // line left

	renderer.SetDrawColor(utils.ColorBlanco.R, utils.ColorBlanco.G, utils.ColorBlanco.B, utils.ColorBlanco.A)
	renderer.DebugText(float32(spp+70)+left_padding, 5, "Top: "+strconv.Itoa(m.Led_s.CountSide.Top))
	renderer.DebugText(float32(spp+10)+left_padding, 20, "Left: "+strconv.Itoa(m.Led_s.CountSide.Left))

	// Mostramos los pixeles del modo cine solo si está activado
	if m.Led_s.Cinema {
		for _, p := range m.Led_s.ScaledPointsCinema {
			renderer.RenderLine(float32(p.X+spp)+left_padding, float32(p.Y)+top_padding, float32(p.X+spp+1)+left_padding, float32(p.Y+1)+top_padding)
		}
		return
	}

	for _, p := range m.Led_s.ScaledPoints {
		renderer.RenderLine(float32(p.X+spp)+left_padding, float32(p.Y)+top_padding, float32(p.X+spp+1)+left_padding, float32(p.Y+1)+top_padding)
	}
}

func (m *MenuSystem) renderLines(renderer *sdl.Renderer) {
	spp := 200
	pW := m.Led_s.Width / 4
	pH := m.Led_s.Height / 4

	top_padding := float32(10)
	left_padding := float32(10)
	renderer.SetDrawColor(utils.ColorGris.R, utils.ColorGris.G, utils.ColorGris.B, utils.ColorGris.A)
	if m.Led_s.Pause {
		renderer.SetDrawColor(0, 0, 255, 255)
	}
	renderer.RenderLine(float32(spp)+left_padding, top_padding+2, float32(pW+spp)+left_padding, float32(1)+top_padding) // line top
	if m.Led_s.Pause {
		renderer.SetDrawColor(255, 0, 0, 255)
	}
	renderer.RenderLine(float32(pW+spp)+left_padding, top_padding, float32(pW+spp-1)+left_padding, float32(pH-2)+top_padding) // line right
	if m.Led_s.Pause {
		renderer.SetDrawColor(255, 255, 255, 255)
	}
	renderer.RenderLine(float32(spp)+left_padding, float32(pH)+top_padding, float32(pW+spp)+left_padding, float32(pH-1)+top_padding) // line bottom
	if m.Led_s.Pause {
		renderer.SetDrawColor(0, 255, 0, 255)
	}
	renderer.RenderLine(float32(spp)+left_padding, top_padding+2, float32(spp+1)+left_padding, float32(pH-2)+top_padding) // line left

	renderer.SetDrawColor(utils.ColorBlanco.R, utils.ColorBlanco.G, utils.ColorBlanco.B, utils.ColorBlanco.A)

	cTop := m.Led_s.CountSide.Top
	cRight := cTop + m.Led_s.CountSide.Right
	cBottom := cRight + m.Led_s.CountSide.Bottom
	cLeft := cBottom + m.Led_s.CountSide.Left

	if m.Led_s.Cinema {
		for c, line := range m.Led_s.ScaledLinesCinema {
			if c < cTop {
				renderer.SetDrawColor(0, 0, 255, 255)
			} else if c < cRight {
				renderer.SetDrawColor(255, 0, 0, 255)
			} else if c < cBottom {
				renderer.SetDrawColor(255, 255, 255, 255)
			} else if c < cLeft {
				renderer.SetDrawColor(0, 255, 0, 255)
			}
			for i := 1; i < len(line.Pixels); i++ {
				a := line.Pixels[i-1]
				b := line.Pixels[i]

				renderer.RenderLine(
					float32(a.X+spp)+left_padding,
					float32(a.Y)+top_padding,
					float32(b.X+spp)+left_padding,
					float32(b.Y)+top_padding,
				)
			}
		}
		return
	}

	for c, line := range m.Led_s.ScaledLines {
		if c < cTop {
			renderer.SetDrawColor(0, 0, 255, 255)
		} else if c < cRight {
			renderer.SetDrawColor(255, 0, 0, 255)
		} else if c < cBottom {
			renderer.SetDrawColor(255, 255, 255, 255)
		} else if c < cLeft {
			renderer.SetDrawColor(0, 255, 0, 255)
		}
		for i := 1; i < len(line.Pixels); i++ {
			a := line.Pixels[i-1]
			b := line.Pixels[i]

			renderer.RenderLine(
				float32(a.X+spp)+left_padding,
				float32(a.Y)+top_padding,
				float32(b.X+spp)+left_padding,
				float32(b.Y)+top_padding,
			)
		}
	}
}
