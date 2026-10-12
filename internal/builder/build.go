package builder

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"evenicle-unburger/internal/startup"
)

const overlayRel = "システム／起動警告.webp"

func (p *Project) Clean() error {
	for _, d := range []string{filepath.Join(p.Root, ".local", "build"), filepath.Join(p.Root, ".local", "release-stage"), filepath.Join(p.Root, "output")} {
		if err := os.RemoveAll(d); err != nil {
			return err
		}
	}
	return os.MkdirAll(filepath.Join(p.Root, "output"), 0o755)
}

func (p *Project) Build() error {
	fmt.Printf("[go-build] AliceTools: %s\n", p.Alice.Path)
	if err := p.EnsureBase(); err != nil {
		return err
	}
	approved, ordered, err := p.ReadChanges()
	if err != nil {
		return err
	}
	fmt.Printf("[go-build] public patch validation PASS managed=%d\n", len(ordered))
	if err := p.Clean(); err != nil {
		return err
	}
	b := filepath.Join(p.Root, ".local", "build")
	out := filepath.Join(p.Root, "output")
	src := p.sourceDir()
	edits := filepath.Join(b, "ain_edits.txt")
	if err := WriteEditFile(edits, ordered); err != nil {
		return err
	}
	_ = os.WriteFile(filepath.Join(b, "managed_change_count.txt"), []byte(fmt.Sprintf("%d\n", len(ordered))), 0o644)

	asset := filepath.Join(b, "generated", filepath.FromSlash(overlayRel))
	tempPNG := filepath.Join(b, "generated", "startup-screen.png")
	mode, err := startup.Generate(filepath.Join(p.Root, "patch", "startup_screen.toml"), filepath.Join(p.Root, "version.txt"), asset, tempPNG, func(in, out string) error { return p.Alice.Run("cg", "convert", "-t", "webp", in, out) })
	if err != nil {
		return err
	}
	ash, _ := shaFile(asset)
	fmt.Printf("[go-build] startup render PASS mode=%s sha256=%s\n", mode, ash)

	ainOut := filepath.Join(out, "Evenicle.ain")
	if err := p.Alice.Run("ain", "edit", "-t", edits, "-o", ainOut, filepath.Join(src, "Evenicle.ain")); err != nil {
		return err
	}
	if err := copyFile(filepath.Join(src, "Evenicle.ex"), filepath.Join(out, "Evenicle.ex")); err != nil {
		return err
	}
	if err := p.buildCG2(asset); err != nil {
		return err
	}
	if err := p.verifyAll(asset, approved); err != nil {
		return err
	}
	zpath, sha, err := p.MakeRelease()
	if err != nil {
		return err
	}
	fmt.Printf("[go-build] release %s\n[go-build] release sha256=%s\n", zpath, sha)
	canonical, err := p.VerifyCanonicalDev383Identity(sha)
	if err != nil {
		return err
	}
	if canonical {
		fmt.Printf("[go-build] canonical dev.383 byte identity PASS\n")
	}
	if err := p.VerifyRelease(zpath); err != nil {
		return err
	}
	fmt.Printf("[go-build] release package verify PASS\n")
	if err := p.PublishRelease(zpath, sha); err != nil {
		return err
	}
	fmt.Printf("\nBUILD_PASS\nRelease: %s\nSHA-256: %s\n", filepath.Join(p.Root, "release", filepath.Base(zpath)), filepath.Join(p.Root, "release", filepath.Base(zpath)+".sha256"))
	return nil
}

