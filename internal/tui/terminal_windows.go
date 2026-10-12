//go:build windows

package tui

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

const (
	enableEchoInput             = 0x0004
	enableLineInput             = 0x0002
	enableVirtualTerminalInput  = 0x0200
	enableVirtualTerminalOutput = 0x0004
)

var (
	kernel32       = syscall.NewLazyDLL("kernel32.dll")
	getConsoleMode = kernel32.NewProc("GetConsoleMode")
	setConsoleMode = kernel32.NewProc("SetConsoleMode")
)

type terminal struct {
	inMode, outMode uint32
	saved           bool
}

func newTerminal() *terminal { return &terminal{} }

func consoleMode(fd uintptr) (uint32, error) {
	var mode uint32
	r, _, e := getConsoleMode.Call(fd, uintptr(unsafe.Pointer(&mode)))
	if r == 0 {
		return 0, e
	}
	return mode, nil
}

func setMode(fd uintptr, mode uint32) error {
	r, _, e := setConsoleMode.Call(fd, uintptr(mode))
	if r == 0 {
		return e
	}
	return nil
}

func (t *terminal) raw() error {
	inFD := os.Stdin.Fd()
	outFD := os.Stdout.Fd()
	in, err := consoleMode(inFD)
	if err != nil {
		return fmt.Errorf("GetConsoleMode(stdin): %w", err)
	}
	out, err := consoleMode(outFD)
	if err != nil {
		return fmt.Errorf("GetConsoleMode(stdout): %w", err)
	}
	raw := (in &^ (enableEchoInput | enableLineInput)) | enableVirtualTerminalInput
	if err := setMode(inFD, raw); err != nil {
		return fmt.Errorf("SetConsoleMode(stdin): %w", err)
	}
	if err := setMode(outFD, out|enableVirtualTerminalOutput); err != nil {
		_ = setMode(inFD, in)
		return fmt.Errorf("SetConsoleMode(stdout): %w", err)
	}
	t.inMode, t.outMode, t.saved = in, out, true
	return nil
}

func (t *terminal) restore() {
	if !t.saved {
		return
	}
	_ = setMode(os.Stdin.Fd(), t.inMode)
	_ = setMode(os.Stdout.Fd(), t.outMode)
	t.saved = false
}
