package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/lucasb-eyer/go-colorful"
)

type theme struct {
	name   string
	hot    rgb
	cool   rgb
	liquid rgb
	rim    rgb
}

var themes = []theme{
	{"lava", rgb{255, 138, 40}, rgb{205, 52, 36}, rgb{36, 12, 56}, rgb{255, 214, 120}},
	{"ocean", rgb{100, 210, 255}, rgb{30, 96, 210}, rgb{4, 16, 38}, rgb{210, 245, 255}},
	{"slime", rgb{180, 255, 90}, rgb{40, 165, 60}, rgb{10, 26, 18}, rgb{235, 255, 190}},
	{"grape", rgb{225, 125, 255}, rgb{120, 40, 200}, rgb{18, 8, 30}, rgb{245, 210, 255}},
	{"sunset", rgb{255, 200, 90}, rgb{225, 60, 120}, rgb{38, 10, 44}, rgb{255, 240, 200}},
	{"mono", rgb{240, 240, 240}, rgb{140, 140, 150}, rgb{14, 14, 16}, rgb{255, 255, 255}},
}

type settings struct {
	theme      theme
	themeIndex int

	blobs int
	size  int
	speed int
	rim   int

	shade bool
	soft  bool

	ascii   bool
	charset []rune

	aspect float64
	fps    int
}

const (
	minBlobs, maxBlobs, defaultBlobs = 1, 20, 6
	minLevel, maxLevel, defaultLevel = 1, 10, 5
	maxRim                           = 5
	minAspect, maxAspect             = 0.5, 3.0
	minFPS, maxFPS                   = 5, 120

	defaultChars = "@%#*+=-:."
)

func (st *settings) sizeScale() float64 {
	return 0.5 + 0.1*float64(st.size)
}

func (st *settings) speedScale() float64 {
	return math.Pow(1.25, float64(st.speed-defaultLevel))
}

func (st *settings) nextTheme() {
	st.themeIndex = (st.themeIndex + 1) % len(themes)
	st.theme = themes[st.themeIndex]
}

var colorNames = []string{
	"red", "orange", "yellow", "green", "cyan", "blue", "purple", "magenta", "pink", "white", "black",
}

var namedColors = map[string]rgb{
	"red":     {230, 50, 40},
	"orange":  {255, 120, 30},
	"yellow":  {250, 210, 50},
	"green":   {60, 200, 80},
	"cyan":    {40, 210, 220},
	"blue":    {50, 110, 240},
	"purple":  {150, 70, 230},
	"magenta": {230, 60, 200},
	"pink":    {255, 120, 170},
	"white":   {240, 240, 240},
	"black":   {20, 20, 24},
}

func parseColor(s string) (rgb, error) {
	s = strings.ToLower(strings.TrimSpace(s))
	if c, ok := namedColors[s]; ok {
		return c, nil
	}
	hex := strings.TrimPrefix(s, "#")
	if len(hex) == 3 {
		hex = string([]byte{hex[0], hex[0], hex[1], hex[1], hex[2], hex[2]})
	}
	if len(hex) == 6 {

		if v, err := strconv.ParseUint(hex, 16, 32); err == nil {
			return rgb{float64(v >> 16 & 0xff), float64(v >> 8 & 0xff), float64(v & 0xff)}, nil
		}
	}
	return rgb{}, fmt.Errorf("unknown color %q: use a name (%s) or a hex code like ff6020",
		s, strings.Join(colorNames, ", "))
}

func themeFromColor(c rgb, base theme) theme {
	l, chroma, hue := toColorful(c).OkLch()
	shade := func(dl, dc float64) rgb {
		return fromColorful(colorful.OkLch(clamp(l+dl, 0, 1), chroma*dc, hue).Clamped())
	}
	return theme{
		name:   "custom",
		hot:    shade(+0.08, 1),
		cool:   shade(-0.12, 1),
		liquid: base.liquid,
		rim:    shade(+0.25, 0.5),
	}
}

func toColorful(c rgb) colorful.Color {
	return colorful.Color{R: c.r / 255, G: c.g / 255, B: c.b / 255}
}

func fromColorful(c colorful.Color) rgb {
	return rgb{c.R * 255, c.G * 255, c.B * 255}
}

var keyHelp = []string{
	"space    pause / resume",
	"c        next color theme",
	"i / d    bigger / smaller blobs",
	"m / l    more / fewer blobs",
	"+ / -    faster / slower",
	"I / D    thicker / thinner rim",
	"s        depth shading on / off",
	"e        smooth / hard edges",
	"a        pixels / characters",
	"[ / ]    adjust aspect ratio",
	"?        show this help",
	"q, Esc   quit",
}

