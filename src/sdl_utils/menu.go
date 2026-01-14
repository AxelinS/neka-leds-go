package sdl_utils

import (
	"go-neka-leds/src/screen"
	"go-neka-leds/src/utils"
	"log"
	"math"
	"strconv"
	"time"

	"github.com/Zyko0/go-sdl3/sdl"
)

const (
	MenuPrincipal MenuState = iota
	MenuConfiguracion
)

// Inicializar el sistema de menu
func NewMenuSystem(winConfig WindowConfig, window *sdl.Window, led_settings *screen.LedSettings) *MenuSystem {
	tc := TitleConfig{
		X:     30,
		Y:     10,
		Scale: 1.3,
	}
	menu := &MenuSystem{
		WindowConfig:  winConfig,
		Led_s:         led_settings,
		State:         MenuPrincipal,
		AnimationTime: 0,
		FocusedField:  -1,
		TitleConf:     tc,
		Window:        window,
		Blocker:       false,
		Keys:          make([]bool, 512),
		Fotogramas:    uint32(60),
		VerModo:       false,
	}

	menu.setupMainMenu()
	return menu
}

// Configurar menu principal
func (m *MenuSystem) setupMainMenu() {
	m.Buttons = []AnimatedButton{
		{
			X: 10, Y: 50, Width: 150, Height: 50,
			Text:       "Modo Test",
			Color:      utils.ColorAzulTron,
			HoverColor: utils.ColorAzulNeon,
			Scale:      1.0, Alpha: 255,
			Action: m.ModeTestPoints,
		},
		{
			X: 10, Y: 120, Width: 150, Height: 50,
			Text:       "Ver modo",
			Color:      utils.ColorAzulTron,
			HoverColor: utils.ColorAzulNeon,
			Scale:      1.0, Alpha: 255,
			Action: func() {
				m.VerModo = !m.VerModo
			},
		},
		{
			X: 10, Y: 190, Width: 150, Height: 50,
			Text:       "Configuracion",
			Color:      utils.ColorAzulTron,
			HoverColor: utils.ColorAzulNeon,
			Scale:      1.0, Alpha: 255,
			Action: m.openConfigMenu,
		},
	}
	m.InputFields = []InputField{
		{
			X: 10, Y: 260, Width: 130, Height: 40,
			Text:        "30",
			Placeholder: "FPS",
			Title:       "FPS",
			Type:        0,
		},
	}
}

// Configurar menu de configuracion
func (m *MenuSystem) setupConfigMenu() {
	m.Buttons = []AnimatedButton{
		{
			X: 180, Y: 5, Width: 55, Height: 28,
			Text:       "Menu",
			Color:      utils.ColorAzulTron,
			HoverColor: utils.ColorAzulNeon,
			Scale:      1.0, Alpha: 255,
			Action: m.backToMainMenu,
		},
	}
	m.InputFields = []InputField{
		{
			X: 10, Y: 60, Width: 230, Height: 40,
			Text:        "...",
			Placeholder: "Combinacion de teclas",
			Title:       "Modo comando",
			Type:        1, // Combinaciones
		}, {
			X: 10, Y: 120, Width: 230, Height: 40,
			Text:        "...",
			Placeholder: "Comando por defecto",
			Title:       "Texto de comando",
			Type:        0, // Texto
		},
	}
}

// Actualizar animaciones y logica del menu
func (m *MenuSystem) Update(deltaTime float64) {
	m.AnimationTime += deltaTime

	// Actualizar animaciones de botones
	for i := range m.Buttons {
		btn := &m.Buttons[i]

		// Animacion de hover
		if btn.IsHovered {
			btn.Scale = float32(1.0 + 0.1*math.Sin(m.AnimationTime*8))
			btn.GlowIntensity = float32(0.5 + 0.3*math.Sin(m.AnimationTime*4))
		} else {
			btn.Scale = 1.0
			btn.GlowIntensity = 0.0
		}

		// Animacion de pulsacion
		if btn.IsPressed {
			btn.Scale *= 0.95
		}
	}

	// Actualizar cursor parpadeante en campos de texto
	for i := range m.InputFields {
		field := &m.InputFields[i]
		if field.IsFocused && time.Since(field.CursorBlink) > 500*time.Millisecond {
			field.CursorBlink = time.Now()
		}
	}
}

// Renderizar el menu
func (m *MenuSystem) Render(renderer *sdl.Renderer) {
	switch m.State {
	case MenuPrincipal:
		m.renderMainMenu(renderer)
	case MenuConfiguracion:
		m.renderConfigMenu(renderer)
	}
}

func (m *MenuSystem) renderDivisionMark(renderer *sdl.Renderer) {
	startOfPreviewPage := 200
	renderer.SetDrawColor(utils.ColorAzulNeon.R, utils.ColorAzulNeon.G, utils.ColorAzulNeon.B, 255)
	renderer.RenderLine(float32(196), float32(0), float32(startOfPreviewPage), float32(m.Height))
}

