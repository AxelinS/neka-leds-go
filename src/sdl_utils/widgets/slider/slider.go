package slider

import "go-neka-leds/src/utils"

// Estructura para sliders controlables
type Slider struct {
	X, Y            float32
	Width, Height   float32
	Value           float64     // Valor actual del slider (0.0 - 1.0 o rango personalizado)
	MinValue        float64     // Valor mínimo
	MaxValue        float64     // Valor máximo
	Title           string      // Título del slider
	IsFocused       bool        // Si el slider está siendo arrastrado
	Color           utils.Color // Color del slider
	HoverColor      utils.Color // Color cuando está en hover
	ThumbColor      utils.Color // Color del icono del slider (thumb)
	ThumbHoverColor utils.Color // Color del thumb en hover
	BarColor        utils.Color // Color de la barra de fondo
	OnChange        func()      // Callback cuando cambia el valor
	Precision       int         // Decimales de precisión para mostrar
	ShowValue       bool        // Si mostrar el valor actual
	IsHovered       bool        // Si el mouse está sobre el slider
}

// NewSlider crea un nuevo slider con valores por defecto
func NewSlider(x, y, width, height float32, minVal, maxVal, currentVal float64, title string) *Slider {
	return &Slider{
		X:               x,
		Y:               y,
		Width:           width,
		Height:          height,
		MinValue:        minVal,
		MaxValue:        maxVal,
		Value:           currentVal,
		Title:           title,
		IsFocused:       false,
		IsHovered:       false,
		Color:           utils.ColorAzulTron,
		HoverColor:      utils.ColorAzulNeon,
		ThumbColor:      utils.ColorAzulTron,
		ThumbHoverColor: utils.ColorAzulNeon,
		BarColor:        utils.Color{R: 50, G: 50, B: 50, A: 255},
		OnChange:        func() {},
		Precision:       2,
		ShowValue:       true,
	}
}

// GetNormalizedValue retorna el valor normalizado entre 0 y 1
func (s *Slider) GetNormalizedValue() float64 {
	if s.MaxValue == s.MinValue {
		return 0
	}
	normalized := (s.Value - s.MinValue) / (s.MaxValue - s.MinValue)
	if normalized < 0 {
		return 0
	}
	if normalized > 1 {
		return 1
	}
	return normalized
}

// SetNormalizedValue establece el valor a partir de un valor normalizado (0-1)
func (s *Slider) SetNormalizedValue(normalized float64) {
	if normalized < 0 {
		normalized = 0
	}
	if normalized > 1 {
		normalized = 1
	}
	s.Value = s.MinValue + normalized*(s.MaxValue-s.MinValue)
	s.OnChange()
}

// GetThumbPosition retorna la posición X del thumb (icono) del slider
func (s *Slider) GetThumbPosition() float32 {
	normalized := float32(s.GetNormalizedValue())
	return s.X + normalized*s.Width
}

// HandleMouseMotion actualiza el hover state y actualiza el valor si está siendo arrastrado
func (s *Slider) HandleMouseMotion(mx, my int32) {
	x, y := float32(mx), float32(my)

	// Verificar si el mouse está sobre el slider
	isOver := x >= s.X && x <= s.X+s.Width && y >= s.Y-s.Height/2 && y <= s.Y+s.Height*1.5
	s.IsHovered = isOver

	// Si está siendo arrastrado, actualizar el valor
	if s.IsFocused {
		normalized := float64((x - s.X) / s.Width)
		s.SetNormalizedValue(normalized)
	}
}

// HandleMousePress maneja el clic del mouse (inicio del arrastre)
func (s *Slider) HandleMousePress(mx, my int32, pressed bool) {
	x := float32(mx)
	// Verificar si el clic está sobre el thumb o la barra
	inThumb := x >= s.GetThumbPosition()-s.Height && x <= s.GetThumbPosition()+s.Height
	inBar := x >= s.X && x <= s.X+s.Width
	// confirma que este dentro del slider
	inSlider := x >= s.X && x <= s.X+s.Width && float32(my) >= s.Y-s.Height/2 && float32(my) <= s.Y+s.Height*1.5
	if pressed && (inThumb || inBar) && inSlider {
		s.IsFocused = true
		// Si hace clic en la barra (no en el thumb), mover el thumb
		if inBar && !inThumb {
			normalized := float64((x - s.X) / s.Width)
			s.SetNormalizedValue(normalized)
		}
	} else if !pressed {
		s.IsFocused = false
	}
}
