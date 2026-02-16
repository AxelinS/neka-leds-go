package main

import (
	"fmt"
	"go-neka-leds/src/esp32"
	"go-neka-leds/src/screen"
	"go-neka-leds/src/utils"
	"go-neka-leds/src/win"
	"time"
)

func StartSuspendManager(led_s *screen.LedsManager, chn *utils.Canales, cap *win.ScreenCapturer) {
	for {
		led_s.Suspend = <-chn.Suspended
		fmt.Println("Estado de suspencion: ", led_s.Suspend)
		if led_s.Suspend {
			led_s.TurnOff()
			continue
		}
		c := win.NewScreenCapturer(led_s.Width, led_s.Height)
		cap = c
		wd, err := esp32.ConnectESP32Wifi(led_s.WifiDev.IP, led_s.WifiDev.Port)
		if err != nil || wd.Conn == nil {
			fmt.Println("No se pudo conectar al ESP32 por Wifi")
		}
		led_s.WifiDev.Conn = wd.Conn
		led_s.WifiDev.Connected = true
		led_s.Devs = esp32.DiscoverESP32()
	}
}

func StartAutoReconnect(led_s *screen.LedsManager, chn *utils.Canales) {
	for {
		time.Sleep(30 * time.Second)

		if led_s.S.UsingWifi {
			if led_s.WifiDev.Conn == nil {
				connected := led_s.WifiDev.Reconnect()
				if connected {
					fmt.Println("ESP32 reconectado por Wifi")
				}
			}
			return
		}
		if len(led_s.Devs) == 0 {
			led_s.Devs = esp32.DiscoverESP32()
			if len(led_s.Devs) > 0 {
				fmt.Println("ESP32 conectado por USB")
			}
		}
	}
}

func StartLedsSenderManager(led_s *screen.LedsManager, chn *utils.Canales, cap *win.ScreenCapturer) {
	ticker := time.NewTicker(time.Second / time.Duration(led_s.S.FPS))
	defer ticker.Stop()

	for range ticker.C {
		if !led_s.Pause && led_s.S.Switch && !led_s.Suspend { // Si esta en pausa o apagado no captura ni envia nada
			var values []byte
			switch led_s.S.Mode {
			case 0:
				img := cap.Capture(led_s.S.WinCaptureMode)
				// Skip frame if capture returns nil (timeout or error)
				if img == nil {
					continue
				}
				values = led_s.GetLedValues(img, led_s.Width, led_s.Height, led_s.Points)
			case 1:
				//values = led_s.GetAudioReactiveValues()
				values = screen.GetValues(225, 225, 225, led_s.S.LedsCount)
			case 2:
				values = screen.GetValues(225, 225, 225, led_s.S.LedsCount)
			default:
				values = screen.GetValues(255, 255, 255, led_s.S.LedsCount)
			}
			if led_s.S.UsingWifi {
				led_s.WifiDev.SendLEDValues(values)
				continue
			}
			for _, dev := range led_s.Devs {
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
				} else {
					if dev.Reconnect() {
						fmt.Println("[RECONNECTED]", dev.Id)
					}
				}
			}
		}
	}
}