// Renderizar menu principal
func (m *MenuSystem) renderMainMenu(renderer *sdl.Renderer) {
	// Titulo con glow effect
	m.renderGlowText(renderer, "Neka-Leds", m.TitleConf.X, m.TitleConf.Y, utils.ColorAzulNeon, m.TitleConf.Scale)
	m.renderDivisionMark(renderer)
	// Estado del modo comando con indicador visual
	statusText := "OFF"
	if m.Led_s.Pause {
		statusText = "ON"
	}
	glowIntensity := float32(0)

	// Indicador del estado Command
	m.renderStatusIndicator(renderer, 180, 62, true, glowIntensity)
	renderer.DebugText(170, 80, statusText)

	// Renderizar botones
	for _, btn := range m.Buttons {
		m.renderAnimatedButton(renderer, btn)
	}
	for _, field := range m.InputFields {
		renderer.DebugText(field.X, field.Y-10, field.Title)
		m.renderInputField(renderer, field)
	}

	// Vista previa de puntos
	if m.VerModo {
		m.renderLines(renderer)
	} else {
		m.renderPoints(renderer)
	}
}

// Renderizar menu de configuracion
func (m *MenuSystem) renderConfigMenu(renderer *sdl.Renderer) {
	// Titulo
	m.renderGlowText(renderer, "Configuracion", m.TitleConf.X, m.TitleConf.Y, utils.ColorAzulNeon, m.TitleConf.Scale)

	// Etiquetas usando DebugText
	renderer.SetDrawColor(utils.ColorBlanco.R, utils.ColorBlanco.G, utils.ColorBlanco.B, utils.ColorBlanco.A)
	// Renderizar campos de entrada
	for _, field := range m.InputFields {
		renderer.DebugText(field.X, field.Y-10, field.Title)
		m.renderInputField(renderer, field)
	}

	// Renderizar botones (flecha de retorno)
	for _, btn := range m.Buttons {
		m.renderAnimatedButton(renderer, btn)
	}
}

// Renderizar boton animado
func (m *MenuSystem) renderAnimatedButton(renderer *sdl.Renderer, btn AnimatedButton) {
	// Calcular posicion y tamaño con escala
	scaledW := btn.Width * btn.Scale
	scaledH := btn.Height * btn.Scale
	offsetX := (btn.Width - scaledW) / 2
	offsetY := (btn.Height - scaledH) / 2

	// Efecto de glow si esta en hover
	if btn.GlowIntensity > 0 {
		glowSize := int(btn.GlowIntensity * 10)
		glowAlpha := uint8(btn.GlowIntensity * 50)

		renderer.SetDrawColor(btn.HoverColor.R, btn.HoverColor.G, btn.HoverColor.B, glowAlpha)

		// Crear efecto de glow con multiples rectangulos
		for i := range glowSize {
			rect := sdl.FRect{
				X: btn.X + offsetX - float32(i),
				Y: btn.Y + offsetY - float32(i),
				W: scaledW + float32(i*2),
				H: scaledH + float32(i*2),
			}
			renderer.RenderRect(&rect)
		}
	}

	// Boton principal
	color := btn.Color
	if btn.IsHovered {
		color = btn.HoverColor
	}

	renderer.SetDrawColor(color.R, color.G, color.B, color.A)

	// Relleno del boton
	fillRect := sdl.FRect{
		X: btn.X + offsetX,
		Y: btn.Y + offsetY,
		W: scaledW,
		H: scaledH,
	}
	renderer.RenderFillRect(&fillRect)

	// Borde con animacion
	borderAlpha := uint8(255)
	if btn.IsHovered {
		borderAlpha = uint8(255 * (0.8 + 0.2*math.Sin(m.AnimationTime*6)))
	}

	renderer.SetDrawColor(utils.ColorAzulNeon.R, utils.ColorAzulNeon.G, utils.ColorAzulNeon.B, borderAlpha)
	renderer.RenderRect(&fillRect)

	// Renderizar texto del boton usando DebugText
	textX := btn.X + btn.Width/2 - float32(len(btn.Text)*3) // Aproximacion para centrar
	textY := btn.Y + btn.Height/2 - 8
	renderer.SetDrawColor(utils.ColorBlanco.R, utils.ColorBlanco.G, utils.ColorBlanco.B, 255)
	renderer.DebugText(textX, textY, btn.Text)
}

