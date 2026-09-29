package design

import (
	"fmt"
	"sniffer/internal/reassemble"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
)

type viewMode int 

const (
    modeTable viewMode = iota 
    modeDetail
)

type MessageReq struct{ 
    Req *reassemble.HttpRequest
}

type Model struct{ 
    table table.Model
    requests []*reassemble.HttpRequest
    selected *reassemble.HttpRequest
    mode viewMode
    width int 
    height int 
    incoming <-chan *reassemble.HttpRequest
}

func NewModel(incoming <-chan *reassemble.HttpRequest) Model { 
    t := table.New(
            table.WithColumns(
                []table.Column{
                    {Title: "METHOD", Width: 8},
                    {Title: "HOST", Width: 20},
                    {Title: "URL", Width: 35},
                    {Title: "CLIENT IP", Width: 35},
                    {Title: "PROTOCOL", Width: 8},
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

func waitForNextRequest(s <-chan *reassemble.HttpRequest) tea.Cmd{ 
    return func() tea.Msg {
        req, ok := <-s
        if !ok { 
            return nil 
        }
        return MessageReq{Req:req}
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
			if m.mode == modeTable && len(m.requests) > 0 {
				idx := m.table.Cursor()
				if idx >= 0 && idx < len(m.requests) {
					m.selected = m.requests[idx]
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

	case MessageReq:
		// Append new request to state
		m.requests = append(m.requests, msg.Req)

		// Create table row
		row := table.Row{
			msg.Req.Method,
			msg.Req.Host,
			msg.Req.URL,
			msg.Req.IP,
			msg.Req.Protocol,
		}

		rows := m.table.Rows()
		rows = append(rows, row)
		m.table.SetRows(rows)

		// Continue listening on channel
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

	// Render Main Table View
	header := titleStyle.Render("LIVE NETWORK TRAFFIC SNIFFER")
	help := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("\n  ↑/↓: Navigate • Enter: Inspect • q: Quit")

	return fmt.Sprintf("\n %s\n\n%s\n%s", header, m.table.View(), help)
}

func (m Model) renderDetailView() string {
	req := m.selected

	var methodStyled string
	if req.Method == "GET" {
		methodStyled = methodGetStyle.Render(req.Method)
	} else {
		methodStyled = methodPostStyle.Render(req.Method)
	}

	content := fmt.Sprintf(
		"METHOD:    %s\n"+
			"HOST:      %s\n"+
			"URL:       %s\n"+
			"CLIENT IP: %s\n"+
			"PROTOCOL:  %s\n",
		methodStyled,
		req.Host,
		req.URL,
		req.IP,
		req.Protocol,
	)

	box := detailBoxStyle.Render(content)
	header := titleStyle.Render("REQUEST INSPECTOR")
	help := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("\n  Esc / q: Back to list")

	return fmt.Sprintf("\n %s\n\n%s\n%s", header, box, help)
}