package win

import (
	"syscall"
	"unsafe"
)

var (
	dxgiDLL          = syscall.NewLazyDLL("dxgi_capture.dll")
	procDXGIInit     = dxgiDLL.NewProc("DXGI_Init")
	procDXGIGetFrame = dxgiDLL.NewProc("DXGI_GetFrame")
	procDXGIClose    = dxgiDLL.NewProc("DXGI_Close")
)

type dxgiCapturer struct {
	width, height int
	buf           []byte
	failCount     int
	alive         bool
}

const dxgiMaxFailures = 60 // number of consecutive failed frames before giving up

func newDXGICapturer(w, h int) *dxgiCapturer {
	// Attempt to load DLL
	if err := dxgiDLL.Load(); err != nil {
		return nil
	}
	ret, _, _ := procDXGIInit.Call(uintptr(w), uintptr(h))
	if ret == 0 {
		return nil
	}
	return &dxgiCapturer{width: w, height: h, buf: make([]byte, w*h*4), alive: true}
}

func (d *dxgiCapturer) Capture() []byte {
	if d == nil {
		return nil
	}
	if len(d.buf) == 0 {
		return nil
	}
	ret, _, _ := procDXGIGetFrame.Call(uintptr(unsafe.Pointer(&d.buf[0])), uintptr(len(d.buf)))
	v := int32(ret)
	if v > 0 {
		// New frame copied
		d.failCount = 0
		return d.buf
	}
	// v == 0 means timeout (no new frame), negative means error or duplication lost
	if v == 0 {
		// no new frame; count as minor failure
		d.failCount++
	} else {
		// error or duplication lost
		d.failCount += 5
		// Try to reinit if duplication lost (-2)
		if v == -2 {
			// call init again to attempt to recover
			procDXGIInit.Call(uintptr(d.width), uintptr(d.height))
		}
	}
	if d.failCount > dxgiMaxFailures {
		d.alive = false
	}
	return d.buf
}

func (d *dxgiCapturer) Close() {
	if d == nil {
		return
	}
	procDXGIClose.Call()
}

func (d *dxgiCapturer) IsAlive() bool {
	if d == nil {
		return false
	}
	return d.alive
}
