package builder

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Supported struct{ Version, SHA, Note string }

func (p *Project) supported() (map[string]Supported, error) {
	f, err := os.Open(filepath.Join(p.Root, "data", "supported_bases.tsv"))
	if err != nil {
		return nil, err
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	first := true
	out := map[string]Supported{}
	for s.Scan() {
		if first {
			first = false
			continue
		}
		cols := strings.Split(s.Text(), "\t")
		if len(cols) < 3 {
			continue
		}
		n := cols[0]
		note := ""
		if len(cols) > 3 {
			note = cols[3]
		}
		out[n] = Supported{cols[1], cols[2], note}
	}
	return out, s.Err()
}

func shaFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func (p *Project) sourceDir() string { return filepath.Join(p.Root, ".local", "source", "en") }

func (p *Project) EnsureBase() error {
	src := p.sourceDir()
	required := []string{"Evenicle.ain", "Evenicle.ex", "EvenicleCG2.afa", "Version.txt"}
	missing := false
	for _, n := range required {
		if !fileExists(filepath.Join(src, n)) {
			missing = true
			break
		}
	}
	if missing {
		if p.GameRoot == "" {
			return fmt.Errorf("missing captured pristine EN source under .local/source/en; set EVENICLE_EN_ROOT / --game-root and run capture")
		}
		if err := p.CaptureBase(); err != nil {
			return err
		}
	}
	return p.validateCapturedBase()
}

func (p *Project) validateBaseDir(src string) error {
	verRaw, err := os.ReadFile(filepath.Join(src, "Version.txt"))
	if err != nil {
		return err
	}
	ver := strings.TrimSpace(string(verRaw))
	if ver != "1.04" {
		return fmt.Errorf("unsupported captured Version.txt %q; expected 1.04", ver)
	}
	sup, err := p.supported()
	if err != nil {
		return err
	}
	for _, n := range []string{"Evenicle.ain", "Evenicle.ex", "EvenicleCG2.afa"} {
		want, ok := sup[n]
		if !ok {
			return fmt.Errorf("supported base table has no %s entry", n)
		}
		got, err := shaFile(filepath.Join(src, n))
		if err != nil {
			return err
		}
		if got != want.SHA {
			return fmt.Errorf("unsupported %s SHA-256: %s; expected %s", n, got, want.SHA)
		}
	}
	return nil
}

func (p *Project) validateCapturedBase() error { return p.validateBaseDir(p.sourceDir()) }

func (p *Project) selectPristineFile(game, name string, wantSHA string) (string, error) {
	candidates := []string{filepath.Join(game, name), filepath.Join(game, name+".unburger-backup")}
	var seen []string
	for _, path := range candidates {
		if !fileExists(path) {
			continue
		}
		seen = append(seen, path)
		if name == "Version.txt" {
			raw, err := os.ReadFile(path)
			if err == nil && strings.TrimSpace(string(raw)) == "1.04" {
				return path, nil
			}
			continue
		}
		got, err := shaFile(path)
		if err == nil && got == wantSHA {
			return path, nil
		}
	}
	if len(seen) == 0 {
		return "", fmt.Errorf("required pristine %s not found in %s (nor verified .unburger-backup)", name, game)
	}
	return "", fmt.Errorf("%s is not the supported pristine 1.04 file and no verified .unburger-backup was found", name)
}

// preparePristineCapture builds a fully validated staging tree. Core files are
// selected from the live install when pristine or from verified
// .unburger-backup files when the game is already patched by this builder.
// The existing captured source is never modified here.
func (p *Project) preparePristineCapture(game string) (string, error) {
	sup, err := p.supported()
	if err != nil {
		return "", err
	}
	parent := filepath.Join(p.Root, ".local", "source")
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return "", err
	}
	tmp := filepath.Join(parent, "en.__capture_tmp")
	if err := os.RemoveAll(tmp); err != nil {
		return "", err
	}
	if err := os.MkdirAll(tmp, 0o755); err != nil {
		return "", err
	}
	fail := func(err error) (string, error) {
		_ = os.RemoveAll(tmp)
		return "", err
	}

	for _, n := range []string{"Evenicle.ain", "Evenicle.ex", "EvenicleCG2.afa"} {
		want, ok := sup[n]
		if !ok {
			return fail(fmt.Errorf("supported base table has no %s entry", n))
		}
		from, err := p.selectPristineFile(game, n, want.SHA)
		if err != nil {
			return fail(err)
		}
		if err := copyFile(from, filepath.Join(tmp, n)); err != nil {
			return fail(err)
		}
	}
	ver, err := p.selectPristineFile(game, "Version.txt", "")
	if err != nil {
		return fail(err)
	}
	if err := copyFile(ver, filepath.Join(tmp, "Version.txt")); err != nil {
		return fail(err)
	}

	// These files are useful for local development/reference but are not part
	// of the byte-identity contract. Copy them when present without weakening
	// validation of the four required pristine inputs above.
	optional := []string{"Evenicle.exe", "EvenicleFlat.afa", "EveniclePact.afa", "EvenicleFont.fnl", "AliceStart.ini", "AsraFrameworkDLL.inc", "Patchnote.txt"}
	for _, n := range optional {
		from := filepath.Join(game, n)
		if fileExists(from) {
			if err := copyFile(from, filepath.Join(tmp, n)); err != nil {
				return fail(err)
			}
		}
	}
	if err := p.validateBaseDir(tmp); err != nil {
		return fail(err)
	}
	return tmp, nil
}

