package win

import "syscall"

var (
	user32dll            = syscall.NewLazyDLL("user32.dll")
	procGetSystemMetrics = user32dll.NewProc("GetSystemMetrics")
)

const (
	SM_CXSCREEN = 0
	SM_CYSCREEN = 1
)

func GetPrimaryMonitor() (left, top, width, height int) {
	w, _, _ := procGetSystemMetrics.Call(SM_CXSCREEN)
	h, _, _ := procGetSystemMetrics.Call(SM_CYSCREEN)

	return 0, 0, int(w), int(h)
}
