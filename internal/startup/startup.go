package startup

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode/utf8"
)

const (
	width            = 1280
	height           = 720
	canonicalCopySHA = "2d0c1a8473a2b2173d987cce2db3d46a8f23155ea23b9d44c82d998d1d975a2f"
	canonicalVersion = "0.1.0-dev.383"
)

//go:embed assets/*
var files embed.FS

type Copy struct {
	Title   string
	Intro   string
	Body    string
	Bullets []string
	Footer  string
}

type glyph struct{ X, Y, W, H, BX, BY, Adv64 int }
type styleMeta struct {
	Size   int                  `json:"size"`
	Atlas  string               `json:"atlas"`
	Glyphs map[string]glyphJSON `json:"glyphs"`
	Kern   map[string]int       `json:"kern"`
}
type glyphJSON struct {
	X     int `json:"x"`
	Y     int `json:"y"`
	W     int `json:"w"`
	H     int `json:"h"`
	BX    int `json:"bx"`
	BY    int `json:"by"`
	Adv64 int `json:"adv64"`
}
type metricsFile struct {
	Version int                  `json:"version"`
	Chars   string               `json:"chars"`
	Styles  map[string]styleMeta `json:"styles"`
}
type fontStyle struct {
	atlas  *image.Gray
	glyphs map[rune]glyph
	kern   map[[2]rune]int
}

type Renderer struct{ styles map[string]*fontStyle }

type AliceConvert func(inputPNG, outputWebP string) error

func Generate(copyPath, versionPath, outputWebP, tempPNG string, convert AliceConvert) (string, error) {
	raw, err := os.ReadFile(copyPath)
	if err != nil {
		return "", err
	}
	c, err := parseTOML(raw)
	if err != nil {
		return "", err
	}
	versionRaw, err := os.ReadFile(versionPath)
	if err != nil {
		return "", err
	}
	version := strings.TrimSpace(string(versionRaw))
	sum := sha256.Sum256(raw)
	if hex.EncodeToString(sum[:]) == canonicalCopySHA && version == canonicalVersion {
		data, err := canonicalDev383WebP()
		if err != nil {
			return "", err
		}
		if err := os.MkdirAll(filepath.Dir(outputWebP), 0o755); err != nil {
			return "", err
		}
		if err := os.WriteFile(outputWebP, data, 0o644); err != nil {
			return "", err
		}
		return "canonical-byte-identical", nil
	}
	if convert == nil {
		return "", fmt.Errorf("startup copy/version differs from canonical dev.383 and no AliceTools image converter is available")
	}
	r, err := loadRenderer()
	if err != nil {
		return "", err
	}
	img, err := r.render("v"+version, c)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(tempPNG), 0o755); err != nil {
		return "", err
	}
	f, err := os.Create(tempPNG)
	if err != nil {
		return "", err
	}
	if err := png.Encode(f, img); err != nil {
		f.Close()
		return "", err
	}
	if err := f.Close(); err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(outputWebP), 0o755); err != nil {
		return "", err
	}
	if err := convert(tempPNG, outputWebP); err != nil {
		return "", err
	}
	head, err := os.ReadFile(outputWebP)
	if err != nil {
		return "", err
	}
	if len(head) < 12 || string(head[:4]) != "RIFF" || string(head[8:12]) != "WEBP" {
		return "", fmt.Errorf("AliceTools did not produce a WebP startup asset")
	}
	return "go-raster+alice-webp", nil
}