func (p *Project) buildCG2(asset string) error {
	b := filepath.Join(p.Root, ".local", "build")
	src := p.sourceDir()
	tree := filepath.Join(b, "cg2-src")
	manifest := filepath.Join(b, "cg2-manifest.txt")
	_ = os.RemoveAll(tree)
	if err := os.MkdirAll(tree, 0o755); err != nil {
		return err
	}
	if err := p.Alice.RunQuiet("ar", "extract", "--raw", "--manifest", manifest, "-o", tree, filepath.Join(src, "EvenicleCG2.afa")); err != nil {
		return err
	}
	dst := filepath.Join(tree, filepath.FromSlash(overlayRel))
	if err := copyFile(asset, dst); err != nil {
		return err
	}
	raw, err := os.ReadFile(manifest)
	if err != nil {
		return err
	}
	normalized := strings.ReplaceAll(string(raw), "\r\n", "\n")
	// Match Python str.splitlines(): a terminal newline is not an extra manifest
	// record. AliceTools treats a blank record before an appended member as the
	// end of the payload list.
	normalized = strings.TrimRight(normalized, "\n")
	lines := strings.Split(normalized, "\n")
	found := false
	for _, ln := range lines {
		if ln == overlayRel {
			found = true
			break
		}
	}
	if !found {
		lines = append(lines, overlayRel)
	}
	safe := filepath.Join(b, "manifests", "EvenicleCG2.afa.manifest")
	if err := os.MkdirAll(filepath.Dir(safe), 0o755); err != nil {
		return err
	}
	out := filepath.Join(p.Root, "output", "EvenicleCG2.afa")
	manifestDir := filepath.Dir(safe)
	treeRel, err := filepath.Rel(manifestDir, tree)
	if err != nil {
		return err
	}
	outRel, err := filepath.Rel(manifestDir, out)
	if err != nil {
		return err
	}
	// AliceTools resolves manifest paths relative to the manifest file, but its
	// option parser tokenizes on whitespace and does not honor shell quotes.
	// Using only relative paths avoids embedding the (possibly spaced) project
	// root and works on both Linux and Windows. The extracted manifest's first
	// line only supplies --src-dir for this archive, so replace it completely.
	first := "#ALICEPACK --src-dir=" + filepath.ToSlash(treeRel)
	if len(lines) > 0 && strings.HasPrefix(lines[0], "#ALICEPACK") {
		lines[0] = first
	} else {
		lines = append([]string{first}, lines...)
	}
	replaced := false
	for i := 1; i < len(lines); i++ {
		ln := strings.TrimSpace(lines[i])
		if ln != "" && !strings.HasPrefix(ln, "#") {
			lines[i] = filepath.ToSlash(outRel)
			replaced = true
			break
		}
	}
	if !replaced {
		return fmt.Errorf("archive manifest has no destination line")
	}
	if err := os.WriteFile(safe, []byte(strings.Join(lines, "\n")), 0o644); err != nil {
		return err
	}
	if err := p.Alice.Run("ar", "pack", safe); err != nil {
		return err
	}
	if !fileExists(out) {
		return fmt.Errorf("archive pack produced no output: %s", out)
	}
	return nil
}

func (p *Project) verifyAll(asset string, approved map[string]Change) error {
	b := filepath.Join(p.Root, ".local", "build")
	v := filepath.Join(b, "verify")
	_ = os.RemoveAll(v)
	_ = os.MkdirAll(v, 0o755)
	src := p.sourceDir()
	baseDump := filepath.Join(v, "base_ain.txt")
	builtDump := filepath.Join(v, "built_ain.txt")
	if err := p.Alice.RunQuiet("ain", "dump", "-t", "-o", baseDump, filepath.Join(src, "Evenicle.ain")); err != nil {
		return err
	}
	if err := p.Alice.RunQuiet("ain", "dump", "-t", "-o", builtDump, filepath.Join(p.Root, "output", "Evenicle.ain")); err != nil {
		return err
	}
	actual, ap, noops, err := VerifyAINDumps(baseDump, builtDump, approved)
	if err != nil {
		return err
	}
	fmt.Printf("[go-build] AIN managed-delta verify PASS actual_changed=%d approved=%d no_op_approved=%d\n", actual, ap, noops)
	exout, err := p.Alice.Capture("ex", "compare", filepath.Join(src, "Evenicle.ex"), filepath.Join(p.Root, "output", "Evenicle.ex"))
	if err != nil {
		return err
	}
	if !strings.Contains(exout, "EX files match") {
		return fmt.Errorf("EX compare failed: %s", strings.TrimSpace(exout))
	}
	baseDir := filepath.Join(v, "cg2-base")
	builtDir := filepath.Join(v, "cg2-built")
	_ = os.RemoveAll(baseDir)
	_ = os.RemoveAll(builtDir)
	_ = os.MkdirAll(baseDir, 0o755)
	_ = os.MkdirAll(builtDir, 0o755)
	if err := p.Alice.RunQuiet("ar", "extract", "--raw", "--manifest", filepath.Join(v, "cg2-base.manifest"), "-o", baseDir, filepath.Join(src, "EvenicleCG2.afa")); err != nil {
		return err
	}
	if err := p.Alice.RunQuiet("ar", "extract", "--raw", "--manifest", filepath.Join(v, "cg2-built.manifest"), "-o", builtDir, filepath.Join(p.Root, "output", "EvenicleCG2.afa")); err != nil {
		return err
	}
	nb, no, sha, err := VerifyCG2(baseDir, builtDir, asset, overlayRel)
	if err != nil {
		return err
	}
	fmt.Printf("[go-build] CG2 overlay verify PASS base_payloads=%d built_payloads=%d overlay=%s sha256=%s\n", nb, no, overlayRel, sha)
	res := filepath.Join(v, "RESULT.txt")
	return os.WriteFile(res, []byte("REBUILD_VERIFY_PASS\n"), 0o644)
}

