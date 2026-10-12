package builder

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type Options struct {
	Root      string
	Alice     string
	GameRoot  string
	SkipAlice bool
}

type Project struct {
	Root     string
	Alice    *Alice
	GameRoot string
}

func FindRoot(start string) (string, error) {
	if start == "" {
		var err error
		start, err = os.Getwd()
		if err != nil {
			return "", err
		}
	}
	p, err := filepath.Abs(start)
	if err != nil {
		return "", err
	}
	for {
		if fileExists(filepath.Join(p, "version.txt")) && fileExists(filepath.Join(p, "patch", "changes.tsv")) {
			return p, nil
		}
		parent := filepath.Dir(p)
		if parent == p {
			break
		}
		p = parent
	}
	return "", fmt.Errorf("could not locate project root from %s", start)
}

func Open(opts Options) (*Project, error) {
	root, err := FindRoot(opts.Root)
	if err != nil {
		return nil, err
	}
	cfg, _ := readConfig(filepath.Join(root, ".local", "config.env"))
	gameRoot := opts.GameRoot
	if gameRoot == "" {
		gameRoot = os.Getenv("EVENICLE_EN_ROOT")
	}
	if gameRoot == "" {
		gameRoot = cfg["EVENICLE_EN_ROOT"]
	}

	if opts.SkipAlice {
		return &Project{Root: root, GameRoot: gameRoot}, nil
	}

	explicitAlice := opts.Alice
	if explicitAlice == "" {
		explicitAlice = os.Getenv("ALICE_BIN")
	}
	if explicitAlice == "" {
		explicitAlice = cfg["ALICE_BIN"]
	}
	alice, err := ResolveAlice(root, explicitAlice)
	if err != nil {
		return nil, err
	}
	return &Project{Root: root, Alice: alice, GameRoot: gameRoot}, nil
}

func readConfig(path string) (map[string]string, error) {
	out := map[string]string{}
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return out, nil
		}
		return nil, err
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		i := strings.IndexByte(line, '=')
		if i <= 0 {
			continue
		}
		k := strings.TrimSpace(line[:i])
		v := strings.TrimSpace(line[i+1:])
		if len(v) >= 2 && ((v[0] == '"' && v[len(v)-1] == '"') || (v[0] == '\'' && v[len(v)-1] == '\'')) {
			v = v[1 : len(v)-1]
			v = strings.ReplaceAll(v, `\"`, `"`)
		}
		out[k] = v
	}
	return out, s.Err()
}

func fileExists(path string) bool { st, err := os.Stat(path); return err == nil && !st.IsDir() }
func dirExists(path string) bool  { st, err := os.Stat(path); return err == nil && st.IsDir() }

func executableName() string {
	if runtime.GOOS == "windows" {
		return "alice.exe"
	}
	return "alice"
}
