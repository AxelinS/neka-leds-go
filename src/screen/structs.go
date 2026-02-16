package screen

import (
	"go-neka-leds/src/esp32"
	"go-neka-leds/src/settings"
	"go-neka-leds/src/utils"
)

type MonitorSettings struct {
	Width, Height int
}

type LedsManager struct {
	Chn  *utils.Canales
	Devs []esp32.ESP32

	WifiDev       *esp32.ESP32WIFI
	WifiConnected bool

	MonitorSettings
	CountSide SideCount
	Pause     bool
	Suspend   bool

	S settings.Settings

	CCC [3]int // Current Cicle Color - RGB

	Points       []Point
	ScaledPoints []Point

	SampleLines []SampleLine
	PixelLines  []PixelLine
	ScaledLines []PixelLine

	PointsCinema       []Point
	ScaledPointsCinema []Point
	LinesCinema        []SampleLine
	PixelCinema        []PixelLine
	ScaledLinesCinema  []PixelLine
}

type Side int

const (
	Top Side = iota
	Right
	Bottom
	Left
)

type SampleLine struct {
	Offsets []int
}
