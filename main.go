package main

import (
	"fmt"
	"go-neka-leds/src/esp32"
	"go-neka-leds/src/screen"
	"go-neka-leds/src/sdl_utils"
	"go-neka-leds/src/settings"
	"go-neka-leds/src/win"
	"time"
)

func main() {

	s := settings.InitializeSettings()

	if s.Brightness < 0 || s.Brightness > 1 {
		s.Brightness = 0.4
	}
	if s.FPS <= 0 {
		s.FPS = 30
	}

	fmt.Println("Buscando ESP32...")
	devs := esp32.DiscoverESP32()
	if len(devs) == 0 {
		fmt.Println("No se encontraron ESP32")
	}
	fmt.Printf("Se encontraron %v dispositivos\n", len(devs))
	for i, d := range devs {
		fmt.Printf("%v:%v\n", i, d.Id)
	}

	_, _, width, height := win.GetPrimaryMonitor()
	if s.Padding < 1 || s.Padding > height/3 {
		s.Padding = 1
	}

	// Inicializa los pixeles a recoger del modo standard
	inner, outer := screen.GetInnerOuterVals(width, height, s.LedsCount, s.Padding, s.LineLen, s.StartPoint)

	pixelLines := screen.BuildPixelLinesBetweenPerimeters(
		outer,
		inner,
		width,
		height,
		s.LineThickness,
	)
	// Inicializa los pixeles a recoger del modo cine
	o_cine := screen.ApplyCinemaPadding(outer, width, height, s.Padding, s.CinePaddingY)
	i_cine := screen.ApplyCinemaPadding(inner, width, height, s.Padding+s.LineLen, s.CinePaddingY+(s.LineLen))

	pixelLinesCine := screen.BuildPixelLinesBetweenPerimeters(
		o_cine,
		i_cine,
		width,
		height,
		s.LineThickness,
	)

	// Inicializa el gestor de leds
	led_s := screen.LedsManager{
		Devs:            devs,
		MonitorSettings: screen.MonitorSettings{Width: width, Height: height},
		CountSide:       screen.CountSides(outer, width, height, s.Padding),
		Pause:           false,

		// settings
		S: s,
		//
		CCC: [3]int{0, 0, 0},

		Points:      outer,
		PixelLines:  pixelLines,
		SampleLines: screen.PixelLinesToSampleLines(pixelLines, width),

		PointsCinema: o_cine,
		PixelCinema:  pixelLinesCine,
		LinesCinema:  screen.PixelLinesToSampleLines(pixelLinesCine, width),
	}

	go func() {
		cap := win.NewScreenCapturer(width, height)
		defer cap.Close()
		ticker := time.NewTicker(time.Second / time.Duration(led_s.S.FPS))
		defer ticker.Stop()

		for range ticker.C {
			if !led_s.Pause && led_s.S.Switch { // Si esta en pausa o apagado no captura ni envia nada
				var values string
				switch led_s.S.Mode {
				case 0:
					img := cap.Capture(led_s.S.WinCaptureMode)
					// Skip frame if capture returns nil (timeout or error)
					if img == nil {
						continue
					}
					values = led_s.GetLedValues(img, width, height, outer)
				case 1:
					//values = led_s.GetAudioReactiveValues()
				case 2:
					values = screen.GetValuesColor(225, 225, 225, s.LedsCount)
				default:
					values = screen.GetValuesColor(255, 255, 255, s.LedsCount)
				}
				for _, dev := range devs {
					if dev.Connected {
						dev.SafeWrite("RGB " + values + "\n")
					} else {
						if dev.Reconnect() {
							fmt.Println("[RECONNECTED]", dev.Id)
						}
					}
				}
			}
		}
	}()

	sdl_utils.WindowLoop(&led_s)
}
