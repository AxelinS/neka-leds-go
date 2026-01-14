package sdl_utils

import "github.com/Zyko0/go-sdl3/sdl"

// Toma un valor Int32 y devuelve el String con el name de la key de SDL Scancode
func IntToScancodeName(i int32) string {
	sc := sdl.Scancode(i)
	return sc.Name()
}

// Toma un slice []int32 y devuelve un string compuesto por estos caracteres
func ScancodeNameChain(ic []int32) string {
	s := ""
	for _, w := range ic {
		s += IntToScancodeName(w) + " "
	}
	return s
}
