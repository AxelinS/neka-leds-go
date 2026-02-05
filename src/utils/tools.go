package utils

import (
	"log"
	"strconv"
)

func StrToFloat(s string) float64 {
	// Convertimos el string a float
	val, err := strconv.ParseFloat(s, len(s))
	if err != nil {
		log.Printf("Error al convertir la Temperature: %v", err)
		val = 0.0
	}
	return val
}

func StrToInt(s string) int {
	// Convertimos el string a entero
	val, err := strconv.Atoi(s)
	if err != nil {
		log.Printf("Error al convertir los FPS: %v", err)
		val = 30
	}
	return val
}
