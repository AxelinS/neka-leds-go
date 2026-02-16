package utils

type Canales struct {
	Suspended chan bool
	Stop      chan struct{}
}
