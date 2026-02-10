//go:build !windows

package terminal

import (
	"os"
	"regexp"
	"syscall"
	"unsafe"

	"golang.org/x/term"
)

// collapseMultiLine converts a multi-line command (with \ continuations) into a single line
func collapseMultiLine(cmd string) string {
	// Replace backslash followed by newline and any leading whitespace with a single space
	// Handles both `\\\n` and `\\ \n` patterns
	re := regexp.MustCompile(`\\\s*\n\s*`)
	return re.ReplaceAllString(cmd, " ")
}

// InsertInput inserts a string into the terminal input buffer
// using TIOCSTI ioctl. This makes the text appear as if typed by the user.
func InsertInput(cmd string) error {
	// Collapse multi-line commands into single line for proper insertion
	cmd = collapseMultiLine(cmd)

	// Put terminal in raw mode to disable echo during insert
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return err
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

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
