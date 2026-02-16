package win

import (
	"syscall"
)

const (
	WM_POWERBROADCAST      = 0x0218
	PBT_APMSUSPEND         = 0x0004
	PBT_APMRESUMEAUTOMATIC = 0x0012
	PBT_APMRESUMESUSPEND   = 0x0007
	PBT_APMRESUMECRITICAL  = 0x0013
)

// capture
var (
	user32 = syscall.NewLazyDLL("user32.dll")
	gdi32  = syscall.NewLazyDLL("gdi32.dll")

	procGetDC     = user32.NewProc("GetDC")
	procReleaseDC = user32.NewProc("ReleaseDC")

	procCreateCompatibleDC     = gdi32.NewProc("CreateCompatibleDC")
	procCreateCompatibleBitmap = gdi32.NewProc("CreateCompatibleBitmap")
	procSelectObject           = gdi32.NewProc("SelectObject")
	procBitBlt                 = gdi32.NewProc("BitBlt")
	procDeleteDC               = gdi32.NewProc("DeleteDC")
	procDeleteObject           = gdi32.NewProc("DeleteObject")
	procGetDIBits              = gdi32.NewProc("GetDIBits")
)

type BITMAPINFOHEADER struct {
	Size          uint32
	Width         int32
	Height        int32
	Planes        uint16
	BitCount      uint16
	Compression   uint32
	SizeImage     uint32
	XPelsPerMeter int32
	YPelsPerMeter int32
	ClrUsed       uint32
	ClrImportant  uint32
}

type BITMAPINFO struct {
	Header BITMAPINFOHEADER
	Colors [1]uint32
}

type captureImpl interface { // Implementation interface for different capture methods
	Capture() []byte
	Close()
	IsAlive() bool
}

type ScreenCapturer struct { // Wrapper used by the rest of the program
	impl          captureImpl
	width, height int
	lastMode      int // 0=auto, 1=GDI forced
}

type gdiImpl struct { // GDI implementation (original behavior)
	width, height int
	hdc, memDC    uintptr
	bitmap        uintptr
	buf           []byte
	bmi           BITMAPINFO
}
