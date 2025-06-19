package main

// A simple program demonstrating the text area component from the Bubbles
// component library.

import (
	"bufio"
	"embed"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const gap = "\n\n"

var display_name, server_address string
var conn net.Conn

//go:embed assets/*.mp3
var soundAssets embed.FS

func main() {
	rd := bufio.NewReader(os.Stdin)

	fmt.Print("\nEnter Display Name: ")
	raw, _ := rd.ReadString('\n')
	display_name = strings.TrimSpace(raw)

	fmt.Print("\nEnter Server Address: ")
	raw_addy, _ := rd.ReadString('\n')
	server_address = strings.TrimSpace(raw_addy) + ":8000"

	c, err := net.Dial("tcp", server_address)
	if err != nil {
		log.Fatal(err)
	}
	conn = c

	playSound("assets/zelda_secret.mp3")

	p := tea.NewProgram(initialModel())

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

type (
	errMsg      error
	incomingMsg string
)

type model struct {
	viewport    viewport.Model
	messages    []string
	textarea    textarea.Model
	senderStyle lipgloss.Style
	serverStyle lipgloss.Style
	aiStyle     lipgloss.Style
	err         error
}

func initialModel() model {
	ta := textarea.New()
	ta.Placeholder = "Send a message..."
	ta.Focus()

	ta.Prompt = "┃ "
	ta.CharLimit = 280

	ta.SetWidth(30)
	ta.SetHeight(3)

	// Remove cursor line styling
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()

	ta.ShowLineNumbers = false

	vp := viewport.New(30, 5)
	vp.SetContent(`
 ______     ______        ______     __  __     ______     ______
/\  ___\   /\  __ \      /\  ___\   /\ \_\ \   /\  __ \   /\__  _\
\ \ \__ \  \ \ \/\ \     \ \ \____  \ \  __ \  \ \  __ \  \/_/\ \/
 \ \_____\  \ \_____\     \ \_____\  \ \_\ \_\  \ \_\ \_\    \ \_\
  \/_____/   \/_____/      \/_____/   \/_/\/_/   \/_/\/_/     \/_/

`)

	_, err := conn.Write([]byte(display_name + "\n"))
	if err != nil {
		fmt.Printf("error sending message: %q", err)
	}

	fmt.Printf("\nJoined room as %s\n", display_name)

	ta.KeyMap.InsertNewline.SetEnabled(false)

	return model{
		textarea:    ta,
		messages:    []string{},
		viewport:    vp,
		senderStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("5")),
		serverStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("3")),
		aiStyle:     lipgloss.NewStyle().Foreground(lipgloss.Color("12")),
		err:         nil,
	}
}

func readIncoming(conn net.Conn) tea.Cmd {
	return func() tea.Msg {
		rd := bufio.NewReader(conn)
		line, err := rd.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				return incomingMsg("<server disconnected>")
			}
			return errMsg(err)
		}

		playSound("assets/noti.mp3")

		return incomingMsg(strings.TrimRight(line, "\r\n"))
	}
}

func playSound(sound string) {
	f, err := soundAssets.Open(sound)
	if err != nil {
		log.Fatal("sound open:", err)
	}
	streamer, format, err := mp3.Decode(f)
	if err != nil {
		log.Fatal("mp3 decode:", err)
	}

	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))

	speaker.Play(beep.Seq(
		streamer,
		beep.Callback(func() {
			streamer.Close()
		}),
	))
}

func extractName(msg string) (t bool, name, text string) {
	name_index := strings.Index(msg, ":")

	if name_index != -1 {
		name = msg[:name_index]
		text = msg[name_index+2:]
	} else {
		return false, "", ""
	}

	return true, name, text
}

func (m model) Init() tea.Cmd {
	return tea.Batch(textarea.Blink, readIncoming(conn))
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)

	m.textarea, tiCmd = m.textarea.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.viewport.Width = msg.Width
		m.textarea.SetWidth(msg.Width)
		m.viewport.Height = msg.Height - m.textarea.Height() - lipgloss.Height(gap)

		if len(m.messages) > 0 {
			// Wrap content before setting it.
			m.viewport.SetContent(lipgloss.NewStyle().Width(m.viewport.Width).Render(strings.Join(m.messages, "\n")))
		}
		m.viewport.GotoBottom()
	case incomingMsg:
		if t, name, text := extractName(string(msg)); t {
			if name == "AI" {
				m.messages = append(m.messages, m.aiStyle.Render(name+": ")+text)
			} else {
				m.messages = append(m.messages, m.senderStyle.Render(name+": ")+text)
			}
		} else {
			m.messages = append(m.messages, m.serverStyle.Render(string(msg)))
		}
		m.viewport.SetContent(lipgloss.NewStyle().Width(m.viewport.Width).Render(strings.Join(m.messages, "\n")))
		m.textarea.Reset()
		m.viewport.GotoBottom()
		return m, readIncoming(conn)
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			fmt.Println(m.textarea.Value())
			return m, tea.Quit
		case tea.KeyEnter:
			msg := m.textarea.Value()
			m.messages = append(m.messages, m.senderStyle.Render(display_name+": ")+msg)
			_, err := conn.Write([]byte(display_name + ": " + msg + "\n"))
			if err != nil {
				fmt.Printf("error sending message: %q", err)
				break
			}
			m.viewport.SetContent(lipgloss.NewStyle().Width(m.viewport.Width).Render(strings.Join(m.messages, "\n")))
			m.textarea.Reset()
			m.viewport.GotoBottom()
		}

	case errMsg:
		m.err = msg
		return m, nil
	}

	return m, tea.Batch(tiCmd, vpCmd)
}

func (m model) View() string {
	return fmt.Sprintf(
		"%s%s%s",
		m.viewport.View(),
		gap,
		m.textarea.View(),
	)
}
