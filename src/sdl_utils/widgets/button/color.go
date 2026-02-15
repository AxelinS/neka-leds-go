package buttons

import "go-neka-leds/src/utils"

var (
	ColorAzulTronOff = utils.Color{R: 32, G: 74, B: 135, A: 255}
	ColorAzulTronOn  = utils.Color{R: 41, G: 128, B: 185, A: 255}
	ColorDisabled    = utils.Color{R: 100, G: 100, B: 100, A: 255}
)

func GetBtnColor(switchState bool, Active bool) utils.Color {
	if !Active {
		return ColorDisabled
	}
	if switchState {
		return ColorAzulTronOn
	}
	return ColorAzulTronOff
}

func (b *AnimatedButton) UpdateColor() {
	if !b.Active {
		b.Color = ColorDisabled
		b.HoverColor = ColorDisabled
		return
	}
	if b.Switch {
		b.Color = ColorAzulTronOn
		b.HoverColor = ColorAzulTronOff
		return
	}
	b.Color = ColorAzulTronOff
	b.HoverColor = ColorAzulTronOn
}
