package settings

import "os"

const SETTINGS_FILE = "/settings.json"

type Settings struct {
	Temperature    float64
	Brightness     float64
	LedsCount      int
	Mode           int
	KernelSize     int
	Padding        int
	LineLen        int
	FPS            int
	LineTickness   int
	CinePaddingY   int
	WinCaptureMode int
	Cinema         bool
}

func GetDefaultSettings() Settings {
	return Settings{
		Temperature:    0.7,
		Brightness:     0.4,
		LedsCount:      84,
		Mode:           1,
		KernelSize:     3,
		Padding:        1,
		LineLen:        80,
		FPS:            30,
		LineTickness:   3,
		CinePaddingY:   200,
		WinCaptureMode: 0,
		Cinema:         false,
	}
}

func InitializeSettings() Settings {
	// vemos si existe el archivo de SETTINGS_FILE
	_, err := os.Open(SETTINGS_FILE)
	if err != nil {
		// si no existe, lo creamos con los valores por defecto
		s := GetDefaultSettings()
		SaveSettings(s)
		return s
	}
	// si existe, lo cargamos
	return LoadSettings()
}
