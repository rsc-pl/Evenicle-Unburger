package tui

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"evenicle-unburger/internal/builder"
)

type key int

const (
	keyOther key = iota
	keyUp
	keyDown
	keyEnter
	keyQuit
)

const uiWidth = 88

type App struct {
	root  string
	term  *terminal
	in    *bufio.Reader
	color bool
}

type menuItem struct {
	Label       string
	Description string
	Run         func() error
}

const (
	ansiReset   = "\x1b[0m"
	ansiBold    = "\x1b[1m"
	ansiDim     = "\x1b[2m"
	ansiCyan    = "\x1b[36m"
	ansiGreen   = "\x1b[32m"
	ansiYellow  = "\x1b[33m"
	ansiRed     = "\x1b[31m"
	ansiBlue    = "\x1b[34m"
	ansiInverse = "\x1b[7m"
)

func Run(root string) error {
	r, err := builder.FindRoot(root)
	if err != nil {
		return err
	}
	app := &App{
		root:  r,
		term:  newTerminal(),
		in:    bufio.NewReader(os.Stdin),
		color: colorsEnabled(),
	}
	if err := app.term.raw(); err != nil {
		return fmt.Errorf("could not enter interactive terminal mode: %w", err)
	}
	defer func() {
		app.term.restore()
		fmt.Print("\x1b[?25h" + ansiReset)
	}()
	fmt.Print("\x1b[?25l")
	return app.mainLoop()
}

func colorsEnabled() bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	return !strings.EqualFold(strings.TrimSpace(os.Getenv("TERM")), "dumb")
}

func (a *App) c(code, s string) string {
	if !a.color {
		return s
	}
	return code + s + ansiReset
}

func (a *App) mainLoop() error {
	selected := 0
	for {
		st, err := builder.InspectSetup(a.root)
		if err != nil {
			return err
		}
		items := []menuItem{
			{
				"Setup / reconfigure",
				"Configure game path, capture pristine files and validate AliceTools",
				a.setupWizard,
			},
			{
				"Build patch",
				"Create the verified release ZIP without touching the game install",
				func() error {
					return a.runOperation("Build patch", func() error {
						p, err := builder.Open(builder.Options{Root: a.root})
						if err != nil {
							return err
						}
						return p.Build()
					})
				},
			},
			{
				"Build + install patch",
				"Build from pristine files, back up the game and install the patch",
				func() error {
					return a.runOperation("Build + install patch", func() error {
						p, err := builder.Open(builder.Options{Root: a.root})
						if err != nil {
							return err
						}
						return p.Install()
					})
				},
			},
			{
				"Recapture pristine base files",
				"Re-read and verify the supported English 1.04 files from the game",
				func() error {
					return a.runOperation("Recapture pristine Evenicle 1.04 base", func() error {
						p, err := builder.Open(builder.Options{Root: a.root, SkipAlice: true})
						if err != nil {
							return err
						}
						return p.CaptureBase()
					})
				},
			},
			{
				"Clean generated build files",
				"Remove generated output while keeping setup and pristine captures",
				func() error {
					return a.runOperation("Clean generated build files", func() error {
						p, err := builder.Open(builder.Options{Root: a.root, SkipAlice: true})
						if err != nil {
							return err
						}
						return p.Clean()
					})
				},
			},
			{"Exit", "Close the builder", nil},
		}
		if selected >= len(items) {
			selected = len(items) - 1
		}
		a.draw(st, items, selected)
		k, err := readKey()
		if err != nil {
			return err
		}
		switch k {
		case keyUp:
			selected--
			if selected < 0 {
				selected = len(items) - 1
			}
		case keyDown:
			selected++
			if selected >= len(items) {
				selected = 0
			}
		case keyEnter:
			if items[selected].Run == nil {
				return nil
			}
			if err := items[selected].Run(); err != nil {
				return err
			}
		case keyQuit:
			return nil
		}
	}
}

