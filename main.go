package main

import (
	"flag"
	"fmt"
	"go-neka-leds/src/esp32"
	"go-neka-leds/src/screen"
	"go-neka-leds/src/sdl_utils"
	"go-neka-leds/src/win"
	"time"
)

var (
	argTemperature float64
	argBrightness  float64
	argFPS         int
	argMode        int
	argKernel      int
	argPadding     int
	argLinelen     int
	argLeds        int
)

func init() {
	flag.Float64Var(&argTemperature, "temperature", 0, "Color temperature (-1.0 frío | 0.0 neutro | +2.0 cálido)")
	flag.Float64Var(&argBrightness, "brightness", 0.4, "Brightness global (0.0 .. 1.0)")
	flag.IntVar(&argFPS, "fps", 30, "Frames per second")
	flag.IntVar(&argMode, "mode", 1, "Modo de operación")
	flag.IntVar(&argKernel, "kernel", 3, "Tamaño de kernel")
	flag.IntVar(&argPadding, "padding", 100, "Separacion de pixeles al borde")
	flag.IntVar(&argLinelen, "linelen", 50, "Largo de las lineas")
	flag.IntVar(&argLeds, "leds_count", 58, "Cantidad de leds")
}

func main() {
	fmt.Println("Iniciando programa...")
	fmt.Println("Buscando ESP32...")
	devs := esp32.DiscoverESP32()
	if len(devs) == 0 {
		fmt.Println("No se encontraron ESP32")
		return
	}
	fmt.Printf("Se encontraron %v dispositivos\n", len(devs))
	for i, d := range devs {
		fmt.Printf("%v:%v\n", i, d.Id)
	}

	_, _, width, height := win.GetPrimaryMonitor()
	points := screen.RectanglePerimeterPoints(width, height, argLeds, argPadding)

	led_s := screen.LedSettings{
		Devs:            devs,
		MonitorSettings: screen.MonitorSettings{Width: width, Height: height},
		CountSide:       screen.CountSides(points, width, height, argPadding),
		Pause:           false,
		LedsCount:       argLeds,
		Temperature:     argTemperature,
		Brightness:      argBrightness,
		Mode:            argMode,
		KernelSize:      argKernel,
		Padding:         argPadding,
		LineLen:         argLinelen,
		Points:          points,

		SampleLines: screen.BuildSampleLines(points, width, height, argPadding, argLinelen),
	}

	go func() {
		cap := win.NewScreenCapturer(width, height)
		defer cap.Close()
		ticker := time.NewTicker(time.Second / time.Duration(argFPS))
		defer ticker.Stop()
		for range ticker.C {
			if !led_s.Pause {
				img := cap.Capture()
				values := led_s.GetLedValues(img, width, height, points)
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
