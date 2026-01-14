package esp32

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/tarm/serial"
)

type ESP32 struct {
	Port      *serial.Port
	ReadBuf   []byte
	Id        string
	PortName  string
	Connected bool
}

func drainSerial(s *serial.Port) {
	buf := make([]byte, 64)
	for {
		n, _ := s.Read(buf)
		if n == 0 {
			return
		}
	}
}

func readLine(s *serial.Port, timeout time.Duration) (string, bool) {
	deadline := time.Now().Add(timeout)
	var line []byte
	buf := make([]byte, 1)

	for time.Now().Before(deadline) {
		n, _ := s.Read(buf)
		if n == 1 {
			if buf[0] == '\n' {
				return strings.TrimSpace(string(line)), true
			}
			line = append(line, buf[0])
		}
	}
	return "", false
}

func (d *ESP32) Reconnect() bool {
	cfg := &serial.Config{
		Name:        d.PortName,
		Baud:        115200,
		ReadTimeout: 200 * time.Millisecond,
	}

	s, err := serial.OpenPort(cfg)
	if err != nil {
		return false
	}

	time.Sleep(2 * time.Second)
	drainSerial(s)

	s.Write([]byte("WHO\n"))
	line, ok := readLine(s, 2*time.Second)
	if !ok || !strings.HasPrefix(line, "leds-") {
		s.Close()
		return false
	}

	d.Port = s
	d.Connected = true
	return true
}

func (d *ESP32) SafeWrite(cmd string) {
	if !d.Connected {
		return
	}

	_, err := d.Port.Write([]byte(cmd))
	if err != nil {
		fmt.Println("[DISCONNECT]", d.PortName)
		d.Port.Close()
		d.Connected = false
	}
}

func DiscoverESP32() []ESP32 {
	var devices []ESP32

	for i := range 20 {
		port := "COM" + strconv.Itoa(i)
		cfg := &serial.Config{
			Name:        port,
			Baud:        115200,
			ReadTimeout: 200 * time.Millisecond,
		}

		s, err := serial.OpenPort(cfg)
		if err != nil {
			continue
		}

		time.Sleep(2000 * time.Millisecond)

		drainSerial(s)

		for range 3 {
			s.Write([]byte("WHO\n"))
			time.Sleep(100 * time.Millisecond)
		}

		line, ok := readLine(s, 2*time.Second)
		fmt.Println("[SCANNING]", port, line)
		if ok && strings.HasPrefix(line, "leds-") {
			fmt.Println("[FOUND]", port, line)
			devices = append(devices, ESP32{
				Port:      s,
				ReadBuf:   make([]byte, 64),
				Id:        line,
				PortName:  port,
				Connected: true,
			})

		} else {
			s.Close()
		}
	}
	return devices
}
