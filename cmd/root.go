package cmd

import (
	"fmt"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"strconv"
	"strings"
	"time"
)

const (
	padding  = 2
	maxWidth = 80
)

var BigDigits = map[rune][]string{
	'0': {
		" 00000 ",
		"0     0",
		"0     0",
		"0     0",
		"0     0",
		"0     0",
		" 00000 ",
	},
	'1': {
		"   1   ",
		"  11   ",
		" 111   ",
		"   1   ",
		"   1   ",
		"   1   ",
		" 11111 ",
	},
	'2': {
		" 22222 ",
		"2     2",
		"     2 ",
		"    2  ",
		"   2   ",
		"  2    ",
		"2222222",
	},
	'3': {
		" 33333 ",
		"3     3",
		"      3",
		"  3333 ",
		"      3",
		"3     3",
		" 33333 ",
	},
	'4': {
		"4     4",
		"4     4",
		"4     4",
		"4444444",
		"      4",
		"      4",
		"      4",
	},
	'5': {
		"5555555",
		"5      ",
		"5      ",
		"555555 ",
		"      5",
		"      5",
		"555555 ",
	},
	'6': {
		" 66666 ",
		"6      ",
		"6      ",
		"666666 ",
		"6     6",
		"6     6",
		" 66666 ",
	},
	'7': {
		"7777777",
		"      7",
		"     7 ",
		"    7  ",
		"   7   ",
		"  7    ",
		" 7     ",
	},
	'8': {
		" 88888 ",
		"8     8",
		"8     8",
		" 88888 ",
		"8     8",
		"8     8",
		" 88888 ",
	},
	'9': {
		" 99999 ",
		"9     9",
		"9     9",
		" 999999",
		"      9",
		"      9",
		" 99999 ",
	},
}

var Colon = []string{
	"     ",
	"  *  ",
	"  *  ",
	"     ",
	"  *  ",
	"  *  ",
	"     ",
}

var (
	name                string
	altscreen           bool
	winHeight, winWidth int
	version             = "dev"
	quitKeys            = key.NewBinding(key.WithKeys("esc", "q"))
	intKeys             = key.NewBinding(key.WithKeys("ctrl+c"))
	pause               = key.NewBinding(key.WithKeys(" "))
	boldStyle           = lipgloss.NewStyle().Bold(true)
	italicStyle         = lipgloss.NewStyle().Italic(true)

	Root = &cobra.Command{
		Use:          "timer",
		Short:        "timer is like sleep, but with progress report",
		Version:      version,
		SilenceUsage: true,
		Args:         cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			addSuffixIfArgIsNumber(&(args[0]), "s")
			duration, err := time.ParseDuration(args[0])
			if err != nil {
				return err
			}
			var opts []tea.ProgramOption
			if altscreen {
				opts = append(opts, tea.WithAltScreen())
			}
			interval := time.Second
			if duration < time.Minute {
				interval = 100 * time.Millisecond
			}
			mode := args[1]
			m, err := tea.NewProgram(model{
				duration:  duration,
				timer:     timer.NewWithInterval(duration, interval),
				progress:  progress.New(progress.WithDefaultGradient()),
				name:      name,
				altscreen: altscreen,
				start:     time.Now(),
				mode:      mode,
			}, opts...).Run()
			if err != nil {
				return err
			}
			if m.(model).interrupting {
				return fmt.Errorf("interrupted")
			}
			if name != "" {
				cmd.Printf("%s ", name)
			}
			cmd.Printf("finished!\n")
			return nil
		},
	}
)

type model struct {
	name         string
	altscreen    bool
	duration     time.Duration
	passed       time.Duration
	start        time.Time
	timer        timer.Model
	progress     progress.Model
	quitting     bool
	interrupting bool
	paused       bool
	termWidth    int
	termHeight   int
	mode         string
}

func init() {
	Root.Flags().StringVarP(&name, "name", "n", "", "timer name")
	Root.Flags().BoolVarP(&altscreen, "fullscreen", "f", false, "fullscreen")
}

