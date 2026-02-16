package main

import (
	"fmt"
	"go-neka-leds/src/esp32"
	"go-neka-leds/src/screen"
	"go-neka-leds/src/sdl_utils"
	"go-neka-leds/src/settings"
	"go-neka-leds/src/utils"
	"go-neka-leds/src/win"
)

func main() {
	s := settings.InitializeSettings()

	if s.Brightness < 0 || s.Brightness > 1 {
		s.Brightness = 0.4
	}
	if s.FPS <= 0 || s.FPS > 240 {
		s.FPS = 30
	}
	if s.LineThickness < 1 || s.LineThickness > 20 {
		s.LineThickness = 3
	}
	if s.CinePaddingY < 0 || s.CinePaddingY > 1000 {
		s.CinePaddingY = 200
	}
	if s.LedsCount < 1 || s.LedsCount > 999 {
		s.LedsCount = 1
	}
	if s.PixelMethod < 0 || s.PixelMethod > 2 {
		s.PixelMethod = 0
	}
	if s.KernelSize < 1 || s.KernelSize > 15 {
		s.KernelSize = 3
	}
	if s.Padding < 1 || s.Padding > 800 {
		s.Padding = 1
	}

	fmt.Println("Buscando ESP32...")
	wifidevs := esp32.ESP32WIFI{IP: s.IP, Port: s.Port}
	if s.UsingWifi {
		fmt.Printf("Intentando conectar al ESP32 por Wifi en %s:%s...\n", s.IP, s.Port)
		wd, err := esp32.ConnectESP32Wifi(s.IP, s.Port)
		if err != nil || wd.Conn == nil {
			fmt.Println("No se pudo conectar al ESP32 por Wifi")
		}
		wifidevs.Conn = wd.Conn
		wifidevs.Connected = true
		fmt.Println("Usando Wifi para controlar los LEDs")
	}
	devs := esp32.DiscoverESP32()
	if len(devs) == 0 {
		fmt.Println("No se encontraron ESP32 conectados por USB")
	}
	fmt.Printf("Se encontraron %d ESP32 conectados por USB\n", len(devs))

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

	chn := utils.Canales{
		Suspended: make(chan bool, 2),
		Stop:      make(chan struct{}),
	}
	// Inicializa el gestor de leds
	led_s := screen.LedsManager{
		Chn:             &chn,
		Devs:            devs,
		WifiDev:         &wifidevs,
		MonitorSettings: screen.MonitorSettings{Width: width, Height: height},
		CountSide:       screen.CountSides(outer, width, height, s.Padding),
		Pause:           false,
		Suspend:         false,

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

	cap := win.NewScreenCapturer(led_s.Width, led_s.Height)
	defer cap.Close()

	go func() {
		win.StartWindowsEvents(&chn)
	}()

	go func() {
		StartSuspendManager(&led_s, &chn, cap)
	}()

	go func() {
		StartAutoReconnect(&led_s, &chn)
	}()

	go func() {
		StartLedsSenderManager(&led_s, &chn, cap)
	}()

	sdl_utils.WindowLoop(&led_s)
}
