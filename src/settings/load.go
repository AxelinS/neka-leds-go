package settings

import (
	"encoding/json"
	"os"
)

func LoadSettings() Settings {
	file, err := os.Open(SETTINGS_FILE)
	if err != nil {
		panic("No se pudo abrir el archivo de configuración: " + err.Error())
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	var s Settings
	err = decoder.Decode(&s)
	if err != nil {
		panic("No se pudo decodificar el archivo de configuración: " + err.Error())
	}
	return s
}