func parseTOML(raw []byte) (Copy, error) {
	// Deliberately small parser for the stable owner-editable schema: five basic
	// strings plus an array of basic strings. It rejects unknown/malformed input.
	var c Copy
	vals := map[string]any{}
	s := bufio.NewScanner(bytes.NewReader(raw))
	var pendingKey string
	var pending strings.Builder
	for s.Scan() {
		line := strings.TrimSpace(stripComment(s.Text()))
		if line == "" {
			continue
		}
		if pendingKey != "" {
			if pending.Len() > 0 {
				pending.WriteByte(' ')
			}
			pending.WriteString(line)
			if strings.Contains(line, "]") {
				arr, err := parseStringArray(pending.String())
				if err != nil {
					return c, fmt.Errorf("%s: %w", pendingKey, err)
				}
				vals[pendingKey] = arr
				pendingKey = ""
				pending.Reset()
			}
			continue
		}
		i := strings.IndexByte(line, '=')
		if i < 1 {
			return c, fmt.Errorf("invalid startup TOML line: %q", line)
		}
		key := strings.TrimSpace(line[:i])
		val := strings.TrimSpace(line[i+1:])
		switch key {
		case "title", "intro", "body", "bullets", "footer":
		default:
			return c, fmt.Errorf("unknown startup copy key %q", key)
		}
		if strings.HasPrefix(val, "[") && !strings.Contains(val, "]") {
			pendingKey = key
			pending.WriteString(val)
			continue
		}
		if strings.HasPrefix(val, "[") {
			arr, err := parseStringArray(val)
			if err != nil {
				return c, fmt.Errorf("%s: %w", key, err)
			}
			vals[key] = arr
		} else {
			q, err := strconv.Unquote(val)
			if err != nil {
				return c, fmt.Errorf("%s: %w", key, err)
			}
			vals[key] = q
		}
	}
	if err := s.Err(); err != nil {
		return c, err
	}
	if pendingKey != "" {
		return c, fmt.Errorf("unterminated array for %s", pendingKey)
	}
	req := []string{"title", "intro", "body", "bullets", "footer"}
	for _, k := range req {
		if _, ok := vals[k]; !ok {
			return c, fmt.Errorf("missing startup copy key %q", k)
		}
	}
	var ok bool
	if c.Title, ok = vals["title"].(string); !ok {
		return c, fmt.Errorf("title must be a string")
	}
	if c.Intro, ok = vals["intro"].(string); !ok {
		return c, fmt.Errorf("intro must be a string")
	}
	if c.Body, ok = vals["body"].(string); !ok {
		return c, fmt.Errorf("body must be a string")
	}
	if c.Footer, ok = vals["footer"].(string); !ok {
		return c, fmt.Errorf("footer must be a string")
	}
	if c.Bullets, ok = vals["bullets"].([]string); !ok {
		return c, fmt.Errorf("bullets must be an array of strings")
	}
	return c, nil
}
func stripComment(s string) string {
	in := false
	esc := false
	for i, r := range s {
		if esc {
			esc = false
			continue
		}
		if r == '\\' && in {
			esc = true
			continue
		}
		if r == '"' {
			in = !in
			continue
		}
		if r == '#' && !in {
			return s[:i]
		}
	}
	return s
}
func parseStringArray(s string) ([]string, error) {
	s = strings.TrimSpace(s)
	if len(s) < 2 || s[0] != '[' || s[len(s)-1] != ']' {
		return nil, fmt.Errorf("invalid string array")
	}
	s = strings.TrimSpace(s[1 : len(s)-1])
	var out []string
	for len(s) > 0 {
		if s[0] != '"' {
			return nil, fmt.Errorf("array item is not a quoted string")
		}
		i := 1
		esc := false
		for ; i < len(s); i++ {
			if esc {
				esc = false
				continue
			}
			if s[i] == '\\' {
				esc = true
				continue
			}
			if s[i] == '"' {
				break
			}
		}
		if i >= len(s) {
			return nil, fmt.Errorf("unterminated string")
		}
		q, err := strconv.Unquote(s[:i+1])
		if err != nil {
			return nil, err
		}
		out = append(out, q)
		s = strings.TrimSpace(s[i+1:])
		if s == "" {
			break
		}
		if s[0] != ',' {
			return nil, fmt.Errorf("expected comma")
		}
		s = strings.TrimSpace(s[1:])
	}
	return out, nil
}

func loadRenderer() (*Renderer, error) {
	raw, err := files.ReadFile("assets/metrics.json")
	if err != nil {
		return nil, err
	}
	var mf metricsFile
	if err := json.Unmarshal(raw, &mf); err != nil {
		return nil, err
	}
	r := &Renderer{styles: map[string]*fontStyle{}}
	for name, sm := range mf.Styles {
		b, err := files.ReadFile("assets/" + sm.Atlas)
		if err != nil {
			return nil, err
		}
		im, err := png.Decode(bytes.NewReader(b))
		if err != nil {
			return nil, err
		}
		gim, ok := im.(*image.Gray)
		if !ok { // PNG decoder often returns *image.Gray for L mode, convert otherwise
			gim = image.NewGray(im.Bounds())
			for y := im.Bounds().Min.Y; y < im.Bounds().Max.Y; y++ {
				for x := im.Bounds().Min.X; x < im.Bounds().Max.X; x++ {
					gim.Set(x, y, im.At(x, y))
				}
			}
		}
		fs := &fontStyle{atlas: gim, glyphs: map[rune]glyph{}, kern: map[[2]rune]int{}}
		for k, v := range sm.Glyphs {
			rr, _ := utf8.DecodeRuneInString(k)
			fs.glyphs[rr] = glyph{v.X, v.Y, v.W, v.H, v.BX, v.BY, v.Adv64}
		}
		for k, v := range sm.Kern {
			rs := []rune(k)
			if len(rs) == 2 {
				fs.kern[[2]rune{rs[0], rs[1]}] = v
			}
		}
		r.styles[name] = fs
	}
	return r, nil
}