func (a *App) draw(st builder.SetupStatus, items []menuItem, selected int) {
	clearScreen()

	a.topBorder()
	title := "EVENICLE UNBURGER BUILDER"
	version := st.Version
	if version == "" {
		version = "version unknown"
	}
	a.boxedStyled("  "+title+"  "+version, "  "+a.c(ansiBold+ansiCyan, title)+"  "+a.c(ansiDim, version))
	a.boxedStyled("  Setup  •  Build  •  Install  •  Verify", "  "+a.c(ansiDim, "Setup  •  Build  •  Install  •  Verify"))
	a.bottomBorder()

	fmt.Println()
	a.section("ENVIRONMENT")
	a.statusRow("Game installation", st.GameRootOK, valueOr(st.GameRoot, "not configured"))
	baseDetail := filepath.Join(a.root, ".local", "source", "en")
	if st.BaseValid {
		a.statusRow("Pristine 1.04 base", true, baseDetail)
	} else if st.BaseCaptured {
		a.statusRow("Pristine 1.04 base", false, "captured but invalid")
		if st.BaseError != "" {
			a.infoRow("", shorten(st.BaseError, 68), ansiRed)
		}
	} else {
		a.statusRow("Pristine 1.04 base", false, "not captured")
	}
	if st.AliceValid {
		a.statusRow("AliceTools", true, st.AlicePath)
	} else {
		a.statusRow("AliceTools", false, "not configured / not found")
		if st.AliceError != "" {
			a.infoRow("", shorten(st.AliceError, 68), ansiYellow)
		}
	}
	if st.ReleaseExists {
		a.infoRow("Latest local release", st.ReleasePath, ansiGreen)
	} else {
		a.infoRow("Latest local release", "not built yet", ansiDim)
	}
	a.sectionBottom()

	buildReady := st.BaseValid && st.AliceValid
	installReady := buildReady && st.GameRootOK
	fmt.Println()
	a.section("READINESS")
	if buildReady {
		a.readinessRow("Build", "READY", "Pristine base and AliceTools validated", ansiGreen)
	} else {
		a.readinessRow("Build", "SETUP NEEDED", "Run Setup / reconfigure first", ansiYellow)
	}
	if installReady {
		a.readinessRow("Install", "READY", "Game folder is available", ansiGreen)
	} else if buildReady {
		a.readinessRow("Install", "GAME PATH NEEDED", "Build-only is still available", ansiYellow)
	} else {
		a.readinessRow("Install", "SETUP NEEDED", "Complete setup before installing", ansiYellow)
	}
	a.sectionBottom()

	fmt.Println()
	a.section("ACTIONS")
	for i, it := range items {
		a.menuRow(i == selected, it.Label, it.Description)
	}
	a.sectionBottom()

	fmt.Println()
	hint := "↑/↓ Navigate   Enter Select   Q Quit"
	fmt.Println("  " + a.c(ansiDim, hint))
}

func (a *App) topBorder() {
	fmt.Println(a.c(ansiCyan, "╭"+strings.Repeat("─", uiWidth-2)+"╮"))
}

func (a *App) bottomBorder() {
	fmt.Println(a.c(ansiCyan, "╰"+strings.Repeat("─", uiWidth-2)+"╯"))
}

func (a *App) section(title string) {
	left := "─ " + title + " "
	fill := uiWidth - 2 - len([]rune(left))
	if fill < 0 {
		fill = 0
	}
	fmt.Println(a.c(ansiBlue, "┌"+left+strings.Repeat("─", fill)+"┐"))
}

func (a *App) sectionBottom() {
	fmt.Println(a.c(ansiBlue, "└"+strings.Repeat("─", uiWidth-2)+"┘"))
}

func (a *App) boxedStyled(plain, styled string) {
	plain = shorten(plain, uiWidth-4)
	pad := uiWidth - 2 - len([]rune(plain))
	if pad < 0 {
		pad = 0
	}
	fmt.Printf("%s%s%s%s%s\n", a.c(ansiCyan, "│"), styled, strings.Repeat(" ", pad), "", a.c(ansiCyan, "│"))
}

func (a *App) sectionLine(plain, styled string) {
	plain = shorten(plain, uiWidth-4)
	pad := uiWidth - 2 - len([]rune(plain))
	if pad < 0 {
		pad = 0
	}
	fmt.Printf("%s%s%s%s\n", a.c(ansiBlue, "│"), styled, strings.Repeat(" ", pad), a.c(ansiBlue, "│"))
}

func (a *App) statusRow(label string, ok bool, value string) {
	icon := "!"
	iconStyle := ansiYellow + ansiBold
	if ok {
		icon = "✓"
		iconStyle = ansiGreen + ansiBold
	}
	value = shorten(value, 53)
	plain := fmt.Sprintf("  %s  %-21s %s", icon, label, value)
	styled := fmt.Sprintf("  %s  %-21s %s", a.c(iconStyle, icon), a.c(ansiBold, label), value)
	a.sectionLine(plain, styled)
}

func (a *App) infoRow(label, value, style string) {
	value = shorten(value, 53)
	if label == "" {
		plain := fmt.Sprintf("     %-21s %s", "", value)
		styled := fmt.Sprintf("     %-21s %s", "", a.c(style, value))
		a.sectionLine(plain, styled)
		return
	}
	plain := fmt.Sprintf("  •  %-21s %s", label, value)
	styled := fmt.Sprintf("  %s  %-21s %s", a.c(style, "•"), a.c(ansiBold, label), a.c(style, value))
	a.sectionLine(plain, styled)
}

