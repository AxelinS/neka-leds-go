package settings

import "os"

const SETTINGS_FILE = "./settings.json"

type Settings struct {
	// === Sistema de Calibración Profesional RGB ===
	// Método estándar: monitores profesionales, calibradores de cámara, postproducción

	Temperature float64 // Color Temperature: -1.0 (frío/azul) a +1.0 (cálido/rojo)
	// Rango típico: [-0.5, 0.5] para uso general
	// Positivo: aumenta rojo, disminuye azul (tungsteno)
	// Negativo: disminuye rojo, aumenta azul (daylight)

	RCal float64 // Red Gain: 0.5 a 2.0 (recomendado 0.8 a 1.2)
	// Corrige dominancia de rojo

	GCal float64 // Green Gain: 0.5 a 2.0 (recomendado 0.8 a 1.2)
	// Corrige dominancia de verde (menos común)

	BCal float64 // Blue Gain: 0.5 a 2.0 (recomendado 0.8 a 1.2)
	// Corrige dominancia de azul

	Brightness float64 // Multiplicador final: 0.1 a 2.0 (típico 0.3 a 0.8)
	// Afecta uniformemente los 3 canales
	Saturation float64 // Saturación: 0.0 (gris) a 2.0 (muy saturado). 1.0 = sin cambio
	Gamma      float64 // Gamma de salida: típicamente 2.2 (usa >0)

	// === Resto de configuración ===
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
	IP               string
	Port             string
	UsingWifi        bool
}

func GetDefaultSettings() Settings {
	return Settings{
		// === Calibración RGB profesional (valores por defecto) ===
		Temperature: 0.0, // Neutro (sin ajuste de temperatura)
		RCal:        1.0, // Rojo: ganancia normal
		GCal:        1.0, // Verde: ganancia normal
		BCal:        1.0, // Azul: ganancia normal
		Brightness:  0.4, // 40% de brillo base (ajustable vía UI)
		Saturation:  1.0, // Saturación por defecto (sin cambio)
		Gamma:       2.2, // Gamma por defecto

		// === Configuración de LEDs ===
		LedsCount:        84,
		PixelMethod:      1, // 1 = líneas (recomendado para rendimiento)
		KernelSize:       3,
		Padding:          1,
		LineLen:          80,
		FPS:              30,
		LineThickness:    3,
		CinePaddingY:     200,
		WinCaptureMode:   0,
		Cinema:           false,
		AudioSensitivity: 1.0,
		Mode:             0, // 0 = screen capture
		RGBMode:          0, // 0 = static
		RGBStaticColor:   8, // 8 = blanco
		StartPoint:       0, // 0 = arriba-izquierda
		Switch:           true,
		IP:               "0.0.0.0",
		Port:             "7770",
		UsingWifi:        false,
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
