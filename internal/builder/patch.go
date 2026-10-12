package builder

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Change struct {
	Kind string
	ID   int
	Text string
	Note string
}

func (p *Project) ReadChanges() (map[string]Change, []Change, error) {
	path := filepath.Join(p.Root, "patch", "changes.tsv")
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()
	r := csv.NewReader(f)
	r.Comma = '\t'
	r.FieldsPerRecord = -1
	h, err := r.Read()
	if err != nil {
		return nil, nil, err
	}
	idx := map[string]int{}
	for i, v := range h {
		idx[v] = i
	}
	for _, k := range []string{"kind", "id", "text"} {
		if _, ok := idx[k]; !ok {
			return nil, nil, fmt.Errorf("changes.tsv missing column %s", k)
		}
	}
	seen := map[string]Change{}
	var ordered []Change
	for {
		row, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, nil, err
		}
		get := func(k string) string {
			i := idx[k]
			if i >= len(row) {
				return ""
			}
			return row[i]
		}
		kind := strings.TrimSpace(get("kind"))
		ids := strings.TrimSpace(get("id"))
		if kind == "" && ids == "" {
			continue
		}
		if kind != "m" && kind != "s" {
			return nil, nil, fmt.Errorf("bad change kind %q", kind)
		}
		id, err := strconv.Atoi(ids)
		if err != nil {
			return nil, nil, fmt.Errorf("bad %s id %q", kind, ids)
		}
		key := fmt.Sprintf("%s[%d]", kind, id)
		if _, ok := seen[key]; ok {
			return nil, nil, fmt.Errorf("duplicate change %s", key)
		}
		note := ""
		if j, ok := idx["note"]; ok && j < len(row) {
			note = row[j]
		}
		c := Change{kind, id, get("text"), note}
		seen[key] = c
		ordered = append(ordered, c)
	}
	if len(ordered) == 0 {
		return nil, nil, fmt.Errorf("refusing build: no managed patch records")
	}
	return seen, ordered, nil
}
func escapeEdit(s string) string {
	return strings.NewReplacer("\\", "\\\\", "\"", "\\\"", "\r", "\\r", "\n", "\\n", "\t", "\\t").Replace(s)
}
func WriteEditFile(path string, changes []Change) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	var b strings.Builder
	for _, c := range changes {
		fmt.Fprintf(&b, "%s[%d] = \"%s\"\n", c.Kind, c.ID, escapeEdit(c.Text))
	}
	return os.WriteFile(path, []byte(b.String()), 0o644)
}