func (a *App) readinessRow(label, state, detail, stateStyle string) {
	detail = shorten(detail, 43)
	plain := fmt.Sprintf("  %-9s %-16s %s", label, state, detail)
	styled := fmt.Sprintf("  %-9s %-16s %s", a.c(ansiBold, label), a.c(stateStyle+ansiBold, state), a.c(ansiDim, detail))
	a.sectionLine(plain, styled)
}

func (a *App) menuRow(active bool, label, desc string) {
	label = shorten(label, 27)
	desc = shorten(desc, 47)
	plain := fmt.Sprintf("  %-28s %s", label, desc)
	if active {
		plain = "▶ " + plain[2:]
		content := fmt.Sprintf("▶ %-26s %s", label, desc)
		styled := a.c(ansiCyan+ansiBold+ansiInverse, content)
		a.sectionLine(plain, styled)
		return
	}
	styled := fmt.Sprintf("  %-28s %s", label, a.c(ansiDim, desc))
	a.sectionLine(plain, styled)
}

func valueOr(v, fallback string) string {
	if strings.TrimSpace(v) == "" {
		return fallback
	}
	return v
}

func shorten(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	if n < 4 {
		return string(r[:n])
	}
	return string(r[:n-3]) + "..."
}

func (a *App) pageTitle(title, subtitle string) {
	clearScreen()
	a.topBorder()
	a.boxedStyled("  "+title, "  "+a.c(ansiBold+ansiCyan, title))
	if subtitle != "" {
		a.boxedStyled("  "+subtitle, "  "+a.c(ansiDim, subtitle))
	}
	a.bottomBorder()
	fmt.Println()
}

func (a *App) setupWizard() error {
	a.term.restore()
	defer a.term.raw()
	a.pageTitle("FIRST-TIME SETUP", "Validate the game, capture pristine files and configure AliceTools")
	fmt.Println("  " + a.c(ansiDim, "Nothing here is written into the public repository; settings stay under .local/."))
	fmt.Println()

	cfg, _ := builder.LoadLocalConfig(a.root)
	current, _ := builder.InspectSetup(a.root)
	gameDefault := cfg.GameRoot
	for {
		game, err := a.prompt("Evenicle installation folder (Q to cancel)", gameDefault)
		if err != nil {
			return err
		}
		if strings.EqualFold(strings.TrimSpace(game), "q") {
			return a.pause(a.c(ansiYellow, "Setup cancelled.") + " Press Enter to return to the main menu.")
		}
		if strings.TrimSpace(game) == "" {
			fmt.Println(a.c(ansiRed, "  A game folder is required."))
			fmt.Println()
			continue
		}
		abs, err := filepath.Abs(strings.TrimSpace(game))
		if err != nil {
			fmt.Printf("%s %v\n\n", a.c(ansiRed+ansiBold, "Invalid path:"), err)
			continue
		}

		sameConfiguredRoot := false
		if cfg.GameRoot != "" {
			if oldAbs, err := filepath.Abs(cfg.GameRoot); err == nil {
				sameConfiguredRoot = filepath.Clean(oldAbs) == filepath.Clean(abs)
			}
		}
		if current.BaseValid && sameConfiguredRoot {
			recapture, _ := a.confirm("A verified pristine base is already captured. Recapture it?", false)
			if !recapture {
				cfg.GameRoot = abs
				fmt.Println(a.c(ansiGreen, "  ✓ Keeping the existing verified pristine base."))
				fmt.Println()
				break
			}
		}

		fmt.Println("\n  " + a.c(ansiCyan+ansiBold, "Validating and capturing pristine 1.04 files..."))
		p := &builder.Project{Root: a.root, GameRoot: abs}
		if err := p.CaptureBase(); err != nil {
			fmt.Printf("\n%s %v\n", a.c(ansiRed+ansiBold, "Capture failed:"), err)
			retry, _ := a.confirm("Try another game folder?", true)
			if retry {
				gameDefault = abs
				fmt.Println()
				continue
			}
			return a.pauseWithError("Setup stopped before changing saved configuration.")
		}
		cfg.GameRoot = abs
		fmt.Printf("%s %s\n\n", a.c(ansiGreen+ansiBold, "  ✓ Pristine base captured:"), filepath.Join(a.root, ".local", "source", "en"))
		break
	}

	aliceDefault := cfg.AliceBin
	if detected, err := builder.ResolveAlice(a.root, aliceDefault); err == nil {
		fmt.Printf("%s %s\n", a.c(ansiGreen+ansiBold, "  ✓ AliceTools detected:"), detected.Path)
		use, _ := a.confirm("Use this AliceTools executable?", true)
		if use {
			cfg.AliceBin = detected.Path
		} else {
			cfg.AliceBin = ""
		}
	}
	for cfg.AliceBin == "" {
		path, err := a.prompt("Path to alice / alice.exe (blank = auto-detect, Q = cancel)", "")
		if err != nil {
			return err
		}
		if strings.EqualFold(strings.TrimSpace(path), "q") {
			return a.pause(a.c(ansiYellow, "Setup cancelled.") + " Press Enter to return to the main menu.")
		}
		candidate := strings.TrimSpace(path)
		detected, err := builder.ResolveAlice(a.root, candidate)
		if err != nil {
			fmt.Printf("%s %v\n\n", a.c(ansiRed+ansiBold, "AliceTools validation failed:"), err)
			continue
		}
		cfg.AliceBin = detected.Path
		fmt.Printf("%s %s\n", a.c(ansiGreen+ansiBold, "  ✓ AliceTools OK:"), detected.Path)
	}

	if err := builder.SaveLocalConfig(a.root, cfg); err != nil {
		return err
	}
	fmt.Println()
	fmt.Printf("%s %s\n", a.c(ansiGreen+ansiBold, "  ✓ Settings saved:"), filepath.Join(a.root, ".local", "config.env"))
	fmt.Println("  " + a.c(ansiDim, "The TUI and CLI now use the same machine-local configuration."))

	buildNow, _ := a.confirm("Build the patch now?", true)
	if buildNow {
		fmt.Println()
		fmt.Println(a.c(ansiCyan+ansiBold, "  Building verified patch..."))
		p, err := builder.Open(builder.Options{Root: a.root})
		if err != nil {
			return a.pauseWithError(err.Error())
		}
		if err := p.Build(); err != nil {
			return a.pauseWithError(err.Error())
		}
		fmt.Println("\n" + a.c(ansiGreen+ansiBold, "  ✓ Build completed successfully."))
	}
	return a.pause("Press Enter to return to the main menu.")
}

