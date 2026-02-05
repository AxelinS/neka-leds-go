package sdl_utils

import "github.com/Zyko0/go-sdl3/sdl"

// Toma un valor Int32 y devuelve el String con el name de la key de SDL Scancode
func IntToScancodeName(i int32) string {
	sc := sdl.Scancode(i)
	return sc.Name()
}

// Toma un slice []int32 y devuelve un string compuesto por estos caracteres
func ScancodeNameChain(ic []int32) string {
	s := ""
	for _, w := range ic {
		s += IntToScancodeName(w) + " "
	}
	return s
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
