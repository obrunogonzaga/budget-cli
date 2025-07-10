package style

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	Primary   = lipgloss.Color("#7D56F4")
	Secondary = lipgloss.Color("#F97316")
	Success   = lipgloss.Color("#10B981")
	Danger    = lipgloss.Color("#EF4444")
	Warning   = lipgloss.Color("#F59E0B")
	Info      = lipgloss.Color("#3B82F6")
	Muted     = lipgloss.Color("#6B7280")

	Background = lipgloss.Color("#1F2937")
	Surface    = lipgloss.Color("#374151")
	Border     = lipgloss.Color("#4B5563")
	Text       = lipgloss.Color("#F3F4F6")
	TextMuted  = lipgloss.Color("#9CA3AF")
)

var (
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(Primary).
			MarginBottom(1)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(TextMuted).
			MarginBottom(1)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(Danger).
			Bold(true)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(Success).
			Bold(true)

	WarningStyle = lipgloss.NewStyle().
			Foreground(Warning).
			Bold(true)

	InfoStyle = lipgloss.NewStyle().
			Foreground(Info)

	BorderStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(Border).
			Padding(1, 2)

	MenuItemStyle = lipgloss.NewStyle().
			PaddingLeft(2)

	SelectedMenuItemStyle = lipgloss.NewStyle().
				Foreground(Primary).
				Bold(true).
				PaddingLeft(2)

	HeaderStyle = lipgloss.NewStyle().
			Background(Surface).
			Foreground(Text).
			Bold(true).
			Padding(0, 1)

	TableHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(Primary).
				BorderBottom(true).
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(Border)

	HelpStyle = lipgloss.NewStyle().
			Foreground(TextMuted).
			MarginTop(1)

	InputStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(Border).
			Padding(0, 1)

	FocusedInputStyle = InputStyle.Copy().
				BorderForeground(Primary)

	ButtonStyle = lipgloss.NewStyle().
			Background(Primary).
			Foreground(Text).
			Padding(0, 2).
			MarginRight(1)

	SecondaryButtonStyle = lipgloss.NewStyle().
				Background(Surface).
				Foreground(Text).
				Padding(0, 2).
				MarginRight(1)
)
