package sdl_utils

import (
	"strings"

	"github.com/Zyko0/go-sdl3/sdl"
)

// Toma un valor Int32 y devuelve el String con el name de la key de SDL Scancode
func IntToScancodeName(i int32) string {
	sc := sdl.Scancode(i)
	return sc.Name()
}

// Toma un slice []int32 y devuelve un string compuesto por estos caracteres
func ScancodeNameChain(ic []int32) string {
	var s strings.Builder
	for _, w := range ic {
		s.WriteString(IntToScancodeName(w) + " ")
	}
	return s.String()
}

func (m *MenuSystem) GetBtnModeActive(name string) bool {
	switch name {
	case "Screen":
		if m.Led_s.S.Mode == 0 {
			return true
		}
	case "Audio":
		if m.Led_s.S.Mode == 1 {
			return true
		}
	case "Static":
		if m.Led_s.S.Mode == 2 {
			return true
		}
	}
	return false
}
func (m *MenuSystem) UpdateModeBtns() {
	m.Buttons["ScreenMode"].Active = m.GetBtnModeActive("Screen")
	m.Buttons["ScreenMode"].UpdateColor()

	m.Buttons["AudioMode"].Active = m.GetBtnModeActive("Audio")
	m.Buttons["AudioMode"].UpdateColor()

	m.Buttons["StaticMode"].Active = m.GetBtnModeActive("Static")
	m.Buttons["StaticMode"].UpdateColor()
}

func (m *MenuSystem) RestartLedsScales() {
	pW := m.Led_s.Width / 4
	pH := m.Led_s.Height / 4
	pP := m.Led_s.S.Padding / 4

	newH := m.Height
	if m.Height < pH-20 {
		newH = pH + 20
	}
	m.Window.SetSize(int32(m.Width+pW+20), int32(newH))

	// Escalamos los pixeles para su presentacion de sdl
	m.Led_s.ScaledPoints = ScalePoints(
		m.Led_s.Points,
		m.Led_s.Width, m.Led_s.Height, // resolución real
		pW, pH, // ventana SDL
		pP,
	)
	m.Led_s.ScaledLines = ScalePixelLines(m.Led_s.PixelLines, m.Led_s.Width, m.Led_s.Height, pW, pH, pP)

	m.Led_s.ScaledPointsCinema = ScalePoints(
		m.Led_s.PointsCinema,
		m.Led_s.Width, m.Led_s.Height,
		pW, pH,
		pP,
	)
	m.Led_s.ScaledLinesCinema = ScalePixelLines(m.Led_s.PixelCinema, m.Led_s.Width, m.Led_s.Height, pW, pH, pP)
}