func replaceDirAtomic(staged, dst string) error {
	old := dst + ".__capture_old"
	_ = os.RemoveAll(old)
	hadOld := dirExists(dst)
	if hadOld {
		if err := os.Rename(dst, old); err != nil {
			return err
		}
	}
	if err := os.Rename(staged, dst); err != nil {
		if hadOld {
			_ = os.Rename(old, dst)
		}
		return err
	}
	if hadOld {
		_ = os.RemoveAll(old)
	}
	return nil
}

func (p *Project) CaptureBase() error {
	if p.GameRoot == "" {
		return fmt.Errorf("EVENICLE_EN_ROOT / --game-root is required for capture")
	}
	game, err := filepath.Abs(p.GameRoot)
	if err != nil {
		return err
	}
	if !dirExists(game) {
		return fmt.Errorf("English game root not found: %s", game)
	}

	staged, err := p.preparePristineCapture(game)
	if err != nil {
		return err
	}
	src := p.sourceDir()
	if err := replaceDirAtomic(staged, src); err != nil {
		_ = os.RemoveAll(staged)
		return err
	}
	if err := p.EnsureCapturedCore(); err != nil {
		return err
	}
	if err := p.validateCapturedBase(); err != nil {
		return err
	}

	rep := filepath.Join(p.Root, ".local", "reports")
	_ = os.MkdirAll(rep, 0o755)
	entries, _ := os.ReadDir(src)
	var names []string
	for _, e := range entries {
		if !e.IsDir() {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	var sums, cap strings.Builder
	cap.WriteString("EN_SOURCE_READY\nversion=1.04\nroot=" + game + "\n")
	ainsha, _ := shaFile(filepath.Join(src, "Evenicle.ain"))
	cap.WriteString("ain_sha256=" + ainsha + "\n")
	for _, n := range names {
		path := filepath.Join(src, n)
		sha, _ := shaFile(path)
		st, _ := os.Stat(path)
		sums.WriteString(sha + "  " + path + "\n")
		fmt.Fprintf(&cap, "%s\t%d bytes\n", n, st.Size())
	}
	_ = os.WriteFile(filepath.Join(rep, "en_source_sha256.txt"), []byte(sums.String()), 0o644)
	_ = os.WriteFile(filepath.Join(rep, "en_source_capture.txt"), []byte(cap.String()), 0o644)
	fmt.Print(cap.String())
	return nil
}

func (p *Project) EnsureCapturedCore() error {
	for _, n := range []string{"Evenicle.ain", "Evenicle.ex", "EvenicleCG2.afa", "Version.txt"} {
		if !fileExists(filepath.Join(p.sourceDir(), n)) {
			return fmt.Errorf("%s missing from captured pristine source", n)
		}
	}
	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	_, e := io.Copy(out, in)
	ce := out.Close()
	if e != nil {
		return e
	}
	return ce
}
