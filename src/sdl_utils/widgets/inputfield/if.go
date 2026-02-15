package ifields

import "time"

// Estructura para campos de texto
type InputField struct {
	X, Y          float32
	Width, Height float32
	Text          string
	Placeholder   string
	IsFocused     bool
	CursorPos     int
	CursorBlink   time.Time
	Title         string
	Type          int8 // 0 texto - 1 combinaciones de teclas
}
