package sdl_utils

import (
	"fmt"
	"go-neka-leds/src/screen"
	"time"
)

func (m *MenuSystem) ModeTestPoints() {
	m.Led_s.Pause = !m.Led_s.Pause

	values := screen.GetValuesColor(0, 0, 180, m.Led_s.CountSide.Top)
	values += screen.GetValuesColor(180, 0, 0, m.Led_s.CountSide.Right)
	values += screen.GetValuesColor(180, 180, 180, m.Led_s.CountSide.Bottom)
	values += screen.GetValuesColor(0, 180, 0, m.Led_s.CountSide.Left)
	// Pintamos de diferente color los leds de cada lado
	time.Sleep(200 * time.Millisecond)
	for _, dev := range m.Led_s.Devs {
		if dev.Connected {
			dev.SafeWrite("RGB " + values + "\n")
		} else {
			if dev.Reconnect() {
				fmt.Println("[RECONNECTED]", dev.Id)
			}
		}
	}
}

func (m *MenuSystem) TurnOff() {
	values := screen.GetValuesColor(0, 0, 0, m.Led_s.CountSide.Top)
	values += screen.GetValuesColor(0, 0, 0, m.Led_s.CountSide.Right)
	values += screen.GetValuesColor(0, 0, 0, m.Led_s.CountSide.Bottom)
	values += screen.GetValuesColor(0, 0, 0, m.Led_s.CountSide.Left)
	for _, dev := range m.Led_s.Devs {
		if dev.Connected {
			dev.SafeWrite("RGB " + values + "\n")
		} else {
			if dev.Reconnect() {
				fmt.Println("[RECONNECTED]", dev.Id)
			}
		}
	}
}
