package win

import (
	"fmt"
	"go-neka-leds/src/utils"
	"log"
	"sync"
	"syscall"
	"unsafe"
)

// EventType representa el tipo de evento de energía detectado
type EventType int

const (
	EventSuspend         EventType = iota // Sistema entrando en suspensión
	EventResumeAutomatic                  // Reanudación automática (wake timer)
	EventResumeSuspend                    // Reanudación desde suspensión
	EventMonitorOff                       // Monitor apagado (DPMS)
	EventMonitorOn                        // Monitor encendido (DPMS)
)

// PowerEvent contiene información sobre un evento de energía
type PowerEvent struct {
	Type      EventType
	Timestamp int64
}

// EventListener escucha eventos de energía de Windows
type EventListener struct {
	eventChan  chan PowerEvent
	stopChan   chan struct{}
	mu         sync.Mutex
	isRunning  bool
	windowName string
	hwnd       uintptr
	done       chan struct{}
}

var (
	// DLLs necesarias
	user32DLL = syscall.NewLazyDLL("user32.dll")
	kernel32  = syscall.NewLazyDLL("kernel32.dll")

	// Procesos de user32.dll
	procRegisterClassW                   = user32DLL.NewProc("RegisterClassW")
	procCreateWindowExW                  = user32DLL.NewProc("CreateWindowExW")
	procDestroyWindow                    = user32DLL.NewProc("DestroyWindow")
	procDefWindowProcW                   = user32DLL.NewProc("DefWindowProcW")
	procGetMessageW                      = user32DLL.NewProc("GetMessageW")
	procTranslateMessage                 = user32DLL.NewProc("TranslateMessage")
	procDispatchMessageW                 = user32DLL.NewProc("DispatchMessageW")
	procPostQuitMessage                  = user32DLL.NewProc("PostQuitMessage")
	procPostMessageW                     = user32DLL.NewProc("PostMessageW")
	procRegisterPowerSettingNotification = user32DLL.NewProc("RegisterPowerSettingNotification")

	// Variable global para almacenar el listener actual
	currentListener *EventListener
	listenerMutex   sync.Mutex
)

// Constantes de Windows
const (
	WM_APP                      = 0x8000
	CS_VREDRAW                  = 0x0001
	CS_HREDRAW                  = 0x0002
	WS_OVERLAPPED               = 0x00000000
	CW_USEDEFAULT               = 0x80000000
	WM_DESTROY                  = 0x0002
	WM_SETTINGCHANGE            = 0x001A
	DEVICE_NOTIFY_WINDOW_HANDLE = 0x00000000
	WM_POWERBROADCAST_STOP      = WM_APP + 1000 // Mensaje personalizado para detener
)

// GUID para monitor power setting
var (
	GUID_MONITOR_POWER_ON = [16]byte{
		0xE0, 0x86, 0xCE, 0x02, 0x5B, 0x3C, 0x46, 0x42,
		0x8D, 0x50, 0x2C, 0xC8, 0xD1, 0xFD, 0xF2, 0xFE,
	}
)

// WNDCLASS es la estructura de clase de ventana
type WNDCLASS struct {
	Style         uint32
	LpfnWndProc   uintptr
	CbClsExtra    int32
	CbWndExtra    int32
	HInstance     uintptr
	HIcon         uintptr
	HCursor       uintptr
	HbrBackground uintptr
	LpszMenuName  *uint16
	LpszClassName *uint16
}

// MSG es la estructura de mensaje de Windows
type MSG struct {
	Hwnd    uintptr
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      struct {
		X int32
		Y int32
	}
}

// NewEventListener crea un nuevo listener de eventos
func NewEventListener(bufferSize int) *EventListener {
	return &EventListener{
		eventChan:  make(chan PowerEvent, bufferSize),
		stopChan:   make(chan struct{}),
		done:       make(chan struct{}),
		windowName: "NEKALEDsEventWindow",
	}
}

// Start inicia el listener en una goroutine
func (l *EventListener) Start() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.isRunning {
		return fmt.Errorf("listener ya está corriendo")
	}

	listenerMutex.Lock()
	currentListener = l
	listenerMutex.Unlock()

	l.isRunning = true

	go l.messageLoop()
	return nil
}

// Stop detiene el listener
func (l *EventListener) Stop() {
	l.mu.Lock()
	defer l.mu.Unlock()

	if !l.isRunning {
		return
	}

	close(l.stopChan)

	listenerMutex.Lock()
	currentListener = nil
	listenerMutex.Unlock()

	// Esperar a que el loop se cierre
	<-l.done
	l.isRunning = false
}

// GetEventChannel retorna el canal donde se envían los eventos
func (l *EventListener) GetEventChannel() <-chan PowerEvent {
	return l.eventChan
}