func parseSettings(args []string) (*settings, error) {
	fs := flag.NewFlagSet("lavalamp", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.Usage = func() {}

	themeName := fs.String("theme", themes[0].name, "color theme: "+themeList())
	waxColor := fs.String("c", "", "wax color, a name or hex code like ff6020 (replaces the theme's wax colors)")
	rimColor := fs.String("k", "", "rim color; turns the rim on at thickness 2 unless -R is given")
	bgColor := fs.String("bg", "", "liquid (background) color")
	blobs := fs.Int("b", defaultBlobs, fmt.Sprintf("number of blobs, %d to %d", minBlobs, maxBlobs))
	size := fs.Int("r", defaultLevel, "blob size, 1 to 10")
	speed := fs.Int("s", defaultLevel, "speed, 1 to 10")
	rim := fs.Int("R", 0, "rim thickness, 0 (off) to 5")
	shade := fs.Bool("shade", false, "depth shading, so blobs look round")
	hard := fs.Bool("hard", false, "hard edges instead of smooth ones")
	chars := fs.String("F", "", "draw with these characters instead of pixels, core first, like \""+defaultChars+"\"")
	aspect := fs.Float64("aspect", 1.0, "pixel height ÷ width: raise it if blobs look tall, lower it if they look wide")
	fps := fs.Int("fps", 30, fmt.Sprintf("frames per second, %d to %d", minFPS, maxFPS))
	showVersion := fs.Bool("version", false, "print the version and exit")

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			printUsage(fs)
		}
		return nil, err
	}
	if *showVersion {
		fmt.Println("lavalamp", currentVersion())
		return nil, errVersionShown
	}
	if fs.NArg() > 0 {
		return nil, fmt.Errorf("unexpected argument %q", fs.Arg(0))
	}

	checks := []struct {
		name   string
		value  int
		lo, hi int
	}{
		{"b", *blobs, minBlobs, maxBlobs},
		{"r", *size, minLevel, maxLevel},
		{"s", *speed, minLevel, maxLevel},
		{"R", *rim, 0, maxRim},
		{"fps", *fps, minFPS, maxFPS},
	}
	for _, c := range checks {
		if c.value < c.lo || c.value > c.hi {
			return nil, fmt.Errorf("-%s must be from %d to %d, got %d", c.name, c.lo, c.hi, c.value)
		}
	}
	if *aspect < minAspect || *aspect > maxAspect {
		return nil, fmt.Errorf("-aspect must be from %.1f to %.1f, got %g", minAspect, maxAspect, *aspect)
	}

	st := &settings{
		blobs:   *blobs,
		size:    *size,
		speed:   *speed,
		rim:     *rim,
		shade:   *shade,
		soft:    !*hard,
		charset: []rune(defaultChars),
		aspect:  *aspect,
		fps:     *fps,
	}

	st.themeIndex = -1
	for i, t := range themes {
		if t.name == strings.ToLower(*themeName) {
			st.themeIndex = i
		}
	}
	if st.themeIndex < 0 {
		return nil, fmt.Errorf("unknown theme %q: choose from %s", *themeName, themeList())
	}
	st.theme = themes[st.themeIndex]

	if *waxColor != "" {
		c, err := parseColor(*waxColor)
		if err != nil {
			return nil, fmt.Errorf("-c: %w", err)
		}
		st.theme = themeFromColor(c, st.theme)
		st.themeIndex = -1
	}
	if *bgColor != "" {
		c, err := parseColor(*bgColor)
		if err != nil {
			return nil, fmt.Errorf("-bg: %w", err)
		}
		st.theme.liquid = c
		st.themeIndex = -1
	}
	if *rimColor != "" {
		c, err := parseColor(*rimColor)
		if err != nil {
			return nil, fmt.Errorf("-k: %w", err)
		}
		st.theme.rim = c
		st.themeIndex = -1
		if !wasSet(fs, "R") {
			st.rim = 2
		}
	}

	if *chars != "" {
		st.ascii = true
		st.charset = []rune(*chars)
	}
	return st, nil
}

func wasSet(fs *flag.FlagSet, name string) bool {
	found := false
	fs.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

func themeList() string {
	names := make([]string, len(themes))
	for i, t := range themes {
		names[i] = t.name
	}
	return strings.Join(names, ", ")
}

func printUsage(fs *flag.FlagSet) {
	out := os.Stdout
	fmt.Fprintln(out, "lavalamp: a lava lamp for your terminal")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Usage: lavalamp [options]")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Options:")
	fs.SetOutput(out)
	fs.PrintDefaults()
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Colors are names ("+strings.Join(colorNames, ", ")+")")
	fmt.Fprintln(out, "or hex codes like ff6020.")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Keys:")
	for _, line := range keyHelp {
		fmt.Fprintln(out, "  "+line)
	}
}
