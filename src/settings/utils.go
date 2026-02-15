package settings

import "os"

const SETTINGS_FILE = "./settings.json"

type Settings struct {
	Temperature      float64
	Brightness       float64
	LedsCount        int
	PixelMethod      int // Modo de obtencion de pixeles - 0 kernel, 1 lineas, 2 pixel
	KernelSize       int
	Padding          int
	LineLen          int
	FPS              int
	LineThickness    int
	CinePaddingY     int
	WinCaptureMode   int
	Cinema           bool
	AudioSensitivity float64
	Mode             int // 0 screen capture, 1 audio reactive, 2 static
	RGBMode          int // 0 static, 1 cicle static fade, 2 fade static, 3 static cicling
	RGBStaticColor   int // 0 rojo, 1 naranja, 2 amarillo, 3 verde, 4 cyan, 5 azul, 6 magenta, 7 rosa, 8 blanco
	StartPoint       int // 0 arriba-izq, 1 arriba-der, 2 abajo-der, 3 abajo-izq
	Switch           bool
}

func GetDefaultSettings() Settings {
	return Settings{
		Temperature:      0.7,
		Brightness:       0.4,
		LedsCount:        84,
		PixelMethod:      1,
		KernelSize:       3,
		Padding:          1,
		LineLen:          80,
		FPS:              30,
		LineThickness:    3,
		CinePaddingY:     200,
		WinCaptureMode:   0,
		Cinema:           false,
		AudioSensitivity: 1.0,
		Mode:             0,
		RGBMode:          0,
		RGBStaticColor:   8,
		StartPoint:       0,
		Switch:           true,
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
