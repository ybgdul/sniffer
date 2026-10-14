package design

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sniffer/internal/reassemble"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type viewMode int 

const (
    modeTable viewMode = iota 
    modeDetail
)

type MessageTranscation struct{ 
    Tx *reassemble.HttpTranscation
}

type Model struct{ 
    table table.Model
    transcations []*reassemble.HttpTranscation
    selected *reassemble.HttpTranscation
    mode viewMode
    width int 
    height int 
    incoming <-chan *reassemble.HttpTranscation
}

func NewModel(incoming <-chan *reassemble.HttpTranscation) Model { 
    t := table.New(
            table.WithColumns(
                []table.Column{
					{Title: "STATUS", Width: 8},
					{Title: "METHOD", Width: 8},
					{Title: "HOST", Width: 20},
					{Title: "URL", Width: 25},
					{Title: "LATENCY", Width: 10},
					{Title: "CLIENT IP", Width: 15},
                },
            ),
            table.WithFocused(true),
            table.WithHeight(15),
        )
    s := table.DefaultStyles()
    s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(true)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(true)
	t.SetStyles(s)
	return Model{
        incoming: incoming,
        table: t,
        mode: modeTable,
	}
}

func waitForNextRequest(s <-chan *reassemble.HttpTranscation) tea.Cmd{ 
    return func() tea.Msg {
        transcation, ok := <-s
        if !ok { 
            return nil 
        }
        return MessageTranscation{Tx:transcation}
    }
}

func (m Model) Init() tea.Cmd { 
	return waitForNextRequest(m.incoming)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.mode == modeDetail {
				m.mode = modeTable
				return m, nil
			}
			return m, tea.Quit

		case "enter":
			if m.mode == modeTable && len(m.transcations) > 0 {
				idx := m.table.Cursor()
				if idx >= 0 && idx < len(m.transcations) {
					m.selected = m.transcations[idx]
					m.mode = modeDetail
				}
			}

		case "esc":
			if m.mode == modeDetail {
				m.mode = modeTable
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.table.SetWidth(msg.Width - 4)
		m.table.SetHeight(msg.Height - 6)

	case MessageTranscation:
		m.transcations = append(m.transcations, msg.Tx)

		var statusStr string
		if msg.Tx.StatusCode > 0 {
			switch {
			case msg.Tx.StatusCode >= 200 && msg.Tx.StatusCode < 300:
				statusStr = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render(fmt.Sprintf("%d OK", msg.Tx.StatusCode))
			case msg.Tx.StatusCode >= 400 && msg.Tx.StatusCode < 500:
				statusStr = lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Render(fmt.Sprintf("%d ERR", msg.Tx.StatusCode))
			case msg.Tx.StatusCode >= 500:
				statusStr = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render(fmt.Sprintf("%d FAIL", msg.Tx.StatusCode))
			default:
				statusStr = fmt.Sprintf("%d", msg.Tx.StatusCode)
			}
		} else { 
			statusStr = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("PENDING")
		}

		row := table.Row{
			statusStr,
			msg.Tx.Method,
			msg.Tx.Host,
			msg.Tx.URL,
			msg.Tx.Duration.Round(time.Millisecond).String(),
			msg.Tx.ClientIP,
		}

		rows := m.table.Rows()
		rows = append(rows, row)
		m.table.SetRows(rows)

		return m, waitForNextRequest(m.incoming)
	}

	if m.mode == modeTable {
		m.table, cmd = m.table.Update(msg)
	}

	return m, cmd
}

func (m Model) View() string {
	if m.mode == modeDetail && m.selected != nil {
		return m.renderDetailView()
	}

	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1).
		Render("LIVE HTTP TRANSACTION SNIFFER")

	help := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("\n  ↑/↓: Navigate • Enter: Inspect • q: Quit")

	return fmt.Sprintf("\n %s\n\n%s\n%s", header, m.table.View(), help)
}

func (m Model) parseJson(raw string) string { 
	if strings.TrimSpace(raw) == ""{
		return "(No Body Payload)"
	}

	var formattedJson bytes.Buffer
	err := json.Indent(&formattedJson, []byte(raw), "", " ")
	if err != nil {
		return raw
	}
	return formattedJson.String()
}
func (m Model) renderDetailView() string {
	tx := m.selected

	var reqHeaders []string
	for k, v := range tx.ReqHeader {
		reqHeaders = append(reqHeaders, fmt.Sprintf("  %s: %s", k, strings.Join(v, ", ")))
	}

	var respHeaders []string
	for k, v := range tx.RespHeader {
		respHeaders = append(respHeaders, fmt.Sprintf("  %s: %s", k, strings.Join(v, ", ")))
	}

	content := fmt.Sprintf(
		"METHOD:      %s\n"+
			"URL:         %s\n"+
			"STATUS:      %d %s\n"+
			"LATENCY:     %s\n"+
			"CLIENT IP:   %s\n\n"+
			"── REQUEST HEADERS ──────────────────────────\n%s\n\n"+
			"── REQUEST BODY ─────────────────────────────\n%s\n\n"+
			"── RESPONSE HEADERS ─────────────────────────\n%s\n\n"+
			"── RESPONSE BODY ────────────────────────────\n%s\n",
		tx.Method,
		tx.URL,
		tx.StatusCode, tx.StatusText,
		tx.Duration.Round(time.Millisecond).String(),
		tx.ClientIP,
		strings.Join(reqHeaders, "\n"),
		tx.ReqBody,
		strings.Join(respHeaders, "\n"),
		tx.RespBody,
	)

	detailBoxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(1, 2)

	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1).
		Render("TRANSACTION INSPECTOR")

	help := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("\n  Esc / q: Back to list")

	return fmt.Sprintf("\n %s\n\n%s\n%s", header, detailBoxStyle.Render(content), help)
}