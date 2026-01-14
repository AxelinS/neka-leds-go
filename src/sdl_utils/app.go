package sdl_utils

import "go-neka-leds/src/screen"

type WindowApp struct {
	Focus bool
	WindowConfig
	MSys        MenuSystem
	LedSettings *screen.LedSettings
}

func WindowLoop(ls *screen.LedSettings) {
	windowConfig := WindowConfig{Width: 200, Height: 400}
	w := WindowApp{WindowConfig: windowConfig, Focus: true, LedSettings: ls}
	w.RenderApp()
}
