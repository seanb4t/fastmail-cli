package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"

	"github.com/seanb4t/fastmail-cli/pkg/fastmail"
)

var (
	headerLabelStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	headerValueStyle = lipgloss.NewStyle()
	separatorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	readerHelpStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
)

// emailBodyLoadedMsg is sent when the full email body has been fetched.
type emailBodyLoadedMsg struct {
	email fastmail.Email
}

// emailReaderModel displays a single email with scrollable body.
type emailReaderModel struct {
	viewport viewport.Model
	email    fastmail.Email
	loading  bool
	goBack   bool
	ready    bool
	width    int
	height   int
}

func newEmailReaderModel(email fastmail.Email) emailReaderModel {
	return emailReaderModel{
		email:   email,
		loading: true,
	}
}

func (m *emailReaderModel) setSize(width, height int) {
	m.width = width
	m.height = height
	if m.ready {
		headerHeight := m.headerHeight()
		m.viewport.Width = width
		m.viewport.Height = height - headerHeight - 1 // 1 for help bar
	}
}

func (m *emailReaderModel) headerHeight() int {
	return len(strings.Split(m.renderHeaders(), "\n"))
}

func (m emailReaderModel) update(msg tea.Msg) (emailReaderModel, tea.Cmd) {
	switch msg := msg.(type) {
	case emailBodyLoadedMsg:
		m.email = msg.email
		m.loading = false

		// Initialize viewport now that we have content
		headerHeight := m.headerHeight()
		m.viewport = viewport.New(m.width, m.height-headerHeight-1)
		m.viewport.SetContent(m.renderBody())
		m.ready = true
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			m.goBack = true
			return m, nil
		case "d":
			m.viewport.HalfPageDown()
			return m, nil
		case "u":
			m.viewport.HalfPageUp()
			return m, nil
		case "g":
			m.viewport.GotoTop()
			return m, nil
		case "G":
			m.viewport.GotoBottom()
			return m, nil
		}
	}

	if m.ready {
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m emailReaderModel) view() string {
	if m.loading {
		return "\n  Loading email..."
	}

	header := m.renderHeaders()
	help := readerHelpStyle.Render("  j/k scroll • d/u half-page • g/G top/bottom • q back")

	return header + "\n" + m.viewport.View() + "\n" + help
}

func (m emailReaderModel) renderHeaders() string {
	var b strings.Builder

	from := m.email.From.String()
	if from == "" {
		from = "(unknown)"
	}
	fmt.Fprintf(&b, "  %s %s\n", headerLabelStyle.Render("From:"), headerValueStyle.Render(from))

	if len(m.email.To) > 0 {
		to := formatAddresses(m.email.To)
		fmt.Fprintf(&b, "  %s   %s\n", headerLabelStyle.Render("To:"), headerValueStyle.Render(to))
	}

	if len(m.email.Cc) > 0 {
		cc := formatAddresses(m.email.Cc)
		fmt.Fprintf(&b, "  %s   %s\n", headerLabelStyle.Render("Cc:"), headerValueStyle.Render(cc))
	}

	subject := m.email.Subject
	if subject == "" {
		subject = "(no subject)"
	}
	fmt.Fprintf(&b, "  %s %s\n", headerLabelStyle.Render("Subj:"), headerValueStyle.Render(subject))

	if !m.email.ReceivedAt.IsZero() {
		date := m.email.ReceivedAt.Format("Mon, 02 Jan 2006 15:04 MST")
		fmt.Fprintf(&b, "  %s %s\n", headerLabelStyle.Render("Date:"), headerValueStyle.Render(date))
	}

	b.WriteString(separatorStyle.Render("  " + strings.Repeat("─", max(m.width-4, 20))))

	return b.String()
}

func (m emailReaderModel) renderBody() string {
	body := m.email.Body
	if body == "" {
		return "  (no body content)"
	}

	// If the body looks like HTML, try glamour rendering
	if strings.Contains(body, "<") && strings.Contains(body, ">") {
		width := max(m.width-4, 40)
		r, err := glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
			glamour.WithWordWrap(width),
		)
		if err == nil {
			if rendered, err := r.Render(body); err == nil {
				return rendered
			}
		}
	}

	// Plain text — wrap with indentation
	return "  " + strings.ReplaceAll(body, "\n", "\n  ")
}

func formatAddresses(addrs []fastmail.EmailAddress) string {
	parts := make([]string, len(addrs))
	for i, a := range addrs {
		parts[i] = a.String()
	}
	return strings.Join(parts, ", ")
}
