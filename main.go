package main

import (
	"errors"
	"flag"
	"fmt"
	"math"
	"os"
	"time"

	"github.com/gdamore/tcell/v2"
)

func main() {
	st, err := parseSettings(os.Args[1:])
	if errors.Is(err, flag.ErrHelp) || errors.Is(err, errVersionShown) {
		return
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "lavalamp:", err)
		fmt.Fprintln(os.Stderr, "Run `lavalamp -h` to see the options.")
		os.Exit(2)
	}

	if err := run(st); err != nil {
		fmt.Fprintln(os.Stderr, "lavalamp:", err)
		os.Exit(1)
	}
}

func run(st *settings) error {
	scr, err := tcell.NewScreen()
	if err != nil {
		return err
	}

	if err := scr.Init(); err != nil {
		return err
	}

	defer func() {
		crash := recover()
		scr.Fini()
		if crash != nil {
			panic(crash)
		}
	}()

	scr.HideCursor()

	events := make(chan tcell.Event, 16)
	stop := make(chan struct{})
	defer close(stop)
	go scr.ChannelEvents(events, stop)

	cols, rows := scr.Size()
	w, h, _ := worldSize(cols, rows, st.aspect)
	s := newSim(w, h, st.blobs, st.sizeScale())

	refit := func() {
		cols, rows := scr.Size()
		w, h, _ := worldSize(cols, rows, st.aspect)
		s.resize(w, h)
	}

	ticker := time.NewTicker(time.Second / time.Duration(st.fps))
	defer ticker.Stop()

	var overlay []string
	var overlayUntil time.Time
	show := func(seconds float64, lines ...string) {
		overlay = lines
		overlayUntil = time.Now().Add(time.Duration(seconds * float64(time.Second)))
	}

	paused := false
	last := time.Now()

	for {
		select {
		case ev := <-events:
			switch ev := ev.(type) {
			case *tcell.EventKey:
				if isQuit(ev) {
					return nil
				}
				if ev.Key() != tcell.KeyRune {
					break
				}
				switch ev.Rune() {
				case ' ':
					paused = !paused
					show(2, onOff("paused", paused))
				case '?', 'h':
					show(8, keyHelp...)
				case '[', ']':
					step := 0.05
					if ev.Rune() == '[' {
						step = -step
					}
					st.aspect = clamp(st.aspect+step, minAspect, maxAspect)
					refit()
					show(2, fmt.Sprintf("aspect %.2f  (run with -aspect %.2f to keep it)", st.aspect, st.aspect))
				default:
					if msg := handleSettingKey(ev.Rune(), st, s); msg != "" {
						show(2, msg)
					}
				}
			case *tcell.EventResize:
				scr.Sync()
				refit()
			}

		case now := <-ticker.C:

			dt := math.Min(now.Sub(last).Seconds(), 0.1)
			last = now
			if !paused {

				s.advance(dt * st.speedScale())
			}
			if now.After(overlayUntil) {
				overlay = nil
			}
			draw(scr, s, st, overlay)
		}
	}
}

func handleSettingKey(key rune, st *settings, s *sim) string {
	switch key {
	case 'c':
		st.nextTheme()
		return "theme " + st.theme.name

	case 'i', 'd':
		st.size = step(st.size, key == 'i', minLevel, maxLevel)
		s.size = st.sizeScale()
		return fmt.Sprintf("size %d/%d", st.size, maxLevel)

	case 'm', 'M', 'l', 'L':
		st.blobs = step(st.blobs, key == 'm' || key == 'M', minBlobs, maxBlobs)
		s.setCount(st.blobs)
		return fmt.Sprintf("blobs %d", st.blobs)

	case '+', '=', '-', '_':
		st.speed = step(st.speed, key == '+' || key == '=', minLevel, maxLevel)
		return fmt.Sprintf("speed %d/%d", st.speed, maxLevel)

	case 'I', 'D':
		st.rim = step(st.rim, key == 'I', 0, maxRim)
		if st.rim == 0 {
			return "rim off"
		}
		return fmt.Sprintf("rim %d/%d", st.rim, maxRim)

	case 's':
		st.shade = !st.shade
		return onOff("shading", st.shade)

	case 'e':
		st.soft = !st.soft
		if st.soft {
			return "smooth edges"
		}
		return "hard edges"

	case 'a':
		st.ascii = !st.ascii
		if st.ascii {
			return "characters: " + string(st.charset)
		}
		return "pixels"
	}
	return ""
}

func step(v int, up bool, lo, hi int) int {
	if up {
		return min(v+1, hi)
	}
	return max(v-1, lo)
}

func onOff(name string, on bool) string {
	if on {
		return name + " on"
	}
	return name + " off"
}

func isQuit(ev *tcell.EventKey) bool {
	switch ev.Key() {
	case tcell.KeyEscape, tcell.KeyCtrlC:
		return true
	case tcell.KeyRune:
		return ev.Rune() == 'q'
	}
	return false
}