func (m model) Init() tea.Cmd {
	return m.timer.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case timer.TickMsg:
		var cmds []tea.Cmd
		var cmd tea.Cmd
		if !m.paused {
			m.passed += m.timer.Interval
			pct := m.passed.Milliseconds() * 100 / m.duration.Milliseconds()
			cmds = append(cmds, m.progress.SetPercent(float64(pct)/100))

			m.timer, cmd = m.timer.Update(msg)
			cmds = append(cmds, cmd)
		}
		return m, tea.Batch(cmds...)

	case tea.WindowSizeMsg:
		m.progress.Width = msg.Width - padding*2 - 4
		winHeight, winWidth = msg.Height, msg.Width
		if !m.altscreen && m.progress.Width > maxWidth {
			m.progress.Width = maxWidth
		}
		m.termWidth = msg.Width
		m.termHeight = msg.Height
		return m, nil

	case timer.StartStopMsg:
		var cmd tea.Cmd
		m.timer, cmd = m.timer.Update(msg)
		return m, cmd

	case timer.TimeoutMsg:
		m.quitting = true
		save2(m)
		return m, tea.Quit

	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		return m, cmd

	case tea.KeyMsg:
		if key.Matches(msg, quitKeys) {
			m.quitting = true
			return m, tea.Quit
		}
		if key.Matches(msg, intKeys) {
			m.interrupting = true
			return m, tea.Quit
		}
		if key.Matches(msg, pause) {
			m.paused = !m.paused
			cmd := m.timer.Toggle()
			return m, cmd
		}
	}

	return m, nil
}

func (m model) View() string {
	switch m.mode {
	case "bar":
		return m.progressBar()
	case "count":
		return m.countDownTimer()
	default:
		return "Error: mode is required and must be either 'bar' or 'count'."
	}
}

func (m model) progressBar() string {
	if m.quitting || m.interrupting {
		return "\n"
	}

	result := boldStyle.Render(m.start.Format(time.Kitchen))
	if m.name != "" {
		result += ": " + italicStyle.Render(m.name)
	}
	result += " - " + boldStyle.Render(m.timer.View()) + "\n" + m.progress.View()
	if m.altscreen {
		textWidth, textHeight := lipgloss.Size(result)
		return lipgloss.NewStyle().Margin((winHeight-textHeight)/2, (winWidth-textWidth)/2).Render(result)
	}
	return result
}

func addSuffixIfArgIsNumber(s *string, suffix string) {
	_, err := strconv.ParseFloat(*s, 64)
	if err == nil {
		*s = *s + suffix
	}
}

func (m model) countDownTimer() string {
	if m.quitting || m.interrupting {
		return "\n"
	}

	remaining := m.duration - m.passed
	hours := remaining / time.Hour
	remaining -= hours * time.Hour
	minutes := remaining / time.Minute
	remaining -= minutes * time.Minute
	seconds := remaining / time.Second

	// Convert hours, minutes, and seconds to strings
	hStr := fmt.Sprintf("%02d", hours)
	mStr := fmt.Sprintf("%02d", minutes)
	sStr := fmt.Sprintf("%02d", seconds)

	// Build the big digit representation for each component of the time
	bigHour := buildBigDigitString(hStr)
	bigMinute := buildBigDigitString(mStr)
	bigSecond := buildBigDigitString(sStr)

	// Combine the big digit lines into a single string
	bigTime := ""
	for i := 0; i < 7; i++ {
		bigTime += bigHour[i] + Colon[i] + bigMinute[i] + Colon[i] + bigSecond[i] + "\n"
	}

	bigTimeLines := strings.Split(bigTime, "\n")
	bigTimeHeight := len(bigTimeLines)
	bigTimeWidth := len(bigTimeLines[0])

	topPadding := max(0, (m.termHeight-bigTimeHeight)/2)
	leftPadding := max(0, (m.termWidth-bigTimeWidth)/2)

	// Clear the screen
	clearScreen := "\033[H\033[2J"
	// Build the centered bigTime with padding
	centeredBigTime := clearScreen
	for i := 0; i < topPadding; i++ {
		centeredBigTime += "\n"
	}
	for _, line := range bigTimeLines {
		centeredBigTime += strings.Repeat(" ", leftPadding) + line + "\n"
	}

	//return clearScreen + "\n" + centeredBigTime
	return centeredBigTime
}

// max returns the maximum of two integers.
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// buildBigDigitString takes a string representation of a number (like "12")
// and returns a slice of strings where each string is a line of the big digit representation
func buildBigDigitString(numberStr string) []string {
	var result [7]string
	for line := 0; line < 7; line++ {
		for _, digit := range numberStr {
			result[line] += BigDigits[digit][line] + " "
		}
	}
	return result[:]
}
