//go:build !windows

package terminal

import (
	"syscall"
	"unsafe"
)

// InsertInput inserts a string into the terminal input buffer
// using TIOCSTI ioctl. This makes the text appear as if typed by the user.
func InsertInput(cmd string) error {
	for _, c := range cmd {
		char := byte(c)
		_, _, errno := syscall.Syscall(
			syscall.SYS_IOCTL,
			uintptr(0), // stdin
			syscall.TIOCSTI,
			uintptr(unsafe.Pointer(&char)),
		)
		if errno != 0 {
			return errno
		}
	}
	return nil
}
