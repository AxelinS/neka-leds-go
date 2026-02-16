package screen

import (
	"fmt"
	"time"
)

func GetInnerOuterVals(width, height, argLeds, padding, lineLen, startPoint int) ([]Point, []Point) {
	inner := RectanglePerimeterPoints(width, height, argLeds, padding+lineLen, startPoint)
	outer := RectanglePerimeterPoints(width, height, argLeds, padding, startPoint)
	return inner, outer
}

func (l *LedsManager) LedsSwitch() {
	l.S.Switch = !l.S.Switch
	if !l.S.Switch {
		l.TurnOff()
	}
}

func (l *LedsManager) TurnOff() {
	values := GetValues(0, 0, 0, l.S.LedsCount)
	for _, dev := range l.Devs {
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
		if l.WifiDev.Connected {
			l.WifiDev.SendLEDValues(values)
		}
	}
}
