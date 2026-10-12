package builder

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

type Alice struct {
	Path string
	Root string
	Env  []string
}

func ResolveAlice(root, explicit string) (*Alice, error) {
	candidates := []string{}
	strictExplicitPath := false
	if explicit != "" {
		// A value containing a path separator is a true path override: if it is
		// wrong, fail instead of silently selecting another AliceTools binary.
		strictExplicitPath = filepath.IsAbs(explicit) || strings.ContainsAny(explicit, `/\\`)
		candidates = append(candidates, explicit)
	}
	if !strictExplicitPath {
		if p, err := exec.LookPath(executableName()); err == nil {
			candidates = append(candidates, p)
		}
		local := filepath.Join(root, ".local", "tools", executableName())
		candidates = append(candidates, local)
	}
	var tried []string
	for _, c := range candidates {
		if c == "" {
			continue
		}
		if !filepath.IsAbs(c) && strings.ContainsAny(c, `/\`) {
			c = filepath.Join(root, c)
		}
		p := c
		if !strings.ContainsAny(c, `/\`) {
			if lp, err := exec.LookPath(c); err == nil {
				p = lp
			}
		}
		p, _ = filepath.Abs(p)
		tried = append(tried, p)
		if !fileExists(p) {
			continue
		}
		a := &Alice{Path: p, Root: root, Env: os.Environ()}
		if runtime.GOOS != "windows" {
			lib := filepath.Join(root, ".local", "tools", "lib")
			if dirExists(lib) {
				old := os.Getenv("LD_LIBRARY_PATH")
				value := lib
				if old != "" {
					value += string(os.PathListSeparator) + old
				}
				a.Env = setEnv(a.Env, "LD_LIBRARY_PATH", value)
			}
		}
		if _, err := a.Capture("--help"); err != nil {
			return nil, fmt.Errorf("AliceTools found at %s but could not be invoked: %w", p, err)
		}
		return a, nil
	}
	return nil, fmt.Errorf("AliceTools (%s) not found; set ALICE_BIN / --alice or put it in PATH (tried: %s)", executableName(), strings.Join(tried, ", "))
}

func setEnv(env []string, key, val string) []string {
	prefix := key + "="
	out := make([]string, 0, len(env)+1)
	found := false
	for _, e := range env {
		if strings.HasPrefix(e, prefix) {
			if !found {
				out = append(out, prefix+val)
				found = true
			}
			continue
		}
		out = append(out, e)
	}
	if !found {
		out = append(out, prefix+val)
	}
	return out
}

func (a *Alice) cmd(args ...string) *exec.Cmd {
	c := exec.Command(a.Path, args...)
	c.Env = a.Env
	c.Dir = a.Root
	return c
}
func (a *Alice) Run(args ...string) error {
	c := a.cmd(args...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		return fmt.Errorf("%s %s: %w", a.Path, strings.Join(args, " "), err)
	}
	return nil
}
func (a *Alice) RunQuiet(args ...string) error {
	c := a.cmd(args...)
	var errb bytes.Buffer
	c.Stdout = &bytes.Buffer{}
	c.Stderr = &errb
	if err := c.Run(); err != nil {
		return fmt.Errorf("%s %s: %w: %s", a.Path, strings.Join(args, " "), err, strings.TrimSpace(errb.String()))
	}
	return nil
}
func (a *Alice) Capture(args ...string) (string, error) {
	c := a.cmd(args...)
	var b bytes.Buffer
	c.Stdout = &b
	c.Stderr = &b
	err := c.Run()
	if err != nil {
		return b.String(), fmt.Errorf("%s %s: %w: %s", a.Path, strings.Join(args, " "), err, strings.TrimSpace(b.String()))
	}
	return b.String(), nil
}
