package builder

import (
	"fmt"
	"path/filepath"
)

const (
	canonicalDev383Version      = "0.1.0-dev.383"
	canonicalDev383ChangesSHA   = "ef65e89fcf69bf9b3c21960829947e50bb9347c6e2687fb553f1287dc1d942ea"
	canonicalDev383StartupSHA   = "2d0c1a8473a2b2173d987cce2db3d46a8f23155ea23b9d44c82d998d1d975a2f"
	canonicalDev383AINOutputSHA = "f7dc624cb5941df445d77ee738c661eed384ab40eadeca1054f28a617c7e3b87"
	canonicalDev383EXOutputSHA  = "208a9cef095e7b89e70ae6d3ef8e752244315fa11a156776a883dfc9ab5d9271"
	canonicalDev383CG2OutputSHA = "5dab1008c024d61a1229b1213f2718327e4b105d1654978def48690cfd583c78"
	canonicalDev383ReleaseSHA   = "1d82d0d18e0f0d4ff892d5a277ef95d685b0f3a3d4bc9eb672e6e366d2483423"
)

// VerifyCanonicalDev383Identity turns the owner-runtime-tested dev.383 bytes
// into a hard build gate. The gate applies only when the editable canonical
// inputs are still the exact dev.383 inputs; later deliberate edits are built
// normally and establish their own release identity.
func (p *Project) VerifyCanonicalDev383Identity(releaseSHA string) (bool, error) {
	version, err := p.version()
	if err != nil {
		return false, err
	}
	changesSHA, err := shaFile(filepath.Join(p.Root, "patch", "changes.tsv"))
	if err != nil {
		return false, err
	}
	startupSHA, err := shaFile(filepath.Join(p.Root, "patch", "startup_screen.toml"))
	if err != nil {
		return false, err
	}
	if version != canonicalDev383Version || changesSHA != canonicalDev383ChangesSHA || startupSHA != canonicalDev383StartupSHA {
		return false, nil
	}
	wantFiles := map[string]string{
		"Evenicle.ain":    canonicalDev383AINOutputSHA,
		"Evenicle.ex":     canonicalDev383EXOutputSHA,
		"EvenicleCG2.afa": canonicalDev383CG2OutputSHA,
	}
	for name, want := range wantFiles {
		got, err := shaFile(filepath.Join(p.Root, "output", name))
		if err != nil {
			return true, err
		}
		if got != want {
			return true, fmt.Errorf("canonical dev.383 byte identity failed for %s: got %s want %s", name, got, want)
		}
	}
	if releaseSHA != canonicalDev383ReleaseSHA {
		return true, fmt.Errorf("canonical dev.383 release byte identity failed: got %s want %s", releaseSHA, canonicalDev383ReleaseSHA)
	}
	return true, nil
}
