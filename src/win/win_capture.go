package win

import (
	"fmt"
	"unsafe"
)

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
	procDeleteObject.Call(g.bitmap) // Clean up bitmap resources
	procDeleteDC.Call(g.memDC)
	procReleaseDC.Call(0, g.hdc)
}

func (g *gdiImpl) IsAlive() bool { return true }

// NewScreenCapturer tries DXGI first (if available) and falls back to GDI
func NewScreenCapturer(w, h int) *ScreenCapturer {
	sc := &ScreenCapturer{width: w, height: h}

	// Try DXGI first - it will auto-detect monitor resolution
	if dx := newDXGICapturer(); dx != nil {
		sc.impl = dx
		// Update dimensions to actual monitor size from DXGI
		sc.width = dx.width
		sc.height = dx.height
		return sc
	}

	// Fallback to GDI with provided dimensions
	sc.impl = newGDIImpl(w, h)
	return sc
}

// Captura la pantalla y devuelve los bytes en formato BGRA
// mode: 0=auto (DXGI if available, else GDI), 1=GDI forced
func (s *ScreenCapturer) Capture(mode int) []byte {
	if s.lastMode != mode {
		fmt.Printf("[win] switching capture mode to %d\n", mode)
	}

	// if switching from forced mode to auto, try DXGI again
	if mode == 0 && s.lastMode == 1 {
		if dx := newDXGICapturer(); dx != nil {
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

	// impl.Capture() may return nil if:
	// - DXGI: no new frame (timeout) or error
	// - GDI: should always return data
	return s.impl.Capture()
}

func (s *ScreenCapturer) Close() {
	if s.impl != nil {
		s.impl.Close()
	}
}
