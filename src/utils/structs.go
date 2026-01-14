package utils

type Canales struct {
	Twitch_MSG    chan []Mensaje
	TStopChat     chan bool
	TwitchCommand chan bool
	StopTTS       chan struct{}
}

type Mensaje struct {
	Nombre string
	Texto  string
}
