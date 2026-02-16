package esp32

import (
	"log"
	"net"
	"time"
)

type ESP32WIFI struct {
	IP        string
	Port      string
	Connected bool
	Conn      net.Conn
}

func ConnectESP32Wifi(ip, port string) (*ESP32WIFI, error) {
	wifiDev := ESP32WIFI{
		IP:        ip,
		Port:      port,
		Connected: false,
		Conn:      nil,
	}
	conn, err := net.Dial("udp", ip+":"+port)
	if err != nil {
		log.Println("Error al conectar al ESP32 por wifi:", err.Error())
		return &wifiDev, err
	}
	wifiDev.Conn = conn
	wifiDev.Connected = true
	return &wifiDev, nil
}

func (w *ESP32WIFI) Reconnect() bool {
	time.Sleep(300 * time.Millisecond)
	conn, err := net.Dial("udp", w.IP+":"+w.Port)
	if err != nil {
		log.Println("Error al reconectar al ESP32 por wifi:", err.Error())
		w.Connected = false
		w.Conn = nil
		return false
	}
	w.Conn = conn
	w.Connected = true
	return true
}

func (w *ESP32WIFI) SendLEDValues(values []byte) error {
	if !w.Connected || w.Conn == nil {
		w.Reconnect()
		return net.ErrClosed
	}
	_, err := w.Conn.Write(values)
	return err
}

func Test_ESP32Connection(ip, port string) string {
	// Implementa una función para probar la conexión al ESP32 usando TCP
	conn, err := net.Dial("udp", ip+":"+port)
	if err != nil {
		log.Println("Error test al conectar al ESP32 por wifi:", err.Error())
		return ""
	}
	defer conn.Close()
	frame := make([]byte, 84*3)
	// llenar frame
	for i := range 84 {
		frame[i*3] = 0
		frame[i*3+1] = 255
		frame[i*3+2] = 0
	}
	for range 100 {
		_, err = conn.Write(frame)
		if err != nil {
			log.Println("Error al enviar datos al ESP32 por wifi:", err.Error())
			return ""
		}
		time.Sleep(16 * time.Millisecond)
	}
	return "Conexión exitosa al ESP32 wifi test"
}
