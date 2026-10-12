package builder

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var dumpEntry = regexp.MustCompile(`^\s*;?\s*([ms])\[(\d+)\]\s*=\s*"(.*)"\s*$`)

func decodeDump(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' && i+1 < len(s) {
			i++
			switch s[i] {
			case 'n':
				b.WriteByte('\n')
			case 'r':
				b.WriteByte('\r')
			case 't':
				b.WriteByte('\t')
			case '"':
				b.WriteByte('"')
			case '\\':
				b.WriteByte('\\')
			default:
				b.WriteByte('\\')
				b.WriteByte(s[i])
			}
		} else {
			b.WriteByte(s[i])
		}
	}
	return b.String()
}
func parseDump(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	out := map[string]string{}
	s := bufio.NewScanner(f)
	buf := make([]byte, 64*1024)
	s.Buffer(buf, 16*1024*1024)
	for s.Scan() {
		m := dumpEntry.FindStringSubmatch(s.Text())
		if m == nil {
			continue
		}
		out[m[1]+"["+m[2]+"]"] = decodeDump(m[3])
	}
	return out, s.Err()
}
func VerifyAINDumps(basePath, builtPath string, approved map[string]Change) (int, int, int, error) {
	base, err := parseDump(basePath)
	if err != nil {
		return 0, 0, 0, err
	}
	built, err := parseDump(builtPath)
	if err != nil {
		return 0, 0, 0, err
	}
	actual := 0
	keys := map[string]bool{}
	for k := range base {
		keys[k] = true
	}
	for k := range built {
		keys[k] = true
	}
	var unexpected []string
	for k := range keys {
		if base[k] != built[k] {
			actual++
			if _, ok := approved[k]; !ok {
				unexpected = append(unexpected, k)
			}
		}
	}
	var missing []string
	noop := 0
	for k, c := range approved {
		if built[k] != c.Text {
			missing = append(missing, k)
		}
		if base[k] == c.Text {
			noop++
		}
	}
	if len(unexpected) > 0 || len(missing) > 0 {
		return actual, len(approved), noop, fmt.Errorf("AIN delta mismatch unexpected=%v missing=%v", head(unexpected, 20), head(missing, 20))
	}
	return actual, len(approved), noop, nil
}
func head(a []string, n int) []string {
	if len(a) > n {
		return a[:n]
	}
	return a
}
func treeFiles(root string) (map[string]string, error) {
	out := map[string]string{}
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		sum, err := shaFile(path)
		if err != nil {
			return err
		}
		out[filepath.ToSlash(rel)] = sum
		return nil
	})
	return out, err
}
func VerifyCG2(baseDir, builtDir, asset, rel string) (int, int, string, error) {
	bf, err := treeFiles(baseDir)
	if err != nil {
		return 0, 0, "", err
	}
	of, err := treeFiles(builtDir)
	if err != nil {
		return 0, 0, "", err
	}
	expected := map[string]bool{}
	for n := range bf {
		expected[n] = true
	}
	expected[rel] = true
	if len(of) != len(expected) {
		return len(bf), len(of), "", fmt.Errorf("CG2 inventory count mismatch")
	}
	for n := range expected {
		if _, ok := of[n]; !ok {
			return len(bf), len(of), "", fmt.Errorf("CG2 missing %s", n)
		}
	}
	for n, h := range bf {
		if n == rel {
			continue
		}
		if of[n] != h {
			return len(bf), len(of), "", fmt.Errorf("CG2 untouched payload mismatch: %s", n)
		}
	}
	ah, err := shaFile(asset)
	if err != nil {
		return 0, 0, "", err
	}
	if of[rel] != ah {
		return len(bf), len(of), ah, fmt.Errorf("CG2 startup overlay mismatch")
	}
	return len(bf), len(of), ah, nil
}
func shaBytes(b []byte) string      { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func parseID(s string) (int, error) { return strconv.Atoi(s) }
