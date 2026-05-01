package main

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Fish species with their art and colors
type Species struct {
	rightFrames []string // animation frames facing right
	leftFrames  []string // animation frames facing left
	color       lipgloss.Color
	speed       float64
	name        string
}

var species = []Species{
	{
		rightFrames: []string{"><>", "><>"},
		leftFrames:  []string{"<><", "<><"},
		color:       lipgloss.Color("#FF6B35"),
		speed:       1.0,
		name:        "Goldfish",
	},
	{
		rightFrames: []string{">><>", "> <>"},
		leftFrames:  []string{"<><<", "<> <"},
		color:       lipgloss.Color("#00D4FF"),
		speed:       1.5,
		name:        "Neon Tetra",
	},
	{
		rightFrames: []string{"><))°>", "><))°>"},
		leftFrames:  []string{"<°((><", "<°((><"},
		color:       lipgloss.Color("#FFD700"),
		speed:       0.7,
		name:        "Angelfish",
	},
	{
		rightFrames: []string{"><(((°>", "><(((°>"},
		leftFrames:  []string{"<°)))><", "<°)))><"},
		color:       lipgloss.Color("#FF69B4"),
		speed:       0.5,
		name:        "Betta",
	},
	{
		rightFrames: []string{"/\\oo/\\\n//////", "/\\oo/\\\n||||||"},
		leftFrames:  []string{"/\\oo/\\\n\\\\\\\\\\\\", "/\\oo/\\\n||||||"},
		color:       lipgloss.Color("#FF4500"),
		speed:       0.3,
		name:        "Jellyfish",
	},
}

type Fish struct {
	x, y      float64
	species   Species
	goingLeft bool
	frame     int
	wobbleY   float64
	wobbleSin float64
}

type Bubble struct {
	x, y  float64
	speed float64
	char  string
}

type FoodParticle struct {
	x, y  float64
	speed float64
	life  int
}

type tickMsg time.Time

type model struct {
	width, height int
	fish          []Fish
	bubbles       []Bubble
	food          []FoodParticle
	tick          int
	paused        bool
	showHelp      bool
}

// aquaHeight returns the number of rows available for the aquarium (excluding the status bar).
func (m model) aquaHeight() int {
	h := m.height - 1
	if h < 6 {
		h = 6
	}
	return h
}

func initialModel() model {
	m := model{
		width:  80,
		height: 24,
		fish:   make([]Fish, 0),
	}
	// Start with some fish
	for i := 0; i < 6; i++ {
		m.fish = append(m.fish, newFish(m.width, m.aquaHeight()))
	}
	return m
}

func newFish(width, height int) Fish {
	sp := species[rand.Intn(len(species))]
	goingLeft := rand.Float64() > 0.5
	var x float64
	if goingLeft {
		x = float64(width - 1)
	} else {
		x = 0
	}
	// Keep fish in the water area (not in sand/decorations at bottom)
	waterHeight := height - 3
	if waterHeight < 3 {
		waterHeight = 3
	}
	y := float64(rand.Intn(waterHeight-1) + 1)
	return Fish{
		x:         x,
		y:         y,
		species:   sp,
		goingLeft: goingLeft,
		frame:     0,
		wobbleSin: rand.Float64() * 6.28,
	}
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m model) Init() tea.Cmd {
	return tickCmd()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "f":
			m.fish = append(m.fish, newFish(m.width, m.aquaHeight()))
		case "r":
			if len(m.fish) > 0 {
				m.fish = m.fish[:len(m.fish)-1]
			}
		case " ":
			// Drop food at random x position
			fx := float64(rand.Intn(m.width-4) + 2)
			m.food = append(m.food, FoodParticle{x: fx, y: 1, speed: 0.3, life: 80})
		case "p":
			m.paused = !m.paused
		case "?":
			m.showHelp = !m.showHelp
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tickMsg:
		if !m.paused {
			m.tick++
			m = m.updateFish()
			m = m.updateBubbles()
			m = m.updateFood()
			// Random bubbles from bottom
			if rand.Float64() < 0.15 {
				bx := float64(rand.Intn(m.width))
				m.bubbles = append(m.bubbles, Bubble{
					x: bx, y: float64(m.aquaHeight() - 3),
					speed: 0.3 + rand.Float64()*0.3,
					char:  []string{"°", "o", "O", "·"}[rand.Intn(4)],
				})
			}
		}
		return m, tickCmd()
	}
	return m, nil
}

