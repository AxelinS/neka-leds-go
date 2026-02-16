package sdl_utils

import (
	"fmt"
	"go-neka-leds/src/screen"
	"time"
)

func (m *MenuSystem) ModeTestPoints() {
	m.Led_s.Pause = !m.Led_s.Pause
	values_t := screen.GetValues(0, 0, 180, m.Led_s.CountSide.Top)
	values_r := screen.GetValues(180, 0, 0, m.Led_s.CountSide.Right)
	values_b := screen.GetValues(180, 180, 180, m.Led_s.CountSide.Bottom)
	values_l := screen.GetValues(0, 180, 0, m.Led_s.CountSide.Left)
	values := append(values_t, values_r...)
	values = append(values, values_b...)
	values = append(values, values_l...)
	// Pintamos de diferente color los leds de cada lado
	time.Sleep(200 * time.Millisecond)
	for _, dev := range m.Led_s.Devs {
		if dev.Connected {
			ps := len(values)
			header := []byte{
				0xAA,
				byte(ps >> 8),
				byte(ps),
			}
			if _, err := dev.Port.Write(header); err != nil {
				fmt.Println("[DISCONNECT]", dev.Id)
				dev.Port.Close()
				dev.Connected = false
			}
			_, err := dev.Port.Write(values)
			if err != nil {
				fmt.Println("[DISCONNECT]", dev.Id)
				dev.Port.Close()
				dev.Connected = false
			}
		}
	}
	for range 3 {
		time.Sleep(30 * time.Millisecond)
		if m.Led_s.WifiDev.Connected {
			m.Led_s.WifiDev.SendLEDValues(values)
		}
	}
}
