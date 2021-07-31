package window

import (
	"github.com/go-vgo/robotgo"
	"log"
	"syscall"
	"unsafe"
)

type HWND uintptr

type Rect struct {
	Left, Top, Bottom, Right int32
}

type ApplicationInfo struct {
	Hwnd      HWND
	Pid       int32
	Left, Top int32
}

var (
	user32, _        = syscall.LoadLibrary("user32.dll")
	findWindowW, _   = syscall.GetProcAddress(user32, "FindWindowW")
	getWindowRect, _ = syscall.GetProcAddress(user32, "GetWindowRect")
	getClientRect, _ = syscall.GetProcAddress(user32, "GetClientRect")
)

func FindWindowByTitle(title string) HWND {
	ret, _, _ := syscall.Syscall(
		findWindowW,
		2,
		0,
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(title))),
		0,
	)
	return HWND(ret)
}

func GetWindowDimensions(hwnd HWND) *Rect {
	var rect Rect

	syscall.Syscall(
		getWindowRect,
		2,
		uintptr(hwnd),
		uintptr(unsafe.Pointer(&rect)),
		0,
	)

	return &rect
}

func GetClientDimensions(hwnd HWND) *Rect {
	var rect Rect

	syscall.Syscall(
		getClientRect,
		2,
		uintptr(hwnd),
		uintptr(unsafe.Pointer(&rect)),
		0,
	)

	return &rect
}

func GetApplicationInfo(name string, processName string) ApplicationInfo {
	defer syscall.FreeLibrary(user32)

	hwnd := FindWindowByTitle(name)
	pids, err := robotgo.FindIds(processName)

	if err != nil {
		panic(err)
	}

	if len(pids) > 1 {
		panic("application " + name + " has more than one process")
	}

	if hwnd > 0 {
		pid := pids[0]
		left, top, _, _ := robotgo.GetBounds(pid)
		log.Printf("`%s` found, geometry rect is: left: `%d`, top: `%d`", name, left, top)

		return ApplicationInfo{
			Hwnd: hwnd,
			Pid:  pid,
			Left: int32(left),
			Top:  int32(top),
		}
	}

	panic("application " + name + " not found")
}
