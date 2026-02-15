package utils

// Colores de la paleta - Azul Neon, Azul Tron y Negro
type Color struct {
	R, G, B, A uint8
}

var (
	ColorNegro      = Color{0, 0, 0, 255}
	ColorAzulNeon   = Color{0, 255, 255, 255}  // Cyan brillante
	ColorAzulTron   = Color{41, 128, 185, 255} // Azul medio
	ColorAzulOscuro = Color{23, 32, 42, 255}   // Azul muy oscuro
	ColorBlanco     = Color{255, 255, 255, 255}
	ColorGris       = Color{127, 140, 141, 255}
	White           = Color{200, 200, 200, 255}
)
