package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type composeField int

const (
	fieldTo composeField = iota
	fieldSubject
	fieldBody
)

type emailSentMsg struct {
	emailID string
}

type emailSendErrMsg struct {
	err error
}

type composeModel struct {
	toInput      textinput.Model
	subjectInput textinput.Model
	bodyInput    textarea.Model
	activeField  composeField
	canceled     bool
	send         bool
	width        int
	height       int
	err          string
}

func newComposeModel() composeModel {
	to := textinput.New()
	to.Placeholder = "recipient@example.com"
	to.Focus()
	to.CharLimit = 256
	to.Prompt = "  To:      "

	subj := textinput.New()
	subj.Placeholder = "Subject"
	subj.CharLimit = 256
	subj.Prompt = "  Subject: "

	body := textarea.New()
	body.Placeholder = "Write your message..."
	body.CharLimit = 0 // unlimited
	body.SetWidth(80)
	body.SetHeight(10)

	return composeModel{
		toInput:      to,
		subjectInput: subj,
		bodyInput:    body,
		activeField:  fieldTo,
	}
}

func (m *composeModel) setSize(width, height int) {
	m.width = width
	m.height = height
	m.bodyInput.SetWidth(width - 4)
	bodyHeight := height - 10
	if bodyHeight < 5 {
		bodyHeight = 5
	}
	m.bodyInput.SetHeight(bodyHeight)
}

func (m composeModel) update(msg tea.Msg) (composeModel, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case keyEsc:
			m.canceled = true
			return m, nil
		case "ctrl+s":
			return m.handleSend()
		case "tab":
			m.nextField()
			return m, nil
		case "shift+tab":
			m.prevField()
			return m, nil
		}
	}

	// Forward to active field
	var cmd tea.Cmd
	switch m.activeField {
	case fieldTo:
		m.toInput, cmd = m.toInput.Update(msg)
	case fieldSubject:
		m.subjectInput, cmd = m.subjectInput.Update(msg)
	case fieldBody:
		m.bodyInput, cmd = m.bodyInput.Update(msg)
	}
	return m, cmd
}

func (m composeModel) handleSend() (composeModel, tea.Cmd) {
	to := strings.TrimSpace(m.toInput.Value())
	subject := strings.TrimSpace(m.subjectInput.Value())

	if to == "" {
		m.err = "To field is required"
		return m, nil
	}
	if subject == "" {
		m.err = "Subject is required"
		return m, nil
	}

	m.err = ""
	m.send = true
	return m, nil
}

func (m *composeModel) nextField() {
	m.blurAll()
	switch m.activeField {
	case fieldTo:
		m.activeField = fieldSubject
		m.subjectInput.Focus()
	case fieldSubject:
		m.activeField = fieldBody
		m.bodyInput.Focus()
	case fieldBody:
		m.activeField = fieldTo
		m.toInput.Focus()
	}
}

func (m *composeModel) prevField() {
	m.blurAll()
	switch m.activeField {
	case fieldTo:
		m.activeField = fieldBody
		m.bodyInput.Focus()
	case fieldSubject:
		m.activeField = fieldTo
		m.toInput.Focus()
	case fieldBody:
		m.activeField = fieldSubject
		m.subjectInput.Focus()
	}
}

func (m *composeModel) blurAll() {
	m.toInput.Blur()
	m.subjectInput.Blur()
	m.bodyInput.Blur()
}

var (
	composeLabelStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	composeHelpStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	composeErrStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
)

func (m composeModel) view() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(composeLabelStyle.Render("  New Email"))
	b.WriteString("\n\n")

	b.WriteString(m.toInput.View())
	b.WriteString("\n")
	b.WriteString(m.subjectInput.View())
	b.WriteString("\n\n")
	b.WriteString("  " + m.bodyInput.View())
	b.WriteString("\n")

	if m.err != "" {
		b.WriteString("\n")
		b.WriteString(composeErrStyle.Render("  " + m.err))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(composeHelpStyle.Render("  tab/shift+tab navigate • ctrl+s send • esc cancel"))

	return b.String()
}

// toAddress returns the "To" field value trimmed.
func (m composeModel) toAddress() string {
	return strings.TrimSpace(m.toInput.Value())
}

// subject returns the trimmed subject.
func (m composeModel) subject() string {
	return strings.TrimSpace(m.subjectInput.Value())
}

// body returns the body text.
func (m composeModel) body() string {
	return m.bodyInput.Value()
}
