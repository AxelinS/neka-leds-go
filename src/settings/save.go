package settings

import (
	"encoding/json"
	"fmt"
	"os"
)

func SaveSettings(s Settings) {
	settings, err := json.Marshal(s)
	if err != nil {
		fmt.Println("Error al convertir settings a JSON: ", err)
	}
	err = os.WriteFile(SETTINGS_FILE, settings, 0644)
	if err != nil {
		fmt.Println("Error al guardar settings en archivo: ", err)
	}
}