func (p *Project) RenderStartup() error {
	b := filepath.Join(p.Root, ".local", "build")
	asset := filepath.Join(b, "generated", filepath.FromSlash(overlayRel))
	mode, err := startup.Generate(filepath.Join(p.Root, "patch", "startup_screen.toml"), filepath.Join(p.Root, "version.txt"), asset, filepath.Join(b, "generated", "startup-screen.png"), func(in, out string) error { return p.Alice.Run("cg", "convert", "-t", "webp", in, out) })
	if err != nil {
		return err
	}
	sha, _ := shaFile(asset)
	fmt.Printf("STARTUP_RENDER_PASS mode=%s output=%s sha256=%s\n", mode, asset, sha)
	return nil
}

func (p *Project) Install() error {
	if err := p.Build(); err != nil {
		return err
	}
	if p.GameRoot == "" {
		return fmt.Errorf("EVENICLE_EN_ROOT / --game-root is required for install")
	}
	game, err := filepath.Abs(p.GameRoot)
	if err != nil {
		return err
	}
	if !dirExists(game) {
		return fmt.Errorf("game root not found: %s", game)
	}
	src := p.sourceDir()
	for _, n := range []string{"Evenicle.ain", "Evenicle.ex", "EvenicleCG2.afa"} {
		gameFile := filepath.Join(game, n)
		backup := gameFile + ".unburger-backup"
		baseSha, err := shaFile(filepath.Join(src, n))
		if err != nil {
			return err
		}
		curSha, curErr := shaFile(gameFile)
		backupSha, backupErr := shaFile(backup)
		switch {
		case curErr == nil && curSha == baseSha:
			if os.IsNotExist(backupErr) {
				if err := copyFile(gameFile, backup); err != nil {
					return err
				}
			}
		case backupErr == nil && backupSha == baseSha:
			// Reinstalling over a patched tree is allowed only because a verified
			// pristine backup already exists.
		default:
			return fmt.Errorf("refusing install: %s is not pristine and no verified pristine backup exists", n)
		}
		tmp := gameFile + ".tmp-unburger"
		if err := copyFile(filepath.Join(p.Root, "output", n), tmp); err != nil {
			return err
		}
		if err := os.Rename(tmp, gameFile); err != nil {
			_ = os.Remove(tmp)
			return err
		}
	}
	version, _ := p.version()
	base, _ := p.baseVersion()
	verFile := filepath.Join(game, "Version.txt")
	verBackup := verFile + ".unburger-backup"
	if fileExists(verFile) && !fileExists(verBackup) {
		if err := copyFile(verFile, verBackup); err != nil {
			return err
		}
	}
	verTmp := verFile + ".tmp-unburger"
	if err := os.WriteFile(verTmp, []byte(fmt.Sprintf("%s / UNBURGER PATCH V%s", base, version)), 0o644); err != nil {
		return err
	}
	if err := os.Rename(verTmp, verFile); err != nil {
		_ = os.Remove(verTmp)
		return err
	}
	fmt.Printf("INSTALL_PASS root=%s\n", game)
	return nil
}

func readFirstLine(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	if s.Scan() {
		return s.Text(), nil
	}
	return "", s.Err()
}
