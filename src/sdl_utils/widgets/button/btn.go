package buttons

import "go-neka-leds/src/utils"

// Estructura para botones animados
type AnimatedButton struct {
	X, Y          float32
	Width, Height float32
	Text          string
	Color         utils.Color
	HoverColor    utils.Color
	IsHovered     bool
	IsPressed     bool
	Scale         float32
	Alpha         float32
	GlowIntensity float32
	Action        func()
	Active        bool
	Switch        bool
}
