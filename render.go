package main

import (
	"math"

	"github.com/gdamore/tcell/v2"
)

type rgb struct{ r, g, b float64 }

func (c rgb) tcell() tcell.Color {
	return tcell.NewRGBColor(int32(c.r+0.5), int32(c.g+0.5), int32(c.b+0.5))
}

func (c rgb) scale(k float64) rgb {
	return rgb{c.r * k, c.g * k, c.b * k}
}

func mix(a, b rgb, t float64) rgb {
	return rgb{
		a.r + (b.r-a.r)*t,
		a.g + (b.g-a.g)*t,
		a.b + (b.b-a.b)*t,
	}
}

var white = rgb{255, 255, 255}

const (
	threshold = 1.0

	softness = 0.2

	rimStep = 0.35
)

func worldSize(cols, rows int, aspect float64) (w, h, unit float64) {

	pw := float64(max(cols, 1))
	ph := float64(max(rows, 1)) * 2 * aspect
	unit = math.Min(pw, ph)
	return pw / unit, ph / unit, unit
}

func field(x, y float64, blobs []blob) (sum, temp float64) {
	for _, b := range blobs {
		dx, dy := (x-b.x)/b.sx, (y-b.y)/b.sy

		d2 := math.Max(dx*dx+dy*dy, 1e-12)
		w := b.r * b.r / d2
		sum += w
		temp += w * b.temp
	}
	if sum > 0 {
		temp /= sum
	}
	return sum, temp
}

func smoothstep(lo, hi, v float64) float64 {
	if v <= lo {
		return 0
	}
	if v >= hi {
		return 1
	}
	t := (v - lo) / (hi - lo)
	return t * t * (3 - 2*t)
}

func edge(f, at, width float64, soft bool) float64 {
	if !soft {
		if f >= at {
			return 1
		}
		return 0
	}
	return smoothstep(at-width, at+width, f)
}

func waxAt(f, temp float64, st *settings) (rgb, float64) {
	cover := edge(f, threshold, softness, st.soft)
	if cover == 0 {
		return rgb{}, 0
	}
	th := &st.theme

	wax := mix(th.cool, th.hot, temp)

	if st.shade {
		depth := math.Sqrt(math.Max(0, 1-threshold/f))
		highlight := mix(mix(wax, th.hot, 0.5), white, 0.12)
		wax = mix(wax.scale(0.5), highlight, depth)
	}

	if st.rim > 0 {
		inner := threshold * (1 + rimStep*float64(st.rim))
		width := 0.5 * softness * math.Pow(inner, 1.5)
		wax = mix(th.rim, wax, edge(f, inner, width, st.soft))
	}
	return wax, cover
}

func draw(scr tcell.Screen, s *sim, st *settings, overlay []string) {
	if st.ascii {
		drawChars(scr, s, st)
	} else {
		drawPixels(scr, s, st)
	}
	drawOverlay(scr, overlay)

	scr.Show()
}

func drawPixels(scr tcell.Screen, s *sim, st *settings) {
	cols, rows := scr.Size()
	_, _, unit := worldSize(cols, rows, st.aspect)

	pixel := func(x, y float64) tcell.Color {
		f, temp := field(x, y, s.blobs)
		wax, cover := waxAt(f, temp, st)
		return mix(st.theme.liquid, wax, cover).tcell()
	}

	for row := 0; row < rows; row++ {
		yTop := (float64(row)*2 + 0.5) * st.aspect / unit
		yBottom := (float64(row)*2 + 1.5) * st.aspect / unit
		for col := 0; col < cols; col++ {
			x := (float64(col) + 0.5) / unit
			style := tcell.StyleDefault.Foreground(pixel(x, yTop)).Background(pixel(x, yBottom))
			scr.Put(col, row, "▀", style)
		}
	}
}

func drawChars(scr tcell.Screen, s *sim, st *settings) {
	cols, rows := scr.Size()
	_, _, unit := worldSize(cols, rows, st.aspect)
	chars := st.charset
	n := len(chars)
	liquid := st.theme.liquid.tcell()
	empty := tcell.StyleDefault.Background(liquid)

	for row := 0; row < rows; row++ {

		y := (float64(row) + 0.5) * 2 * st.aspect / unit
		for col := 0; col < cols; col++ {
			x := (float64(col) + 0.5) / unit
			f, temp := field(x, y, s.blobs)
			wax, cover := waxAt(f, temp, st)
			if cover == 0 {
				scr.Put(col, row, " ", empty)
				continue
			}
			ring := math.Sqrt(threshold / f)
			i := min(int(ring*float64(n)), n-1)

			fg := mix(st.theme.liquid, wax, cover).tcell()
			scr.Put(col, row, string(chars[i]), empty.Foreground(fg))
		}
	}
}

func drawOverlay(scr tcell.Screen, lines []string) {
	if len(lines) == 0 {
		return
	}
	width := 0
	for _, line := range lines {
		width = max(width, len([]rune(line)))
	}
	style := tcell.StyleDefault.
		Foreground(tcell.NewRGBColor(255, 255, 255)).
		Background(tcell.NewRGBColor(0, 0, 0))
	for row, line := range lines {
		runes := []rune(line)

		for col := 0; col < width+2; col++ {
			ch := " "
			if col >= 1 && col-1 < len(runes) {
				ch = string(runes[col-1])
			}
			scr.Put(col, row, ch, style)
		}
	}
}
