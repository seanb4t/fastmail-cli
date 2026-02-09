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
	viewport        viewport.Model
	email           fastmail.Email
	loading         bool
	goBack          bool
	ready           bool
	width           int
	height          int
	status          statusModel
	pendingDelete   bool
	action          *emailAction
	showThread      bool
	showAttachments bool
	reply           bool
	replyAll        bool
	isPreview       bool
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

	case statusClearMsg:
		m.status.update(msg)
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	if m.ready {
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m emailReaderModel) handleKey(msg tea.KeyMsg) (emailReaderModel, tea.Cmd) {
	key := msg.String()

	// Cancel pending delete on any key that isn't "x"
	if m.pendingDelete && key != "x" {
		m.pendingDelete = false
		cmd := m.status.setStatus("", false)
		m.status.visible = false
		return m, cmd
	}

	switch key {
	case "q", keyEsc:
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
	case "a":
		m.action = &emailAction{kind: "archive", email: m.email}
		return m, nil
	case "x":
		if m.pendingDelete {
			m.pendingDelete = false
			m.action = &emailAction{kind: "delete", email: m.email}
			return m, nil
		}
		m.pendingDelete = true
		return m, m.status.setStatus("Press x again to delete", false)
	case "r":
		m.reply = true
		return m, nil
	case "R":
		m.replyAll = true
		return m, nil
	case ".":
		m.action = &emailAction{kind: "toggleRead", email: m.email}
		return m, nil
	case "f":
		m.action = &emailAction{kind: "toggleFlag", email: m.email}
		return m, nil
	case "m":
		m.action = &emailAction{kind: "move", email: m.email}
		return m, nil
	case "A":
		if len(m.email.Attachments) > 0 {
			m.showAttachments = true
		}
		return m, nil
	case "t":
		if m.email.ThreadID != "" {
			m.showThread = true
		}
		return m, nil
	}

	return m, nil
}

func (m emailReaderModel) view() string {
	if m.loading {
		if m.isPreview {
			return "Loading preview..."
		}
		return "\n  Loading email..."
	}

	header := m.renderHeaders()

	if m.isPreview {
		return header + "\n" + m.viewport.View()
	}

	var bottom string
	if s := m.status.view(); s != "" {
		bottom = s
	} else {
		bottom = readerHelpStyle.Render("  j/k scroll • d/u half-page • g/G top/bottom • r reply • R reply-all • a archive • x delete • . read • f flag • m move • A att • t thread • q back")
	}

	return header + "\n" + m.viewport.View() + "\n" + bottom
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

	if len(m.email.Attachments) > 0 {
		count := fmt.Sprintf("%d attachment(s)", len(m.email.Attachments))
		fmt.Fprintf(&b, "  %s  %s\n", headerLabelStyle.Render("Att:"), headerValueStyle.Render(count))
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
