package design

import (

	"charm.land/lipgloss/v2"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1)

	detailBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")).
			Padding(1, 2)

	methodGetStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("42"))
	methodPostStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("214"))
)