func (m model) updateFish() model {
	waterHeight := m.aquaHeight() - 3
	if waterHeight < 3 {
		waterHeight = 3
	}
	newFishList := make([]Fish, 0, len(m.fish))
	for _, f := range m.fish {
		// Move horizontally
		speed := f.species.speed * 0.5
		if f.goingLeft {
			f.x -= speed
		} else {
			f.x += speed
		}

		// Wobble vertically
		f.wobbleSin += 0.1
		f.wobbleY = 0.5 * math.Sin(f.wobbleSin)

		// Check if fish is attracted to food
		for i := range m.food {
			dx := m.food[i].x - f.x
			dy := m.food[i].y - f.y
			dist := dx*dx + dy*dy
			if dist < 100 { // attracted within range
				if dx > 0 {
					f.x += 0.3
					f.goingLeft = false
				} else {
					f.x -= 0.3
					f.goingLeft = true
				}
				if dy > 0 {
					f.y += 0.2
				} else if dy > -2 {
					f.y -= 0.2
				}
			}
			// Eat food if close enough
			if dist < 4 {
				m.food[i].life = 0
			}
		}

		// Animate
		if m.tick%5 == 0 {
			if f.goingLeft {
				f.frame = (f.frame + 1) % len(f.species.leftFrames)
			} else {
				f.frame = (f.frame + 1) % len(f.species.rightFrames)
			}
		}

		// Wrap around or reverse direction
		fishLen := 0
		for _, line := range strings.Split(f.species.rightFrames[0], "\n") {
			if l := len([]rune(line)); l > fishLen {
				fishLen = l
			}
		}
		if f.x < float64(-fishLen) {
			f.goingLeft = false
		} else if f.x > float64(m.width+fishLen) {
			f.goingLeft = true
		}

		// Keep in vertical bounds
		if f.y < 1 {
			f.y = 1
		}
		if f.y > float64(waterHeight-1) {
			f.y = float64(waterHeight - 1)
		}

		// Random direction change
		if rand.Float64() < 0.005 {
			f.goingLeft = !f.goingLeft
		}

		newFishList = append(newFishList, f)
	}
	m.fish = newFishList
	return m
}

func (m model) updateBubbles() model {
	newBubbles := make([]Bubble, 0, len(m.bubbles))
	for _, b := range m.bubbles {
		b.y -= b.speed
		b.x += (rand.Float64() - 0.5) * 0.3
		if b.y > 0 {
			newBubbles = append(newBubbles, b)
		}
	}
	m.bubbles = newBubbles
	return m
}

func (m model) updateFood() model {
	newFood := make([]FoodParticle, 0, len(m.food))
	bottomY := float64(m.aquaHeight() - 3)
	for _, f := range m.food {
		f.life--
		if f.y < bottomY {
			f.y += f.speed
		}
		if f.life > 0 {
			newFood = append(newFood, f)
		}
	}
	m.food = newFood
	return m
}

// cellWidth returns the number of terminal cells s occupies, pessimistically
// counting every non-ASCII rune as 2 cells. That matches how most modern
// terminals (Windows Terminal, iTerm2, etc.) render emoji.
func cellWidth(s string) int {
	w := 0
	for _, r := range s {
		if r < 0x80 {
			w++
		} else {
			w += 2
		}
	}
	return w
}

// fitToWidth truncates or pads s so it occupies exactly width terminal cells.
func fitToWidth(s string, width int) string {
	var b strings.Builder
	used := 0
	for _, r := range s {
		rw := 1
		if r >= 0x80 {
			rw = 2
		}
		if used+rw > width {
			break
		}
		b.WriteRune(r)
		used += rw
	}
	if used < width {
		b.WriteString(strings.Repeat(" ", width-used))
	}
	return b.String()
}

// pickStatusLine returns the richest candidate that fits in width cells,
// falling back to the most compact one (which fitToWidth will then truncate).
func pickStatusLine(candidates []string, width int) string {
	for _, c := range candidates {
		if cellWidth(c) <= width {
			return c
		}
	}
	return candidates[len(candidates)-1]
}