// Renderizar campo de entrada de texto
func (m *MenuSystem) renderInputField(renderer *sdl.Renderer, field InputField) {
	// Fondo del campo
	bgColor := utils.ColorAzulOscuro
	if field.IsFocused {
		bgColor = utils.ColorAzulTron
	}

	renderer.SetDrawColor(bgColor.R, bgColor.G, bgColor.B, bgColor.A)

	fillRect := sdl.FRect{X: field.X, Y: field.Y, W: field.Width, H: field.Height}
	renderer.RenderFillRect(&fillRect)

	// Borde
	borderColor := utils.ColorGris
	if field.IsFocused {
		borderColor = utils.ColorAzulNeon
		// Efecto de glow para campo activo
		glowAlpha := uint8(100 * (0.5 + 0.5*math.Sin(m.AnimationTime*4)))
		renderer.SetDrawColor(utils.ColorAzulNeon.R, utils.ColorAzulNeon.G, utils.ColorAzulNeon.B, glowAlpha)

		glowRect := sdl.FRect{
			X: field.X - 2, Y: field.Y - 2,
			W: field.Width + 4, H: field.Height + 4,
		}
		renderer.RenderRect(&glowRect)
	}

	renderer.SetDrawColor(borderColor.R, borderColor.G, borderColor.B, borderColor.A)
	renderer.RenderRect(&fillRect)

	// Renderizar texto del campo
	renderer.SetDrawColor(utils.ColorBlanco.R, utils.ColorBlanco.G, utils.ColorBlanco.B, 255)
	if field.Text != "" {
		renderer.DebugText(field.X+5, field.Y+10, field.Text)
	} else if !field.IsFocused {
		renderer.SetDrawColor(utils.ColorGris.R, utils.ColorGris.G, utils.ColorGris.B, 150)
		renderer.DebugText(field.X+5, field.Y+10, field.Placeholder)
	}

	// Cursor parpadeante (si esta enfocado)
	if field.IsFocused && time.Since(field.CursorBlink) < 250*time.Millisecond {
		cursorX := field.X + 4 + float32(len(field.Text)*8) // Aproximacion
		renderer.SetDrawColor(utils.ColorBlanco.R, utils.ColorBlanco.G, utils.ColorBlanco.B, 255)
		renderer.RenderLine(cursorX, field.Y+5, cursorX, field.Y+field.Height-5)
	}
}

// Renderizar indicador de estado circular
func (m *MenuSystem) renderStatusIndicator(renderer *sdl.Renderer, x, y float32, active bool, glowIntensity float32) {
	centerX, centerY := x, y
	radius := float32(8)

	color := utils.ColorGris
	if active {
		color = utils.ColorAzulNeon

		// Efecto de glow usando multiples circulos
		if glowIntensity > 0 {
			glowRadius := radius * (1 + glowIntensity)
			glowAlpha := uint8(glowIntensity * 100)

			// Dibujar circulos concentricos para el glow
			for r := radius; r <= glowRadius; r += 1 {
				alpha := glowAlpha * uint8(glowRadius-r) / uint8(glowRadius-radius)
				renderer.SetDrawColor(color.R, color.G, color.B, alpha)
				m.drawCircle(renderer, centerX, centerY, r)
			}
		}
	}

	// Circulo principal
	renderer.SetDrawColor(color.R, color.G, color.B, color.A)
	m.drawFilledCircle(renderer, centerX, centerY, radius)

	// Borde
	renderer.SetDrawColor(utils.ColorBlanco.R, utils.ColorBlanco.G, utils.ColorBlanco.B, 255)
	m.drawCircle(renderer, centerX, centerY, radius)
}

// Funcion auxiliar para dibujar circulo usando funciones C optimizadas
func (m *MenuSystem) drawCircle(renderer *sdl.Renderer, centerX, centerY, radius float32) {
	// Fallback usando SDL puro
	for angle := 0.0; angle < 2*math.Pi; angle += 0.1 {
		x := centerX + radius*float32(math.Cos(angle))
		y := centerY + radius*float32(math.Sin(angle))
		renderer.RenderPoint(x, y)
	}

}

// Funcion auxiliar para dibujar circulo relleno usando funciones C optimizadas
func (m *MenuSystem) drawFilledCircle(renderer *sdl.Renderer, centerX, centerY, radius float32) {
	// Fallback usando SDL puro
	radiusInt := int(radius)
	for y := -radiusInt; y <= radiusInt; y++ {
		for x := -radiusInt; x <= radiusInt; x++ {
			if float32(x*x+y*y) <= radius*radius {
				renderer.RenderPoint(centerX+float32(x), centerY+float32(y))
			}
		}

	}
}

