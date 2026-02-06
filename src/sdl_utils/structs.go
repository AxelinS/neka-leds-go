package sdl_utils

import (
	"go-neka-leds/src/screen"
	"go-neka-leds/src/utils"
	. "go-neka-leds/src/utils"
	"time"

	"github.com/Zyko0/go-sdl3/sdl"
)

type WindowConfig struct {
	Width  int
	Height int
}

// Estados del menu
type MenuState int

// Estructura para botones animados
type AnimatedButton struct {
	X, Y          float32
	Width, Height float32
	Text          string
	Color         Color
	HoverColor    Color
	IsHovered     bool
	IsPressed     bool
	Scale         float32
	Alpha         float32
	GlowIntensity float32
	Action        func()
}

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

// Estructura principal del menu
type MenuSystem struct {
	WindowConfig
	Led_s          *screen.LedsManager
	State          MenuState
	Buttons        map[string]*AnimatedButton
	InputFields    map[string]*InputField
	AnimationTime  float64
	MouseX, MouseY int32
	Keys           []bool
	FocusedField   string // Para manejar que campo tiene el foco
	TitleConf      TitleConfig
	Canales        utils.Canales
	Window         *sdl.Window
	Blocker        bool
	Fotogramas     uint32

	VerModo bool
}

type TitleConfig struct {
	X      float32
	Y      float32
	Width  int
	Height int
	Scale  float32
}

// Estructura extendida para incluir el sistema de menu
type MenuConfig struct {
	ComandoModo   bool
	ComandoTexto  string
	TeclasComando string
}
