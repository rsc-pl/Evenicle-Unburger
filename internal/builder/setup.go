package builder

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type LocalConfig struct {
	GameRoot string
	AliceBin string
}

type SetupStatus struct {
	Root          string
	Version       string
	ConfigPath    string
	GameRoot      string
	GameRootOK    bool
	BaseCaptured  bool
	BaseValid     bool
	BaseError     string
	AlicePath     string
	AliceValid    bool
	AliceError    string
	ReleasePath   string
	ReleaseExists bool
}

func LoadLocalConfig(root string) (LocalConfig, error) {
	cfg, err := readConfig(filepath.Join(root, ".local", "config.env"))
	if err != nil {
		return LocalConfig{}, err
	}
	return LocalConfig{GameRoot: cfg["EVENICLE_EN_ROOT"], AliceBin: cfg["ALICE_BIN"]}, nil
}

func SaveLocalConfig(root string, cfg LocalConfig) error {
	if err := os.MkdirAll(filepath.Join(root, ".local"), 0o755); err != nil {
		return err
	}
	quote := func(s string) string {
		// config.env is parsed by this builder, not sourced by a shell. Preserve
		// Windows backslashes literally and only escape embedded quotes.
		return `"` + strings.ReplaceAll(s, `"`, `\"`) + `"`
	}
	var b strings.Builder
	b.WriteString("# Machine-local paths. This file is ignored by Git.\n")
	b.WriteString("EVENICLE_PROJECT_ROOT=" + quote(root) + "\n")
	if cfg.GameRoot != "" {
		b.WriteString("EVENICLE_EN_ROOT=" + quote(cfg.GameRoot) + "\n")
	}
	if cfg.AliceBin != "" {
		b.WriteString("ALICE_BIN=" + quote(cfg.AliceBin) + "\n")
	}
	path := filepath.Join(root, ".local", "config.env")
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, []byte(b.String()), 0o644); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return nil
}

func InspectSetup(root string) (SetupStatus, error) {
	r, err := FindRoot(root)
	if err != nil {
		return SetupStatus{}, err
	}
	cfg, err := LoadLocalConfig(r)
	if err != nil {
		return SetupStatus{}, err
	}
	st := SetupStatus{Root: r, ConfigPath: filepath.Join(r, ".local", "config.env"), GameRoot: cfg.GameRoot}
	if v, err := os.ReadFile(filepath.Join(r, "version.txt")); err == nil {
		st.Version = strings.TrimSpace(string(v))
	}
	if cfg.GameRoot != "" {
		if abs, err := filepath.Abs(cfg.GameRoot); err == nil {
			st.GameRoot = abs
		}
		st.GameRootOK = dirExists(st.GameRoot)
	}
	p := &Project{Root: r, GameRoot: st.GameRoot}
	src := p.sourceDir()
	st.BaseCaptured = fileExists(filepath.Join(src, "Evenicle.ain")) && fileExists(filepath.Join(src, "Evenicle.ex")) && fileExists(filepath.Join(src, "EvenicleCG2.afa")) && fileExists(filepath.Join(src, "Version.txt"))
	if st.BaseCaptured {
		if err := p.validateCapturedBase(); err != nil {
			st.BaseError = err.Error()
		} else {
			st.BaseValid = true
		}
	}
	explicit := cfg.AliceBin
	if explicit == "" {
		explicit = os.Getenv("ALICE_BIN")
	}
	if a, err := ResolveAlice(r, explicit); err != nil {
		st.AliceError = err.Error()
	} else {
		st.AliceValid = true
		st.AlicePath = a.Path
	}
	if st.Version != "" {
		st.ReleasePath = filepath.Join(r, "release", "Evenicle-Unburger-v"+st.Version+".zip")
		st.ReleaseExists = fileExists(st.ReleasePath)
	}
	return st, nil
}

func (p *Project) ValidateGameRoot(game string) error {
	if game == "" {
		return fmt.Errorf("game directory is empty")
	}
	abs, err := filepath.Abs(game)
	if err != nil {
		return err
	}
	if !dirExists(abs) {
		return fmt.Errorf("game directory not found: %s", abs)
	}
	tmp, err := p.preparePristineCapture(abs)
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)
	return p.validateBaseDir(tmp)
}
