# lavalamp

A lava lamp for your terminal, written in Go. Runs natively on Windows, macOS and Linux.

Wax blobs heat up at the bottom, rise, cool at the top and sink again. They merge and split as they pass each other, drawn with smooth edges in true color. Inspired by [lavat](https://github.com/AngelJumbo/lavat).

## Install

With Go 1.24 or newer:

```sh
go install github.com/eligibilityy/lavalamp@latest
```

This puts `lavalamp` in Go's bin folder (`%USERPROFILE%\go\bin` on Windows, `~/go/bin` elsewhere). Make sure that folder is on your PATH.

## Usage

```sh
lavalamp                          # the default lava lamp
lavalamp -theme ocean -shade      # blue, with rounded shading
lavalamp -c pink -R 2             # your own wax color, with a rim
lavalamp -F "@%#*+=-:."           # draw with characters
lavalamp -h                       # every option and key
```

If the blobs look stretched, adjust the aspect ratio with `[` and `]` while it runs, then pass the value it shows with `-aspect`.

### Options

| Flag | Meaning | Default |
|---|---|---|
| `-theme` | `lava`, `ocean`, `slime`, `grape`, `sunset`, `mono` | `lava` |
| `-c` | wax color: a name or hex code like `ff6020` | from theme |
| `-k` | rim color (turns the rim on) | from theme |
| `-bg` | liquid color | from theme |
| `-b` | number of blobs, 1 to 20 | 6 |
| `-r` | blob size, 1 to 10 | 5 |
| `-s` | speed, 1 to 10 | 5 |
| `-R` | rim thickness, 0 (off) to 5 | 0 |
| `-shade` | depth shading | off |
| `-hard` | hard edges | off |
| `-F` | draw with these characters, core first | off |
| `-aspect` | pixel height ÷ width | 1.0 |
| `-fps` | frames per second, 5 to 120 | 30 |
| `-version` | print the version | |

Color names: red, orange, yellow, green, cyan, blue, purple, magenta, pink, white, black.

### Keys

| Key | Action |
|---|---|
| `space` | pause / resume |
| `c` | next color theme |
| `i` / `d` | bigger / smaller blobs |
| `m` / `l` | more / fewer blobs |
| `+` / `-` | faster / slower |
| `I` / `D` | thicker / thinner rim |
| `s` | depth shading on / off |
| `e` | smooth / hard edges |
| `a` | pixels / characters |
| `[` / `]` | adjust aspect ratio |
| `?` | show the keys |
| `q`, `Esc` | quit |

## Build from source

```sh
git clone https://github.com/eligibilityy/lavalamp
cd lavalamp
go build
```

Best in a terminal with true color, such as Windows Terminal, iTerm2, Kitty, Alacritty or WezTerm.

## License

MIT

## Credits

- [lavat](https://github.com/AngelJumbo/lavat) by AngelJumbo, the terminal lava lamp that inspired this project. Many of lavalamp's flags and keys follow lavat's.
- [tcell](https://github.com/gdamore/tcell) by Garrett D'Amore, which handles the terminal on every platform, including Windows.
- [go-colorful](https://github.com/lucasb-eyer/go-colorful) by Lucas Beyer, used for color math in OkLch.
- The metaball technique comes from Jim Blinn's "blobby objects" (1982).
