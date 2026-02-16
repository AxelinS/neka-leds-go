package sdl_utils

import (
	"go-neka-leds/src/screen"
	btn "go-neka-leds/src/sdl_utils/widgets/button"
	ifield "go-neka-leds/src/sdl_utils/widgets/inputfield"
	slider "go-neka-leds/src/sdl_utils/widgets/slider"
	"go-neka-leds/src/utils"

	"github.com/Zyko0/go-sdl3/sdl"
)

type WindowConfig struct {
	Width  int
	Height int
}

// Estados del menu
type MenuState int

// Estructura principal del menu
type MenuSystem struct {
	WindowConfig
	Led_s          *screen.LedsManager
	State          MenuState
	Buttons        map[string]*btn.AnimatedButton
	InputFields    map[string]*ifield.InputField
	Sliders        map[string]*slider.Slider
	AnimationTime  float64
	MouseX, MouseY int32
	Keys           []bool
	FocusedField   string // Para manejar que campo tiene el foco
	TitleConf      TitleConfig
	Canales        utils.Canales
	Window         *sdl.Window
	Blocker        bool
	Fotogramas     uint32

	Visual bool
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
