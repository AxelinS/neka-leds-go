package win

import (
	"syscall"
	"unsafe"
)

var (
	dxgiDLL          = syscall.NewLazyDLL("dxgi_capture.dll")
	procDXGIInit     = dxgiDLL.NewProc("DXGI_Init")
	procDXGIGetFrame = dxgiDLL.NewProc("DXGI_GetFrame")
	procDXGIGetSize  = dxgiDLL.NewProc("DXGI_GetSize")
	procDXGIIsAlive  = dxgiDLL.NewProc("DXGI_IsAlive")
	procDXGIClose    = dxgiDLL.NewProc("DXGI_Close")
)

type dxgiCapturer struct {
	width, height int
	buf           []byte
	failCount     int
	alive         bool
}

const dxgiMaxFailures = 60 // number of consecutive failed frames before giving up

func newDXGICapturer() *dxgiCapturer {
	// Attempt to load DLL
	if err := dxgiDLL.Load(); err != nil {
		return nil
	}

	// Initialize without size - DLL will detect it
	ret, _, _ := procDXGIInit.Call()
	if ret == 0 {
		return nil
	}

	// Get actual size from DLL
	var w, h int32
	procDXGIGetSize.Call(uintptr(unsafe.Pointer(&w)), uintptr(unsafe.Pointer(&h)))

	if w <= 0 || h <= 0 {
		return nil
	}

	return &dxgiCapturer{
		width:  int(w),
		height: int(h),
		buf:    make([]byte, int(w)*int(h)*4),
		alive:  true,
	}
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
		// New frame copied successfully
		d.failCount = 0
		return d.buf
	}

	// v == 0 means timeout (no new frame)
	// v < 0 means error or duplication lost
	if v == 0 {
		// No new frame - don't return stale data
		d.failCount++
		return nil // Signal: no new frame
	}

	// Negative error code
	d.failCount += 5

	switch v {
	case -2:
		// Duplication lost - attempt to reinit DLL and resize buffer if needed
		ret, _, _ := procDXGIInit.Call()
		if ret == 0 {
			d.alive = false
			return nil
		}
		// Get new size
		var w, h int32
		procDXGIGetSize.Call(uintptr(unsafe.Pointer(&w)), uintptr(unsafe.Pointer(&h)))
		if w > 0 && h > 0 {
			newLen := int(w) * int(h) * 4
			if newLen != len(d.buf) {
				d.buf = make([]byte, newLen)
			}
			d.width = int(w)
			d.height = int(h)
			d.failCount = 0
			d.alive = true
		}
		return nil
	case -3:
		// Caller buffer too small - this is a programmer error in Go side
		panic("DXGI_GetFrame: buffer too small")
	default:
		if d.failCount > dxgiMaxFailures {
			d.alive = false
		}
		return nil // Error occurred - no data
	}
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
	ret, _, _ := procDXGIIsAlive.Call()
	return ret != 0
}
