package builder

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"evenicle-unburger/internal/reprozip"
)

func (p *Project) version() (string, error) {
	b, err := os.ReadFile(filepath.Join(p.Root, "version.txt"))
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(b)), nil
}
func (p *Project) baseVersion() (string, error) {
	s, err := p.supported()
	if err != nil {
		return "", err
	}
	v, ok := s["Evenicle.ain"]
	if !ok {
		return "", fmt.Errorf("could not resolve supported Evenicle base version")
	}
	return v.Version, nil
}
func releaseReadme(version string) []byte {
	return []byte(fmt.Sprintf("Evenicle Unburger v%s\n\nFor the supported English PC release of Evenicle 1.04.\n\nInstallation:\n1. Back up your original Evenicle.ain, Evenicle.ex, EvenicleCG2.afa and Version.txt.\n2. Copy Evenicle.ain, Evenicle.ex, EvenicleCG2.afa and Version.txt from this archive into the game folder.\n3. Overwrite the existing files.\n\nVersion.txt adds the Unburger patch version to Help/About and the game window title.\nEvenicleCG2.afa adds the Unburger/Fidelity patch information screen at startup.\n\nTo uninstall, restore your original files.\n", version))
}
func (p *Project) MakeRelease() (string, string, error) {
	version, err := p.version()
	if err != nil {
		return "", "", err
	}
	base, err := p.baseVersion()
	if err != nil {
		return "", "", err
	}
	badge := []byte(fmt.Sprintf("%s / UNBURGER PATCH V%s", base, version))
	if len(badge) >= 256 {
		return "", "", fmt.Errorf("generated Version.txt exceeds engine CGameVersion limit")
	}
	outdir := filepath.Join(p.Root, "output")
	names := []string{"Evenicle.ain", "Evenicle.ex", "EvenicleCG2.afa"}
	entries := make([]reprozip.Entry, 0, 5)
	for _, n := range names {
		b, err := os.ReadFile(filepath.Join(outdir, n))
		if err != nil {
			return "", "", err
		}
		entries = append(entries, reprozip.Entry{Name: n, Data: b, Mode: 0o644})
	}
	entries = append(entries, reprozip.Entry{Name: "Version.txt", Data: badge, Mode: 0o644}, reprozip.Entry{Name: "README.txt", Data: releaseReadme(version), Mode: 0o644})
	zipname := fmt.Sprintf("Evenicle-Unburger-v%s.zip", version)
	zpath := filepath.Join(outdir, zipname)
	if err := reprozip.WritePythonCompatible(zpath, entries); err != nil {
		return "", "", err
	}
	sha, err := shaFile(zpath)
	if err != nil {
		return "", "", err
	}
	shaPath := zpath + ".sha256"
	if err := os.WriteFile(shaPath, []byte(fmt.Sprintf("%s  %s\n", sha, zipname)), 0o644); err != nil {
		return "", "", err
	}
	return zpath, sha, nil
}
func (p *Project) VerifyRelease(zpath string) error {
	version, err := p.version()
	if err != nil {
		return err
	}
	base, err := p.baseVersion()
	if err != nil {
		return err
	}
	badge := []byte(fmt.Sprintf("%s / UNBURGER PATCH V%s", base, version))
	zr, err := zip.OpenReader(zpath)
	if err != nil {
		return err
	}
	defer zr.Close()
	want := []string{"Evenicle.ain", "Evenicle.ex", "EvenicleCG2.afa", "Version.txt", "README.txt"}
	if len(zr.File) != len(want) {
		return fmt.Errorf("release member mismatch: %d", len(zr.File))
	}
	for i, f := range zr.File {
		if f.Name != want[i] {
			return fmt.Errorf("release member mismatch at %d: %q", i, f.Name)
		}
		rc, err := f.Open()
		if err != nil {
			return err
		}
		b, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			return err
		}
		switch f.Name {
		case "Version.txt":
			if string(b) != string(badge) {
				return fmt.Errorf("release Version.txt mismatch")
			}
		case "README.txt":
			s := string(b)
			for _, q := range []string{"Version.txt", "Help/About", "game window title", "EvenicleCG2.afa", "startup"} {
				if !strings.Contains(s, q) {
					return fmt.Errorf("release README missing %q", q)
				}
			}
		default:
			exp, err := os.ReadFile(filepath.Join(p.Root, "output", f.Name))
			if err != nil {
				return err
			}
			if shaBytes(b) != shaBytes(exp) {
				return fmt.Errorf("release %s does not match rebuilt output", f.Name)
			}
		}
	}
	return nil
}
func (p *Project) PublishRelease(zpath, sha string) error {
	release := filepath.Join(p.Root, "release")
	if err := os.MkdirAll(release, 0o755); err != nil {
		return err
	}
	ents, _ := os.ReadDir(release)
	for _, e := range ents {
		if strings.HasPrefix(e.Name(), "Evenicle-Unburger-v") && (strings.HasSuffix(e.Name(), ".zip") || strings.HasSuffix(e.Name(), ".zip.sha256")) {
			_ = os.Remove(filepath.Join(release, e.Name()))
		}
	}
	name := filepath.Base(zpath)
	if err := copyFile(zpath, filepath.Join(release, name)); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(release, name+".sha256"), []byte(fmt.Sprintf("%s  %s\n", sha, name)), 0o644)
}
