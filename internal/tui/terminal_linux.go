//go:build linux

package tui

import (
	"os"
	"syscall"
	"unsafe"
)

type terminal struct {
	old   syscall.Termios
	saved bool
}

func newTerminal() *terminal { return &terminal{} }

func (t *terminal) raw() error {
	fd := os.Stdin.Fd()
	var old syscall.Termios
	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fd, uintptr(syscall.TCGETS), uintptr(unsafe.Pointer(&old))); errno != 0 {
		return errno
	}
	raw := old
	raw.Lflag &^= syscall.ECHO | syscall.ICANON
	raw.Iflag &^= syscall.IXON | syscall.ICRNL
	raw.Cc[syscall.VMIN] = 1
	raw.Cc[syscall.VTIME] = 0
	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fd, uintptr(syscall.TCSETS), uintptr(unsafe.Pointer(&raw))); errno != 0 {
		return errno
	}
	t.old = old
	t.saved = true
	return nil
}

func (t *terminal) restore() {
	if !t.saved {
		return
	}
	fd := os.Stdin.Fd()
	_, _, _ = syscall.Syscall(syscall.SYS_IOCTL, fd, uintptr(syscall.TCSETS), uintptr(unsafe.Pointer(&t.old)))
	t.saved = false
}
