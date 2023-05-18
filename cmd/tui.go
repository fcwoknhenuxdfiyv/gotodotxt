/*
Copyright © 2022 Jason Quigley <jason@jasonquigley.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/

package cmd

import (
	"fmt"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"

	"gotodotxt/tdt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gookit/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type fileChangedMsg struct {
	Boo bool
}

var (
	projectsRegex = regexp.MustCompile(`\s+\+(\w+)\b`)
	contextsRegex = regexp.MustCompile(`\s+@(\w+)\b`)
	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#25A065")).
			Render
	grey = color.Gray.Render
)

type model struct {
	file         *tdt.TaskFile
	rows         Rows
	cursor       int
	beginningEnd int
	endBeginning int
	selected     map[int]struct{}
	command      string
	textInput    textinput.Model
}

func newModel(fn string) model {
	ti := textinput.New()
	ti.Placeholder = "Parameters"
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 20

	opts := tdt.Opts{
		ShowFuture: showFuture,
		SortOrder:  sortOrder,
	}
	m := model{
		file:      tdt.Watch(fn, opts).Sort().Filter(),
		selected:  make(map[int]struct{}),
		textInput: ti,
	}
	m.rows = renderTasks(m.file, false)
	return m
}

func (m model) Init() tea.Cmd {
	return tea.Batch(waitForFileChanges(m.file.Events), textinput.Blink)
}

var winHeight, winWidth int

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		winHeight = msg.Height
		winWidth = msg.Width
		m.textInput.Width = winWidth - 4

	case tdt.FileChangedEvent:
		// m.eep = msg.EventName
		m.file = tdt.Watch(m.file.Path, m.file.Opts).Sort().Filter()
		m.rows = renderTasks(m.file, false)
		return m, waitForFileChanges(m.file.Events)

	case tea.KeyMsg:
		if m.command == "" {
			// m.eep = msg.String()
			switch msg.String() {

			case "1":
				m.file = tdt.Watch(viper.GetString("file"), m.file.Opts)
				m.reset(false)
				return m, waitForFileChanges(m.file.Events)

			case "2", "3", "4", "5", "6", "7", "8", "9", "0":
				num, err := strconv.Atoi(msg.String())
				if err != nil {
					return m, nil
				}
				if len(viper.GetStringSlice("other-files")) > num-2 {
					m.file = tdt.Watch(viper.GetStringSlice("other-files")[num-2], m.file.Opts)
					m.reset(false)
					return m, waitForFileChanges(m.file.Events)
				}

			case "A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K",
				"L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V",
				"W", "X", "Y", "Z":
				m.file.Edit("("+msg.String()+")", true, m.getSelected()...)
				m.refresh(true)

			case "a":
				m.command = "archive"
				m.textInput.Placeholder = "Type yes to archive"
				m.textInput.Reset()

			case "e":
				if len(m.selected) == 0 {
					m.command = "editOne"
					// Log(log.Debug, m.cursor, len(m.file.Tasks), len(m.rows), m.rows[m.cursor].LineNumber)
					orig := m.file.Original(m.rows[m.cursor].LineNumber)
					m.textInput.Placeholder = orig
					m.textInput.SetValue(orig)
					m.textInput.SetCursor(0)
				} else {
					m.command = "editMany"
					m.textInput.Placeholder = "(A) t:today due:today rec:1w"
					m.textInput.Reset()
				}

			case "f":
				m.file.Opts.ShowFuture = !m.file.Opts.ShowFuture
				m.reset(false)

			case "g":
				m.cursor = 0

			case "ctrl+g":
				m.cursor = len(m.rows) - 1

			case "j", "down":
				if m.cursor < len(m.rows)-1 {
					m.cursor++
				}

			case "k", "up":
				if m.cursor > 0 {
					m.cursor--
				}

			case "n":
				m.command = "new"
				m.textInput.Placeholder = "Write new task"
				m.textInput.Reset()

			case "q", "ctrl+c":
				m.file.Write()
				return m, tea.Quit

			case "x":
				m.file.Toggle(m.getSelected()...)
				m.refresh(true)

			case "backspace", "delete":
				m.command = "delete"
				m.textInput.Placeholder = "Type yes to delete"
				m.textInput.Reset()

			case "enter", " ":
				ln := m.rows[m.cursor].LineNumber
				_, ok := m.selected[ln]
				if ok {
					delete(m.selected, ln)
				} else {
					m.selected[ln] = struct{}{}
				}

			case "esc":
				m.selected = make(map[int]struct{})
				m.command = ""

			case "[":
				ids := m.getSelected()
				m.file.Edit("t:1d due:1d", false, ids...)
				m.refresh(true)

			case "{":
				ids := m.getSelected()
				m.file.Edit("t:1d due:1d", true, ids...)
				m.refresh(true)

			case "]":
				ids := m.getSelected()
				m.file.Edit("t:1w due:1w", false, ids...)
				m.refresh(true)

			case "}":
				ids := m.getSelected()
				m.file.Edit("t:1w due:1w", true, ids...)
				m.refresh(true)

			}

		} else {
			switch msg.String() {

			case "enter":
				switch m.command {
				case "editOne":
					ids := m.getSelected()
					if len(ids) == 1 {
						m.file.Replace(m.textInput.Value(), ids[0])
						m.refresh(true)
					}
				case "editMany":
					ids := m.getSelected()
					m.file.Edit(m.textInput.Value(), true, ids...)
					m.refresh(true)
				case "new":
					m.file.Add(m.textInput.Value())
					m.refresh(true)
				case "archive":
					if isYes(m.textInput.Value()) {
						m.file.Archive()
						m.reset(true)
					}
				case "delete":
					if isYes(m.textInput.Value()) {
						m.file.Delete(m.getSelected()...)
						m.reset(true)
					}
				}
				m.command = ""

			case "esc":
				m.command = ""

			}

			m.textInput, cmd = m.textInput.Update(msg)
			return m, cmd
		}
	}

	// Logf(log.Noticef, "%d", runtime.NumGoroutine())
	// Logf(log.Debugf, "%d - %dx%d", m.cursor, winWidth, winHeight)
	return m, nil
}

func (m model) View() string {
	return m.header() + m.body() + m.footer()
}

func (m *model) refresh(writeFile bool) {
	m.file.Sort().Filter()
	m.rows = renderTasks(m.file, false)
	if writeFile {
		go m.file.Write()
	}
}

func (m *model) reset(writeFile bool) {
	m.selected = make(map[int]struct{})
	m.cursor = 0
	m.refresh(writeFile)
}

func (m model) header() string {
	selected := "Last change: " + m.file.LastUpdate.Local().Format("15:04:05")
	if len(m.selected) > 0 {
		selected = fmt.Sprintf("%d selected", len(m.selected))
	}
	s := fmt.Sprintf("  %s • %d tasks • %s \n\n",
		path.Base(m.file.Path),
		len(m.rows),
		selected,
	)
	return s
}

func (m model) footer() string {
	if m.command != "" {
		return fmt.Sprintf("\n  %s\n  %s", m.textInput.View(), "(esc to cancel)")
	} else {
		s := "\n      " + grey("sort: "+sortOrder) + "\n"
		s += "      spc:select n:new x:toggle e:edit q:quit a:archive\n"
		s += "      f:future A-Z:pri z:no pri [:+1 day ]:+1 week"
		return s
	}
}

func (m model) body() string {
	maxRows := winHeight - 6

	if len(m.rows) == 0 {
		if winHeight > 0 {
			return strings.Repeat("\n", maxRows)
		} else {
			return ""
		}
	}

	if m.cursor >= len(m.rows) {
		m.cursor = len(m.rows) - 1
	}

	start := 0
	finish := 0
	rowCount := 0
	padTop := false

	rowCount = 0
	for i := 0; i < len(m.rows); i++ {
		r := m.rows[i]
		finish = i
		rowCount += r.Lines
		if rowCount > maxRows {
			finish--
			break
		}
	}

	if m.cursor > finish {
		padTop = true
		finish = m.cursor
		rowCount = 0
		for i := m.cursor; i >= 0; i-- {
			r := m.rows[i]
			start = i + 1
			rowCount += r.Lines
			if rowCount > maxRows {
				break
			}
		}
	}

	s := ""
	rowCount = 0
	for i := start; i <= finish; i++ {
		r := m.rows[i]
		line1 := r.Line1
		line2 := r.Line2

		cursor := " "
		if m.cursor == i {
			cursor = "│"
		}

		selected := " "
		if _, ok := m.selected[int(r.LineNumber)]; ok {
			selected = "*"
			// line1 = selectedStyle(line1)
			// line2 = selectedStyle(line2)
		}
		// Render the row
		s += fmt.Sprintf("%s%s%s\n", cursor, selected, line1)
		rowCount++
		if r.Lines == 2 {
			s += fmt.Sprintf("%s %s\n", cursor, line2)
			rowCount++
		}
	}

	if maxRows-rowCount > 0 {
		if padTop {
			s = strings.Repeat("\n", maxRows-rowCount) + s
		} else {
			s += strings.Repeat("\n", maxRows-rowCount)
		}
	}

	return s
}

func isYes(txt string) bool {
	return strings.ToLower(txt) == "yes"
}

func (m *model) getSelected() []int {
	var selected []int
	if len(m.selected) == 0 {
		selected = append(selected, m.rows[m.cursor].LineNumber)
	} else {
		for k := range m.selected {
			selected = append(selected, k)
		}
	}
	return selected
}

func waitForFileChanges(changed chan tdt.FileChangedEvent) tea.Cmd {
	return func() tea.Msg {
		return tdt.FileChangedEvent(<-changed)
	}
}

func tui() {
	printOnExit = false
	davMode = checkDavMode()

	p := tea.NewProgram(newModel(viper.GetString("file")), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("%v", err)
		os.Exit(1)
	}
}

var tuiAliases = []string{"tui2"}
var tuiCmd = &cobra.Command{
	Use:     "tui",
	Aliases: tuiAliases,
	Short:   "Run in interactive mode",
	Long: `Run in interactive mode

Add stuff about keystrokes, etc.
blah, blah, blah.`,
	Run: func(cmd *cobra.Command, args []string) {
		tui()
	},
}

func init() {
	rootCmd.AddCommand(tuiCmd)
}
