//go:build !linux && !windows

package tui

import "fmt"

type terminal struct{}

func newTerminal() *terminal { return &terminal{} }
func (t *terminal) raw() error {
	return fmt.Errorf("interactive TUI is currently supported on Linux and Windows")
}
func (t *terminal) restore() {}