// messageLoop es el loop principal que procesa mensajes de Windows
func (l *EventListener) messageLoop() {
	defer func() {
		if l.hwnd != 0 {
			procDestroyWindow.Call(l.hwnd)
			l.hwnd = 0
		}
		close(l.eventChan)
		close(l.done)
	}()

	// Registrar clase de ventana
	className, _ := syscall.UTF16PtrFromString(l.windowName + "_Class")

	wndClass := WNDCLASS{
		Style:         CS_VREDRAW | CS_HREDRAW,
		LpfnWndProc:   syscall.NewCallback(windowProc),
		LpszClassName: className,
	}

	atom, _, _ := procRegisterClassW.Call(uintptr(unsafe.Pointer(&wndClass)))
	if atom == 0 {
		log.Println("[EventListener] Error: No se pudo registrar la clase de ventana")
		return
	}

	// Crear ventana oculta
	windowTitle, _ := syscall.UTF16PtrFromString(l.windowName)
	hwnd, _, _ := procCreateWindowExW.Call(
		0,                                    // dwExStyle
		uintptr(unsafe.Pointer(className)),   // lpClassName
		uintptr(unsafe.Pointer(windowTitle)), // lpWindowName
		WS_OVERLAPPED,                        // dwStyle
		0, 0, 0, 0,                           // x, y, nWidth, nHeight
		0, // hWndParent
		0, // hMenu
		0, // hInstance
		0, // lpParam
	)

	if hwnd == 0 {
		log.Println("[EventListener] Error: No se pudo crear la ventana oculta")
		return
	}

	l.hwnd = hwnd
	log.Printf("[EventListener] Ventana creada: %v\n", hwnd)

	// Registrar para recibir notificaciones de eventos de energía
	ret, _, _ := procRegisterPowerSettingNotification.Call(hwnd, uintptr(unsafe.Pointer(&GUID_MONITOR_POWER_ON)), DEVICE_NOTIFY_WINDOW_HANDLE)
	if ret == 0 {
		log.Println("[EventListener] Advertencia: RegisterPowerSettingNotification falló")
	} else {
		log.Println("[EventListener] Registrado para eventos de energía")
	}

	// Enviar un mensaje a nuestra propia ventana para activarla
	procPostMessageW.Call(hwnd, WM_APP, 0, 0)

	// Loop de mensajes
	var msg MSG
	for {
		ret, _, _ := procGetMessageW.Call(uintptr(unsafe.Pointer(&msg)), hwnd, 0, 0)
		if ret == 0 {
			break
		}

		select {
		case <-l.stopChan:
			procPostQuitMessage.Call(0)
			// Continuar procesando para que GetMessageW devuelva 0
			continue
		default:
		}

		if int(ret) == -1 {
			log.Println("[EventListener] Error en GetMessageW")
			break
		}

		procTranslateMessage.Call(uintptr(unsafe.Pointer(&msg)))
		procDispatchMessageW.Call(uintptr(unsafe.Pointer(&msg)))
	}
}

// windowProc es el callback de la ventana
func windowProc(hwnd uintptr, msg uint32, wParam uintptr, lParam uintptr) uintptr {
	switch msg {
	case WM_POWERBROADCAST:
		//log.Printf("[windowProc] WM_POWERBROADCAST recibido: wParam=%v, lParam=%v\n", wParam, lParam)
		event := powerBroadcastToEvent(wParam)
		if event != nil {
			listenerMutex.Lock()
			if currentListener != nil {
				select {
				case currentListener.eventChan <- *event:
					//log.Printf("[windowProc] Evento enviado: %v\n", event.Type)
				default:
					log.Println("[windowProc] Advertencia: Canal lleno")
				}
			}
			listenerMutex.Unlock()
		}
		return 1

	case WM_SETTINGCHANGE:
		// Detectar cambios de monitor
		//log.Printf("[windowProc] WM_SETTINGCHANGE recibido: wParam=%v\n", wParam)
		return 0

	case WM_DESTROY:
		procPostQuitMessage.Call(0)
		return 0

	default:
		ret, _, _ := procDefWindowProcW.Call(hwnd, uintptr(msg), wParam, lParam)
		return ret
	}
}

// powerBroadcastToEvent convierte un wParam de WM_POWERBROADCAST a un PowerEvent
func powerBroadcastToEvent(wParam uintptr) *PowerEvent {
	var eventType EventType
	var isValid bool

	switch wParam {
	case PBT_APMSUSPEND:
		eventType = EventSuspend
		isValid = true
	case PBT_APMRESUMEAUTOMATIC:
		eventType = EventResumeAutomatic
		isValid = true
	case PBT_APMRESUMESUSPEND:
		eventType = EventResumeSuspend
		isValid = true
	case PBT_APMRESUMECRITICAL:
		eventType = EventResumeSuspend
		isValid = true
	}

	if !isValid {
		log.Printf("[powerBroadcastToEvent] wParam desconocido: %v\n", wParam)
		return nil
	}

	return &PowerEvent{
		Type:      eventType,
		Timestamp: int64(0),
	}
}

func StartWindowsEvents(chn *utils.Canales) {
	listener := NewEventListener(10)
	err := listener.Start()
	if err != nil {
		log.Println("No se pudo iniciar el lector de eventos de windows")
		return
	}
	defer listener.Stop()

	eventChan := listener.GetEventChannel()
	log.Println("Escuchando eventos de Windows (suspensión, reanudación, etc.)")

	for event := range eventChan {
		log.Printf("[StartWindowsEvents] Evento detectado: %v\n", event.Type)
		switch event.Type {
		case EventSuspend: // Sistema entrando en suspensión
			log.Println("Sistema entrando en suspensión")
			chn.Suspended <- true
		case EventResumeAutomatic: // Sistema reactivado por timer
			log.Println("Sistema reactivado por timer")
			chn.Suspended <- false
		case EventResumeSuspend: // Sistema reactivado por usuario
			log.Println("Sistema reactivado")
			chn.Suspended <- false
		case EventMonitorOff: // Monitor apagado
			log.Println("Monitor apagado")
			chn.Suspended <- true
		case EventMonitorOn: // Monitor encendido
			log.Println("Monitor encendido")
			chn.Suspended <- false
		default:
			fmt.Printf("Evento desconocido: %v\n", event.Type)
		}
	}
}