// Simular texto con glow usando rectagulos
func (m *MenuSystem) renderGlowText(renderer *sdl.Renderer, text string, x, y float32, color utils.Color, scale float32) {
	textWidth := float32(len(text)) * 7 * scale
	textHeight := 15 * scale

	// Efecto glow
	glowIntensity := float32(0.3 + 0.2*math.Sin(m.AnimationTime*2))
	glowSize := glowIntensity * 6
	glowAlpha := uint8(glowIntensity * 80)

	renderer.SetDrawColor(color.R, color.G, color.B, glowAlpha)
	glowRect := sdl.FRect{
		X: x - glowSize, Y: y - glowSize,
		W: textWidth + glowSize*2, H: textHeight + glowSize*2,
	}
	renderer.RenderRect(&glowRect)

	// Usar DebugText para el texto real
	renderer.SetDrawColor(color.R, color.G, color.B, color.A)
	renderer.DebugText(x, y, text)
}

// Manejar eventos del mouse
func (m *MenuSystem) HandleMouseMotion(x, y int32) {
	m.MouseX, m.MouseY = x, y

	// Verificar hover en botones
	for i := range m.Buttons {
		btn := &m.Buttons[i]
		btn.IsHovered = m.isPointInButton(x, y, *btn)
	}
}

func (m *MenuSystem) HandleMouseClick(x, y int32, pressed bool) {
	for i := range m.Buttons {
		btn := &m.Buttons[i]
		if m.isPointInButton(x, y, *btn) {
			btn.IsPressed = pressed
			if !pressed && btn.Action != nil {
				btn.Action()
			}
		}
	}

	// Manejar clics en campos de texto (solo en menu de configuracion)
	if !pressed {
		for i := range m.InputFields {
			field := &m.InputFields[i]
			if m.isPointInInputField(x, y, *field) {
				// Desenfocar otros campos
				m.Window.StopTextInput()
				for j := range m.InputFields {
					m.InputFields[j].IsFocused = false
				}
				field.IsFocused = true
				if field.Type != 1 {
					m.Window.StartTextInput()
				}
				m.Blocker = true
				field.CursorBlink = time.Now()
				m.FocusedField = i
			}
		}
	}
}

// Verificar si un punto esta dentro de un boton
func (m *MenuSystem) isPointInButton(x, y int32, btn AnimatedButton) bool {
	return float32(x) >= btn.X && float32(x) <= btn.X+btn.Width &&
		float32(y) >= btn.Y && float32(y) <= btn.Y+btn.Height
}

// Verificar si un punto esta dentro de un campo de entrada
func (m *MenuSystem) isPointInInputField(x, y int32, field InputField) bool {
	return float32(x) >= field.X && float32(x) <= field.X+field.Width &&
		float32(y) >= field.Y && float32(y) <= field.Y+field.Height
}

// Manejar entrada de texto
func (m *MenuSystem) HandleTextInput(text string) {
	if m.FocusedField >= 0 && m.FocusedField < len(m.InputFields) {
		field := &m.InputFields[m.FocusedField]
		if field.IsFocused {
			field.Text += text
			field.CursorBlink = time.Now()
		}
	}
}

// Manejar entrada de teclado
func (m *MenuSystem) HandleKeyInput(keysc sdl.Scancode, pressed bool) {
	// Si no es de tipo combinacion de teclas entonces salimos

	keycode := int32(keysc)
	if keycode >= 0 && keycode < int32(len(m.Keys)) {
		m.Keys[keycode] = pressed
	}

	// Manejar teclas especiales para campos de texto
	if pressed && m.FocusedField >= 0 && m.FocusedField < len(m.InputFields) {
		field := &m.InputFields[m.FocusedField]

		if field.IsFocused {
			switch keycode {
			case 42: // tecla backspace
				if len(field.Text) > 0 {
					if field.Type == 1 {
						field.Text = "..."
					} else {
						field.Text = field.Text[:len(field.Text)-1]
					}
					field.CursorBlink = time.Now()
				}
			case 41: // tecla escape
				field.IsFocused = false
				m.FocusedField = -1
				m.Blocker = false
			case 40: // tecla enter
				field.IsFocused = false

				if field.Title == "Velocidad" {
					// Convertimos el string a entero
					val, err := strconv.Atoi(field.Text)
					if err != nil {
						log.Printf("Error al convertir velocidad: %v", err)
						val = 60
					}
					m.Fotogramas = uint32(1000.0 / val)
				}

				if field.Type != 1 {
					//config.SaveSettings(config.Settings{Any: "..."})
					m.FocusedField = -1
				} else {
					//config.SaveSettings(config.Settings{Any: "..."})
					m.FocusedField = -1
				}
				m.Blocker = false
			default: // Otras teclas
				field.CursorBlink = time.Now()
			}
		}
	}
}

func (m *MenuSystem) openConfigMenu() {
	m.State = MenuConfiguracion
	m.setupConfigMenu()
}

func (m *MenuSystem) backToMainMenu() {
	m.State = MenuPrincipal
	m.setupMainMenu()
	m.FocusedField = -1
	m.Blocker = false
}
