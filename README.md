# bubblequarium

A cozy terminal aquarium written in Go. Fish swim, bubbles rise, you can drop food and watch them chase it.

A small homage to [asciiquarium](https://robobunny.com/projects/asciiquarium/html/), built on [Bubble Tea](https://github.com/charmbracelet/bubbletea) and [Lip Gloss](https://github.com/charmbracelet/lipgloss) — hence the name.

## Install

```sh
go install github.com/cmoles/bubblequarium@latest
```

Or clone and run:

```sh
git clone https://github.com/cmoles/bubblequarium
cd bubblequarium
go run .
```

Requires Go 1.24 or newer and a terminal that supports 24-bit color.

## Controls

| Key   | Action          |
| ----- | --------------- |
| `f`   | Add a fish      |
| `r`   | Remove a fish   |
| space | Drop food       |
| `p`   | Pause / resume  |
| `?`   | Toggle help     |
| `q`   | Quit            |

Fish are attracted to nearby food.

## License

MIT — see [LICENSE](LICENSE).
