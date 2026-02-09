package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"

	"github.com/seanb4t/fastmail-cli/pkg/fastmail"
)

// threadLoadedMsg carries the emails in a thread.
type threadLoadedMsg struct {
	emails []fastmail.Email
}

// threadItem implements list.DefaultItem for thread emails.
type threadItem struct {
	email fastmail.Email
}

func (i threadItem) Title() string {
	from := i.email.From.Email
	if i.email.From.Name != "" {
		from = i.email.From.Name
	}
	return from
}

func (i threadItem) Description() string {
	date := ""
	if !i.email.ReceivedAt.IsZero() {
		date = formatAge(i.email.ReceivedAt)
	}
	return fmt.Sprintf("%s  %s", date, truncate(i.email.Preview, 60))
}

func (i threadItem) FilterValue() string {
	return i.email.Subject + " " + i.email.From.String()
}

type threadViewModel struct {
	list     list.Model
	email    fastmail.Email // original email that opened thread
	loading  bool
	goBack   bool
	viewport viewport.Model
	viewing  *fastmail.Email // currently viewing body
	ready    bool
	width    int
	height   int
}

func newThreadViewModel(email fastmail.Email) threadViewModel {
	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(lipgloss.Color("12")).
		BorderForeground(lipgloss.Color("12"))
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(lipgloss.Color("8")).
		BorderForeground(lipgloss.Color("12"))

	l := list.New(nil, delegate, 0, 0)
	l.Title = fmt.Sprintf("Thread: %s", email.Subject)
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(true)

	return threadViewModel{
		list:    l,
		email:   email,
		loading: true,
	}
}

func (m *threadViewModel) setSize(width, height int) {
	m.width = width
	m.height = height
	m.list.SetSize(width, height)
	if m.ready {
		m.viewport.Width = width
		m.viewport.Height = height - 5 // room for header + help
	}
}

func (m threadViewModel) update(msg tea.Msg) (threadViewModel, tea.Cmd) {
	switch msg := msg.(type) {
	case threadLoadedMsg:
		items := make([]list.Item, len(msg.emails))
		for i, e := range msg.emails {
			items[i] = threadItem{email: e}
		}
		m.loading = false
		cmd := m.list.SetItems(items)
		return m, cmd

	case tea.KeyMsg:
		if m.viewing != nil {
			return m.handleViewingKey(msg.String())
		}
		if !m.list.SettingFilter() {
			if m.handleKey(msg.String()) {
				return m, nil
			}
		}
	}

	if m.viewing != nil {
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		return m, cmd
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *threadViewModel) handleKey(key string) bool {
	switch key {
	case keyEsc, "q":
		m.goBack = true
		return true
	case keyEnter:
		if item, ok := m.list.SelectedItem().(threadItem); ok {
			email := item.email
			m.viewing = &email
			m.viewport = viewport.New(m.width, m.height-5)
			m.viewport.SetContent(m.renderThreadEmail(email))
			m.ready = true
		}
		return true
	}
	return false
}

func (m *threadViewModel) handleViewingKey(key string) (threadViewModel, tea.Cmd) {
	switch key {
	case keyEsc, "q":
		m.viewing = nil
		m.ready = false
		return *m, nil
	case "d":
		m.viewport.HalfPageDown()
		return *m, nil
	case "u":
		m.viewport.HalfPageUp()
		return *m, nil
	case "g":
		m.viewport.GotoTop()
		return *m, nil
	case "G":
		m.viewport.GotoBottom()
		return *m, nil
	}
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)})
	return *m, cmd
}

func (m threadViewModel) renderThreadEmail(email fastmail.Email) string {
	var b strings.Builder

	from := email.From.String()
	if from == "" {
		from = "(unknown)"
	}
	fmt.Fprintf(&b, "  %s %s\n", headerLabelStyle.Render("From:"), headerValueStyle.Render(from))

	if len(email.To) > 0 {
		to := formatAddresses(email.To)
		fmt.Fprintf(&b, "  %s   %s\n", headerLabelStyle.Render("To:"), headerValueStyle.Render(to))
	}

	fmt.Fprintf(&b, "  %s %s\n", headerLabelStyle.Render("Subj:"), headerValueStyle.Render(email.Subject))

	if !email.ReceivedAt.IsZero() {
		date := email.ReceivedAt.Format("Mon, 02 Jan 2006 15:04 MST")
		fmt.Fprintf(&b, "  %s %s\n", headerLabelStyle.Render("Date:"), headerValueStyle.Render(date))
	}

	b.WriteString(separatorStyle.Render("  " + strings.Repeat("\u2500", max(m.width-4, 20))))
	b.WriteString("\n\n")

	body := email.Body
	if body == "" {
		body = email.Preview
	}

	switch {
	case body == "":
		b.WriteString("  (no body content)")
	case strings.Contains(body, "<") && strings.Contains(body, ">"):
		width := max(m.width-4, 40)
		r, err := glamour.NewTermRenderer(glamour.WithAutoStyle(), glamour.WithWordWrap(width))
		if err == nil {
			if rendered, rerr := r.Render(body); rerr == nil {
				b.WriteString(rendered)
			} else {
				b.WriteString("  " + strings.ReplaceAll(body, "\n", "\n  "))
			}
		}
	default:
		b.WriteString("  " + strings.ReplaceAll(body, "\n", "\n  "))
	}

	return b.String()
}

func (m threadViewModel) view() string {
	if m.loading {
		return fmt.Sprintf("\n  Loading thread for %q...", m.email.Subject)
	}
	if m.viewing != nil {
		help := readerHelpStyle.Render("  d/u half-page \u2022 g/G top/bottom \u2022 q/esc back to thread")
		return m.viewport.View() + "\n" + help
	}
	return m.list.View()
}
