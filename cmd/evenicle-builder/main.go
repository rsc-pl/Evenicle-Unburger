package main

import (
	"flag"
	"fmt"
	"os"

	"evenicle-unburger/internal/builder"
	"evenicle-unburger/internal/tui"
)

func usage() {
	fmt.Fprintf(os.Stderr, `Evenicle Unburger Go builder

Usage:
  go run ./cmd/evenicle-builder <command> [options]

Commands:
  build            Build, verify and package the release (default)
  tui              Interactive setup/build frontend
  capture          Capture and validate pristine English 1.04 base files
  render-startup   Generate the startup-screen WebP
  clean            Remove generated build/output state
  install          Build and install into the configured game directory

Common options:
  --root PATH       Project root (auto-detected from current directory)
  --alice PATH      Explicit alice/alice.exe path (overrides config/PATH)
  --game-root PATH  Installed pristine English Evenicle 1.04 directory
`)
}

func main() {
	cmd := "build"
	args := os.Args[1:]
	if len(args) > 0 && args[0] != "" && args[0][0] != '-' {
		cmd = args[0]
		args = args[1:]
	}
	fs := flag.NewFlagSet(cmd, flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	root := fs.String("root", "", "project root")
	alice := fs.String("alice", "", "AliceTools path")
	game := fs.String("game-root", "", "English Evenicle game root")
	if err := fs.Parse(args); err != nil {
		os.Exit(2)
	}
	if cmd == "help" || cmd == "-h" || cmd == "--help" {
		usage()
		return
	}
	if cmd == "tui" {
		if err := tui.Run(*root); err != nil {
			fmt.Fprintln(os.Stderr, "ERROR:", err)
			os.Exit(1)
		}
		return
	}

	p, err := builder.Open(builder.Options{Root: *root, Alice: *alice, GameRoot: *game, SkipAlice: cmd == "clean" || cmd == "capture"})
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERROR:", err)
		os.Exit(1)
	}
	switch cmd {
	case "build":
		err = p.Build()
	case "capture":
		err = p.CaptureBase()
	case "render-startup":
		err = p.RenderStartup()
	case "clean":
		err = p.Clean()
	case "install":
		err = p.Install()
	default:
		usage()
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERROR:", err)
		os.Exit(1)
	}
}
