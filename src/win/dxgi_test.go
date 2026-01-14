package win

import (
	"os"
	"testing"
)

func TestDXGIInitAndCapture(t *testing.T) {
	// Skip if DLL not present next to package
	if _, err := os.Stat("dxgi_capture.dll"); os.IsNotExist(err) {
		t.Skip("dxgi_capture.dll not found; skipping DXGI integration test")
	}

	w, h := 100, 100
	dx := newDXGICapturer(w, h)
	if dx == nil {
		t.Skip("DXGI Init failed; possibly unsupported hardware or VM")
	}
	defer dx.Close()

	buf := dx.Capture()
	if buf == nil || len(buf) == 0 {
		t.Skip("No frame available; skipping")
	}
	if len(buf) < 4*w*h {
		t.Fatalf("Frame smaller than expected: got %d, want %d", len(buf), 4*w*h)
	}
}
