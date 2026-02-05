package win

import (
	"fmt"
	"syscall"
	"unsafe"
)

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

// Implementation interface for different capture methods
type captureImpl interface {
	Capture() []byte
	Close()
	IsAlive() bool
}

// Wrapper used by the rest of the program
type ScreenCapturer struct {
	impl          captureImpl
	width, height int
	lastMode      int // 0=auto, 1=GDI forced
}

// GDI implementation (original behavior)
type gdiImpl struct {
	width, height int
	hdc, memDC    uintptr
	bitmap        uintptr
	buf           []byte
	bmi           BITMAPINFO
}

func newGDIImpl(w, h int) *gdiImpl {
	hdc, _, _ := procGetDC.Call(0)
	memDC, _, _ := procCreateCompatibleDC.Call(hdc)
	bitmap, _, _ := procCreateCompatibleBitmap.Call(hdc, uintptr(w), uintptr(h))
	procSelectObject.Call(memDC, bitmap)

	g := &gdiImpl{
		width:  w,
		height: h,
		hdc:    hdc,
		memDC:  memDC,
		bitmap: bitmap,
		buf:    make([]byte, w*h*4),
	}

	g.bmi.Header = BITMAPINFOHEADER{
		Size:     uint32(unsafe.Sizeof(BITMAPINFOHEADER{})),
		Width:    int32(w),
		Height:   -int32(h), // top-down
		Planes:   1,
		BitCount: 32,
	}

	return g
}

func (g *gdiImpl) Capture() []byte {
	procBitBlt.Call(
		g.memDC, 0, 0,
		uintptr(g.width), uintptr(g.height),
		g.hdc, 0, 0, 0x00CC0020,
	)

	procGetDIBits.Call(
		g.memDC,
		g.bitmap,
		0,
		uintptr(g.height),
		uintptr(unsafe.Pointer(&g.buf[0])),
		uintptr(unsafe.Pointer(&g.bmi)),
		0,
	)

	return g.buf
}

func (g *gdiImpl) Close() {
	procDeleteDC.Call(g.memDC)
	procReleaseDC.Call(0, g.hdc)
}

func (g *gdiImpl) IsAlive() bool { return true }

// NewScreenCapturer tries DXGI first (if available) and falls back to GDI
func NewScreenCapturer(w, h int) *ScreenCapturer {
	sc := &ScreenCapturer{width: w, height: h}
	if dx := newDXGICapturer(w, h); dx != nil {
		sc.impl = dx
		return sc
	}

	// Fallback to GDI
	sc.impl = newGDIImpl(w, h)
	return sc
}

// Captura la pantalla y devuelve los bytes en formato BGRA
// mode: 0=auto (DXGI if available, else GDI), 1=GDI forced
func (s *ScreenCapturer) Capture(mode int) []byte {
	if s.lastMode != mode {
		fmt.Printf("[win] switching capture mode to %d\n", mode)
	}

	// si cambia de modo de forzado a auto, volvemos a intentar DXGI
	if mode == 0 && s.lastMode == 1 {
		if dx := newDXGICapturer(s.width, s.height); dx != nil {
			s.impl.Close()
			s.impl = dx
		}
	}
	s.lastMode = mode

	if mode == 1 {
		// Forced GDI
		if _, ok := s.impl.(*gdiImpl); !ok {
			s.impl.Close()
			s.impl = newGDIImpl(s.width, s.height)
		}
		return s.impl.Capture()
	}

	// If current implementation has died (e.g., DXGI duplication lost), switch to GDI
	if s.impl != nil && !s.impl.IsAlive() {
		fmt.Println("[win] capture impl not alive, switching to GDI fallback")
		s.impl.Close()
		s.impl = newGDIImpl(s.width, s.height)
	}
	return s.impl.Capture()
}

func (s *ScreenCapturer) Close() {
	if s.impl != nil {
		s.impl.Close()
	}
}