func (a *App) runOperation(title string, fn func() error) error {
	a.term.restore()
	defer a.term.raw()
	a.pageTitle(strings.ToUpper(title), "Running the canonical Go builder pipeline")
	fmt.Println(a.c(ansiDim, "  Builder output follows below. No verification gate is bypassed by the TUI."))
	fmt.Println()
	err := fn()
	fmt.Println()
	if err != nil {
		fmt.Printf("%s %v\n", a.c(ansiRed+ansiBold, "✗ FAILED:"), err)
	} else {
		fmt.Println(a.c(ansiGreen+ansiBold, "✓ DONE"))
	}
	_ = a.pause("Press Enter to return to the main menu.")
	return nil
}

func (a *App) prompt(label, def string) (string, error) {
	if def != "" {
		fmt.Printf("%s %s: ", a.c(ansiBold, label), a.c(ansiDim, "["+def+"]"))
	} else {
		fmt.Printf("%s: ", a.c(ansiBold, label))
	}
	line, err := a.in.ReadString('\n')
	if err != nil {
		return "", err
	}
	line = strings.TrimSpace(line)
	if line == "" {
		return def, nil
	}
	return trimOuterQuotes(line), nil
}

func (a *App) confirm(label string, def bool) (bool, error) {
	suffix := "[Y/n]"
	if !def {
		suffix = "[y/N]"
	}
	fmt.Printf("%s %s: ", a.c(ansiBold, label), a.c(ansiCyan, suffix))
	line, err := a.in.ReadString('\n')
	if err != nil {
		return false, err
	}
	v := strings.ToLower(strings.TrimSpace(line))
	if v == "" {
		return def, nil
	}
	return v == "y" || v == "yes", nil
}

func trimOuterQuotes(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 && ((s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'')) {
		return s[1 : len(s)-1]
	}
	return s
}

func (a *App) pause(msg string) error {
	fmt.Print("\n" + msg + " ")
	_, err := a.in.ReadString('\n')
	return err
}

func (a *App) pauseWithError(msg string) error {
	fmt.Printf("\n%s %s\n", a.c(ansiRed+ansiBold, "✗"), msg)
	return a.pause("Press Enter to return to the main menu.")
}

func clearScreen() { fmt.Print("\x1b[2J\x1b[H") }

func readKey() (key, error) {
	var b [1]byte
	if _, err := os.Stdin.Read(b[:]); err != nil {
		return keyOther, err
	}
	switch b[0] {
	case 'q', 'Q':
		return keyQuit, nil
	case '\r', '\n':
		return keyEnter, nil
	case 0x1b:
		var seq [2]byte
		if _, err := os.Stdin.Read(seq[:]); err != nil {
			return keyOther, err
		}
		if seq[0] == '[' {
			if seq[1] == 'A' {
				return keyUp, nil
			}
			if seq[1] == 'B' {
				return keyDown, nil
			}
		}
	}
	return keyOther, nil
}
