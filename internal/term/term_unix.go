//go:build darwin || linux

// Package term reports the size of the controlling terminal and detects
// whether a file descriptor is a terminal, using only the standard library.
package term

import (
	"os"
	"syscall"
	"unsafe"
)

// winsize mirrors struct winsize from <sys/ioctl.h>.
type winsize struct {
	Rows uint16
	Cols uint16
	X    uint16
	Y    uint16
}

// winsizeFD issues TIOCGWINSZ on fd. ok reports whether fd is a terminal.
func winsizeFD(fd uintptr) (ws winsize, ok bool) {
	_, _, errno := syscall.Syscall6(syscall.SYS_IOCTL, fd, syscall.TIOCGWINSZ,
		uintptr(unsafe.Pointer(&ws)), 0, 0, 0)
	return ws, errno == 0
}

// Size reports the controlling terminal's rows and columns. It returns
// ok=false when there is no controlling terminal or the size cannot be
// determined.
func Size() (rows, cols int, ok bool) {
	f, err := os.OpenFile("/dev/tty", os.O_RDONLY|syscall.O_NOCTTY, 0)
	if err != nil {
		return 0, 0, false
	}
	defer f.Close()
	ws, isTTY := winsizeFD(f.Fd())
	if !isTTY || ws.Rows == 0 || ws.Cols == 0 {
		return 0, 0, false
	}
	return int(ws.Rows), int(ws.Cols), true
}

// IsTerminal reports whether f refers to a terminal.
func IsTerminal(f *os.File) bool {
	if f == nil {
		return false
	}
	_, ok := winsizeFD(f.Fd())
	return ok
}