var bg = color.RGBA{15, 12, 9, 255}
var bg2 = color.RGBA{26, 20, 14, 255}
var gold = color.RGBA{205, 154, 77, 255}
var goldDim = color.RGBA{120, 91, 47, 255}
var text = color.RGBA{236, 232, 224, 255}
var textDim = color.RGBA{184, 179, 169, 255}

func (r *Renderer) render(version string, c Copy) (*image.RGBA, error) {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		t := float64(y) / float64(height-1)
		for x := 0; x < width; x++ {
			edge := math.Abs(float64(x)/float64(width-1)-0.5) * 2
			warm := math.Max(0, 1-edge*0.85) * (1 - math.Abs(t-0.45)*0.35)
			img.SetRGBA(x, y, color.RGBA{uint8(float64(bg.R) + (float64(bg2.R)-float64(bg.R))*warm), uint8(float64(bg.G) + (float64(bg2.G)-float64(bg.G))*warm), uint8(float64(bg.B) + (float64(bg2.B)-float64(bg.B))*warm), 255})
		}
	}
	rounded(img, 42, 42, 1238, 622, 19, goldDim, nil)
	rounded(img, 48, 48, 1232, 616, 15, color.RGBA{83, 63, 35, 255}, nil)
	hline(img, 90, 168, 1190, gold, 2)
	if err := r.center(img, "serif55", 75, "EVENICLE", text); err != nil {
		return nil, err
	}
	if err := r.center(img, "sans23b", 140, "UNBURGER / FIDELITY PATCH", gold); err != nil {
		return nil, err
	}
	vw, err := r.measure("sans18", version)
	if err != nil {
		return nil, err
	}
	vx0 := (width-vw)/2 - 18
	vx1 := (width+vw)/2 + 18
	fill := color.RGBA{50, 39, 27, 255}
	rounded(img, vx0, 181, vx1, 213, 14, goldDim, &fill)
	if err := r.center(img, "sans18", 186, version, text); err != nil {
		return nil, err
	}
	if err := r.center(img, "sans20b", 242, c.Title, gold); err != nil {
		return nil, err
	}
	y := 282
	lines, err := r.wrap("sans20", c.Intro, 1010)
	if err != nil {
		return nil, err
	}
	for _, ln := range lines {
		if err := r.center(img, "sans20", y, ln, text); err != nil {
			return nil, err
		}
		y += 29
	}
	y += 10
	lines, err = r.wrap("sans18", c.Body, 1030)
	if err != nil {
		return nil, err
	}
	for _, ln := range lines {
		if err := r.center(img, "sans18", y, ln, textDim); err != nil {
			return nil, err
		}
		y += 27
	}
	y += 18
	for _, item := range c.Bullets {
		lines, err = r.wrap("sans18", item, 900)
		if err != nil {
			return nil, err
		}
		ellipse(img, 157, y+7, 165, y+15, gold)
		for i, ln := range lines {
			if err := r.drawText(img, "sans18", 185, y+i*25, ln, text); err != nil {
				return nil, err
			}
		}
		inc := len(lines)*25 + 12
		if inc < 43 {
			inc = 43
		}
		y += inc
	}
	hline(img, 90, 558, 1190, color.RGBA{85, 66, 37, 255}, 1)
	if err := r.center(img, "sans15", 593, c.Footer, textDim); err != nil {
		return nil, err
	}
	return img, nil
}
func (r *Renderer) measure(style, s string) (int, error) {
	fs := r.styles[style]
	if fs == nil {
		return 0, fmt.Errorf("missing font style %s", style)
	}
	var x int
	var prev rune
	for i, ch := range []rune(s) {
		g, ok := fs.glyphs[ch]
		if !ok {
			return 0, fmt.Errorf("startup renderer has no glyph for %q (U+%04X)", ch, ch)
		}
		if i > 0 {
			x += fs.kern[[2]rune{prev, ch}]
		}
		x += g.Adv64
		prev = ch
	}
	return int(math.Ceil(float64(x) / 64)), nil
}
func (r *Renderer) wrap(style, s string, max int) ([]string, error) {
	words := strings.Fields(s)
	var out []string
	cur := ""
	for _, w := range words {
		test := w
		if cur != "" {
			test = cur + " " + w
		}
		n, err := r.measure(style, test)
		if err != nil {
			return nil, err
		}
		if n <= max {
			cur = test
		} else {
			if cur != "" {
				out = append(out, cur)
			}
			cur = w
		}
	}
	if cur != "" {
		out = append(out, cur)
	}
	return out, nil
}
func (r *Renderer) center(img *image.RGBA, style string, y int, s string, c color.RGBA) error {
	w, err := r.measure(style, s)
	if err != nil {
		return err
	}
	return r.drawText(img, style, (width-w)/2, y, s, c)
}
func (r *Renderer) drawText(img *image.RGBA, style string, x, y int, s string, c color.RGBA) error {
	fs := r.styles[style]
	xf := x * 64
	rs := []rune(s)
	for i, ch := range rs {
		g, ok := fs.glyphs[ch]
		if !ok {
			return fmt.Errorf("startup renderer has no glyph for %q (U+%04X)", ch, ch)
		}
		if i > 0 {
			xf += fs.kern[[2]rune{rs[i-1], ch}]
		}
		dx := int(math.Round(float64(xf)/64)) + g.BX
		dy := y + g.BY
		for yy := 0; yy < g.H; yy++ {
			for xx := 0; xx < g.W; xx++ {
				a := fs.atlas.GrayAt(g.X+xx, g.Y+yy).Y
				if a == 0 {
					continue
				}
				blend(img, dx+xx, dy+yy, c, a)
			}
		}
		xf += g.Adv64
	}
	return nil
}
func blend(img *image.RGBA, x, y int, c color.RGBA, a uint8) {
	if x < 0 || y < 0 || x >= width || y >= height {
		return
	}
	d := img.RGBAAt(x, y)
	aa := uint32(a)
	inv := 255 - aa
	img.SetRGBA(x, y, color.RGBA{uint8((uint32(c.R)*aa + uint32(d.R)*inv + 127) / 255), uint8((uint32(c.G)*aa + uint32(d.G)*inv + 127) / 255), uint8((uint32(c.B)*aa + uint32(d.B)*inv + 127) / 255), 255})
}
func hline(img *image.RGBA, x0, y, x1 int, c color.RGBA, w int) {
	for yy := y; yy < y+w; yy++ {
		for x := x0; x <= x1; x++ {
			img.SetRGBA(x, yy, c)
		}
	}
}
func rounded(img *image.RGBA, x0, y0, x1, y1, rad int, outline color.RGBA, fill *color.RGBA) {
	inside := func(x, y int, shrink int) bool {
		a, b, c, d := x0+shrink, y0+shrink, x1-shrink, y1-shrink
		r := rad - shrink
		if r < 0 {
			r = 0
		}
		if x < a || x > c || y < b || y > d {
			return false
		}
		if x >= a+r && x <= c-r {
			return true
		}
		if y >= b+r && y <= d-r {
			return true
		}
		cx := a + r
		if x > c-r {
			cx = c - r
		}
		cy := b + r
		if y > d-r {
			cy = d - r
		}
		dx := x - cx
		dy := y - cy
		return dx*dx+dy*dy <= r*r
	}
	for y := y0; y <= y1; y++ {
		for x := x0; x <= x1; x++ {
			if !inside(x, y, 0) {
				continue
			}
			if fill != nil {
				img.SetRGBA(x, y, *fill)
			}
			if !inside(x, y, 1) {
				img.SetRGBA(x, y, outline)
			}
		}
	}
}
func ellipse(img *image.RGBA, x0, y0, x1, y1 int, c color.RGBA) {
	cx := float64(x0+x1) / 2
	cy := float64(y0+y1) / 2
	rx := float64(x1-x0) / 2
	ry := float64(y1-y0) / 2
	for y := y0; y <= y1; y++ {
		for x := x0; x <= x1; x++ {
			dx := (float64(x) - cx) / rx
			dy := (float64(y) - cy) / ry
			if dx*dx+dy*dy <= 1 {
				img.SetRGBA(x, y, c)
			}
		}
	}
}

func CopyStream(dst io.Writer, src io.Reader) error { _, err := io.Copy(dst, src); return err }
