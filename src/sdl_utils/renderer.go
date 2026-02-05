package sdl_utils

import (
	. "go-neka-leds/src/utils"
	"log"
	"time"

	"github.com/Zyko0/go-sdl3/img"
	"github.com/Zyko0/go-sdl3/sdl"
)

func (w *WindowApp) RenderApp() {
	// Inicializa SDL
	sdl.LoadLibrary(sdl.Path()) // "SDL3.dll", "libSDL3.so.0", "libSDL3.dylib"
	defer sdl.Quit()
	if err := sdl.Init(sdl.INIT_VIDEO | sdl.INIT_JOYSTICK); err != nil {
		panic(err)
	}

	// Crear ventana mas grande para el menu
	window, renderer, err := sdl.CreateWindowAndRenderer("neka-leds", w.Width, w.Height, sdl.WINDOW_OPENGL)
	if err != nil {
		log.Println(err)
		return
	}
	w.Window = window
	img.LoadLibrary(img.Path())
	iconSurf, err := img.Load("./icon.png")
	if err != nil {
		log.Fatalln("No se cargó el icono: ", err)
	}
	window.SetIcon(iconSurf)

	defer window.Destroy()
	defer renderer.Destroy()
	defer iconSurf.Destroy()

	// Inicializar el sistema de menu
	menuSystem := NewMenuSystem(w.WindowConfig, window, w.LedsManager)
	menuSystem.RestartLedsScales()

	// Variables para el control de tiempo
	lastTime := time.Now()
	// Event loop
	sdl.RunLoop(func() error {
		currentTime := time.Now()
		deltaTime := currentTime.Sub(lastTime).Seconds()
		lastTime = currentTime

		var event sdl.Event
		for sdl.PollEvent(&event) {
			switch event.Type {
			case sdl.EVENT_QUIT:
				return sdl.EndLoop
			case sdl.EVENT_WINDOW_FOCUS_GAINED:
				w.Focus = true
			case sdl.EVENT_WINDOW_FOCUS_LOST:
				w.Focus = false
			case sdl.EVENT_MOUSE_MOTION:
				evt := event.MouseMotionEvent()
				menuSystem.HandleMouseMotion(int32(evt.X), int32(evt.Y))

			case sdl.EVENT_MOUSE_BUTTON_DOWN:
				evt := event.MouseButtonEvent()
				if evt.Button == uint8(sdl.BUTTON_LEFT) {
					menuSystem.HandleMouseClick(int32(evt.X), int32(evt.Y), true)
				}

			case sdl.EVENT_MOUSE_BUTTON_UP:
				evt := event.MouseButtonEvent()
				if evt.Button == uint8(sdl.BUTTON_LEFT) {
					menuSystem.HandleMouseClick(int32(evt.X), int32(evt.Y), false)
				}

			case sdl.EVENT_KEY_DOWN:
				evt := event.KeyboardEvent()
				menuSystem.HandleKeyInput(evt.Scancode, true)

			case sdl.EVENT_KEY_UP:
				evt := event.KeyboardEvent()
				menuSystem.HandleKeyInput(evt.Scancode, false)

			case sdl.EVENT_TEXT_INPUT:
				evt := event.TextInputEvent()
				textBytes := evt.Text[:]
				var textStr string
				for i, b := range textBytes {
					if b == 0 {
						textStr = string(textBytes[:i])
						break
					}
				}
				if textStr == "" {
					textStr = string(textBytes[:])
				}
				menuSystem.HandleTextInput(textStr)
			}
		}

		if w.Focus {
			menuSystem.Update(deltaTime)
			renderer.SetDrawColor(ColorNegro.R, ColorNegro.G, ColorNegro.B, ColorNegro.A)
			renderer.Clear()
			menuSystem.Render(renderer)
			renderer.Present()
		}
		sdl.Delay(menuSystem.Fotogramas) // 64 = 16 FPS
		return nil
	})
}