func (m model) View() string {
	if m.width < 20 || m.height < 10 {
		return "Terminal too small!"
	}

	// Build the screen buffer (reserve 2 rows for status bars)
	aquaHeight := m.aquaHeight()

	buf := make([][]rune, aquaHeight)
	colors := make([][]lipgloss.Color, aquaHeight)
	for i := range buf {
		buf[i] = make([]rune, m.width)
		colors[i] = make([]lipgloss.Color, m.width)
		for j := range buf[i] {
			buf[i][j] = ' '
			colors[i][j] = ""
		}
	}

	bottomStart := aquaHeight - 2

	// Draw water background with subtle waves
	for y := 0; y < bottomStart; y++ {
		for x := 0; x < m.width; x++ {
			// Subtle water shimmer
			if rand.Float64() < 0.003 {
				buf[y][x] = '~'
				colors[y][x] = lipgloss.Color("#2a5a8c")
			}
		}
	}

	// Draw top border (water surface)
	for x := 0; x < m.width; x++ {
		waveChars := []rune{'~', '~', '≈', '~', '≈'}
		idx := (x + m.tick/2) % len(waveChars)
		buf[0][x] = waveChars[idx]
		colors[0][x] = lipgloss.Color("#4a9eff")
	}

	// Draw sand bottom
	for y := bottomStart; y < aquaHeight; y++ {
		for x := 0; x < m.width; x++ {
			sandChars := []rune{'░', '▒', '░', '·', '░'}
			idx := (x + y) % len(sandChars)
			buf[y][x] = sandChars[idx]
			colors[y][x] = lipgloss.Color("#C2B280")
		}
	}

	// Draw seaweed/plants
	plantPositions := []int{5, 15, 28, 45, 60, 72}
	for _, px := range plantPositions {
		if px >= m.width {
			continue
		}
		plantHeight := 3 + rand.Intn(2)
		if m.tick%20 < 10 { // sway
			for py := 0; py < plantHeight; py++ {
				row := bottomStart - 1 - py
				if row > 0 && row < aquaHeight && px < m.width {
					sway := 0
					if py > 1 && m.tick%20 < 5 {
						sway = 1
					}
					col := px + sway
					if col < m.width {
						plantChars := []rune{'⌇', '⌇', '❀'}
						if py == plantHeight-1 {
							buf[row][col] = plantChars[2]
							colors[row][col] = lipgloss.Color("#FF6B9D")
						} else {
							buf[row][col] = plantChars[0]
							colors[row][col] = lipgloss.Color("#2ECC71")
						}
					}
				}
			}
		} else {
			for py := 0; py < plantHeight; py++ {
				row := bottomStart - 1 - py
				if row > 0 && row < aquaHeight && px < m.width {
					if py == plantHeight-1 {
						buf[row][px] = '❀'
						colors[row][px] = lipgloss.Color("#FF6B9D")
					} else {
						buf[row][px] = '⌇'
						colors[row][px] = lipgloss.Color("#2ECC71")
					}
				}
			}
		}
	}

	// Draw decorations (rocks, shells)
	decoPositions := []struct {
		x    int
		char rune
		col  lipgloss.Color
	}{
		{10, '⌂', lipgloss.Color("#808080")},
		{22, '◇', lipgloss.Color("#C0C0C0")},
		{38, '⍟', lipgloss.Color("#FFD700")},
		{52, '⌂', lipgloss.Color("#696969")},
		{67, '◇', lipgloss.Color("#C0C0C0")},
	}
	for _, d := range decoPositions {
		if d.x < m.width {
			row := bottomStart
			if row < aquaHeight {
				buf[row][d.x] = d.char
				colors[row][d.x] = d.col
			}
		}
	}

	// Draw a little castle
	if m.width > 40 {
		castleX := 35
		castleRow := bottomStart - 1
		castle := []string{

      "  ╔╗    ╔╗  ",
      " ╔╣╠╗╔╗╔╣╠╗ ",
      " ║╚╝╚╝╚╝╚╝║ ",
      "╔╝        ╚╗",
      "║ ╗╔╗╔╗╔╗╔ ║",
      "║ ╚╝╚╝╚╝╚╝ ║",
      "║  ▐████▌  ║",
      "║  ▐█  █▌  ║",
      "╚══════════╝",
			// " ][ ][ ][ ",
			// "=][=][=][=",
			// "[        ]",
			// "[  [  ]  ]",
			// "[  [  ]  ]",
			// "[========]",
		}
		for dy, line := range castle {
			for dx, ch := range []rune(line) {
				ry := castleRow - len(castle) + 1 + dy
				rx := castleX + dx
				if ry > 0 && ry < aquaHeight && rx >= 0 && rx < m.width {
					buf[ry][rx] = ch
					colors[ry][rx] = lipgloss.Color("#B8860B")
				}
			}
		}
	}

	// Draw bubbles
	for _, b := range m.bubbles {
		bx, by := int(b.x), int(b.y)
		if bx >= 0 && bx < m.width && by >= 0 && by < aquaHeight {
			for _, r := range b.char {
				buf[by][bx] = r
				colors[by][bx] = lipgloss.Color("#87CEEB")
				break
			}
		}
	}

	// Draw food
	for _, f := range m.food {
		fx, fy := int(f.x), int(f.y)
		if fx >= 0 && fx < m.width && fy >= 0 && fy < aquaHeight {
			buf[fy][fx] = '•'
			colors[fy][fx] = lipgloss.Color("#FFA500")
		}
	}

	// Draw fish
	for _, f := range m.fish {
		var art string
		if f.goingLeft {
			art = f.species.leftFrames[f.frame%len(f.species.leftFrames)]
		} else {
			art = f.species.rightFrames[f.frame%len(f.species.rightFrames)]
		}
		fy := int(f.y + f.wobbleY)
		fx := int(f.x)
		if fy < 1 {
			fy = 1
		}
		if fy >= aquaHeight-1 {
			continue
		}
		lines := strings.Split(art, "\n")
		for dy, line := range lines {
			ry := fy + dy
			for i, ch := range []rune(line) {
				cx := fx + i
				if cx >= 0 && cx < m.width && ry > 0 && ry < aquaHeight-2 {
					buf[ry][cx] = ch
					colors[ry][cx] = f.species.color
				}
			}
		}
	}

	// Render buffer to string
	var sb strings.Builder
	for y := 0; y < aquaHeight; y++ {
		for x := 0; x < m.width; x++ {
			ch := string(buf[y][x])
			c := colors[y][x]
			if c != "" {
				style := lipgloss.NewStyle().Foreground(c)
				sb.WriteString(style.Render(ch))
			} else {
				// Dark blue water background
				style := lipgloss.NewStyle().Foreground(lipgloss.Color("#1a3a5c"))
				sb.WriteString(style.Render(ch))
			}
		}
		sb.WriteString("\n")
	}

	// Status bar
	statusStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#1a1a2e")).
		Foreground(lipgloss.Color("#e0e0e0")).
		Bold(true)

	now := time.Now()
	n := len(m.fish)
	fullClock := now.Format("3:04:05 PM")
	shortClock := now.Format("3:04 PM")
	miniClock := now.Format("15:04")
	fullDate := now.Format("Mon Jan 2, 2006")
	medDate := now.Format("Jan 2, 2006")
	shortDate := now.Format("Jan 2")

	// Richest to most compact. The first that fits in m.width wins.
	statusLine := pickStatusLine([]string{
		fmt.Sprintf(" 🐟%d  🕐 %s  📅 %s  ?:❓", n, fullClock, fullDate),
		fmt.Sprintf(" %d fish  %s  %s  ?", n, fullClock, fullDate),
		fmt.Sprintf(" %d fish  %s  %s  ?", n, shortClock, medDate),
		fmt.Sprintf(" %d fish  %s  %s  ?", n, shortClock, shortDate),
		fmt.Sprintf(" %df  %s  %s  ?", n, shortClock, shortDate),
		fmt.Sprintf(" %df %s ?", n, miniClock),
	}, m.width)
	statusLine = fitToWidth(statusLine, m.width)

	sb.WriteString(statusStyle.Render(statusLine))

	// Help overlay
	if m.showHelp {
		help := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#4a9eff")).
			Padding(1, 2).
			Background(lipgloss.Color("#1a1a2e")).
			Foreground(lipgloss.Color("#e0e0e0")).
			Render(
				"🐠 bubblequarium 🐠\n\n" +
					"F     - Add a fish\n" +
					"R     - Remove a fish\n" +
					"Space - Drop food\n" +
					"P     - Pause/Resume\n" +
					"Q     - Quit\n\n" +
					"Fish are attracted to food!\n" +
					"Press ? to close this help.",
			)
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, help,
			lipgloss.WithWhitespaceBackground(lipgloss.Color("#0a1628")))
	}

	return sb.String()
}

func main() {
	p := tea.NewProgram(
		initialModel(),
		tea.WithAltScreen(),
	)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}
