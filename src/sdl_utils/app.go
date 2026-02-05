package sdl_utils

import (
	"go-neka-leds/src/screen"

	"github.com/Zyko0/go-sdl3/sdl"
)

type WindowApp struct {
	Focus bool
	WindowConfig
	Window      *sdl.Window
	MSys        MenuSystem
	LedsManager *screen.LedsManager
}

func WindowLoop(ls *screen.LedsManager) {
	windowConfig := WindowConfig{Width: 200, Height: 400}
	w := WindowApp{WindowConfig: windowConfig, Focus: true, LedsManager: ls}
	w.RenderApp()
}
